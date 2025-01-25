package starter

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Get environment variables
func getEnvVars() error {
	if _, err := os.Stat(".env"); err == nil {
		// Initialize Viper from .env file
		viper.SetConfigFile(".env") // Specify the name of your .env file

		// Read the .env file
		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}

	// Enable reading environment variables
	viper.AutomaticEnv()

	// get username from Viper
	username = viper.GetString("USERNAME")
	if username == "" {
		return fmt.Errorf("username must be provided")
	}

	return nil
}
