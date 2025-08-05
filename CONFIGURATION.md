# Configuration

The messaging service can be configured using environment variables. All configuration values have sensible defaults for development.

## Environment Variables

### Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Port number for the HTTP server |
| `SERVER_READ_TIMEOUT` | `30s` | Maximum duration for reading the entire request |
| `SERVER_WRITE_TIMEOUT` | `30s` | Maximum duration before timing out writes of the response |
| `SERVER_IDLE_TIMEOUT` | `60s` | Maximum amount of time to wait for the next request |

### Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | Database host address |
| `DB_PORT` | `5432` | Database port number |
| `DB_NAME` | `messaging_service` | Database name |
| `DB_USER` | `messaging_user` | Database username |
| `DB_PASSWORD` | `messaging_password` | Database password |
| `DB_SSL_MODE` | `disable` | SSL mode for database connection |

### Database Connection Pool Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_MAX_OPEN_CONNS` | `25` | Maximum number of open connections to the database |
| `DB_MAX_IDLE_CONNS` | `25` | Maximum number of idle connections in the pool |
| `DB_CONN_MAX_LIFETIME` | `5m` | Maximum amount of time a connection may be reused |

## Example Configuration

```bash
# Development
export PORT=8080
export DB_HOST=localhost
export DB_PASSWORD=my_secure_password

# Production
export PORT=443
export DB_HOST=my-db.example.com
export DB_PASSWORD=production_password
export DB_SSL_MODE=require
export SERVER_READ_TIMEOUT=60s
export SERVER_WRITE_TIMEOUT=60s
```

## Duration Format

Time-based configuration values (timeouts) use Go's duration format:

- `30s` - 30 seconds
- `5m` - 5 minutes
- `1h` - 1 hour
- `24h` - 24 hours

## Validation

The configuration is validated at startup. If any required values are missing or invalid, the application will fail to start with a descriptive error message. 