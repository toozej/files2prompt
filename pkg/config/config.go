package config

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Paths           []string `mapstructure:"paths"`
	Extensions      []string `mapstructure:"extensions"`
	IncludeHidden   bool     `mapstructure:"include_hidden"`
	IgnoreGitignore bool     `mapstructure:"ignore_gitignore"`
	IgnorePatterns  []string `mapstructure:"ignore_patterns"`
	OutputFile      string   `mapstructure:"output_file"`
	ClaudeXML       bool     `mapstructure:"claude_xml"`
	LineNumbers     bool     `mapstructure:"line_numbers"`
}

// Get environment variables
func GetEnvVars() Config {
	if _, err := os.Stat(".env"); err == nil {
		// Initialize Viper from .env file
		viper.SetConfigFile(".env") // Specify the name of your .env file

		// Read the .env file
		if err := viper.ReadInConfig(); err != nil {
			log.Warn("Error loading .env file: ", err)
		}
	}

	// Enable reading environment variables
	viper.AutomaticEnv()
	// Setup conf struct with items from environment variables
	var conf Config
	if err := viper.Unmarshal(&conf); err != nil {
		log.Fatalf("Error unmarshalling Viper conf: %s\n", err)
	}

	// print config for debugging purposes
	log.Debugf("config pkg Config struct contains: %v\n", conf)

	return conf
}
