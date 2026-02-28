# log-service

A high-performance logging microservice built with Go that provides real-time log streaming via Server-Sent Events (SSE). Designed following Domain-Driven Design (DDD) principles with clean architecture patterns.

## Technical Overview

**log-service** is a centralized logging service that enables applications to send structured log entries and receive real-time updates through SSE connections. The service implements DDD patterns with a clean separation of concerns across domain, application, and infrastructure layers.

### Core Technologies

- **Go 1.25.3** - Primary runtime and development language
- **MongoDB** - Document database for log persistence with flexible schema support
- **Server-Sent Events (SSE)** - Real-time log streaming using the r3labs/sse library
- **Chi Router** - HTTP routing with middleware support for CORS and logging
- **Docker** - Containerization with multi-stage builds for production deployment
- **Swagger/OpenAPI** - Comprehensive API documentation with interactive testing interface

### Architecture Patterns

- **Domain-Driven Design (DDD)** - Business logic encapsulated in domain entities and value objects
- **Clean Architecture** - Dependency inversion with clear layer boundaries
- **Repository Pattern** - Data access abstraction with MongoDB implementation
- **Value Objects** - Type-safe log levels with validation and priority handling
- **CQRS-ready** - Separated command and query responsibilities

## API Endpoints

### Log Management
- `POST /api/v1/logs` - Create new log entries with structured metadata
- `GET /api/v1/events/{applicationID}` - SSE endpoint for real-time log streaming

### Documentation
- `GET /swagger/index.html` - Interactive Swagger UI documentation
- `GET /docs/swagger.json` - OpenAPI specification in JSON format

## Quick Start

### Prerequisites
- Go 1.25.3 or higher
- Docker and Docker Compose
- MongoDB (via Docker or local installation)

### Local Development Setup

1. **Clone and prepare environment**
```bash
git clone https://github.com/rubensantoniorosa2704/log-service.git
cd 
cp .env.example .env
```

2. **Configure environment variables**
```bash
# .env file
MONGO_URI=mongodb://root:rootpassword@localhost:27017/loggingdb?authSource=admin
MONGO_DB_NAME=loggingdb
PORT=8080
APP_PORT=8080
```

3. **Start MongoDB service**
```bash
docker-compose up -d mongodb
```

4. **Install dependencies and run application**
```bash
go mod download
go build ./cmd/api
./api
```

### Docker Deployment

**Complete stack deployment:**
```bash
docker-compose up -d
```

**Application-only deployment:**
```bash
docker build -t  .
docker run -p 8080:8080 --env-file .env 
```

## Usage Examples

### Creating Log Entries

**Basic log entry:**
```bash
curl -X POST http://localhost:8080/api/v1/logs \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "message": "User authentication successful",
    "level": "INFO"
  }'
```

**Enhanced log with metadata:**
```bash
curl -X POST http://localhost:8080/api/v1/logs \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "message": "Database connection timeout",
    "level": "ERROR",
    "source": "DatabaseService",
    "tags": {
      "module": "database",
      "error_type": "timeout",
      "severity": "high"
    },
    "metadata": {
      "connection_timeout": "30s",
      "retry_count": 3,
      "database_host": "prod-db-01.company.com"
    }
  }'
```

### Real-time Log Monitoring

**Connect to SSE stream:**
```bash
curl -N -H "Accept: text/event-stream" \
  "http://localhost:8080/api/v1/events/550e8400-e29b-41d4-a716-446655440000"
```

**Monitor with application filtering:**
```bash
curl -N -H "Accept: text/event-stream" \
  "http://localhost:8080/api/v1/events/{your-application-id}?stream={your-application-id}"
```

## Log Level Specifications

The service supports six hierarchical log levels with automatic validation:

| Level | Priority | Use Case |
|-------|----------|----------|
| TRACE | 1 | Detailed execution flow for debugging |
| DEBUG | 2 | Development and diagnostic information |
| INFO  | 3 | General application flow and events |
| WARN  | 4 | Potential issues that don't affect functionality |
| ERROR | 5 | Application errors that impact functionality |
| FATAL | 6 | Critical errors that may cause application termination |

## Data Schema

### Log Entry Structure
```json
{
  "id": "uuid",
  "application_id": "uuid",
  "user_id": "uuid", 
  "message": "string",
  "level": "INFO|WARN|ERROR|DEBUG|TRACE|FATAL",
  "source": "string (optional)",
  "tags": {
    "key": "value"
  },
  "metadata": {
    "flexible": "schema"
  },
  "timestamp": "2025-10-28T10:30:00Z"
}
```

## Development

### Project Structure
```
internal/
├── application/          # Application services and DTOs
│   └── log/             # Log-specific use cases and data transfer objects
├── domain/              # Business logic and domain entities
│   ├── log/            # Log domain entities and interfaces
│   └── valueobjects/   # Domain value objects (LogLevel, etc.)
└── infrastructure/      # External integrations and frameworks
    ├── db/             # Database connections and configurations
    ├── http/           # HTTP routing, controllers, and middleware
    └── repository/     # Data persistence implementations
```

### Building from Source
```bash
# Generate Swagger documentation
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g ./cmd/api/main.go -o ./docs

# Build application
go build -o app ./cmd/api

# Run tests
go test ./...
```

## API Documentation

Interactive API documentation is available at `http://localhost:8080/swagger/index.html` when the service is running. The documentation includes request/response schemas, example payloads, and a testing interface.

## Contributing

This project follows clean architecture principles and DDD patterns. When contributing:

1. Maintain clear separation between domain, application, and infrastructure layers
2. Implement comprehensive tests for business logic
3. Update Swagger documentation for API changes
4. Follow Go conventions and gofmt standards

## License

MIT License - see LICENSE file for details.
