package service

import (
	"encoding/json"
	"fmt"

	"github.com/uryumtsevaa/gophkeeper/internal/crypto"
	"github.com/uryumtsevaa/gophkeeper/internal/server/interfaces"
)

// cryptoService реализация сервиса шифрования
type cryptoService struct {
	encryptor crypto.Encryptor
}

// NewCryptoService создает новый сервис шифрования
func NewCryptoService(encryptor crypto.Encryptor) interfaces.CryptoService {
	return &cryptoService{
		encryptor: encryptor,
	}
}

// EncryptSecretData шифрует данные секрета
func (c *cryptoService) EncryptSecretData(data interface{}, masterPassword string) ([]byte, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize data: %w", err)
	}
	return c.encryptor.Encrypt(dataBytes, masterPassword)
}

// DecryptSecretData расшифровывает данные секрета
func (c *cryptoService) DecryptSecretData(encryptedData []byte, masterPassword string) (interface{}, error) {
	decryptedBytes, err := c.encryptor.Decrypt(encryptedData, masterPassword)
	if err != nil {
		return nil, err
	}
	
	var data interface{}
	if err := json.Unmarshal(decryptedBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to deserialize data: %w", err)
	}
	
	return data, nil
}