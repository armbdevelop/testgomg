# Mortgage Calculator

Сервис ипотечного расчёта: создание расчёта с графиком платежей (аннуитет), дедупликация одинаковых запросов через Redis, график считается отдельным воркером.

## Диаграммы

[Диаграмма последовательностей и диаграмма зависимостей (FigJam)](https://www.figma.com/board/PtwsN3S6ETF21ZdZheKmZB/Untitled?node-id=0-1&t=s9vprZZqttwfF9o6-1)

## Стек

Go, fiber, Postgres (sqlx/pgx), Redis (go-redis), JWT (golang-jwt). Гексагональная архитектура: `domain` не зависит ни от чего, адаптеры реализуют `ports`.

## Запуск

```bash
docker compose up -d                                   # postgres + redis
psql -h localhost -U postgres -d mortgage -f migrations/0001_init.up.sql
cp config/config.example.json config/config.json       # подставить свои значения
go run ./cmd
```

Тестовый токен (sub = id юзера, секрет из конфига):

```bash
go run ./cmd/gentoken dev-secret-for-local-testing 12345
```

## API

### POST /mortgage-profiles

Заголовок `Authorization: Bearer <jwt>`. Создаёт расчёт атомарно (профиль + расчёт в одной транзакции) и ставит задачу воркеру.

```json
{
  "propertyPrice": 10000000,
  "propertyType": "apartment_in_new_building",
  "downPaymentAmount": 2000000,
  "matCapitalAmount": 500000,
  "matCapitalIncluded": true,
  "mortgageTermYears": 20,
  "interestRate": 8
}
```

Ответ `201`: `{"id": "1"}` (id строкой, по контракту ТЗ). Повторный запрос того же пользователя с идентичными параметрами вернёт тот же `id`. Готовый результат берётся из Redis; при промахе или ошибке кэша Postgres находит существующий расчёт. Незавершённый расчёт ставится в очередь. Worker пропускает вычисление, если предыдущая задача уже сохранила график.

### GET /mortgage-profiles/:id

Ответ `200`: суммы строками (по контракту ТЗ), вычет, рекомендуемый доход, `mortgagePaymentSchedule` и поле `status`:
`pending` — воркер ещё считает (график `null`), `done` — готово. Расчёт доступен только владельцу (проверка `user_id` из JWT в запросе).


## Тесты

```bash
go test ./...
```

Доменные тесты сверяют платёж с аннуитетными калькуляторами (10 млн цена, взнос 2 млн, 8%, 20 лет → 66 916 ₽/мес).

## Линтер

```bash
golangci-lint run ./...
```
