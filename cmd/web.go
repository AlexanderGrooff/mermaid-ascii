package cmd

import (
	"fmt"
	"html/template"
	"net/http"

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
		inputText := c.PostForm("inputText")
		result := generate_map(inputText)
		c.String(http.StatusOK, result)
	})

	return r
}

func generate_map(input string) string {
	mermaidMap, _, err := mermaidFileToMap(input)
	if err != nil {
		return "Failed to parse mermaid file"
	}

	ascii_art := drawMap(mermaidMap, nil)
	escaped_ascii_art := template.HTMLEscapeString(ascii_art)
	html_ascii_art := fmt.Sprintf("<pre>%s</pre>", escaped_ascii_art)
	return html_ascii_art
}
