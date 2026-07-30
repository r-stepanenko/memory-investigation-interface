# Решение

## Архитектура

Backend реализован на Go.

Основные компоненты:

- HTTP API
- Dataset loader
- Search engine
- Scoring engine

## Поиск

Поиск выполняется по:

- user_id
- file_name
- action
- destination_type

Поддерживаются:

- exact match
- substring match
- fuzzy match
- nearby context

## Производительность

Поддерживается работа с наборами:

- 100k событий
- 1M событий

Есть benchmark тесты.