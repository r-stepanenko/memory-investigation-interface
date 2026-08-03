# API Documentation

## POST /api/search

This document describes the HTTP API of the event-memory-search service.

The API follows an OpenAPI-style REST documentation format.

## Overview

Backend API для поиска событий по неполному и неточному описанию аналитика.

Сервис:
- принимает поисковый запрос;
- ищет кандидатов среди событий;
- рассчитывает score;
- возвращает объяснение результата;
- предоставляет контекст события.

Base URL:

```
http://localhost:8080
```

---

# Health Check

## GET /api/health

Проверка состояния сервера.

### Response

```json
{
  "status": "ok"
}
```

---

# Datasets

## GET /api/datasets

Возвращает доступные наборы событий.

### Response

```json
[
  {
    "id": "control",
    "size": 33
  },
  {
    "id": "large100k",
    "size": 100000
  }
]
```

---

# Search

## POST /api/search

Поиск событий по описанию.

### Request

```json
{
  "dataset_id": "control",
  "hints": {
    "user_id": "ivan",
    "file_name": "client_data.zip",
    "action": "download"
  },
  "time": {
    "around": "2025-01-01T12:00:00Z",
    "tolerance": "10m"
  },
  "limit": 10,
  "min_score": 50
}
```

---

## Scoring

Каждый кандидат получает итоговый score:

```
score = hints_score + nearby_score
```

Максимум:

```
100 баллов
```

Распределение:

- совпадения по hints — до 90 баллов;
- nearby/context совпадение — до 10 баллов.

В ответе возвращаются:

- score;
- matched_hints;
- причины попадания кандидата.

---

# Search Result

## GET /api/search/{search_id}

Получение результата ранее выполненного поиска.

### Response

```json
{
  "search_id": "abc123",
  "candidates": [
    {
      "event_id": "evt_1",
      "score": 90,
      "matched_hints": [
        "user_id",
        "file_name"
      ]
    }
  ]
}
```

---

# Event Context

## GET /api/events/{event_id}/context

Возвращает окружение события.

Query parameters:

```
dataset
```

Пример:

```
GET /api/events/evt_32/context?dataset=test
```

Ответ содержит:

- событие;
- события до него;
- события после него.

---

# Explain

## GET /api/search/{search_id}/candidates/{event_id}/explain

Возвращает детализацию расчёта score.

Пример:

```json
{
  "event_id": "evt_1",
  "score": 90,
  "contributions": [
    {
      "field": "user_id",
      "points": 45
    },
    {
      "field": "file_name",
      "points": 45
    }
  ]
}
```

---

# Time Filtering

Поиск поддерживает временное ограничение:

Поля:

```
time.around
time.tolerance
```

Пример:

```json
{
  "time": {
    "around": "2025-01-01T12:00:00Z",
    "tolerance": "5m"
  }
}
```

---

# Nearby Search

Поддерживается поиск событий рядом с указанным событием.

Параметры:

```
require_nearby
before
after
```

Пример:

```json
{
  "context": {
    "require_nearby": {
      "action": "upload"
    },
    "before": "5m",
    "after": "5m"
  }
}
```

---

# Error Responses

API возвращает структурированные ошибки.

Пример:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "dataset not found"
  }
}
```

---

# Frontend Integration

Frontend подключается к backend API:

```
Frontend
   |
   |
HTTP JSON API
   |
   |
Event Memory Search Backend
```

Основные сценарии:

1. загрузка datasets;
2. отправка поиска;
3. просмотр кандидатов;
4. просмотр explain;
5. просмотр контекста события.

---

# Performance

Для проверки производительности используется benchmark:

```
go test -bench .
```

Поддерживаются наборы:

- 100k событий;
- 1M событий.
