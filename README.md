# GophKeeper

GophKeeper - это клиент-серверная система для безопасного хранения паролей, текстовых данных, бинарных файлов и данных банковских карт.

## Возможности

### Сервер
- Регистрация и аутентификация пользователей с JWT токенами
- Безопасное хранение приватных данных с шифрованием AES-GCM
- Синхронизация данных между несколькими клиентами
- Защищенная передача данных по HTTPS
- REST API для всех операций

### Клиент (CLI)
- Кросс-платформенное CLI приложение (Windows, Linux, macOS)
- Управление учетными данными (логин/пароль)
- Хранение произвольных текстовых данных
- Работа с бинарными файлами
- Управление данными банковских карт
- Автоматическая синхронизация с сервером
- Локальное кэширование для офлайн доступа

### Типы данных
- **Credentials** - пары логин/пароль с URL и метаданными
- **Text** - произвольные текстовые данные
- **Binary** - бинарные файлы любого размера
- **Card** - данные банковских карт (номер, срок действия, CVV, держатель)

## Архитектура

Проект использует современные подходы к разработке:

- **Clean Architecture** - разделение на слои (models, storage, service, handlers)
- **Шифрование** - AES-GCM с ключами, производными от Argon2
- **База данных** - PostgreSQL с миграциями
- **HTTP API** - RESTful API с Gin framework
- **CLI** - Cobra для командной строки
- **Конфигурация** - Viper для управления настройками
- **Тестирование** - Комплексные тесты с покрытием 60-100%

## Быстрый старт

### Требования
- Go 1.21+
- PostgreSQL 12+
- Make (опционально)

### Установка и запуск

1. **Клонирование репозитория:**
```bash
git clone https://github.com/uryumtsevaa/gophkeeper.git
cd gophkeeper
```

2. **Установка зависимостей:**
```bash
make deps
# или
go mod download
```

3. **Запуск базы данных (Docker):**
```bash
make db-up
```

4. **Сборка приложений:**
```bash
make build
# или отдельно
make build-server
make build-client
```

5. **Запуск сервера:**
```bash
# С помощью Makefile
make run-server

# Или напрямую
./bin/gophkeeper-server --jwt-secret=your-secret-key \
  --db-password=password
```

6. **Использование клиента:**
```bash
# Регистрация нового пользователя
./bin/gophkeeper-client auth register

# Вход в систему
./bin/gophkeeper-client auth login

# Добавление пароля
./bin/gophkeeper-client secrets add credentials

# Просмотр всех секретов
./bin/gophkeeper-client secrets list

# Синхронизация
./bin/gophkeeper-client sync
```

## Конфигурация

### Сервер

Сервер можно настроить через:
- Файл конфигурации `config.yaml`
- Переменные окружения
- Флаги командной строки

**Основные настройки:**
```yaml
port: 8080
jwt_secret: "your-secret-key"
database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "password"
  name: "gophkeeper"
  sslmode: "prefer"
```

**Переменные окружения:**
```bash
GOPHKEEPER_PORT=8080
GOPHKEEPER_JWT_SECRET=your-secret-key
GOPHKEEPER_DATABASE_HOST=localhost
GOPHKEEPER_DATABASE_PORT=5432
GOPHKEEPER_DATABASE_USER=postgres
GOPHKEEPER_DATABASE_PASSWORD=password
GOPHKEEPER_DATABASE_NAME=gophkeeper
```

### Клиент

Клиент сохраняет настройки в директории `~/.gophkeeper/`:
- `token` - JWT токен авторизации
- `secrets.json` - локальный кэш секретов

## API Документация

### Аутентификация

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "user",
  "email": "user@example.com",
  "password": "password"
}
```

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "user",
  "password": "password"
}
```

### Управление секретами

```http
GET /api/v1/secrets
Authorization: Bearer <token>
X-Master-Password: <master-password>
```

```http
POST /api/v1/secrets
Authorization: Bearer <token>
X-Master-Password: <master-password>
Content-Type: application/json

{
  "type": "credentials",
  "name": "My Login",
  "data": {
    "username": "user",
    "password": "pass",
    "url": "https://example.com"
  },
  "metadata": "Important account"
}
```

### Синхронизация

```http
POST /api/v1/sync
Authorization: Bearer <token>
X-Master-Password: <master-password>
Content-Type: application/json

{
  "last_sync_time": "2023-01-01T00:00:00Z",
  "client_hashes": {
    "uuid": "hash"
  }
}
```

