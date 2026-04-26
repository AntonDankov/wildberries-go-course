# GC Profiler

Go server with endpoits to profile memory

## Run

```bash
go run .
```

Use Insomnia and provided "requests-L4-4-Insomnia.yaml" to run requests

## Endpoints

### GET `/memory_metrics`

Returns current Go runtime memory metrics in Prometheus text exposition format.

**Response format:** `text/plain; version=0.0.4; charset=utf-8`

**Response metrics:**

| Metric | Type | Description |
|---|---|---|
| `go_allocations_amount` | Counter | Total number of heap allocations |
| `go_garbage_collection_amount` | Counter | Total number of GC cycles completed |
| `go_used_memory` | Gauge | Bytes currently allocated on the heap |
| `go_last_time_garbage_collected` | Gauge | Seconds elapsed since the last GC run |
| `go_total_pause_time_gc` | Gauge | Total GC stop-the-world pause time in seconds |
| `go_gc_percentage` | Gauge | Current GC target percentage (GOGC) |

**Example request:**
```bash
curl http://localhost:8080/memory_metrics
```

**Example response:**
```
# HELP go_allocations_amount Amount of allocations
# TYPE go_allocations_amount counter
go_allocations_amount 104321

# HELP go_used_memory Amount of memory allocated by gc in bytes
# TYPE go_used_memory gauge
go_used_memory 2097152
...
```

---

### POST `/gc_percentage`

Dynamically changes the garbage collector target percentage (equivalent to setting `GOGC`).

**Request body (JSON):**

{"percentage": 25}

percentage - int - gc target percentage. Value -1 to disable GC. Value 0 not allowed.

**Example request:**
```bash
curl -X POST http://localhost:8080/gc_percentage \
  -H "Content-Type: application/json" \
  -d '{"percentage": 25}'
```

**Responses:**

`200 OK` - Percentage updated successfully (empty body)
`400 Bad Request` - Invalid/missing body, or value is `0` or less than `-1`

**Error response body:**
```json
{
  "error": "percent must be -1 (disable GC) or a positive integer"
}
```

## Task

Необходимо разработать программу на Go, которая показывает через HTTP-endpoint в формате Prometheus текущую информацию о памяти и сборщике мусора.

Используйте runtime.ReadMemStats, debug.SetGCPercent, профилирование (pprof).

Примеры метрик:

количество аллокаций

количество сборок мусора

используемая память

последнее время GC

другие — по вашему желанию

Результат: директория с кодом сервера, инструкцией по запуску (README), примерами запросов.