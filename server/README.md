# Event Memory Search API — Backend

Backend проекта **Event Memory Search API** реализован на Go и предоставляет REST API для:

* поиска событий по неточному описанию;
* ранжирования кандидатов по score;
* получения контекста события;
* объяснения расчёта score.

---

# Сборка

Собрать проект можно одной из команд:

```bash
go build ./...
```

или:

```bash
make build
```

---

# Генерация больших данных

Для генерации тестовых датасетов:

```bash
make generate
```

Будут созданы файлы:

```
internal/datasets/events100k.jsonl
internal/datasets/events1m.jsonl
```

Данные используются для проверки производительности поиска.

---

# Запуск Backend

Запустить сервер:

```bash
go run ./cmd/event-memory-search-api
```

или:

```bash
make run
```

Также поддерживается запуск через CLI:

```bash
go run ./cmd/event-memory-search-api serve
```

После запуска API доступен:

```
http://localhost:8080
```

---

# API

## Health check

Проверка состояния сервера:

```
GET /api/health
```

Пример:

```
http://localhost:8080/api/health
```

---

## Получение списка датасетов

```
GET /api/datasets
```

Возвращает доступные наборы событий.

---

# Поиск событий

Endpoint:

```
POST /api/search
```

Пример запроса PowerShell:

```powershell
curl -Method POST http://localhost:8080/api/search `
-Headers @{"Content-Type"="application/json"} `
-Body '{
  "dataset_id":"control",
  "hints":{
    "user_id":"ivan"
  }
}'
```

Ответ содержит:

* `search_id` — идентификатор поиска;
* `candidates` — найденные события;
* `score` — итоговую оценку совпадения;
* `matched_hints` — совпавшие условия;
* `contributions` — вклад каждого совпадения.

---

# Получение результатов поиска

После выполнения поиска результаты доступны:

```
GET /api/search/{search_id}
```

где:

```
{search_id}
```

— идентификатор поиска из ответа `/api/search`.

---

# Explain

Подробное объяснение расчёта score:

```
GET /api/search/{search_id}/candidates/{event_id}/explain
```

Ответ содержит:

* итоговый score;
* вклад каждого совпадения (`contributions`);
* несовпавшие условия (`missed_hints`).

---

# Контекст события

Получение событий вокруг выбранного события:

```
GET /api/events/{event_id}/context
```

Поддерживаемые параметры:

* `dataset` — идентификатор датасета;
* `before` — временное окно до события;
* `after` — временное окно после события.

Пример:

```
GET /api/events/evt_1/context?dataset=control&before=30m&after=30m
```

---

# Тестирование

Запуск тестов:

```bash
go test ./...
```

или:

```bash
make test
```

HTTP-тесты:

```bash
go test -v ./internal/http
```

---

# Frontend integration

## Backend

```bash
cd server
go run ./cmd/event-memory-search-api serve
```

или:

```bash
make demo
```

---

## Frontend

```bash
cd front
npm install
npm run dev
```

Frontend подключается к Backend:

```
http://localhost:8080
```

Используемые API:

```
GET  /api/datasets
POST /api/search
GET  /api/events/{id}/context
```

---

# API Documentation

Подробное описание API:

```
docs/api.md
```

Документация содержит:

* описание endpoints;
* форматы запросов;
* форматы ответов;
* обработку ошибок.

API построен в стиле OpenAPI.

---

# Benchmark

Используемые датасеты:

* 100k событий;
* 1M событий.

Запуск:

```bash
make bench
```

Результаты:

| Dataset     | Время поиска |
| ----------- | ------------ |
| 100k events | ~500 ms      |
| 1M events   | ~1.8 s       |

---

# Frontend Demo

Запуск демонстрации:

Backend:

```bash
make demo
```

Frontend:

```bash
cd front
npm install
npm run dev
```

Демонстрация использует:

```
GET  /api/datasets
POST /api/search
GET  /api/events/{id}/context
```

---

# Verification

Проверено:

```bash
go test ./...
```

и:

```powershell
.\check.ps1
```

Результат проверки:

* Backend запускается;
* API endpoints работают;
* поиск возвращает кандидатов;
* score рассчитывается;
* explain возвращает детализацию;
* context работает;
* тесты проходят успешно.
