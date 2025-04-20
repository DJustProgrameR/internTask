# internshipPVZ
Welcome, ladies, lads and gentlemen

## Проделанная работа
1. Реализованы все основные и доп. условия
2. Приложение(как прод так и интеграционное тестирование) собирается при помощи DI контейнера uber-fx
3. Покрытие тестами(всей задействованной логики кроеме grpc и prometheus) - 75,2%. 
4. Покрытие чисто usecase-ов - 88,3%
5. Добавлена конфигурация линтера
6. Насткройка кодогенерации DTO эндопинтов:
```
make generate

или
Устанавливаем генерирующую утилиту
go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@v2.2.0

Создаём конфиг
dto_generator_cfg.yaml

Подтягиваем нужную либу
go get github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen

запускаем go генерацию в файле
generate.go
```

## Инструкции по запуску
### Установка либ
```
go mod download
```
### Запуск линтера
```
make linter
или
golangci-lint run ./cmd/config/...
golangci-lint run ./cmd/initdb/...
golangci-lint run ./internal/...
golangci-lint run ./test/integration/...
```

### Запуск проекта
```
make up

или * c тестами
make start

или 
docker-compose up --build -d
```
#### С логами
```
make up_log

или * c тестами
make start_log

или 
docker-compose up --build

Закрывать в другой консоли или ctrl+C для windows
```
### Остановка
```
make down или stop

или
docker-compose down
```
### Запуск юнит тестов
```
make unit

или
go test ./internal/usecase/...
```
### Запуск интеграционных тестов
```
make integration

или
docker-compose -f docker-compose.test.yml up -d
go test ./test/integration/...
docker-compose down

```

### Запуск тестов с покрытием
```
make coverage_t

или
Из корневой директории
docker-compose -f docker-compose.test.yml up
go test -coverprofile "coverage/cover.out" ./...
go run coverage/filter_coverage.go coverage/cover.out coverage/filtered_coverage.out
go tool cover -func "coverage/filtered_coverage.out"  
docker-compose -f docker-compose.test.yml down
```

### Проверка grpc?
```
grpcurl -plaintext -d '{}' localhost:3000 pvz.v1.PVZService/GetPVZList
```

## Вопросы и объяснение решений
1. Логирование настроил при помощи slog
2. Т.к. сказано удалять товары в порядке LIFO, 
но далее в пояснении говорится о FIFO. Исходя из логики 
работы ПВЗ(последним был добавлен неправильный товар),
реализован принцип LIFO
3. Во время решения появился вопрос "Как организовать сериализуемость
закрытия приёмки незадолго после добавления последнего товара или
нескольких последовательных закрытий одной и той же приёмки пользователем?"
Решено было исходить из того, что нам нужна производительность, 
следовательно, минимизация транзакций(особенно serializable).
Поэтому тот интерфейс, с которым взаимодействует пользователь должен 
отправлять запросы, ожидая ответа предыдущего(синхронно)
P.S. за client-side остаётся возможность кэшировать запросы,
отправляя их по-очереди, если у пользователя медленный интернет, etc.
4. При добавлении grpc в .proto был найден неиспользуемый enum:
решено было закоммитить "for future use" и не засорять генерируемый код.
5. Во время генерации DTO endpoint-ов возник вопрос, что использовать для генерации,
ради лаконичности и функциональности решение приянто в пользу oapi-codegen, несмотря на
недостатки
```
oapi-codegen:
т.к. в схеме некоторые поля на русском, сгенерируются
соответствующие названия переменных, переписать на англйский

так же данная утилита не сгенерировала ответ на запрос с фильтрациями,
о чём указано в её API, так что это DTO нужно дописать руками

прописывает комплексные типы в dto, что позволяет валидировать данные по схеме
в процессе парсинга тела запроса, что удобно
```
```
openapitools/openapi-generator-cli:
пишет много документации

использует примитивы в dto, что вынуждает писать дополнительную валидацию

генерирует много дополнительного кода в самих моделях:
типа Getter/Setter - ов,Маршаизаторов, обёрток

а так же просто неиспользуемый код типа HttpClient

причём создаёт отдельный модуль, который сложнее интегрировать в систему
```
6. Для корректной работы oapi нужно было добавить в swagger.yaml:
```
добавляем в openapi схему в schemas, иначе утилита не сгенерирует нужные файлы:
    GetFilteredResponse:
    GetFilteredResponseReceptions:
              
И теперь ссылаемся на схему ответа в get /pvz:
      responses:
        '200':
          description: Список ПВЗ
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetFilteredResponse'
```