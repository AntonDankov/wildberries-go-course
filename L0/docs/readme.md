# Wildberries L0 project - Order Processing System

## Architecture

The system follows a microservices architecture with the following data flow:

1. **Message Producer** - PowerShell script publishes order JSON to Kafka topic
2. **Message Consumer** - Go service consumes messages, validates data, and persists to database
3. **Caching Layer** - Custom LRU cache implementation 
4. **REST API** - HTTP endpoints for order queries with cache-first strategy
5. **Web Interface** - Simple frontend for testing and order lookup

## API Endpoints

### GET /order/{order_uid}

Retrieves order information by UUID.

**Response**: Order object with delivery, payment, and items details

**Status Codes**:

- `200` - Order found
- `400` - Invalid UUID format
- `404` - Order not found
- `405` - Method not allowed

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21 or higher
- PowerShell (for message producer)


### Installation

1. **Clone the repository**

```bash
git clone https://github.com/AntonDankov/wildberries-go-course.git
cd wildberries-order-system
```
```
2. **Start infrastructure services**

```bash
docker compose up -d
```

3. **Wait for services to be ready** (approximately 30 seconds)
4. **Run the application**

```bash
go run main.go
```

5. **Open the web interface**

Navigate to folder `frontend` and open `index.html` in your browser

6. **Send test order message**
```powershell
# On Windows
cd kafka-producer-script
.\send_kafka_message.ps1
```

7. **Query the order**

Use the UUID from the Kafka message in the web interface to retrieve order information.

## Environment Configuration

Create a `.env` file in the project root:

```env
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=admin
DATABASE_PASSWORD=admin
DATABASE_NAME=wildberries
DATABASE_SSL_MODE=disable
```


## Docker Services

The `docker-compose.yml` configures the following services:

- **Zookeeper** (port 2181) - Kafka coordination service
- **Kafka** (port 9092) - Message broker with topic `delivery-topic`
- **PostgreSQL** (port 5432) - Primary data store with persistent volume


## Database Schema

The system uses the following main tables:

- `orders` - Main order information
- `deliveries` - Delivery address and contact details
- `payments` - Payment transaction information
- `items` - Order line items and product details

Schema migration is automatically applied on application startup.

## Message Format

Orders are consumed as JSON messages from Kafka topic `delivery-topic`. Example message structure:

```json
{
  "order_uid": "e5c179eb-96c3-4a7d-a316-ff3dfd4613a1",
  "track_number": "WBILMTESTTRACK",
  "entry": "WBIL",
  "delivery": {
    "name": "Test Testov",
    "phone": "+9720000000",
    "zip": "2639809",
    "city": "Kiryat Mozkin",
    "address": "Ploshad Mira 15",
    "region": "Kraiot",
    "email": "test@gmail.com"
  },
  "payment": {
    "transaction": "e5c179eb-96c3-4a7d-a316-ff3dfd4613a1",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 1817,
    "payment_dt": 1637907727,
    "bank": "alpha",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
  },
  "items": [
    {
      "chrt_id": 9934930,
      "track_number": "WBILMTESTTRACK",
      "price": 453,
      "rid": "ab4219087a764ae0btest",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 2389212,
      "brand": "Vivienne Sabo",
      "status": 202
    },
    {
      "chrt_id": 9434222,
      "track_number": "WBILMTESTTRACK242",
      "price": 453,
      "rid": "ab4219087a764ae0btest323",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 2389212,
      "brand": "Vivienne Sabo",
      "status": 202
    }
  ],
  "locale": "en",
  "internal_signature": "",
  "customer_id": "test",
  "delivery_service": "meest",
  "shardkey": "9",
  "sm_id": 99,
  "date_created": "2021-11-26T06:22:19Z",
  "oof_shard": "1"
}
```
