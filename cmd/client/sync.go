package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/uryumtsevaa/gophkeeper/internal/models"
)

var syncCmd = &cobra.Command{
	Use:              "sync",
	Short:            "Synchronize secrets with server",
	PersistentPreRun: requireAuth(),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Получаем данные для синхронизации из локального хранилища
		lastSync := localStorage.GetLastSyncTime()
		clientHashes := localStorage.GetHashes()

		fmt.Printf("Starting sync (last sync: %s)...\n", lastSync.Format("2006-01-02 15:04:05"))

		// Отправляем запрос на синхронизацию
		syncReq := &models.SyncRequest{
			LastSyncTime: lastSync,
			ClientHashes: clientHashes,
		}

		response, err := clientInstance.SyncSecrets(ctx, syncReq)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to sync: %v\n", err)
			return
		}

		// Обрабатываем обновленные секреты
		if len(response.UpdatedSecrets) > 0 {
			fmt.Printf("Updating %d secrets...\n", len(response.UpdatedSecrets))
			for _, secret := range response.UpdatedSecrets {
				if err := localStorage.SaveSecret(&secret); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to save secret %s: %v\n", secret.ID, err)
				}
			}
		}

		// Обрабатываем удаленные секреты
		if len(response.DeletedSecrets) > 0 {
			fmt.Printf("Deleting %d secrets...\n", len(response.DeletedSecrets))
			if err := localStorage.DeleteSecrets(response.DeletedSecrets); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to delete secrets: %v\n", err)
			}
		}

		// Обновляем время последней синхронизации
		if err := localStorage.SetLastSyncTime(response.SyncTime); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to update sync time: %v\n", err)
		}

		fmt.Printf("Sync completed successfully at %s\n", response.SyncTime.Format("2006-01-02 15:04:05"))
	},
}
