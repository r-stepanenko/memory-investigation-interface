# Backend

Backend проекта **Event Memory Search API** реализован на Go и предоставляет REST API для поиска событий, получения контекста и объяснения расчёта score.

## Сборка

Собрать проект можно одной из команд:

```bash
go build ./...
```

или

```bash
make -f Makefile build
```

## Запуск

Запустить сервер можно одной из команд:

```bash
go run ./cmd/event-memory-search-api
```

или

```bash
make -f Makefile run
```

После запуска сервер будет доступен по адресу:

```text
http://localhost:8080
```

## Тестирование

Запустить HTTP-тесты:

```bash
go test -v ./internal/http
```

или

```bash
make -f Makefile test
```

## Проверка состояния сервера

Проверить, что сервер успешно запущен:

```text
GET http://localhost:8080/api/health
```

## Выполнение поиска

Пример запроса поиска (PowerShell):

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

* **search_id** — идентификатор поиска;
* **candidates** — найденные события;
* **score** — итоговая оценка совпадения;
* **matched_hints** — совпавшие поисковые условия.

## Получение результатов поиска

После выполнения поиска результаты доступны по адресу:

```text
GET http://localhost:8080/api/search/{search_id}
```

где `{search_id}` — идентификатор, полученный при выполнении поиска.

## Explain

Получить подробное объяснение расчёта score для найденного события:

```text
GET http://localhost:8080/api/search/{search_id}/candidates/{event_id}/explain
```

Ответ включает:

* итоговый score;
* вклад каждого совпадения (contributions);
* список несовпавших условий (missed_hints).

## Контекст события

Получить события, произошедшие до и после выбранного события:

```text
GET http://localhost:8080/api/events/{event_id}/context
```

При необходимости можно указать параметры запроса:

* `dataset` — идентификатор набора данных;
* `before` — окно поиска событий до выбранного события;
* `after` — окно поиска событий после выбранного события.

Пример:

```text
GET http://localhost:8080/api/events/{event_id}/context?dataset=control&before=30m&after=30m
```

# backend
cd server
go run ./cmd/event-memory-search-api

# frontend
cd front
npm install
npm run dev