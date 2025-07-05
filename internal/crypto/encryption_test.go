package crypto

import (
	"testing"
)

func TestAESEncryptor_EncryptDecrypt(t *testing.T) {
	encryptor := NewAESEncryptor()
	password := "test-password"
	plaintext := []byte("Hello, World! This is a test message.")

	// Тест шифрования
	encrypted, err := encryptor.Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	if len(encrypted) == 0 {
		t.Fatal("Encrypted data is empty")
	}

	// Тест расшифровки
	decrypted, err := encryptor.Decrypt(encrypted, password)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("Decrypted data doesn't match original. Got: %s, Expected: %s", string(decrypted), string(plaintext))
	}
}

func TestAESEncryptor_WrongPassword(t *testing.T) {
	encryptor := NewAESEncryptor()
	password := "correct-password"
	wrongPassword := "wrong-password"
	plaintext := []byte("Secret data")

	encrypted, err := encryptor.Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Попытка расшифровки с неверным паролем
	_, err = encryptor.Decrypt(encrypted, wrongPassword)
	if err == nil {
		t.Fatal("Expected decryption to fail with wrong password")
	}
}

func TestAESEncryptor_EncryptString(t *testing.T) {
	encryptor := NewAESEncryptor()
	password := "test-password"
	plaintext := "Hello, World!"

	encrypted, err := encryptor.EncryptString(plaintext, password)
	if err != nil {
		t.Fatalf("String encryption failed: %v", err)
	}

	decrypted, err := encryptor.DecryptString(encrypted, password)
	if err != nil {
		t.Fatalf("String decryption failed: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("Decrypted string doesn't match. Got: %s, Expected: %s", decrypted, plaintext)
	}
}

func TestHashPassword(t *testing.T) {
	password := "test-password"

	hash1 := HashPassword(password)
	hash2 := HashPassword(password)

	// Хеши должны быть разными (из-за случайной соли)
	if hash1 == hash2 {
		t.Fatal("Password hashes should be different due to random salt")
	}

	// Оба хеша должны быть корректными для данного пароля
	if !VerifyPassword(password, hash1) {
		t.Fatal("First hash verification failed")
	}

	if !VerifyPassword(password, hash2) {
		t.Fatal("Second hash verification failed")
	}

	// Неверный пароль не должен проходить проверку
	if VerifyPassword("wrong-password", hash1) {
		t.Fatal("Wrong password should not verify")
	}
}

func TestGenerateSyncHash(t *testing.T) {
	data1 := []byte("test data")
	data2 := []byte("different data")
	data3 := []byte("test data") // такие же данные как data1

	hash1 := GenerateSyncHash(data1)
	hash2 := GenerateSyncHash(data2)
	hash3 := GenerateSyncHash(data3)

	// Разные данные должны давать разные хеши
	if hash1 == hash2 {
		t.Fatal("Different data should produce different hashes")
	}

	// Одинаковые данные должны давать одинаковые хеши
	if hash1 != hash3 {
		t.Fatal("Same data should produce same hashes")
	}

	// Хеш не должен быть пустым
	if hash1 == "" {
		t.Fatal("Hash should not be empty")
	}
}

func BenchmarkEncryption(b *testing.B) {
	encryptor := NewAESEncryptor()
	password := "benchmark-password"
	data := make([]byte, 1024) // 1KB данных
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encryptor.Encrypt(data, password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecryption(b *testing.B) {
	encryptor := NewAESEncryptor()
	password := "benchmark-password"
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	encrypted, err := encryptor.Encrypt(data, password)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encryptor.Decrypt(encrypted, password)
		if err != nil {
			b.Fatal(err)
		}
	}
}