## Команды CLI

### Аутентификация
```bash
# Регистрация
gophkeeper-client auth register -u username -e email@example.com

# Вход
gophkeeper-client auth login -u username

# Выход
gophkeeper-client auth logout
```

### Управление секретами
```bash
# Список секретов
gophkeeper-client secrets list

# Добавление учетных данных
gophkeeper-client secrets add credentials -n "My Account" -u username

# Добавление текста
gophkeeper-client secrets add text -n "My Note" -c "Secret note"

# Добавление файла
gophkeeper-client secrets add binary -n "My File" -f /path/to/file

# Добавление карты
gophkeeper-client secrets add card -n "My Card"

# Просмотр секрета
gophkeeper-client secrets get <secret-id>

# Удаление секрета
gophkeeper-client secrets delete <secret-id>
```

### Синхронизация
```bash
# Синхронизация с сервером
gophkeeper-client sync
```

### Версия
```bash
# Информация о версии
gophkeeper-client version
gophkeeper-server version
```

## Разработка

### Структура проекта
```
.
├── cmd/                    # Точки входа
│   ├── client/            # CLI клиент
│   └── server/            # HTTP сервер
├── internal/              # Внутренние пакеты
│   ├── client/           # Клиентская логика и локальное хранилище
│   ├── config/           # Управление конфигурацией
│   ├── crypto/           # Шифрование и безопасность
│   ├── interfaces/       # Общие интерфейсы
│   ├── models/           # Модели данных
│   ├── router/           # HTTP роутинг и middleware
│   ├── server/           # Серверная бизнес-логика
│   └── storage/          # Работа с БД
├── migrations/           # Миграции БД
├── pkg/                 # Публичные пакеты
│   └── api/             # API клиент и типы
├── config.yaml          # Конфигурация
├── Makefile            # Команды сборки
└── README.md
```

### Команды разработки

```bash
# Тестирование
make test
make test-coverage

# Линтинг
make lint

# Форматирование кода
make fmt

# Сборка для разных платформ
make build-all

# Очистка
make clean
```

### Тестирование

Проект включает комплексные тесты с высоким покрытием:

**Покрытие по компонентам:**
- **internal/router** - 100% (HTTP роутинг, middleware)
- **internal/crypto** - 82.6% (шифрование, безопасность)
- **internal/client** - 79.7% (клиентская логика, локальное хранилище)
- **internal/storage** - 65.8% (работа с БД, конфигурация)
- **internal/server** - 58.2% (серверная бизнес-логика)

**Типы тестов:**
- Юнит-тесты для всех компонентов
- Интеграционные тесты HTTP API
- Тесты шифрования/расшифровки
- Тесты локального хранилища
- Тесты конфигурации и роутинга
- Бенчмарки производительности

```bash
# Запуск всех тестов
go test ./...

# Тесты с покрытием
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Тестирование конкретных компонентов
go test -cover ./internal/crypto    # Шифрование
go test -cover ./internal/client    # Клиент
go test -cover ./internal/server    # Сервер
go test -cover ./internal/router    # Роутинг

# Бенчмарки
go test -bench=. ./internal/crypto/
```

## Безопасность

### Шифрование
- **Алгоритм:** AES-256-GCM
- **Деривация ключей:** Argon2id
- **Аутентификация:** HMAC с проверкой целостности
- **Соли:** Случайные 32-байтовые соли для каждого секрета

### Передача данных
- Все API endpoints защищены JWT токенами
- Мастер-пароль передается в заголовке и не сохраняется на сервере
- Поддержка HTTPS для защищенной передачи

### Хранение
- Пароли пользователей хешируются с Argon2
- Секреты шифруются индивидуально
- Локальное хранение защищено правами доступа

## Производительность

- Шифрование: ~100MB/s на современном CPU
- Синхронизация: только измененные данные
- Локальное кэширование для быстрого доступа
- Эффективная индексация в PostgreSQL

## Лицензия

MIT License - см. файл [LICENSE](LICENSE)

## Вклад в проект

1. Fork репозитория
2. Создайте ветку для новой функции
3. Внесите изменения с тестами
4. Убедитесь, что все тесты проходят
5. Создайте Pull Request

## Поддержка

Если у вас есть вопросы или предложения:
- Создайте Issue в GitHub
- Обратитесь к документации API
- Проверьте существующие тесты для примеров использования
