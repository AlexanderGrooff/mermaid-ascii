package cmd

import (
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(webCmd)
}

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "HTTP server for rendering mermaid diagrams.",
	Run: func(cmd *cobra.Command, args []string) {
		r := setupRouter()
		// Listen and Server in 0.0.0.0:8080
		err := r.Run(":3001")
		if err != nil {
			panic(err)
		}
	},
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{})
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
		log.Infof("Received input %s", c.Request.PostForm.Encode())
		result := generate_map(mermaidString)
		c.String(http.StatusOK, result)
	})

	return r
}

func generate_map(input string) string {
	properties, err := mermaidFileToMap(input, "html")
	log.Infof("Properties: %v", properties)
	if err != nil {
		return "Failed to parse mermaid file"
	}

	return drawMap(properties)
}
