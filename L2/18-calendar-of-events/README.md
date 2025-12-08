# Event Calendar

## Run

```bash
go run . -port=8080
```

## Endpoints

### POST (JSON Body)

* `/create_event` - `{"user_id": 1, "date": "2026-01-01", "description": "name"}`
* `/update_event` - `{"id": 1, "user_id": 1, "date": "2026-01-01", "description": "name"}`
* `/delete_event` - `{"user_id": 1, "id": 1}`

### GET (Query Params: `user_id`, `date`)

* `/events_for_day`
* `/events_for_week`
* `/events_for_month`
