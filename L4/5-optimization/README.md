## Requirements

- Docker
- Go 1.26+

# Оптимизация использует SIMD API доступная лишь с версии 1.26 (также необходимо добавить флаг в env)!

## Run
1. Добавить в env возможность использовать simd пакет
set GOEXPERIMENT=simd
2. Запустить benchmark тесты
Из папки L5/5-optimization запустить команду
go test .\cache\. -run=^$ -bench=. -benchmem

## Об изменениях

Вместо LRU cache с hashmap и linked list, использование псевдо LRU алгоритма и использование SIMD для быстрого нахождения значения в кэше. 
Кэш теперь является массиов, благодаря чему не происходят дополнительные аллокации в сравнении с linked list.

Также алгоритм показывает хороший результат лишь когда происходит нечастая синхронизация между потоками.
Для этого создан массив из шардов, чтобы уменьшить вероятность одновременного попадания в один шард.
В идеале для запросов должны использоваться лишь определенные потоки у которых был бы лишь свой кэш, тогда синхронизацию в целом можно было бы убрать.


## Результат
Отсутствие дополнительных аллокаций в кэше.
Время операции кэша сокращено на ~39% (BenchmarkCache_Parallel-24 58.49ns / BenchmarkTreePseudoLRUCache_Parallel-24 35.67ns).

goos: windows
goarch: amd64
pkg: wildberries-go-course/L0/cache
cpu: AMD Ryzen 9 5900X 12-Core Processor
BenchmarkCache_Put-24                           22662589                53.45 ns/op           23 B/op          0 allocs/op
BenchmarkCache_Get-24                           49573461                23.65 ns/op            0 B/op          0 allocs/op
BenchmarkCache_Parallel-24                      20475508                58.49 ns/op            4 B/op          0 allocs/op
BenchmarkTreePseudoLRUCache_Put-24              42303570                28.33 ns/op            0 B/op          0 allocs/op
BenchmarkTreePseudoLRUCache_Get-24              92172976                13.50 ns/op            0 B/op          0 allocs/op
BenchmarkTreePseudoLRUCache_Parallel-24         33724347                35.67 ns/op            0 B/op          0 allocs/op

## Task

Взять простой HTTP API (например, сервис, который складывает числа или возвращает JSON, или сервис из L0), создать нагрузку и оптимизировать его по CPU и памяти.

pprof и net/http/pprof

benchstat, go test -bench

анализ trace

Результат: проект с кодом API, профилировкой с историей коммитов, README с описанием изменений.