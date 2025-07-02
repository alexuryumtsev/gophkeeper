package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/uryumtsevaa/gophkeeper/internal/models"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage secrets",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := setupAuthentication(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secrets",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		response, err := clientInstance.GetSecrets(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get secrets: %v\n", err)
			return
		}

		if len(response.Secrets) == 0 {
			fmt.Println("No secrets found")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprintln(w, "ID\tNAME\tTYPE\tCREATED")
		for _, secret := range response.Secrets {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				secret.ID.String()[:8]+"...",
				secret.Name,
				secret.Type,
				secret.CreatedAt.Format("2006-01-02 15:04"))
		}
		w.Flush()
	},
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new secret",
}

var addCredentialsCmd = &cobra.Command{
	Use:   "credentials",
	Short: "Add login/password credentials",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		username, _ := cmd.Flags().GetString("username")
		url, _ := cmd.Flags().GetString("url")
		metadata, _ := cmd.Flags().GetString("metadata")

		if name == "" {
			fmt.Print("Name: ")
			fmt.Scanln(&name)
		}

		if username == "" {
			fmt.Print("Username: ")
			fmt.Scanln(&username)
		}

		password, err := readPassword("Password: ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}

		credentials := models.Credentials{
			Name:     name,
			Username: username,
			Password: password,
			URL:      url,
			Metadata: metadata,
		}

		req := &models.SecretRequest{
			Type:     models.SecretTypeCredentials,
			Name:     name,
			Data:     credentials,
			Metadata: metadata,
		}

		ctx := context.Background()
		response, err := clientInstance.CreateSecret(ctx, req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create secret: %v\n", err)
			return
		}

		fmt.Printf("Credentials saved successfully! ID: %s\n", response.ID)
	},
}

var addTextCmd = &cobra.Command{
	Use:   "text",
	Short: "Add text data",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		content, _ := cmd.Flags().GetString("content")
		metadata, _ := cmd.Flags().GetString("metadata")

		if name == "" {
			fmt.Print("Name: ")
			fmt.Scanln(&name)
		}

		if content == "" {
			fmt.Print("Content: ")
			fmt.Scanln(&content)
		}

		textData := models.TextData{
			Name:     name,
			Content:  content,
			Metadata: metadata,
		}

		req := &models.SecretRequest{
			Type:     models.SecretTypeText,
			Name:     name,
			Data:     textData,
			Metadata: metadata,
		}

		ctx := context.Background()
		response, err := clientInstance.CreateSecret(ctx, req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create secret: %v\n", err)
			return
		}

		fmt.Printf("Text data saved successfully! ID: %s\n", response.ID)
	},
}

var addBinaryCmd = &cobra.Command{
	Use:   "binary",
	Short: "Add binary file",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		filename, _ := cmd.Flags().GetString("file")
		metadata, _ := cmd.Flags().GetString("metadata")

		if name == "" {
			fmt.Print("Name: ")
			fmt.Scanln(&name)
		}

		if filename == "" {
			fmt.Print("File path: ")
			fmt.Scanln(&filename)
		}

		data, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read file: %v\n", err)
			return
		}

		binaryData := models.BinaryData{
			Name:     name,
			Filename: filename,
			Data:     data,
			Metadata: metadata,
		}

		req := &models.SecretRequest{
			Type:     models.SecretTypeBinary,
			Name:     name,
			Data:     binaryData,
			Metadata: metadata,
		}

		ctx := context.Background()
		response, err := clientInstance.CreateSecret(ctx, req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create secret: %v\n", err)
			return
		}

		fmt.Printf("Binary file saved successfully! ID: %s\n", response.ID)
	},
}

