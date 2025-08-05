# Messaging Service

A production-grade unified messaging service that supports SMS, MMS, and Email messaging with conversation management, built with Go 1.24 and clean architecture principles.

## 🚀 Features

- **Unified Messaging API**: Send SMS, MMS, and Email messages through a single API
- **Conversation Management**: Automatic grouping of messages into conversations
- **Data Persistence**: PostgreSQL database with proper indexing and constraints
- **Webhook Support**: Handle incoming messages from external providers
- **Error Handling**: Retry logic with exponential backoff for provider errors (500, 429)
- **Production-Ready**: Dockerized with multi-stage builds, health checks, and security
- **API Documentation**: Interactive Swagger/OpenAPI documentation served by the main application
- **Clean Architecture**: Separation of concerns with dependency injection
- **Comprehensive Testing**: Unit, integration, and API tests
- **HTTP Error Handling**: Robust retry logic for provider errors (500, 429, etc.)



## 🛠️ Quick Start

### Prerequisites
- Go 1.24+
- Docker and Docker Compose
- Make (optional, for convenience)

### Development Setup

1. **Clone and setup:**
   ```bash
   git clone <repository-url>
   cd messaging-service
   make setup
   ```

2. **Generate Swagger documentation:**
   ```bash
   make swagger
   ```

3. **Start the application:**
   ```bash
   make run
   ```

4. **Access the API:**
   - API: http://localhost:8080/api
   - Swagger Docs: http://localhost:8080/swagger/index.html
   - Health Check: http://localhost:8080/health

### Docker Setup

1. **Start the full stack (database + application):**
   ```bash
   make docker-up
   ```

2. **Access the application:**
   - API: http://localhost:8080/api
   - Swagger Docs: http://localhost:8080/swagger/index.html
   - Health Check: http://localhost:8080/health

3. **Stop the stack:**
   ```bash
   make docker-down
   ```

## 📚 API Documentation

The API documentation is automatically generated and served by the main application:

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **OpenAPI JSON**: http://localhost:8080/swagger/doc.json
- **OpenAPI YAML**: http://localhost:8080/swagger/doc.yaml

### Available Endpoints

| Method | Endpoint | Description                                         |
|--------|----------|-----------------------------------------------------|
| `POST` | `/api/messages/message` | Send SMS/MMS message                                |
| `POST` | `/api/messages/email` | Send email message                                  |
| `POST` | `/api/webhooks/message` | Handle incoming SMS/MMS                             |
| `POST` | `/api/webhooks/email` | Handle incoming email                               |
| `GET` | `/api/conversations` | List conversations by query - query params required |
| `GET` | `/api/conversations/:id/messages` | Get messages in conversation                        |
| `GET` | `/health` | Health check endpoint                               |

## 🗄️ Database Schema

### Conversations Table
```sql
CREATE TABLE conversations (
    id SERIAL PRIMARY KEY,
    participant1 VARCHAR(255) NOT NULL,
    participant2 VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Messages Table
```sql
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    conversation_id INTEGER REFERENCES conversations(id),
    from_address VARCHAR(255) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    message_type VARCHAR(10) NOT NULL,
    body TEXT NOT NULL,
    attachments JSONB,
    provider_message_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 🧪 Testing

### Run All Tests
```bash
make test
```

### Run Specific Test Types
```bash
# Unit tests only
go test ./internal/... -v

# Integration tests only
go test ./tests/... -v

# API tests
./bin/test.sh
```

## 🐳 Docker Commands

| Command | Description |
|---------|-------------|
| `make docker-build` | Build Docker image |
| `make docker-up` | Start the full stack (database + app) |
| `make docker-down` | Stop the full stack |
| `make docker-logs` | View logs |

## ⚙️ Configuration

The application uses environment variables for configuration. See `CONFIGURATION.md` for all available options.

### Key Environment Variables
```bash
# Server Configuration
SERVER_PORT=8080
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
SERVER_IDLE_TIMEOUT=60s

# Database Configuration
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=messaging_service
DATABASE_USER=messaging_user
DATABASE_PASSWORD=messaging_password
DATABASE_MAX_OPEN_CONNS=25
DATABASE_MAX_IDLE_CONNS=5
DATABASE_CONN_MAX_LIFETIME=5m
```

## 🏗️ Architecture

