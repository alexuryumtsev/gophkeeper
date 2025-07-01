# GophKeeper API Client

Публичный Go SDK для работы с GophKeeper API - безопасным менеджером паролей и данных.

## Установка

```bash
go get github.com/uryumtsevaa/gophkeeper/pkg/api
```

## Быстрый старт

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/uryumtsevaa/gophkeeper/pkg/api"
)

func main() {
    // Создаем клиент
    client, err := api.NewDefaultClient("http://localhost:8080")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Регистрируемся
    registerReq := &api.RegisterRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "password123",
    }

    resp, err := client.Register(ctx, registerReq)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Зарегистрирован пользователь: %s\n", resp.User.Username)

    // Устанавливаем мастер-пароль для шифрования
    client.SetMasterPassword("my-master-password")

    // Создаем секрет
    secret := &api.SecretRequest{
        Type: api.SecretTypeCredentials,
        Name: "My Gmail",
        Data: &api.Credentials{
            Username: "myemail@gmail.com",
            Password: "secret123",
            URL:      "https://gmail.com",
        },
    }

    secretResp, err := client.CreateSecret(ctx, secret)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Создан секрет: %s\n", secretResp.Name)
}
```

## Основные компоненты

### Client

Главный интерфейс для работы с API:

```go
type Client interface {
    // Аутентификация
    Register(ctx context.Context, req *RegisterRequest) (*LoginResponse, error)
    Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
    
    // Управление токенами
    SetToken(token string)
    SetMasterPassword(password string)
    
    // Управление секретами
    CreateSecret(ctx context.Context, req *SecretRequest) (*SecretResponse, error)
    GetSecrets(ctx context.Context) (*SecretsListResponse, error)
    GetSecret(ctx context.Context, secretID uuid.UUID) (*SecretResponse, error)
    UpdateSecret(ctx context.Context, secretID uuid.UUID, req *SecretRequest) (*SecretResponse, error)
    DeleteSecret(ctx context.Context, secretID uuid.UUID) error
    
    // Синхронизация
    SyncSecrets(ctx context.Context, req *SyncRequest) (*SyncResponse, error)
}
```

### Типы секретов

Поддерживаются 4 типа секретов:

1. **Credentials** - логин и пароль
2. **Text** - произвольный текст
3. **Binary** - бинарные данные
4. **Card** - данные банковских карт

```go
// Создание различных типов секретов
credentialsSecret := &api.SecretRequest{
    Type: api.SecretTypeCredentials,
    Name: "GitHub",
    Data: &api.Credentials{
        Username: "myusername",
        Password: "mypassword",
        URL:      "https://github.com",
    },
}

textSecret := &api.SecretRequest{
    Type: api.SecretTypeText,
    Name: "SSH Key",
    Data: &api.TextData{
        Content: "ssh-rsa AAAAB3...",
    },
}

cardSecret := &api.SecretRequest{
    Type: api.SecretTypeCard,
    Name: "My Credit Card",
    Data: &api.CardData{
        Number:      "1234567890123456",
        ExpiryMonth: 12,
        ExpiryYear:  2025,
        CVV:         "123",
        Holder:      "John Doe",
        Bank:        "Example Bank",
    },
}
```

### Конфигурация

```go
// Конфигурация по умолчанию
config := api.DefaultConfig()

// Настройка для продакшена
config := api.ProductionConfig("https://api.gophkeeper.com")

// Настройка для локальной разработки
config := api.LocalConfig()

// Кастомная конфигурация
config := &api.Config{
    BaseURL:   "https://my-server.com",
    Timeout:   time.Minute,
    UserAgent: "MyApp/1.0",
}

