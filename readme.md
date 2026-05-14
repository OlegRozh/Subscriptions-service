## Subscription Service API
#### REST API сервис для управления подписками пользователей. Позволяет хранить информацию о подписках, отслеживать расходы и получать данные по пользователям, сервисам и периодам.
## 
**Возможности**

- ✅ **CRUD операции** — создание, чтение, обновление, удаление подписок
- ✅ **Аналитика расходов** — подсчёт суммы подписок за выбранный период
- ✅ **Фильтрация** — по пользователю, названию сервиса, датам
- ✅ **Контейнеризация** — готов к запуску в Docker
- ✅ **Документация API** — интерактивная Swagger документация

**Использованные технологии**

| Компонент | Технология |
|-----------|--------|
| **Язык** | Go 1.25.1 |
| **Роутер** | chi    |
| **База данных** | PostgreSQL 18.3 |
| **Драйвер БД** | pgxpool |
| **Миграции** | goose  |
| **Логирование** | slog   |
| **Контейнеризация** | Docker, Docker Compose |
| **Документация** | Swagger |

**Для работы приложения необходим файл .env с переменными окружения. Создайте его в корне проекта.**

```env
# PostgreSQL
POSTGRES_DB=Subscriptions
POSTGRES_USER=myuser
POSTGRES_PASSWORD=pg12345

# DATABASE_URL для Docker Compose
DATABASE_URL=postgres://myuser:pg12345@postgres:5432/Subscriptions?sslmode=disable

# Goose (миграции)
GOOSE_DRIVER=postgres
GOOSE_DBSTRING="user=myuser password=pg12345 dbname=Subscriptions host=localhost port=5432 sslmode=disable"
GOOSE_MIGRATION_DIR=./migrations
```
### Запуск через Docker

```bash
# Клонировать репозиторий
git clone https://github.com/OlegRozh/subscriptions-service.git

cd subscriptions-service

# Запустить сервис
docker-compose up --build -d
```
**API Эндпоинты**

| Метод | Эндпоинт | Описание |
|-------|----------|----------|
| POST | `/subscriptions` | Создать подписку |
| GET | `/subscriptions/{id}` | Получить подписку |
| GET | `/subscriptions/sum` | Сумма с фильтрацией |
| PUT | `/subscriptions/{id}` | Обновить подписку |
| DELETE | `/subscriptions/{id}` | Удалить подписку |

**Пример запроса:**

Создание подписки
```bash
curl -X POST http://localhost:8080/subscriptions \
  -H "Content-Type: application/json" \
  -d '{"service_name":"Yandex Plus","price":399,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_month":"2025-01-01T00:00:00Z"}'
```
Получаем сумму трат по созданной записи
```bash
curl "http://localhost:8080/subscriptions/sum?user_id=550e8400-e29b-41d4-a716-446655440000"
```
**Подробнее с документацией можно ознакомится после запуска сервера**
```bash
http://localhost:8080/swagger/index.html
```
**Запуск тестов**

Для программы предусмотрена проверка крайних тестовых случаев, например успешное создание, невалидный JSON, пустые поля и другие.
```bash
go test ./...
```