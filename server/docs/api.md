# Event Memory Search API

API для поиска событий в журнале действий пользователей.

## Base URL

```
http://localhost:8080
```

---

# GET /api/health

Проверка доступности сервиса.

## Response

```json
{
  "status": "ok"
}
```

---

# GET /api/datasets

Возвращает список доступных наборов событий.

## Response

```json
[
  "control"
]
```

---

# POST /api/search

Создаёт поиск событий по заданным критериям.

## Request

```json
{
  "dataset_id": "control",
  "hints": {
    "user_id": "ivan",
    "action": "file"
  },
  "context": {
    "require_nearby": {
      "action": "file_copy"
    }
  },
  "scoring": {
    "limit": 10,
    "min_score": 0
  }
}
```

## Поддерживаемые критерии

### Hints

* `user_id`
* `file_name`
* `action`
* `destination_type`
* `channel`
* `severity`

### Context

* `before`
* `after`
* `require_nearby`

### Scoring

* `limit`
* `min_score`

---

# Response

```json
{
  "search_id": "srch_1784553642878703600",
  "status": "done",
  "dataset_id": "control",
  "total_candidates": 24,
  "candidates": [
    {
      "score": 100,
      "matched_hints": [
        "user_id exact",
        "action exact"
      ],
      "event": {
        "event_id": "evt_32",
        "timestamp": "2026-06-20T11:20:00Z",
        "user_id": "ivan",
        "action": "file_copy",
        "file_name": "client_data.zip",
        "destination_type": "usb"
      }
    }
  ]
}
```

---

# GET /api/search/{search_id}

Возвращает ранее выполненный поиск.

## Response

```json
{
  "search_id": "srch_1784553642878703600",
  "status": "done",
  "dataset_id": "control",
  "total_candidates": 24,
  "candidates": [
    {
      "score": 90,
      "matched_hints": [
        "user_id exact",
        "action exact"
      ],
      "event": {
        "event_id": "evt_32",
        "timestamp": "2026-06-20T11:20:00Z",
        "user_id": "ivan",
        "action": "file_copy",
        "file_name": "client_data.zip",
        "destination_type": "usb"
      }
    },
    {
      "score": 67.5,
      "matched_hints": [
        "user_id exact",
        "action substring"
      ],
      "event": {
        "event_id": "evt_18",
        "timestamp": "2026-06-18T11:00:00Z",
        "user_id": "ivan",
        "action": "file_copy",
        "file_name": "contract.xlsx",
        "destination_type": "usb"
      }
    }
  ]
}
```

---

# GET /api/events/{event_id}/context

Возвращает выбранное событие и его временной контекст.

Ответ содержит:

* `event` — найденное событие;
* `before` — события до него;
* `after` — события после него.

## Response

```json
{
  "event": {},
  "before": [],
  "after": []
}
```

---

# GET /api/search/{search_id}/candidates/{event_id}/explain

Возвращает подробное объяснение расчёта score.

## Response

```json
{
  "search_id": "srch_1784554729740319700",
  "event_id": "evt_32",
  "score": 100,
  "contributions": [
    {
      "hint": "user_id",
      "type": "exact",
      "value": "ivan",
      "query": "ivan",
      "points": 45,
      "matched": true,
      "reason": "exact user id match"
    },
    {
      "hint": "action",
      "type": "exact",
      "value": "file_copy",
      "query": "file_copy",
      "points": 45,
      "matched": true,
      "reason": "exact action match"
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

# HTTP Status Codes

| Code | Description           |
| ---- | --------------------- |
| 200  | Success               |
| 400  | Invalid request       |
| 404  | Not found             |
| 405  | Method not allowed    |
| 500  | Internal server error |

---

# Score Calculation

Максимальный score: **100**

Score состоит из двух частей:

* **90 баллов** — распределяются между заполненными `hints`;
* **10 баллов** — резерв под `nearby`.

---

## Hints weight

Вес каждого hint:

```
weight = 90 / количество заполненных hints
```

---

## Пример: два hints

Запрос:

```json
{
  "hints": {
    "user_id": "ivan",
    "action": "file"
  }
}
```

Количество hints:

```
2
```

Вес каждого:

```
90 / 2 = 45
```

При полном совпадении:

```
user_id exact = +45
action exact = +45
```

Итог:

```
score = 90
```

---

# Частичное совпадение

Частичное совпадение даёт половину веса hint.

Пример:

Запрос:

```json
{
  "hints": {
    "user_id": "ivan"
  }
}
```

Событие:

```json
{
  "user_id": "ivanov"
}
```

Совпадение:

```
user_id substring
```

Расчёт:

```
weight = 90 / 1 = 90

substring = 90 / 2 = 45
```

Итог:

```
score = 45
```

---

# Nearby Bonus

`nearby` добавляет до 10 баллов при выполнении условия `require_nearby`.

Пример:

```json
{
  "hints": {
    "user_id": "ivan"
  },
  "context": {
    "require_nearby": {
      "action": "file_copy"
    }
  }
}
```

Расчёт:

```
user_id exact = +90

nearby event found = +10
```

Итог:

```
score = 100
```

Если nearby-условие не выполнено:

```
score = 90
```

---

# Sorting

Результаты сортируются:

1. По `score` по убыванию.
2. При одинаковом score — по времени события (`timestamp`) от новых к старым.

---

# CLI Search

Поддерживается отдельный CLI-поиск:

```bash
go run ./cmd/event-memory-search-api search
```

## Options

### Dataset

```bash
--datasets ./internal/datasets
```

### Custom events file

```bash
--events ./events.jsonl
```

### Query file

```bash
--query query.json
```

### Save output

```bash
--out result.json
```

Пример:

```bash
go run ./cmd/event-memory-search-api search \
  --query query.json \
  --out result.json
```

---

# Example query.json

```json
{
  "dataset_id": "control",
  "hints": {
    "user_id": "ivan",
    "action": "login"
  },
  "scoring": {
    "limit": 10,
    "min_score": 0
  }
}
```
