# Product Manager with auth

Тестовое задание для inHouseAd

## Начало работы

Эти инструкции позволят вам запустить копию проекта на вашем локальном компьютере для разработки и тестирования.

### Предварительные требования

Что нужно установить на ПК для использования:

Docker, 
Docker Compose

### Установка

Шаги для запуска проекта:

1. Клонируйте репозиторий:
```bash
git clone github.com/KVSH-user/product_manager_with_auth
```

2. Перейдите в директорию проекта
3. Запустите проект с помощью Docker Compose:
```bash
docker-compose up -d
```

### Использование

Доступные ```REST```(запрос - ответ): 

1. Регистрация - ```POST /user/signup```
```
{
    "email" : "my@email.com",
    "password" : "myPass"
}
```

```
{
    "id": 1
}
```

2. Авторизация - ```POST /user/signin```
```
{
    "email" : "my@email.com",
    "password" : "myPass"
}
```

```
{
    "token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDk5MDgwMjQsImlzc3VlZCI6MTcwOTgyMTYyNCwidWlkIjoxfQ.xCpFMzg09xV6S_ZUTrCnqROQ-a-o2xN6WcJjxPr3K3bBZQ6kSsMvYIR0TOAazHISVCXd_Q8HOyv3v4OKCjRhOw"
}
```
3. Создание категории товаров - ```POST /category/create```
```
{
    "category_name" : "Name"
}
```
```
{
    "category_id" : 1,
    "category_name" : "Name"
}
```
4. Редактирование категории - ```PATCH /category/update```
```
{
    "category_id" : 1,
    "new_name" : "Name"
}
```
```
{
    "category_id" : 1,
    "new_name" : "Name"
}
```
5. Удаление категории - ```DELETE /category/delete/{id}```
```
{}
```
```
{
    "category_id" : 1,
    "deleted" : true
}
```
6. Доавление товара с опредленной категорией - ```POST /good/create/{categoryId}```
```
{
    "good_name" : "Name"
}
```
```
{
    "good_id": 5,
    "good_category_id": 1,
    "good_name": "Name",
    "category_name": "Test"
}
```
7. Редактирование доавленного ранее товара - ```PATCH /good/update```
```
// можно поменять название товара/добавить ему категорию(или все вместе)
{
    "good_id" : 1,
    "good_actual_name" : "renamed", //необязательно
    "added_category_id" : 3 //необязательно
}
```
```
{
    "good_id" : 1,
    "good_name" : "renamed",
    "category_name" : "Category Name" //отобразится несколько, если их несколько
}
```
8. Удаление товара - ```DELETE /good/delete/{id}```
```
{}
```
```
{
    "good_id" : 1,
    "deleted" : true
}
```
9. Посмотреть список существующих категорий - ```GET /category/list```
```
{}
```
```
{
    "category_id" : 1,
    "category_name" : "Name"
}
```
10. Посмотреть список товаров конкретной категории - ```GET /good/list/{categoryId}```
```
{}
```
```
{
    "good_id" : 1,
    "good_name" : "Name"
}
```
