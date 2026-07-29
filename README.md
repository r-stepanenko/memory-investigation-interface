# Event Memory Search API

Проект предназначен для поиска событий в наборах данных с использованием системы оценки **score**, анализа контекста событий и объяснения результатов поиска **Explain**.

Проект состоит из двух частей:

* **server/** — backend на Go;
* **front/** — frontend на React + TypeScript + Vite.

---

# Структура проекта

```text
project/
├── server/          # Backend (Go)
├── front/           # Frontend (React + TypeScript + Vite)
├── README.md
├── Makefile
└── check.ps1
```

---

# Требования

Перед запуском убедитесь, что установлены:

* Go 1.24 или новее
* Node.js 20 или новее
* npm
* GNU Make

---

# Быстрый запуск

Полный запуск проекта:

```bash
make demo
```

Если используется Windows и Makefile не определяется автоматически:

```powershell
make -f .\Makefile demo
```

Команда выполняет:

1. установку зависимостей;
2. сборку backend и frontend;
3. запуск тестов;
4. запуск backend и frontend.

После запуска:

Backend:

```text
http://localhost:8080
```

Frontend:

```text
http://localhost:5173
```

---

# Команды Makefile

## Установка зависимостей

```bash
make install
```

Windows:

```powershell
make -f .\Makefile install
```

Устанавливает:

* Go-зависимости backend;
* npm-зависимости frontend.

---

## Сборка проекта

```bash
make build
```

Собирает:

* backend на Go;
* production-сборку frontend.

---

## Запуск backend

```bash
make run-server
```

или вручную:

```bash
cd server
go run ./cmd/event-memory-search-api
```

Backend доступен:

```text
http://localhost:8080
```

---

## Запуск frontend

```bash
make run-front
```

или вручную:

```bash
cd front
npm install
npm run dev
```

Frontend доступен:

```text
http://localhost:5173
```

---

## Одновременный запуск backend и frontend

```bash
make serve
```

Открывает два процесса:

* backend;
* frontend.

---

# API

| Метод | Endpoint                                                | Назначение                          |
| ----- | ------------------------------------------------------- | ---------------------------------   |
| GET   | `/api/health`                                           | Проверка готовности сервера         |
| GET   | `/api/datasets`                                         | Список доступных наборов данных     |
| POST  | `/api/search`                                           | Создать поиск по неточному описанию |
| GET   | `/api/search/{search_id}`                               | Получить результаты поиска          |
| GET   | `/api/events/{event_id}/context`                        | Получить контекст вокруг события    |
| GET   | `/api/search/{search_id}/candidates/{event_id}/explain` | Получить объяснение score           |


---

# Выполнение поиска

Пример запроса:

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
* `score` — итоговую оценку;
* `matched_hints` — совпавшие условия.

---

# Расчёт Score

Максимальный score:

```text
100
```

Распределение:

* до **90 баллов** — совпадения hints;
* до **10 баллов** — совпадения nearby-событий.

Explain содержит:

* итоговый score;
* вклад каждого совпадения;
* matched hints;
* missed hints;
* вклад nearby.

---

# Возможности поиска

Поддерживаются:

* поиск по пользователю;
* поиск по имени файла;
* поиск по действию;
* поиск по каналу;
* поиск по уровню важности;
* поиск по типу назначения;
* поиск по времени;
* поиск с nearby-событиями;
* фильтрация по минимальному score;
* ограничение количества результатов;
* Explain результатов.

---

# Nearby-события

Поддерживается поиск связанных событий:

* `before`
* `after`
* `require_nearby`

Пример:

```json
{
  "context": {
    "before": "30m",
    "after": "30m",
    "require_nearby": [
      {
        "action": "file_download",
        "within": "10m"
      }
    ]
  }
}
```

---

# Контекст событий

Endpoint:

```text
GET /api/events/{event_id}/context
```

Возвращает:

* выбранное событие;
* события до него;
* события после него.

Пример:

```text
GET /api/events/evt_33/context?dataset=control&before=30m&after=30m
```

---

# Индексация поиска

Для ускорения поиска используется индекс пользователей:

* построение индекса выполняется при запуске сервера;
* поиск по `user_id` использует индекс вместо полного перебора событий.

Поддерживаются большие наборы данных:

* 100K событий;
* 1M событий.

---

# Benchmark

Запуск benchmark:

```bash
make bench
```

Windows:

```powershell
make -f .\Makefile bench
```

Benchmark выполняется для:

* `100K` событий;
* `1M` событий.

Пример результатов:

```text
BenchmarkSearch100K
~10 ms/op

BenchmarkSearch1M
~100 ms/op
```

Проверяется:

* скорость поиска через индекс;
* количество аллокаций памяти;
* работа CalculateScore.

---

# Тестирование

Все тесты проекта:

```bash
make test
```

Windows:

```powershell
make -f .\Makefile test
```

---

## Backend

Все Go-тесты:

```bash
cd server
go test ./...
```

HTTP-тесты:

```bash
go test ./internal/http -v
```

Тесты поиска:

```bash
go test ./internal/search -v
```

---

## Frontend

Frontend-тесты:

```bash
cd front
npm run test:run
```

Проверяются:

* поиск;
* datasets;
* Explain;
* Context.

---

# Mock-режим

Frontend поддерживает работу без backend.

Включение:

```env
VITE_USE_MOCK=true
```

В режиме mock используются встроенные тестовые данные.

---

# Используемые технологии

## Backend

* Go
* net/http
* encoding/json

## Frontend

* React
* TypeScript
* Vite
* Vitest
* Testing Library

---