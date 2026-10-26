# Wallet API

## Запуск

```sh
cp .env.example config.env
make up
```

Миграции применяются автоматически, Swagger генерируется при сборке.

- Swagger: http://localhost:8000/swagger/index.html
- Проверка доступности: http://localhost:8000/health

## Тесты

Скопируйте пример настроек тестового окружения:

```sh
cp tests/env.example tests/config.env
make test-integration # Все Go-тесты с PostgreSQL в Docker
make test-load        # 1000 пополнений/с одного кошелька в течение 60 секунд
```

Тесты используют отдельную БД.

## Остальные команды

| Команда | Назначение |
|---|---|
| `make db` | Запустить PostgreSQL |
| `make migrate` | Применить миграции |
| `make migrate-status` | Показать статус миграций |
| `make logs` | Смотреть логи приложения и БД |
| `make ps` | Показать состояние контейнеров |
| `make docs` | Перегенерировать Swagger |
| `make test` | Запустить тесты сервиса и HTTP API с `-race` и покрытием |
| `make test-env` | Поднять тестовую БД и применить миграции |

Для `make docs` и `make test` нужен локально установленный Go.

## Остановка

```sh
make down      # Приложение
make test-down # Тестовое окружение
```
