# Frontend: Интерфейс расследования по памяти

Рабочее место аналитика для поиска событий по неточному описанию.

## Запуск
1. Установите зависимости: `npm install`
2. Запустите в режиме разработки: `npm run dev`

## Конфигурация
Переменные окружения находятся в `src/.env`:
- `VITE_API_URL` — URL backend API (по умолчанию `http://localhost:8080`)
- `VITE_USE_MOCK` — включить mock-режим без бэкенда (`true` / `false`)

## Тесты
- Unit и Component тесты: `npm run test`
- E2E тесты: `npm run test:e2e`

## Сборка
- `npm run build` — сборка production-версии
- `npm run preview` — предпросмотр сборки