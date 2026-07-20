# Event Memory Search API

API для поиска событий в журнале действий пользователей.

## Base URL

```
http://localhost:8080
```

---

## GET /api/health

Проверка доступности сервиса.

### Response

```json
{
  "status": "ok"
}
```

---

## GET /api/datasets

Возвращает список доступных наборов событий.

### Response

```json
[
  "control"
]
```

---

## POST /api/search

Создаёт поиск событий по заданным критериям.

## Request

```json
{
  "dataset_id": "control",
  "hints": {
    "user_id": "ivan",
    "action": "file"
  }
}
```

Поддерживаемые критерии:

- user_id
- file_name
- action
- destination_type

---

## Response

```json
{
  "search_id": "srch_1784553642878703600",
  "status": "done",
  "dataset_id": "control",
  "total_candidates": 24,
  "candidates": [
    {
      "score": 75,
      "matched_hints": [
        "user_id exact",
        "action substring"
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

## GET /api/search/{search_id}

Возвращает ранее выполненный поиск

### Response

```json
{
  "search_id": "srch_1784553642878703600",
  "status": "done",
  "dataset_id": "control",
  "total_candidates": 24,
  "candidates": [
    {
      "score": 75,
      "matched_hints": [
        "user_id exact",
        "action substring"
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
      "score": 50,
      "matched_hints": [
        "user_id substring",
        "action substring"
      ],
      "event": {
        "event_id": "evt_18",
        "timestamp": "2026-06-18T11:00:00Z",
        "user_id": "ivanov",
        "action": "file_copy",
        "file_name": "contract.xlsx",
        "destination_type": "usb"
      }
    }
  ]
}
```

---

## GET /api/events/{event_id}/context

Возвращает выбранное событие и его временной контекст.

Ответ содержит:

- event — найденное событие
- before — события до него
- after — события после него

### Response

```json
{
  "event": {},
  "before": [],
  "after": []
}
```

---

## GET /api/search/{search_id}/candidates/{event_id}/explain

Возвращает объяснение расчёта score.

### Response

```json
{
  "search_id": "srch_1784554729740319700",
  "event_id": "evt_32",
  "score": 75,
  "contributions": [
    {
      "hint": "user_id",
      "type": "exact",
      "value": "ivan",
      "query": "ivan",
      "points": 50
    },
    {
      "hint": "action",
      "type": "substring",
      "value": "file_copy",
      "query": "file",
      "points": 25
    }
  ]
}
```

---

# HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 400 | Invalid request |
| 404 | Not found |
| 405 | Method not allowed |
| 500 | Internal server error |

---

# Score

Максимальный score: 100

Вес каждого hint: weight = 100 / количество заполненных hints

Пример:

2 hints:

```json
{
  "user_id": "ivan",
  "action": "file"
}
```

Каждый hint имеет вес: 100 / 2 = 50
Расчёт: score = user_id + action = 50 + 50 = 100

Частичное совпадение даёт половину веса:

Пример:

user_id = ivan
action = file

Вес каждого hint:

100 / 2 = 50

Запрос: user_id = ivan
Событие: user_id = ivanov
Совпадение частичное: user_id substring = 50 / 2 = +25

Запрос: action = file
Событие: action = file
Совпадение точное: action exact = 50 / 1 = +50

Итоговый score: user_id substring + action exact = 25 + 50 = 75