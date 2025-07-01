// Package main GophKeeper Server
//
//	@title						GophKeeper API
//	@version					1.0
//	@description				Secure password and data manager API
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.url				http://www.swagger.io/support
//	@contact.email				support@swagger.io
//	@license.name				MIT
//	@license.url				https://opensource.org/licenses/MIT
//	@host						localhost:8080
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/uryumtsevaa/gophkeeper/internal/config"
	"github.com/uryumtsevaa/gophkeeper/internal/server"
)

var configFile string

var rootCmd = &cobra.Command{
	Use:   "gophkeeper-server",
	Short: "GophKeeper server - secure password and data manager backend",
	Run: func(cmd *cobra.Command, args []string) {
		serverConfig := config.LoadServerConfig()

		srv, err := server.NewServer(serverConfig)
		if err != nil {
			log.Fatalf("Failed to create server: %v", err)
		}

		if err := srv.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		versionInfo := config.GetDefaultVersion()
		versionInfo.SetVersion(version, buildDate, gitCommit)
		
		fmt.Printf("GophKeeper Server\n")
		fmt.Printf("Version: %s\n", versionInfo.Version)
		fmt.Printf("Build Date: %s\n", versionInfo.BuildDate)
		fmt.Printf("Git Commit: %s\n", versionInfo.GitCommit)
	},
}

var (
	version   = "dev"
	buildDate = "unknown"
	gitCommit = "unknown"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is ./config.yaml)")
	config.SetupServerFlags(rootCmd)

	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	config.InitServerConfig(configFile)
}
