# Решение задачи Event Memory Search API

## Назначение

Backend-сервис выполняет поиск событий по неполному и неточному описанию аналитика.

Система:

- загружает наборы событий из JSONL;
- принимает поисковый запрос;
- выполняет фильтрацию по времени;
- рассчитывает score кандидатов;
- формирует объяснение результата;
- возвращает контекст события.


## Архитектура

Проект разделён на:

- cmd/event-memory-search-api — запуск приложения;
- internal/http — HTTP API;
- internal/search — алгоритм поиска и scoring;
- internal/domain — модели данных;
- internal/datasets — загрузка данных.


## Алгоритм поиска

Поиск выполняется в несколько этапов:

1. Получение запроса.
2. Поиск кандидатов.
3. Расчёт score.
4. Сортировка кандидатов.
5. Формирование explain.


## Расчёт score

Итоговый score:

- 90 баллов — совпадения по hints;
- 10 баллов — nearby context.


## Context

Поддерживается:

- before;
- after;
- nearby action.


## Тестирование

Проверка:
go test ./...


Benchmark:
go test -bench=. ./internal/search



## API

Основные endpoints:

- GET /api/health
- GET /api/datasets
- POST /api/search
- GET /api/search/{search_id}
- GET /api/events/{event_id}/context
- GET /api/search/{search_id}/candidates/{event_id}/explain
