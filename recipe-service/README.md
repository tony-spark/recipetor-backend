# recipe-service

Сервис отвечает за работу с данными о рецептах

## Topics

Читает события из

- `recipes.new` данные о новых рецептах
- `recipes.req` запросы получение информации о рецептах
- `nutritionfacts` расчёты КБЖУ для рецептов


Записывает события в

- `recipes` рецепты
