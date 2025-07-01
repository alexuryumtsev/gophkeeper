package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user",
	Run: func(cmd *cobra.Command, args []string) {
		email, _ := cmd.Flags().GetString("email")
		username, _ := cmd.Flags().GetString("username")

		if username == "" {
			fmt.Print("Username: ")
			fmt.Scanln(&username)
		}

		if email == "" {
			fmt.Print("Email: ")
			fmt.Scanln(&email)
		}

		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
			return
		}
		password := string(passwordBytes)
		fmt.Println()

		ctx := context.Background()
		response, err := clientInstance.Register(ctx, username, email, password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Registration failed: %v\n", err)
			return
		}

		// Сохраняем токен
		if err := saveToken(response.Token); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save token: %v\n", err)
		}

		fmt.Printf("Registration successful! Welcome, %s\n", response.User.Username)
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login with existing credentials",
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("username")

		if username == "" {
			fmt.Print("Username: ")
			fmt.Scanln(&username)
		}

		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
			return
		}
		password := string(passwordBytes)
		fmt.Println()

		ctx := context.Background()
		response, err := clientInstance.Login(ctx, username, password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
			return
		}

		// Сохраняем токен
		if err := saveToken(response.Token); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save token: %v\n", err)
		}

		fmt.Printf("Login successful! Welcome back, %s\n", response.User.Username)
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and clear stored credentials",
	Run: func(cmd *cobra.Command, args []string) {
		if err := removeToken(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove token: %v\n", err)
			return
		}
		fmt.Println("Logged out successfully")
	},
}

func init() {
	registerCmd.Flags().StringP("username", "u", "", "Username")
	registerCmd.Flags().StringP("email", "e", "", "Email address")

	loginCmd.Flags().StringP("username", "u", "", "Username")

	authCmd.AddCommand(registerCmd, loginCmd, logoutCmd)
}

// saveToken сохраняет токен в файл
func saveToken(token string) error {
	// Создаем директорию если она не существует
	if err := os.MkdirAll(clientConfig.ConfigDir, 0700); err != nil {
		return err
	}
	
	tokenPath := filepath.Join(clientConfig.ConfigDir, "token")
	return os.WriteFile(tokenPath, []byte(token), 0600)
}

// loadToken загружает токен из файла
func loadToken() (string, error) {
	tokenPath := filepath.Join(clientConfig.ConfigDir, "token")
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// removeToken удаляет файл с токеном
func removeToken() error {
	tokenPath := filepath.Join(clientConfig.ConfigDir, "token")
	return os.Remove(tokenPath)
}
