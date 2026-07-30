# Event Memory Search API

REST API для поиска и анализа событий пользователей.

## Base URL

```
http://localhost:8080
```

---

# Health Check

## GET /api/health

Проверяет доступность сервиса.

### Response

Status: `200 OK`

```json
{
  "status": "ok"
}
```

---

# Datasets

## GET /api/datasets

Возвращает список доступных наборов событий.

### Response

Status: `200 OK`

```json
[
  "control",
  "test",
  "large100k"
]
```

---

# Search

## POST /api/search

Создаёт поиск событий по заданным критериям.

---

## Request

```json
{
  "dataset_id": "control",
  "time": {
    "around": "2026-06-20T11:00:00Z",
    "tolerance": "10m"
  },
  "hints": {
    "user_id": "ivan",
    "file_name": "report.docx",
    "action": "login",
    "destination_type": "usb",
    "channel": "web",
    "severity": "high"
  },
  "context": {
    "before": "30m",
    "after": "30m",
    "require_nearby": {
      "action": "create_archive",
      "within": "10m"
    }
  },
  "scoring": {
    "limit": 10,
    "min_score": 0
  }
}
```

---

# Search Hints

Поддерживаемые критерии:

| Поле             | Описание          |
| ---------------- | ----------------- |
| user_id          | Пользователь      |
| file_name        | Имя файла         |
| action           | Тип действия      |
| destination_type | Тип назначения    |
| channel          | Канал события     |
| severity         | Уровень опасности |

---

# Time Filter

Фильтр по времени:

```json
{
  "time": {
    "around": "2026-06-20T11:00:00Z",
    "tolerance": "10m"
  }
}
```

События выбираются из диапазона:

```
around - tolerance
до
around + tolerance
```

Пример:

```
11:00 ± 10 минут

10:50 — 11:10
```

---

# Context Search

## GET /api/events/{event_id}/context

Возвращает событие и связанные события вокруг него.

Ответ содержит:

* `event` — выбранное событие;
* `before` — события до него;
* `after` — события после него.

### Response

```json
{
  "event": {},
  "before": [],
  "after": []
}
```

---

# Nearby Search

Поле:

```json
{
  "context": {
    "require_nearby": {
      "action": "file_copy",
      "within": "10m"
    }
  }
}
```

Проверяет наличие связанного события рядом по времени.

При успешном выполнении добавляется бонус:

```
nearby = +10 баллов
```

---

# Search Response

## POST /api/search

### Response

```json
{
  "search_id": "srch_123",
  "status": "done",
  "dataset_id": "control",
  "total_candidates": 10,
  "candidates": [
    {
      "score": 100,
      "matched_hints": [
        "user_id exact",
        "action exact"
      ],
      "event": {
        "event_id": "evt_1",
        "timestamp": "2026-06-20T11:00:00Z",
        "user_id": "ivan",
        "action": "login"
      }
    }
  ]
}
```

---

# Search By ID

## GET /api/search/{search_id}

Возвращает ранее выполненный поиск.

---

# Explain Score

## GET /api/search/{search_id}/candidates/{event_id}/explain

Возвращает детализацию расчёта score.

### Response

```json
{
  "search_id": "srch_123",
  "event_id": "evt_1",
  "score": 100,
  "contributions": [
    {
      "hint": "user_id",
      "type": "exact",
      "points": 90,
      "matched": true,
      "reason": "exact user id match"
    },
    {
      "hint": "nearby",
      "type": "matched",
      "points": 10,
      "matched": true,
      "reason": "required nearby event found"
    }
  ]
}
```

---

# Score Calculation

Максимальный score:

```
100
```

Расчёт:

```
90 баллов — hints
10 баллов — nearby bonus
```

---

## Hint Weight

Вес распределяется между заполненными критериями:

```
weight = 90 / количество hints
```

---

## Exact Match

Пример:

```
user_id = ivan
event.user_id = ivan
```

Результат:

```
+90
```

---

## Partial Match

Подстрочное совпадение:

```
query: ivan
event: ivanov
```

Начисляется:

```
50% веса hint
```

---

# Sorting

Результаты сортируются:

1. По `score` по убыванию.
2. При одинаковом score:

   * новые события выше старых.

---

# Error Responses

Все ошибки возвращаются в JSON формате.

Пример:

```json
{
  "code": "dataset_not_found",
  "message": "dataset unknown not found"
}
```

---

# HTTP Status Codes

| Код | Описание                |
| --- | ----------------------- |
| 200 | Успешный запрос         |
| 400 | Некорректный запрос     |
| 404 | Ресурс не найден        |
| 405 | Метод не поддерживается |
| 500 | Ошибка сервера          |

---

# CLI Search

Поддерживается поиск через CLI:

```bash
go run ./cmd/event-memory-search-api search
```

Параметры:

```
--datasets
--events
--query
--out
```

Пример:

```bash
go run ./cmd/event-memory-search-api search \
--query query.json \
--out result.json
```

---

# Testing

Запуск тестов:

```bash
go test ./...
```

Benchmark:

```bash
go test -bench=.
```

---

# Project Structure

```
cmd/
 └── event-memory-search-api/

internal/
 ├── domain/
 ├── http/
 ├── search/
 └── datasets/

docs/
 └── api.md
```