### System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Client Applications                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │   Web App   │  │  Mobile App │  │   External  │          │
│  │             │  │             │  │   Services   │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────┬───────────────────────────────────────────┘
                      │ HTTP/HTTPS
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Messaging Service API                       │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    Middleware Layer                     │   │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐     │   │
│  │  │ Request ID  │ │  Logging    │ │   Metrics   │     │   │
│  │  │             │ │             │ │             │     │   │
│  │  └─────────────┘ └─────────────┘ └─────────────┘     │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              │                               │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    HTTP Handlers                       │   │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐     │   │
│  │  │   SMS/MMS   │ │    Email    │ │ Webhooks    │     │   │
│  │  │   Handler   │ │   Handler   │ │   Handler   │     │   │
│  │  └─────────────┘ └─────────────┘ └─────────────┘     │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Business Logic Layer                        │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                  Service Layer                          │   │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐     │   │
│  │  │ Messaging   │ │Conversation │ │ Validation  │     │   │
│  │  │  Service    │ │  Service    │ │   Logic     │     │   │
│  │  └─────────────┘ └─────────────┘ └─────────────┘     │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Data Access Layer                           │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                 Repository Layer                        │   │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐     │   │
│  │  │   Message   │ │Conversation │ │   Database  │     │   │
│  │  │ Repository  │ │ Repository  │ │  Connection │     │   │
│  │  └─────────────┘ └─────────────┘ └─────────────┘     │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    External Services                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │   SMS/MMS   │  │    Email    │  │  Database   │          │
│  │  Provider   │  │  Provider   │  │ (PostgreSQL)│          │
│  │             │  │             │  │             │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

### Request Flow

```
1. Client Request
   ↓
2. Middleware Processing
   ├── Request ID Generation
   ├── Structured Logging
   └── Metrics Collection
   ↓
3. Route Matching
   ↓
4. Handler Processing
   ├── Request Validation
   ├── Business Logic
   └── Response Formatting
   ↓
5. Service Layer
   ├── Business Rules
   ├── Data Validation
   └── External Calls
   ↓
6. Repository Layer
   ├── Database Operations
   └── Data Persistence
   ↓
7. External Providers
   ├── SMS/MMS Delivery
   ├── Email Delivery
   └── Webhook Processing
```

### Data Flow Examples

#### **Outbound Message Flow:**
```
Client → POST /api/messages/message
  ↓
Handler.SendSMS() → Validate Request
  ↓
Service.SendSMS() → Business Logic
  ↓
Provider.SendSMS() → External SMS Service
  ↓
Repository.Create() → Save to Database
  ↓
Response → Success/Error
```

#### **Inbound Webhook Flow:**
```
External Service → POST /api/webhooks/message
  ↓
Handler.HandleInboundSMS() → Parse Webhook
  ↓
Service.HandleInboundSMS() → Process Message
  ↓
Repository.Create() → Save to Database
  ↓
Response → Acknowledgment
```

## 📁 Project Structure

```
messaging-service/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── app/                     # Application lifecycle
│   ├── config/                  # Configuration management
│   ├── container/               # Dependency injection
│   ├── domain/                  # Domain models and interfaces
│   ├── handler/                 # HTTP handlers
│   ├── logger/                  # Structured logging
│   ├── middleware/              # HTTP middleware
│   ├── provider/                # External service providers
│   ├── repository/              # Data access layer
│   ├── router/                  # HTTP routing
│   ├── service/                 # Business logic
│   └── telemetry/               # OpenTelemetry setup
├── tests/                       # Integration tests
├── docs/                        # Generated Swagger docs
├── init.sql/                    # Database schema
├── bin/                         # Scripts
├── Dockerfile                   # Multi-stage Docker build
├── docker-compose.yml           # Development environment
├── Makefile                     # Build and deployment commands
└── README.md                    # This file
```

## 🔧 Development Commands

| Command | Description |
|---------|-------------|
| `make setup` | Initialize project dependencies |
| `make run` | Start the messaging service |
| `make test` | Run all tests |
| `make swagger` | Generate Swagger documentation |
| `make docs` | Generate Swagger documentation |
| `make help` | Show all available commands |

## 🚀 Production Deployment

### Docker Compose (Recommended)
```bash
# Start production stack
make docker-up

# View logs
make docker-logs

# Stop stack
make docker-down
```

### Manual Deployment
```bash
# Build the application
go build -o messaging-service ./cmd/server

# Set environment variables
export DATABASE_HOST=your-db-host
export DATABASE_PORT=5432
# ... other environment variables

# Run the application
./messaging-service
```

## 📊 Monitoring

### Health Check
```bash
curl http://localhost:8080/health
```

### Database Connection
The application includes database connection monitoring and will log connection status on startup.

## 🔒 Security

- **Non-root user** in Docker containers
- **Input validation** on all endpoints
- **SQL injection protection** through parameterized queries
- **Boilerplate for Authorization** through api-key Authorization header

## 📊 Observability

- **Request ID tracking** with `X-Request-ID` header
- **Structured logging** with request details and timing
- **OpenTelemetry metrics** with Prometheus exporter
- **Request/response metrics** including duration and size

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## 📄 License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details.