client, err := api.NewClient(config)
```

### Обработка ошибок

SDK предоставляет типизированные ошибки:

```go
_, err := client.Login(ctx, loginReq)
if err != nil {
    switch {
    case api.IsAuthError(err):
        fmt.Println("Ошибка аутентификации:", err)
    case api.IsNetworkError(err):
        fmt.Println("Сетевая ошибка:", err)
    case api.IsValidationError(err):
        fmt.Println("Ошибка валидации:", err)
    default:
        fmt.Println("Другая ошибка:", err)
    }
}
```

## Примеры использования

### Полный цикл работы с секретами

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/uryumtsevaa/gophkeeper/pkg/api"
)

func main() {
    client, err := api.NewDefaultClient("http://localhost:8080")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // 1. Вход в систему
    loginReq := &api.LoginRequest{
        Username: "testuser",
        Password: "password123",
    }

    _, err = client.Login(ctx, loginReq)
    if err != nil {
        log.Fatal(err)
    }

    // 2. Установка мастер-пароля
    client.SetMasterPassword("my-master-password")

    // 3. Создание секрета
    secretReq := &api.SecretRequest{
        Type: api.SecretTypeCredentials,
        Name: "Important Service",
        Data: &api.Credentials{
            Username: "admin",
            Password: "super-secret",
            URL:      "https://important-service.com",
        },
        Metadata: "Production credentials",
    }

    secret, err := client.CreateSecret(ctx, secretReq)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Создан секрет ID: %s\n", secret.ID)

    // 4. Получение всех секретов
    secrets, err := client.GetSecrets(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Всего секретов: %d\n", secrets.Total)

    // 5. Обновление секрета
    updateReq := &api.SecretRequest{
        Type: api.SecretTypeCredentials,
        Name: "Important Service (Updated)",
        Data: &api.Credentials{
            Username: "admin",
            Password: "new-super-secret",
            URL:      "https://important-service.com",
        },
        Metadata: "Updated production credentials",
    }

    updatedSecret, err := client.UpdateSecret(ctx, secret.ID, updateReq)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Обновлен секрет: %s\n", updatedSecret.Name)

    // 6. Удаление секрета
    err = client.DeleteSecret(ctx, secret.ID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Секрет удален")
}
```

### Синхронизация

```go
// Создание запроса на синхронизацию
syncReq := &api.SyncRequest{
    LastSyncTime: time.Now().Add(-24 * time.Hour), // последняя синхронизация 24 часа назад
    ClientHashes: map[uuid.UUID]string{
        // хеши секретов на клиенте
        secretID1: "hash1",
        secretID2: "hash2",
    },
}

syncResp, err := client.SyncSecrets(ctx, syncReq)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Обновлено секретов: %d\n", len(syncResp.UpdatedSecrets))
fmt.Printf("Удалено секретов: %d\n", len(syncResp.DeletedSecrets))
```

### Работа с разными типами данных

```go
// Извлечение данных из ответа
secret, err := client.GetSecret(ctx, secretID)
if err != nil {
    log.Fatal(err)
}

switch secret.Type {
case api.SecretTypeCredentials:
    if creds, ok := secret.GetCredentials(); ok {
        fmt.Printf("Username: %s, Password: %s\n", creds.Username, creds.Password)
    }
case api.SecretTypeText:
    if text, ok := secret.GetTextData(); ok {
        fmt.Printf("Content: %s\n", text.Content)
    }
case api.SecretTypeCard:
    if card, ok := secret.GetCardData(); ok {
        fmt.Printf("Card: %s, Holder: %s\n", card.Number, card.Holder)
    }
}
```

## Безопасность

- **Сквозное шифрование**: Все секреты шифруются на клиенте с помощью мастер-пароля
- **HTTPS**: Используйте HTTPS в продакшене
- **Токены**: JWT токены для аутентификации
- **Валидация**: Автоматическая валидация всех запросов

## Настройки для разработки

```go
// Для локальной разработки с самоподписанными сертификатами
config := api.LocalConfig()
config.WithInsecureSkipVerify(true)

client, err := api.NewClient(config)
```

## Лицензия

MIT License - см. файл [LICENSE](../../LICENSE) для деталей.