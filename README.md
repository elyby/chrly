# Это заготовка для нормального файла

Для настройки Dev-окружения нужно склонировать проект в удобное место,
за тем сделать символьную ссылку в свой GOPATH:

```sh
# Выполнять, находясь внутри директории репозитория
mkdir -p $GOPATH/src/elyby
ln -s $PWD $GOPATH/src/elyby/minecraft-skinsystem
```

Или можно склонировать репозиторий сразу в нужную локацию:

```sh
git clone git@bitbucket.org:elyby/minecraft-skinsystem.git $GOPATH/src/elyby/minecraft-skinsystem
```

Поднять репозиторий можно командой:

```sh
docker-compose up -d
```

Рекомендуемый файл `docker-compose.override.yml` для dev-окружения:

```sh
version: '2'
services:
    app:
        volumes:
            - ./:/go/src/app
        command: ["go", "run", "minecraft-skinsystem.go"]
```

В таком случае, для перезапуска контейнера (при условии, что не появляется
новых зависимостей) будет достаточно выполнить только одну команду:

```sh
docker-compose restart app
```

Если нужно пересобрать весь контейнер, то выполняем следующее:

```
docker-compose stop app  # Останавливаем конейтнер, если он ещё работает
docker-compose rm -f app # Удаляем конейтнер
docker-compose build app # Запускаем билд по новой
docker-compose up -d app # Поднимаем свежесобранный контейнер обратно
```
