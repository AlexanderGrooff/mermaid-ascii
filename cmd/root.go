package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AlexanderGrooff/mermaid-ascii/internal/diagram"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Global flags
var Verbose bool
var Coords bool
var boxBorderPadding = 1
var paddingBetweenX = 5
var paddingBetweenY = 5
var graphDirection = ""
var useAscii = false
var maxWidth = 0

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

		var mermaid []byte
		var err error

		filePath := cmd.Flag("file").Value.String()
		if filePath == "" || filePath == "-" {
			// Read from stdin
			mermaid, err = io.ReadAll(os.Stdin)
			if err != nil {
				log.Fatal("Failed to read from stdin: ", err)
				return
			}
		} else {
			// Read from file
			mermaid, err = os.ReadFile(filePath)
			if err != nil {
				log.Fatal("Failed to read mermaid file: ", err)
				return
			}
		}

		// Create render configuration from flags
		config, err := diagram.NewCLIConfig(
			useAscii,
			Coords,
			Verbose,
			boxBorderPadding,
			paddingBetweenX,
			paddingBetweenY,
				maxWidth,
			graphDirection,
		)
		if err != nil {
			log.Fatalf("Invalid configuration: %v", err)
		}

		// Render diagram (automatically detects type)
		output, err := RenderDiagram(string(mermaid), config)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(output)
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
	rootCmd.PersistentFlags().BoolVarP(&useAscii, "ascii", "a", false, "Don't use extended character set")
	rootCmd.PersistentFlags().BoolVarP(&Coords, "coords", "c", false, "Show coordinates")
	rootCmd.PersistentFlags().IntVarP(&paddingBetweenX, "paddingX", "x", paddingBetweenX, "Horizontal space between nodes")
	rootCmd.PersistentFlags().IntVarP(&paddingBetweenY, "paddingY", "y", paddingBetweenY, "Vertical space between nodes")
	rootCmd.PersistentFlags().IntVarP(&boxBorderPadding, "borderPadding", "p", boxBorderPadding, "Padding between text and border")
    rootCmd.PersistentFlags().IntVarP(&maxWidth, "maxWidth", "w", maxWidth, "Maximum diagram width in characters (0 = unlimited)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringP("file", "f", "", "Mermaid file to parse (use '-' for stdin)")
}
