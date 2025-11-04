# MariaDB Encryption Migration Monitor

A real-time monitoring application built in Go for tracking MariaDB/RDS database encryption migration. This tool monitors data consistency, integrity, replica lag, and performs checksum validation during the encryption migration process.

## Features

- **Real-time Replica Lag Monitoring**: Track replication lag between source and target databases
- **Checksum Validation**: Verify data integrity by comparing table checksums
- **Data Consistency Checks**: Monitor row count consistency across databases
- **Web-based Dashboard**: Access monitoring data through a responsive web interface
- **Automated Alerts**: Get notified when issues are detected
- **WebSocket Updates**: Real-time updates without page refresh
- **Graceful Error Handling**: Continues monitoring even with temporary connection issues

## Prerequisites

- Go 1.21 or higher
- Access to source (unencrypted) and target (encrypted) MariaDB/RDS instances
- Database user with appropriate permissions:
  - `SELECT` on tables to monitor
  - `REPLICATION CLIENT` privilege for replica lag monitoring

## Installation

### From Source

```bash
# Clone the repository
git clone <repository-url>
cd mariadb-encryption-monitor

# Build the application
go build -o monitor ./cmd/monitor

# Run the application
./monitor -config config.yaml
```

### Using Docker

```bash
# Build the Docker image
docker build -t mariadb-monitor .

# Run the container
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/root/config.yaml \
  --name mariadb-monitor \
  mariadb-monitor
```

## Configuration

Create a `config.yaml` file with your database connection details:

```yaml
source_db:
  host: "source-db.example.com"
  port: 3306
  username: "monitor_user"
  password: "your_password"
  database: "your_database"

target_db:
  host: "target-db.example.com"
  port: 3306
  username: "monitor_user"
  password: "your_password"
  database: "your_database"

monitoring_interval: "30s"
replica_lag_threshold: "60s"
web_server_port: 8080

tables_to_monitor:
  - "users"
  - "orders"
  - "products"

log_level: "info"
```

### Environment Variables

You can override sensitive configuration values using environment variables:

- `SOURCE_DB_HOST`
- `SOURCE_DB_USERNAME`
- `SOURCE_DB_PASSWORD`
- `TARGET_DB_HOST`
- `TARGET_DB_USERNAME`
- `TARGET_DB_PASSWORD`

Example:

```bash
export SOURCE_DB_PASSWORD="secret123"
export TARGET_DB_PASSWORD="secret456"
./monitor -config config.yaml
```

## Usage

1. **Start the monitor**:
   ```bash
   ./monitor -config config.yaml
   ```

2. **Access the web interface**:
   Open your browser and navigate to `http://localhost:8080`

3. **Monitor the migration**:
   - View real-time replica lag
   - Check checksum validation results
   - Monitor data consistency
   - Review active alerts

## API Endpoints

The application provides REST API endpoints for integration:

- `GET /`: Web interface
- `GET /ws`: WebSocket endpoint for real-time updates
- `GET /api/metrics`: Current metrics (JSON)
- `GET /api/alerts`: Alert history (JSON)
- `GET /api/health`: Health check endpoint

### Example API Usage

```bash
# Get current metrics
curl http://localhost:8080/api/metrics

# Get alerts
curl http://localhost:8080/api/alerts

# Health check
curl http://localhost:8080/api/health
```

## Monitoring Metrics

### Replica Lag
- Measures replication delay in seconds
- Alerts when lag exceeds configured threshold
- Status indicators: `ok`, `replication_stopped`, `error`, `no_replication`

### Checksum Validation
- Compares table checksums between source and target
- Detects data corruption or replication issues
- Per-table granularity

### Data Consistency
- Compares row counts between databases
- Identifies missing or extra rows
- Helps verify complete data replication

## Alert Severity Levels

- **CRITICAL**: Checksum mismatch, major consistency issues, replication stopped
- **WARNING**: Replica lag exceeds threshold, connection issues
- **INFO**: Connection restored, monitoring events

## Troubleshooting

### Connection Issues

If the monitor cannot connect to databases:

1. Verify database credentials in `config.yaml`
2. Check network connectivity to database hosts
3. Ensure database user has required permissions
4. Check firewall rules

### No Replica Lag Data

If replica lag shows "no_replication":

1. Verify replication is configured on the target database
2. Check that the target is actually a replica
3. Ensure the monitor user has `REPLICATION CLIENT` privilege

### Checksum Errors

If checksum validation fails:

1. Verify tables exist in both databases
2. Check that table names are correct (case-sensitive)
3. Ensure monitor user has `SELECT` permission on tables

## AWS RDS Encryption Migration

This tool is designed to work with the AWS RDS encryption migration process described in:
https://repost.aws/knowledge-center/rds-encrypt-instance-mysql-mariadb

The typical workflow:

1. Create encrypted snapshot of source RDS instance
2. Restore snapshot to new encrypted RDS instance
3. Set up replication from source to target
4. Use this monitor to track migration progress
5. Verify data consistency before cutover

## Performance Considerations

- Monitoring interval: Shorter intervals provide more frequent updates but increase database load
- Table selection: Monitor only critical tables to reduce overhead
- Connection pooling: The application uses connection pooling for efficiency
- Memory usage: Keeps 24 hours of replica lag history in memory

## Security Best Practices

1. Use environment variables for sensitive credentials
2. Create dedicated database users with minimal required permissions
3. Use TLS/SSL connections to databases (configure in DSN)
4. Restrict web interface access using firewall rules
5. Consider adding authentication to the web interface for production use

## License

[Your License Here]

## Support

For issues, questions, or contributions, please [open an issue](link-to-issues).
