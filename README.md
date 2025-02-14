# Установка и запуск
0. Убедитесь, что у вас установлены:
  - Docker(и запущен движок)
  - git
  - make(для запуска Makefile). Если вы не хотите устанавливать make, можете запускать команды из Makefile вручную
1. Клонируем проект и переходим в него
```
sudo git clone https://github.com/nekkkkitch/AvitoTest
```
```
cd AvitoTest
```
2. Стартуем и ждем запуска
```
sudo make start
```

## Доступные руты
- http://localhost:8080/api/auth

Аутентификация.
При первом вводе логин/пароль пользователя добавляется в БД, а также возвращается токен. В будущем для получения токена потребуется ввести первоначальную версию пароля.
Токен потребуется при вводе остальных рутов, он должен быть в header "Authorization":"Bearer {token}"

- http://localhost:8080/api/buy/:item

Покупка мерча.
Вместо :item требуется ввести один из доступных вариантов. В случае, если он существует, будет производиться попытка покупки оного предмета, в случае успеха он добавится в инвентарь пользователя.
При недостаточном количестве средств вернется соответствующая ошибка.

- http://localhost:8080/api/info

Получение информации по пользователю.

- http://localhost:8080/api/sendCoin

Перевод денег другому пользователю. В случае, если существует пользователь с данным логином, будет производиться попытка передачи средств, в случае успеха данные о переводе добавятся в историю. При недостаточном количестве средств вернется соответствующая ошибка.

## Тестирование
В большинстве папок присутствуют test файлы, в основном это unit-тесты, но также присутствую интеграционные тесты в cmd/app. 
internal/db тестируется на моковой БД, запускаемой вместе с остальными сервисами в docker-compose.

Тесты запускаются следующей командой:
```
go test -v
```
# Проблемы и решение
Одной из основных проблем стало взаимодействие с БД, в основном перевод средств. Для того, чтобы не проверять возможность перевода средств, понадобилось привязать транзакции, а вместе с ними и сменить pgx.Conn на pgxpool.Pool, так как для стабильной работы транзакций понадобилось большое количество подключений.

