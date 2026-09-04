# Avito.Kitchen MVP

Backend-сервис агрегатора доставки еды. Приложение позволяет клиенту найти заведения и блюда, создать пустой заказ, постепенно добавить позиции, убрать разрешённые ингредиенты и получить оценку времени.

## Быстрый запуск

### Локально

По умолчанию используется in-memory repository, поэтому для запуска не нужны PostgreSQL, Kafka или дополнительные сервисы:

```bash
go run ./cmd/kitchen-api
```

API будет доступен на `http://localhost:8080`.

### Docker Compose

```bash
docker compose up --build
```

Compose запускает API, PostgreSQL, PostgreSQL exporter, Prometheus, Grafana, Kafka и Zookeeper. Клиентское приложение работает через in-memory repository, а PostgreSQL подготовлен для переключения на SQL-адаптер.

Сервисы:

- API: `http://localhost:8080`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000`
- PostgreSQL: `localhost:5432`

## Пример клиентского сценария

```bash
curl http://localhost:8080/api/v1/shops
curl 'http://localhost:8080/api/v1/dishes?type=main&sort=price_desc&page=1&page_size=20'
curl http://localhost:8080/api/v1/dishes/2
curl -X POST http://localhost:8080/api/v1/orders \
  -H 'Content-Type: application/json' \
  -d '{"shop_id":1}'
curl -X POST http://localhost:8080/api/v1/orders/1/items \
  -H 'Content-Type: application/json' \
  -d '{"dish_id":2,"quantity":1,"excluded_ingredient_ids":[3]}'
curl http://localhost:8080/api/v1/orders/1/estimate
curl http://localhost:8080/api/v1/orders/1
```

## API

Полный контракт находится в [`openapi/kitchen-api.yaml`](openapi/kitchen-api.yaml).

Основные ручки:

- `GET /api/v1/shops` — список заведений.
- `GET /api/v1/shops/{shopId}` — карточка заведения.
- `GET /api/v1/shops/{shopId}/dishes` — меню заведения.
- `GET /api/v1/dishes` — общий каталог с фильтрами `type`, `shop_id`, поиском, сортировкой и пагинацией.
- `GET /api/v1/dishes/{dishId}` — карточка блюда.
- `POST /api/v1/orders` — создание пустого заказа.
- `POST /api/v1/orders/{orderId}/items` — добавление позиции.
- `PATCH /api/v1/orders/{orderId}/items/{itemId}` — изменение количества и исключённых ингредиентов.
- `DELETE /api/v1/orders/{orderId}/items/{itemId}` — удаление позиции.
- `GET /api/v1/orders/{orderId}/estimate` — оценка приготовления и доставки.
- `GET /api/v1/orders/{orderId}` — состояние заказа и его позиций.
- `GET /metrics` — метрики Prometheus.

## Архитектура

Проект использует Clean Architecture с направлением зависимостей внутрь:

```text
cmd/kitchen-api
    -> internal/delivery/http
    -> internal/services
    -> internal/repository
    -> internal/domain
```

- `cmd` содержит только сборку зависимостей и запуск HTTP-сервера.
- `internal/domain` содержит сущности и бизнес-статусы без зависимостей от HTTP или БД.
- `internal/services` содержит use case каталога и заказов.
- `internal/repository` содержит интерфейс-порт, in-memory адаптер и PostgreSQL адаптер.
- `internal/dto` содержит request/response-модели внешнего API и преобразования domain → response.
- `internal/delivery/http` отвечает за маршрутизацию, валидацию HTTP-входа и формирование ответов.
- `internal/kafka` содержит Kafka producer и consumer для обмена событиями с заведениями.
- `internal/metrics` содержит Prometheus middleware.

## Request/response модели и DTO

HTTP-модели не находятся в repository и не смешиваются с domain:

- request-типы находятся в `internal/dto/requests.go`;
- response-типы находятся в `internal/dto/responses.go`;
- функции `ShopFromDomain`, `DishFromDomain` и `OrderFromDomain` преобразуют внутренние сущности в публичный формат.

Repository работает только с domain-моделями. Это позволяет менять JSON-контракт, не изменяя SQL-слой, и не протекать HTTP-деталям в бизнес-логику.

## Chi

Для HTTP-маршрутизации используется `github.com/go-chi/chi/v5`. Router собирается в `internal/delivery/http/handler.go`, подключает middleware метрик и разделяет health, metrics и API-маршруты. Chi выбран как небольшой идиоматичный router для стандартного `net/http`, без привязки application/service-слоёв к конкретному веб-фреймворку.

## Kafka и интеграция с заведением

Use case зависит от интерфейса `EventPublisher`, а не от Kafka-клиента. Все события публикуются в topic `orders.events`:

- `order.created` — создан пустой заказ;
- `order.item.added` — в заказ добавлена позиция;
- `order.accepted` — заведение подтвердило заказ;
- `order.rejected` — заведение не может выполнить заказ, поле `reason` содержит причину.

При запуске через Docker Compose API публикует события в Kafka и одновременно читает ответы ресторана consumer-группой `kitchen-api`. Для подтверждения или отказа ресторан публикует в тот же topic JSON-событие с `type`, `order_id`, `shop_id` и, для отказа, `reason`:

```json
{
  "type": "order.rejected",
  "order_id": 1,
  "shop_id": 1,
  "reason": "Некоторые блюда закончились",
  "created_at": "2026-09-04T09:00:00Z"
}
```

После получения ответа статус заказа меняется на `accepted` или `rejected` и доступен через `GET /api/v1/orders/{orderId}`. Переменные `KAFKA_BROKERS` и `KAFKA_ORDERS_TOPIC` включают Kafka-адаптер. Если `KAFKA_BROKERS` не задан, локальный `go run` использует `LoggingPublisher` и пишет события в лог без внешних зависимостей.

## Repository и sqlc

Для локальной работы по умолчанию используется `InMemoryRepository` с демонстрационными заведениями и блюдами. Переключение на PostgreSQL выполняется настройками:

```bash
REPOSITORY_DRIVER=postgres
DATABASE_URL='postgres://postgres:postgres@localhost:5432/kitchen?sslmode=disable'
```

SQL-схема находится в `migrations/`. Запросы для sqlc — в `internal/repository/queries/`, конфигурация — в `sqlc.yaml`, сгенерированные типы — в `internal/repository/db/`. PostgreSQL-адаптер сохраняет тот же repository-порт, что и in-memory реализация.

## Данные и расчёт времени

Миграции используют нормализованные таблицы `shops`, `dishes`, `ingredients`, `dish_ingredients`, `orders`, `order_items` и `order_item_excluded_ingredients`. Исключить можно только ингредиент, у которого в составе блюда установлен `removable = true`; недопустимое исключение возвращает ошибку.

Время считается как максимальное время приготовления среди позиций плюс время доставки заведения. Это MVP-упрощение модели параллельной кухни: позиции одного заказа считаются готовящимися одновременно.

## Наблюдаемость

API считает количество HTTP-запросов с labels `method`, `route`, `status` и отдаёт их на `/metrics`. PostgreSQL exporter собирает метрики базы, Prometheus хранит и опрашивает метрики API и exporter, Grafana подключается к Prometheus через provisioning-конфигурацию в `monitoring/grafana`.

## Проверка

```bash
go test ./...
go vet ./...
docker compose config --quiet
```

Авторизация, платежи, доставка, полноценный Kafka broker-клиент и административное API в MVP не реализованы, чтобы оставить клиентский сценарий компактным и запускаемым одной командой.
