// Package cmd sets up the CLI commands using Cobra,
// handling flags, subcommands, and execution logic.
package cmd

// Package cmd sets up the CLI commands using Cobra,
// handling flags, subcommands, and execution logic for the files2prompt tool.

import (
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/toozej/files2prompt/internal/files2prompt"
	"github.com/toozej/files2prompt/pkg/config"
	"github.com/toozej/files2prompt/pkg/man"
	"github.com/toozej/files2prompt/pkg/version"
)

var conf config.Config

var rootCmd = &cobra.Command{
	Use:   "files2prompt [paths...]",
	Short: "Crawl and output file contents with various filtering options for AI prompting",
	Long: `files2prompt helps prepare files for AI prompts by crawling directories
and outputting file contents with optional filtering and formatting.`,
	Args:             cobra.ArbitraryArgs,
	PersistentPreRun: rootCmdPreRun,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read paths from stdin if available
		stdinPaths := readPathsFromStdin(conf.Null)
		// Combine args and stdin paths
		conf.Paths = append(args, stdinPaths...)
		if len(conf.Paths) == 0 {
			return fmt.Errorf("no paths provided via arguments or stdin")
		}
		log.Debugf("cmd pkg RunE config config struct contains: %v\n", conf)
		return files2prompt.Run(conf)
	},
}

func rootCmdPreRun(cmd *cobra.Command, args []string) {
	if debug, err := cmd.Flags().GetBool("debug"); err == nil && debug {
		log.SetLevel(log.DebugLevel)
	}
}

func readPathsFromStdin(useNull bool) []string {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil
	}
	// Check if stdin has data
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil // No input
	}
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil
	}
	var paths []string
	if useNull {
		paths = strings.Split(string(content), "\x00")
	} else {
		paths = strings.Fields(string(content))
	}
	// Filter empty
	var filtered []string
	for _, p := range paths {
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// Execute runs the root Cobra command for the CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func init() {
	// get configuration from environment variables
	conf = config.GetEnvVars()

	// create rootCmd-level flags
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug-level logging")

	// override .env configurations with flags+args
	if len(conf.Extensions) == 0 {
		rootCmd.Flags().StringSliceVarP(&conf.Extensions, "extension", "e", []string{}, "File extensions to include")
	}
	if !conf.IncludeHidden {
		rootCmd.Flags().BoolVarP(&conf.IncludeHidden, "include-hidden", "", false, "Include hidden files and folders")
	}
	if !conf.IgnoreGitignore {
		rootCmd.Flags().BoolVarP(&conf.IgnoreGitignore, "ignore-gitignore", "", false, "Ignore .gitignore files")
	}
	if len(conf.IgnorePatterns) == 0 {
		rootCmd.Flags().StringSliceVarP(&conf.IgnorePatterns, "ignore", "", []string{},
			"Patterns to ignore (can be comma-separated or specified multiple times). "+
				"Use '/' suffix to match directories only. Examples: "+
				"'*.test.js', 'test/', 'path/to/ignore/, 'dir1/,dir2/'")
	}
	if conf.OutputFile == "" {
		rootCmd.Flags().StringVarP(&conf.OutputFile, "output", "o", "", "Output file path")
	}
	if !conf.ClaudeXML {
		rootCmd.Flags().BoolVarP(&conf.ClaudeXML, "cxml", "c", false, "Output in XML format for Claude")
	}
	if !conf.LineNumbers {
		rootCmd.Flags().BoolVarP(&conf.LineNumbers, "line-numbers", "n", false, "Display line numbers in output")
	}
	if !conf.Markdown {
		rootCmd.Flags().BoolVarP(&conf.Markdown, "markdown", "m", false, "Output in Markdown format with fenced code blocks")
	}
	if !conf.Null {
		rootCmd.Flags().BoolVarP(&conf.Null, "null", "0", false, "Use NUL character as separator when reading from stdin")
	}

	// add sub-commands
	rootCmd.AddCommand(
		man.NewManCmd(),
		version.Command(),
	)
}
