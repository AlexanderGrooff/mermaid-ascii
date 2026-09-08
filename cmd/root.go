package cmd

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
	"github.com/AlexanderGrooff/mermaid-ascii/pkg/render"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Global flags
var Verbose bool
var Coords bool
var boxBorderPadding = 1
var paddingBetweenX = 5
var paddingBetweenY = 5
var graphDirection = "LR"
var useAscii = false
var maxWidth string

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
			graphDirection,
		)
		if err != nil {
			log.Fatalf("Invalid configuration: %v", err)
		}
		config.MaxWidth, err = resolveMaxWidth(maxWidth, os.Stdout)
		if err != nil {
			log.Fatal(err)
		}

		// Render diagram (automatically detects type)
		output, widthStatus, err := render.RenderDiagramWithStatus(string(mermaid), config)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(output)
		if widthStatus.Requested && widthStatus.Compacted && widthStatus.Met {
			fmt.Fprintf(os.Stderr, "note: graph exceeded --max-width %d; used compact spacing (%d columns)\n", widthStatus.Limit, widthStatus.Width)
		} else if widthStatus.Requested && !widthStatus.Met {
			fmt.Fprintf(os.Stderr, "warning: graph is %d columns wide; it could not fit --max-width %d even with compact spacing\n", widthStatus.Width, widthStatus.Limit)
		}
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

func resolveMaxWidth(value string, stdout *os.File) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	if strings.EqualFold(value, "auto") {
		if stdout == nil || !term.IsTerminal(int(stdout.Fd())) {
			return 0, fmt.Errorf("--max-width=auto requires stdout to be a terminal")
		}
		width, _, err := term.GetSize(int(stdout.Fd()))
		if err != nil || width <= 0 {
			return 0, fmt.Errorf("could not detect terminal width for --max-width=auto")
		}
		return width, nil
	}

	width, err := strconv.Atoi(value)
	if err != nil || width <= 0 {
		return 0, fmt.Errorf("invalid --max-width %q: use a positive number or auto", value)
	}
	return width, nil
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
	rootCmd.PersistentFlags().StringVar(&maxWidth, "max-width", "", "Maximum graph width in terminal columns (positive number or auto)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringP("file", "f", "", "Mermaid file to parse (use '-' for stdin)")
}
