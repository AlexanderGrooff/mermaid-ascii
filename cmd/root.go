package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Global flags
var Verbose bool
var Coords bool
var boxBorderPadding = 1
var paddingBetweenX = 5
var paddingBetweenY = 5
var graphDirection = "LR"
var useExtendedChars = true

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mermaid-ascii",
	Short: "Generate ASCII diagrams from mermaid code.",
	Run: func(cmd *cobra.Command, args []string) {
		if Verbose {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}
		mermaid, err := os.ReadFile(cmd.Flag("file").Value.String())
		if err != nil {
			log.Fatal("Failed to parse mermaid file: ", err)
			return
		}
		properties, err := mermaidFileToMap(string(mermaid), "cli")
		if err != nil {
			log.Fatal("Failed to parse mermaid file: ", err)
		}
		drawMap(properties)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVarP(&useExtendedChars, "extended-chars", "e", false, "Use extended character set")
	rootCmd.PersistentFlags().BoolVarP(&Coords, "coords", "c", false, "Show coordinates")
	rootCmd.PersistentFlags().IntVarP(&paddingBetweenX, "paddingX", "x", paddingBetweenX, "Horizontal space between nodes")
	rootCmd.PersistentFlags().IntVarP(&paddingBetweenY, "paddingY", "y", paddingBetweenY, "Vertical space between nodes")
	rootCmd.PersistentFlags().IntVarP(&boxBorderPadding, "borderPadding", "p", boxBorderPadding, "Padding between text and border")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringP("file", "f", "", "Mermaid file to parse")
}
