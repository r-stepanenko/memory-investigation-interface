build:
go build ./...
make build


run:
go run ./cmd/event-memory-search-api
make run


test:
go test -v ./internal/http
make test


Проверка состояния сервера:
http://localhost:8080/api/health


Проверка поиска события:
curl -Method POST http://localhost:8080/api/search `
-Headers @{"Content-Type"="application/json"} `
-Body '{
  "dataset_id":"control",
  "hints":{
    "user_id":"ivan"
  }
}'

search_id: идентификатор поиска
candidates: найденные события
score: оценка совпадения
matched_hints: совпавшие условия

Получение результатов поиска:
http://localhost:8080/api/search/{search_id}

Объяснение кандидата:
http://localhost:8080/api/search/{search_id}/candidates/{event_id}/explain

Контекст события:
http://localhost:8080/api/events/{event_id}/context