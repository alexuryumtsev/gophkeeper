package server

import (
	"encoding/json"
	"fmt"
)

type Encryptor interface {
	Encrypt(data []byte, masterPassword string) ([]byte, error)
	Decrypt(data []byte, masterPassword string) ([]byte, error)
}

type CryptoService interface {
	EncryptSecretData(data interface{}, masterPassword string) ([]byte, error)
	DecryptSecretData(encryptedData []byte, masterPassword string) (interface{}, error)
}

type cryptoService struct {
	encryptor Encryptor
}

func NewCryptoService(encryptor Encryptor) CryptoService {
	return &cryptoService{
		encryptor: encryptor,
	}
}

func (c *cryptoService) EncryptSecretData(data interface{}, masterPassword string) ([]byte, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize data: %w", err)
	}

	encryptedData, err := c.encryptor.Encrypt(dataBytes, masterPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	return encryptedData, nil
}

func (c *cryptoService) DecryptSecretData(encryptedData []byte, masterPassword string) (interface{}, error) {
	decryptedData, err := c.encryptor.Decrypt(encryptedData, masterPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	var data interface{}
	if err := json.Unmarshal(decryptedData, &data); err != nil {
		return nil, fmt.Errorf("failed to deserialize data: %w", err)
	}

	return data, nil
}