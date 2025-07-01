package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/uryumtsevaa/gophkeeper/internal/client"
	"github.com/uryumtsevaa/gophkeeper/internal/config"
)

var (
	clientConfig   *config.ClientConfig
	clientInstance *client.Client
	localStorage   *client.LocalStorage
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "gophkeeper-client",
	Short: "GophKeeper client - secure password and data manager",
	Long: `GophKeeper client is a command-line interface for managing
your passwords, text data, binary files, and credit card information securely.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Инициализируем клиент
		clientInstance = client.NewClient(clientConfig.ServerURL)

		// Инициализируем локальное хранилище
		var err error
		localStorage, err = client.NewLocalStorage(clientConfig.ConfigDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize local storage: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	var err error
	clientConfig, err = config.SetupClientFlags(rootCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup client config: %v\n", err)
		os.Exit(1)
	}

	// Добавляем команды
	rootCmd.AddCommand(
		authCmd,
		secretsCmd,
		syncCmd,
		versionCmd,
	)
}
