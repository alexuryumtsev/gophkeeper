package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/argon2"
)

const (
	saltSize   = 32
	nonceSize  = 12
	keySize    = 32
	iterations = 3
	memory     = 64 * 1024
	threads    = 4
)

var (
	ErrInvalidData      = errors.New("invalid encrypted data")
	ErrDecryptionFailed = errors.New("decryption failed")
)

// Encryptor интерфейс для шифрования данных
type Encryptor interface {
	Encrypt(data []byte, password string) ([]byte, error)
	Decrypt(encryptedData []byte, password string) ([]byte, error)
	GenerateKey(password string, salt []byte) []byte
}

// AESEncryptor реализует шифрование с использованием AES-GCM
type AESEncryptor struct{}

// NewAESEncryptor создает новый экземпляр AESEncryptor
func NewAESEncryptor() *AESEncryptor {
	return &AESEncryptor{}
}

// GenerateKey генерирует ключ шифрования из пароля и соли
func (e *AESEncryptor) GenerateKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, iterations, memory, threads, keySize)
}

// Encrypt шифрует данные с использованием пароля
func (e *AESEncryptor) Encrypt(data []byte, password string) ([]byte, error) {
	// Генерируем случайную соль
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// Генерируем ключ из пароля и соли
	key := e.GenerateKey(password, salt)

	// Создаем AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Создаем GCM режим
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Генерируем nonce
	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Шифруем данные
	ciphertext := gcm.Seal(nil, nonce, data, nil)

	// Объединяем соль, nonce и зашифрованные данные
	result := make([]byte, saltSize+nonceSize+len(ciphertext))
	copy(result[:saltSize], salt)
	copy(result[saltSize:saltSize+nonceSize], nonce)
	copy(result[saltSize+nonceSize:], ciphertext)

	return result, nil
}

// Decrypt расшифровывает данные с использованием пароля
func (e *AESEncryptor) Decrypt(encryptedData []byte, password string) ([]byte, error) {
	if len(encryptedData) < saltSize+nonceSize {
		return nil, ErrInvalidData
	}

	// Извлекаем соль, nonce и зашифрованные данные
	salt := encryptedData[:saltSize]
	nonce := encryptedData[saltSize : saltSize+nonceSize]
	ciphertext := encryptedData[saltSize+nonceSize:]

	// Генерируем ключ из пароля и соли
	key := e.GenerateKey(password, salt)

	// Создаем AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Создаем GCM режим
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Расшифровываем данные
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// EncryptString шифрует строку и возвращает результат в base64
func (e *AESEncryptor) EncryptString(data, password string) (string, error) {
	encrypted, err := e.Encrypt([]byte(data), password)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptString расшифровывает строку из base64
func (e *AESEncryptor) DecryptString(encryptedData, password string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}

	decrypted, err := e.Decrypt(data, password)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// HashPassword хеширует пароль для хранения в базе данных
func HashPassword(password string) string {
	salt := make([]byte, saltSize)
	rand.Read(salt)

	hash := argon2.IDKey([]byte(password), salt, iterations, memory, threads, keySize)

	// Объединяем соль и хеш
	result := make([]byte, saltSize+keySize)
	copy(result[:saltSize], salt)
	copy(result[saltSize:], hash)

	return base64.StdEncoding.EncodeToString(result)
}

// VerifyPassword проверяет пароль против хеша
func VerifyPassword(password, hashedPassword string) bool {
	data, err := base64.StdEncoding.DecodeString(hashedPassword)
	if err != nil || len(data) != saltSize+keySize {
		return false
	}

	salt := data[:saltSize]
	hash := data[saltSize:]

	testHash := argon2.IDKey([]byte(password), salt, iterations, memory, threads, keySize)

	// Сравниваем хеши
	if len(hash) != len(testHash) {
		return false
	}

	for i := 0; i < len(hash); i++ {
		if hash[i] != testHash[i] {
			return false
		}
	}

	return true
}

// GenerateSyncHash генерирует хеш для синхронизации
func GenerateSyncHash(data []byte) string {
	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}
