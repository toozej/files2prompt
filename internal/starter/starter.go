package starter

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	username string
)

func Run(cmd *cobra.Command, args []string) {
	err := getEnvVars()
	if err != nil {
		log.Fatal("Error gathering required environment variables: ", err)
	}

	fmt.Println("Hello from ", username)
}
