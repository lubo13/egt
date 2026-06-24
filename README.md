# Event pipeline task

## Requirements

### General requirements

```
Build two Go services communicating via Apache Kafka. Provide a docker-compose.yml for local deployment.
Provide example request calls in one of the following formats:
    - Bruno collection (preferred)
    - Postman collection
    - Console curl commands
```

### Tech stack

| Component         | Version / Library      | ✅   |
| ----------------- | ---------------------- | --- |
| Go                | 1.23+                  | ✅   |
| gRPC              | google.golang.org/grpc | ✅   |
| Kafka             | Apache Kafka 3.7       | ✅   |
| PostgreSQL        | 16                     | ✅   |
| Kafka client      | segmentio/kafka-go     | ✅   |
| PostgreSQL driver | jackc/pgx/v5           | ✅   |


### Flow

1. Client sends HTTP POST with JSON payload to Service A.
2. Service A produces the message to a Kafka topic asynchronously.
3. Service B consumes the message from Kafka and inserts it into PostgreSQL.
4. Service B serves stored records via gRPC query endpoints.

#### Service A — Ingestion API

- HTTP endpoint POST /events
- Accepts JSON payload with 5 fields (names and types are up to the candidate)
- Validates payload (all fields required)
- Publishes message to Kafka topic events asynchronously
- Returns 202 Accepted on success, 400 Bad Request on validation failure
- gRPC server exposing a health check endpoint

#### Service B — Consumer & Storage

- Kafka consumer on topic events
- gRPC server with the following endpoints:
- GetEvent — returns a single record by ID
- ListEvents — returns all records (no pagination)
- Synchronously writes each consumed message to PostgreSQL table events
- Table schema defined by the candidate (must match the 5 JSON fields)
- Handles consumer errors gracefully (log and continue)
- Structured logging with request correlation IDs propagated from Service A

### Bonus points

- docker-compose.yml spinning up: both services, Kafka (KRaft mode), PostgreSQL 16
- Migration tool (golang-migrate) for schema setup
- Unit tests for core logic
- Graceful shutdown handling for both services
- Retry with exponential backoff for transient Kafka producer failures
- Dead letter queue (DLQ) topic for messages that fail processing after retries
- Configuration management with env vars and startup validation

---
## Implementation details

- Postman collections and grpcurl collections provided into postman folder -> ./postman
- The task is implemented based on the requirements:
  | Component         | Requested Version / Library | Used Version / Library | ✅   |
  | ----------------- | --------------------------- | ---------------------- | --- |
  | Go                | 1.23+                       | 1.25                   | ✅   |
  | gRPC              | google.golang.org/grpc      | google.golang.org/grpc | ✅   |
  | Kafka             | Apache Kafka 3.7            | Apache Kafka 3.8       | ✅   |
  | PostgreSQL        | 16                          | 18                     | ✅   |
  | Kafka client      | segmentio/kafka-go          | segmentio/kafka-go     | ✅   |
  | PostgreSQL driver | jackc/pgx/v5                | jackc/pgx/v5           | ✅   |

- Bonus points (checked marks are implemented)
- [x] **docker-compose.yml spinning up:** both services, Kafka (KRaft mode), PostgreSQL 16
- [x] **Migration tool:** (golang-migrate) for schema setup
- [ ] **Unit tests** for core logic
- [x] **Graceful shutdown** handling for both services
- [x] **Retry with exponential backoff** for transient Kafka producer failures
- [ ] **Dead letter queue (DLQ)** topic for messages that fail processing after retries
- [x] **Configuration management** with env vars and startup validation

### Application Architecture (onion architecture, separated and decoupled layers)
1. Gateway service (Service A)
    - Expose REST API
    - Validate request
        * if validaion failed return 400 Status code
        * if do not failed send event for background processing and return 201 HTTP Status code
    - Go routine consumes new incomming events from a channel, process and produce message to database table (outbox pattern)
        * for kafka messages proto definition is used for a strict structure
    - Supports **Graceful shutdown**, **Configuration management**, **docker-compose.yml spinning up:**, **Migration tool:**

2. Producer service
    - fetch messages from outbox table, dispatch messages to kafka and delete processed messages from outbox table
    - this service is a global purpose one - with configuration can be use as a sidecar container.
    - Supports **Graceful shutdown**, **Configuration management**, **docker-compose.yml spinning up:**

3. Report service (Service B)
    - consume from kafka, idempotently process and store to messages database
    - expose gRPC service
        * DeviceEventService/GetEvent
        * DeviceEventService/ListEvent
    - Supports **Graceful shutdown**, **Configuration management**, **docker-compose.yml spinning up:**, **Migration tool:**, **Retry with exponential backoff of consumed message for transient error**

4. Graceful shutdown verification
    - once the container is up
    - execute `docker container ps`
    - find container id and replace $CONTAINER_ID$ -> `docker stop -t 1000 -s SIGTERM $CONTAINER_ID$`
    - open container logs and check or check exit code

---
## NOTE: The app is not production-ready - build happens on containers init

## RUN Application
```bash
    ./bootstrap.sh
```

## RUN Application
```bash
    ./bootstrap.sh
```

## Stop and remove applications volumes and generated ProtoBuf definitions
```bash
    ./flush.sh
```
