# Общие бизнес-требования

Система представляет собой HTTP API со следующими требованиями к бизнес-логике:
- регистрация, аутентификация и авторизация пользователей;
- создание, редактирование и публикация рецептов (для авторизованных пользователей);
- просмотр опубликованных рецептов;
- рассчёт КБЖУ для рецептов;
- подбор рецепта по списку ингредиентов.

# Общие технические требования
- для БД должен быть реализован механизм версионности (миграций);
- клиент может поддерживать HTTP-запросы/ответы со сжатием данных;
- система должна проверять запросы клиента, выдавать соответствующие ошибки на некорректные запросы;
- для HTTP API должны быть реализованы тесты;
- при разработке системы должен быть реализован автоматический CI пайплайн, включающий в себя: статический анализ кода, прогон юнит-тестов, прогон  тестов HTTP API
функционал “рассчёт КБЖУ для рецептов” должен быть выполнен в виде отдельного сервиса

# Конфигурирование сервиса
Приложение должно конфигурироваться с помощью переменных окружения и/или флагов.

# Бонусные требования
При необходимости дополнить объём проекта (продолжить проект), можно реализовать следующие требования:
- добавление рецептов в избранное, получение списка избранных рецептов (для авторизованных пользователей);
- роли пользователей (пользователь, модератор, администратор)
- ведение базы данных пользователей (для администраторов)
- ведение базы данных рецептов (для модераторов)
- полнотекстовый поиск по рецептам;
- оценки для рецептов
- отзывы на рецепты
- написать парсер какого-либо сайта с рецептами для пополнения базы сервиса (только чтобы при этом не нарушить правила пользования сайтом)
- категоризатор рецептов
- создание планов питания (меню)
- сбор и анализ логов работы системы
