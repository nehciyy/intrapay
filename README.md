# IntraPay

IntraPay is a lightweight internal transfer system built in Go. It allows users to register internal accounts, query balances, and process secure transactions between accounts. Designed with simplicity, concurrency safety, and modularity in mind, IntraPay is ideal as a foundational financial microservice in fintech or internal wallet systems.

---

## Features

- Create account with initial balance
- Get account balance
- Create transaction between two accounts with balance check and rollback
- Safe transactions using `FOR UPDATE` and retry logic
- Clean architecture: separated API, service, and repository layers
- Full unit test coverage for service and API logic

---

## API Endpoints

### 1. Create Account

**POST** `/accounts`

**Request Body**:

```json
{
  "account_id": 123,
  "initial_balance": 100.0
}
```

---

### 2. Get Account Balance

**GET** `/accounts/{id}`

**Path Parameter**:

- `id`: integer, account ID

**Response**:

```json
{
  "account_id": 123,
  "balance": 100.0
}
```

---

### 3. Create Transaction

**POST** `/transactions`

**Request Body**:

```json
{
  "source_account_id": 1,
  "destination_account_id": 2,
  "amount": 50.0
}
```

---

## Setup & Installation

### 1. Prerequisites

- Docker + Docker Compose
- Go 1.21+ (only needed for development outside containers)

---

### 2. Start the Application

From the root directory, run:

```bash
docker-compose up --build
```

This will:

- Start the PostgreSQL service
- Apply the schema from `migrations/001_init.sql`
- Start the Go application server

Once complete, the app will be running at:
📍 `http://localhost:8080`

**Note: Make sure Docker and PostgreSQL is running on your system.**

### 3. Test API with curl

#### Create Account

```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1, "initial_balance": 100}'
```

#### Get Account

```bash
curl -X GET http://localhost:8080/accounts/1
```

#### Create Transaction

```bash
 curl -X POST http://localhost:8080/transactions \
 -H "Content-Type: application/json" \
 -d '{"source_account_id": 2, "destination_account_id": 1, "amount": 50}'
```

---

That's a useful command for developers! Here's a clean and general way to include it in your documentation:

---

### Accessing the PostgreSQL Database

To inspect your database from the terminal, you can run:

```bash
docker exec -it intrapay_db psql -U <your_user> -d <your_database>
```

For example, if you're using the default credentials in `.env`:

```bash
docker exec -it intrapay_db psql -U postgres -d intrapay
```

**Tip:** Replace `<your_user>` and `<your_database>` with the values from your `.env` file:

- `POSTGRES_USER`
- `POSTGRES_DB`

---

## Run Tests

To run all unit tests (API + service logic):

```bash
go test ./internal/... -v
```

---

## Project Structure

```
.
├── cmd/server             # Application entry point
├── internal
│   ├── api                # HTTP handlers
│   ├── db                 # DB connection setup
│   ├── models             # Request structs
│   ├── service            # Business logic (Service layer)
│   ├── repository         # Data access abstraction
├── migrations             # SQL schema
├── Dockerfile             # Docker image for app
├── docker-compose.yml     # PostgreSQL + app services
├── go.mod / go.sum        # Dependencies
└── README.md
```

---

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
