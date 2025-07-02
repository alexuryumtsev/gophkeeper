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

		password, err := readPassword("Password: ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}

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

		password, err := readPassword("Password: ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}

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

// readPassword читает пароль из stdin, скрывая ввод
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("error reading password: %v", err)
	}
	fmt.Println()
	return string(passwordBytes), nil
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

// setupAuthentication выполняет общую логику аутентификации
func setupAuthentication() error {
	// Загружаем токен
	token, err := loadToken()
	if err != nil {
		return fmt.Errorf("not authenticated. Please login first")
	}
	clientInstance.SetToken(token)

	// Устанавливаем мастер-пароль
	if clientConfig.MasterPassword == "" {
		masterPassword, err := readPassword("Master password: ")
		if err != nil {
			return err
		}
		clientConfig.MasterPassword = masterPassword
	}
	clientInstance.SetMasterPassword(clientConfig.MasterPassword)
	
	return nil
}

// requireAuth возвращает функцию PersistentPreRun для команд, требующих аутентификации
func requireAuth() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if err := setupAuthentication(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}
}
