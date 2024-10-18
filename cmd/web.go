package cmd

import (
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// Add the resultCache variable with additional fields
var (
	resultCache = struct {
		sync.RWMutex
		m map[string]cacheEntry
	}{m: make(map[string]cacheEntry)}
	maxCacheSize = 10000 // Maximum number of entries in the cache
)

type cacheEntry struct {
	value string
}

var (
	gitVersion     string
	gitVersionOnce sync.Once
)

func init() {
	rootCmd.AddCommand(webCmd)
}

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "HTTP server for rendering mermaid diagrams.",
	Run: func(cmd *cobra.Command, args []string) {
		if Verbose {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}
		r := setupRouter()
		// Listen and Server in 0.0.0.0:8080
		err := r.Run(":3001")
		if err != nil {
			panic(err)
		}
	},
}

// Add this function near the top of the file, after the imports
func getGitVersion() string {
	gitVersionOnce.Do(func() {
		log.Info("Getting git version")
		cmd := exec.Command("git", "describe", "--tags", "--always")
		output, err := cmd.Output()
		if err != nil {
			log.Warnf("Failed to get git version: %v", err)
			gitVersion = "unknown"
		} else {
			gitVersion = strings.TrimSpace(string(output))
		}
	})
	return gitVersion
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"Version": getGitVersion(),
		})
	})

	r.POST("/generate", func(c *gin.Context) {
		mermaidString := c.PostForm("mermaid")
		// Parse xPadding and yPadding as integers
		xPadding := c.PostForm("xPadding")
		if xPadding != "" {
			if padding, err := strconv.Atoi(xPadding); err == nil {
				paddingBetweenX = padding
			} else {
				log.Warnf("Invalid xPadding value: %s", xPadding)
			}
		}

		yPadding := c.PostForm("yPadding")
		if yPadding != "" {
			if padding, err := strconv.Atoi(yPadding); err == nil {
				paddingBetweenY = padding
			} else {
				log.Warnf("Invalid yPadding value: %s", yPadding)
			}
		}
		useExtendedCharsData := c.PostForm("useExtendedChars")
		useExtendedChars = useExtendedCharsData != ""
		log.Debugf("Received input %s", c.Request.PostForm.Encode())

		// Create a cache key using the input parameters
		cacheKey := mermaidString + "x" + xPadding + "y" + yPadding + "e" + useExtendedCharsData

		// Check if the result is already in the cache
		resultCache.RLock()
		entry, found := resultCache.m[cacheKey]
		resultCache.RUnlock()

		if found {
			log.Infof("Cache hit for key: %s", cacheKey)
			c.String(http.StatusOK, entry.value)
			return
		}

		// If not in cache or expired, generate the map
		result := generate_map(mermaidString)

		// Store the result in the cache
		resultCache.Lock()
		if len(resultCache.m) >= maxCacheSize {
			log.Infof("Cache is full, removing oldest entry")
			// Remove a random entry if cache is full
			for k := range resultCache.m {
				delete(resultCache.m, k)
				break
			}
		}
		resultCache.m[cacheKey] = cacheEntry{
			value: result,
		}
		resultCache.Unlock()

		c.String(http.StatusOK, result)
	})

	return r
}

func generate_map(input string) string {
	properties, err := mermaidFileToMap(input, "html")
	if err != nil {
		return "Failed to parse mermaid file"
	}

	return drawMap(properties)
}
