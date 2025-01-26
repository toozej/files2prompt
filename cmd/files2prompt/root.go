package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/automaxprocs/maxprocs"

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
	Args:             cobra.MinimumNArgs(1),
	PersistentPreRun: rootCmdPreRun,
	RunE: func(cmd *cobra.Command, args []string) error {
		conf.Paths = args
		if viper.GetBool("debug") {
			log.Debugf("cmd pkg RunE config config struct contains: %v\n", conf)
		}
		return files2prompt.Run(conf)
	},
}

func rootCmdPreRun(cmd *cobra.Command, args []string) {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return
	}
	if viper.GetBool("debug") {
		log.SetLevel(log.DebugLevel)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func init() {
	_, err := maxprocs.Set()
	if err != nil {
		log.Error("Error setting maxprocs: ", err)
	}

	// get configuration from environment variables
	conf = config.GetEnvVars()

	// create rootCmd-level flags
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug-level logging")

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
		rootCmd.Flags().StringSliceVarP(&conf.IgnorePatterns, "ignore", "", []string{}, "Patterns to ignore")
	}
	if conf.OutputFile == "" {
		rootCmd.Flags().StringVarP(&conf.OutputFile, "output", "o", "", "Output file path")
	}
	if !conf.ClaudeXML {
		rootCmd.Flags().BoolVarP(&conf.ClaudeXML, "cxml", "c", false, "Output in XML format for Claude")
	}

	// print config for debugging purposes
	log.Debugf("cmd pkg init function config struct contains: %v\n", conf)

	// add sub-commands
	rootCmd.AddCommand(
		man.NewManCmd(),
		version.Command(),
	)
}
