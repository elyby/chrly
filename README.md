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

Нужно скопировать правильный docker-compose файл для желаемого окружения:

```sh
cp docker-compose.dev.yml docker-compose.yml  # dev env
cp docker-compose.prod.yml docker-compose.yml # prod env
```

И за тем всё это поднять:

```sh
docker-compose up -d
```

Если нужно пересобрать весь контейнер, то выполняем следующее:

```
docker-compose stop app  # Останавливаем конейтнер, если он ещё работает
docker-compose rm -f app # Удаляем конейтнер
docker-compose build app # Запускаем билд по новой
docker-compose up -d app # Поднимаем свежесобранный контейнер обратно
```