var addCardCmd = &cobra.Command{
	Use:   "card",
	Short: "Add credit card data",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		number, _ := cmd.Flags().GetString("number")
		holder, _ := cmd.Flags().GetString("holder")
		bank, _ := cmd.Flags().GetString("bank")
		metadata, _ := cmd.Flags().GetString("metadata")

		if name == "" {
			fmt.Print("Name: ")
			fmt.Scanln(&name)
		}

		if number == "" {
			fmt.Print("Card number: ")
			fmt.Scanln(&number)
		}

		if holder == "" {
			fmt.Print("Cardholder name: ")
			fmt.Scanln(&holder)
		}

		fmt.Print("Expiry month (1-12): ")
		var monthStr string
		fmt.Scanln(&monthStr)
		month, err := strconv.Atoi(monthStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid month: %v\n", err)
			return
		}

		fmt.Print("Expiry year: ")
		var yearStr string
		fmt.Scanln(&yearStr)
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid year: %v\n", err)
			return
		}

		cvv, err := readPassword("CVV: ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}

		cardData := models.CardData{
			Name:        name,
			Number:      number,
			ExpiryMonth: month,
			ExpiryYear:  year,
			CVV:         cvv,
			Holder:      holder,
			Bank:        bank,
			Metadata:    metadata,
		}

		req := &models.SecretRequest{
			Type:     models.SecretTypeCard,
			Name:     name,
			Data:     cardData,
			Metadata: metadata,
		}

		ctx := context.Background()
		response, err := clientInstance.CreateSecret(ctx, req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create secret: %v\n", err)
			return
		}

		fmt.Printf("Card data saved successfully! ID: %s\n", response.ID)
	},
}

var getCmd = &cobra.Command{
	Use:   "get [ID]",
	Short: "Get a specific secret",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		secretID := args[0]

		ctx := context.Background()
		response, err := clientInstance.GetSecret(ctx, secretID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get secret: %v\n", err)
			return
		}

		// Форматированный вывод секрета
		fmt.Printf("ID: %s\n", response.ID)
		fmt.Printf("Name: %s\n", response.Name)
		fmt.Printf("Type: %s\n", response.Type)
		fmt.Printf("Created: %s\n", response.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated: %s\n", response.UpdatedAt.Format("2006-01-02 15:04:05"))

		if response.Metadata != "" {
			fmt.Printf("Metadata: %s\n", response.Metadata)
		}

		fmt.Println("\nData:")
		dataJSON, _ := json.MarshalIndent(response.Data, "", "  ")
		fmt.Println(string(dataJSON))
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [ID]",
	Short: "Delete a secret",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		secretID := args[0]

		fmt.Printf("Are you sure you want to delete secret %s? (y/N): ", secretID)
		var confirm string
		fmt.Scanln(&confirm)

		if confirm != "y" && confirm != "Y" {
			fmt.Println("Cancelled")
			return
		}

		ctx := context.Background()
		err := clientInstance.DeleteSecret(ctx, secretID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete secret: %v\n", err)
			return
		}

		fmt.Println("Secret deleted successfully")
	},
}

func init() {
	// Add command flags
	addCredentialsCmd.Flags().StringP("name", "n", "", "Name for the credentials")
	addCredentialsCmd.Flags().StringP("username", "u", "", "Username")
	addCredentialsCmd.Flags().String("url", "", "URL")
	addCredentialsCmd.Flags().StringP("metadata", "m", "", "Metadata")

	addTextCmd.Flags().StringP("name", "n", "", "Name for the text data")
	addTextCmd.Flags().StringP("content", "c", "", "Text content")
	addTextCmd.Flags().StringP("metadata", "m", "", "Metadata")

	addBinaryCmd.Flags().StringP("name", "n", "", "Name for the binary data")
	addBinaryCmd.Flags().StringP("file", "f", "", "Path to binary file")
	addBinaryCmd.Flags().StringP("metadata", "m", "", "Metadata")

	addCardCmd.Flags().StringP("name", "n", "", "Name for the card")
	addCardCmd.Flags().String("number", "", "Card number")
	addCardCmd.Flags().String("holder", "", "Cardholder name")
	addCardCmd.Flags().String("bank", "", "Bank name")
	addCardCmd.Flags().StringP("metadata", "m", "", "Metadata")

	// Build command tree
	addCmd.AddCommand(addCredentialsCmd, addTextCmd, addBinaryCmd, addCardCmd)
	secretsCmd.AddCommand(listCmd, addCmd, getCmd, deleteCmd)
}
