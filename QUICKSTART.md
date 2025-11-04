# Quick Start Guide

## Prerequisites

Before running the MariaDB Encryption Migration Monitor, ensure you have:

1. **Go 1.21+** installed (for building from source)
2. **Source and Target MariaDB/RDS instances** with:
   - Network connectivity from the monitor host
   - Database user with appropriate permissions
3. **Database Permissions** required:
   - `SELECT` on all tables you want to monitor
   - `REPLICATION CLIENT` for replica lag monitoring

## Setup Database User

Create a dedicated monitoring user on both source and target databases:

```sql
-- Create monitoring user
CREATE USER 'monitor_user'@'%' IDENTIFIED BY 'secure_password';

-- Grant required permissions
GRANT SELECT ON your_database.* TO 'monitor_user'@'%';
GRANT REPLICATION CLIENT ON *.* TO 'monitor_user'@'%';

-- Apply changes
FLUSH PRIVILEGES;
```

## Installation

### Option 1: Build from Source

```bash
# Clone or download the project
cd mariadb-encryption-monitor

# Build the application
go build -o monitor ./cmd/monitor

# The binary is now ready to use
./monitor --help
```

### Option 2: Using Docker

```bash
# Build the Docker image
docker build -t mariadb-monitor .

# Run with your configuration
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/root/config.yaml \
  --name mariadb-monitor \
  mariadb-monitor
```

## Configuration

1. **Copy the example configuration:**
   ```bash
   cp config.example.yaml config.yaml
   ```

2. **Edit config.yaml with your database details:**
   ```yaml
   source_db:
     host: "your-source-db.example.com"
     port: 3306
     username: "monitor_user"
     password: "your_password"
     database: "your_database"

   target_db:
     host: "your-target-db.example.com"
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
   ```

3. **For sensitive credentials, use environment variables:**
   ```bash
   export SOURCE_DB_PASSWORD="your_source_password"
   export TARGET_DB_PASSWORD="your_target_password"
   ```

## Running the Monitor

### Start the Application

```bash
./monitor -config config.yaml
```

You should see output like:

```
2024/11/04 10:00:00 Loading configuration...
2024/11/04 10:00:00 Configuration loaded successfully
2024/11/04 10:00:00 Monitoring interval: 30s
2024/11/04 10:00:00 Starting monitoring engine...
2024/11/04 10:00:00 Successfully connected to source database
2024/11/04 10:00:00 Successfully connected to target database
2024/11/04 10:00:00 Monitoring engine started
2024/11/04 10:00:00 Starting web server on port 8080...
2024/11/04 10:00:00 MariaDB Encryption Migration Monitor is running
2024/11/04 10:00:00 Access the web interface at http://localhost:8080
```

### Access the Web Interface

Open your browser and navigate to:
```
http://localhost:8080
```

You'll see the monitoring dashboard with:
- Connection status for both databases
- Real-time replica lag metrics
- Checksum validation results
- Data consistency status
- Active alerts

## Monitoring During Migration

### Typical AWS RDS Encryption Migration Workflow

1. **Before Migration:**
   - Set up replication from source to target
   - Start the monitor to establish baseline metrics

2. **During Migration:**
   - Monitor replica lag to ensure replication is keeping up
   - Watch for checksum mismatches
   - Check data consistency across tables
   - Review alerts for any issues

3. **Before Cutover:**
   - Verify replica lag is minimal (< 1 second)
   - Confirm all checksums match
   - Ensure row counts are consistent
   - Check that no critical alerts are active

4. **After Cutover:**
   - Continue monitoring for a period to ensure stability
   - Verify application connectivity to new encrypted database

## Understanding the Dashboard

### Connection Status
- **Green dot**: Database is connected and healthy
- **Red dot**: Database connection failed

### Replica Lag
- **Green**: Lag < 10 seconds (healthy)
- **Yellow**: Lag 10-60 seconds (warning)
- **Red**: Lag > 60 seconds (critical)

### Checksum Validation
- **✓ Match**: Table checksums match between source and target
- **✗ Mismatch**: Data integrity issue detected

### Data Consistency
- **✓ Consistent**: Row counts match
- **✗ Inconsistent**: Row count mismatch detected

### Alerts
- **CRITICAL**: Immediate attention required (checksum mismatch, replication stopped)
- **WARNING**: Monitor closely (high lag, connection issues)
- **INFO**: Informational messages

## API Usage

### Get Current Metrics
```bash
curl http://localhost:8080/api/metrics | jq
```

### Get Alert History
```bash
curl http://localhost:8080/api/alerts | jq
```

### Health Check
```bash
curl http://localhost:8080/api/health
```

## Troubleshooting

### "Failed to connect to database"
- Verify database host and port are correct
- Check network connectivity: `telnet <host> <port>`
- Verify credentials are correct
- Check firewall rules

### "No replica lag data"
- Ensure target database is configured as a replica
- Verify user has `REPLICATION CLIENT` privilege
- Check replication status: `SHOW SLAVE STATUS;`

### "Checksum validation error"
- Verify tables exist in both databases
- Check table names (case-sensitive)
- Ensure user has `SELECT` permission

### High Memory Usage
- Reduce monitoring interval
- Monitor fewer tables
- Reduce history retention (modify code if needed)

## Stopping the Monitor

Press `Ctrl+C` to gracefully shutdown the monitor. It will:
1. Stop the monitoring engine
2. Close database connections
3. Shutdown the web server

## Next Steps

- Review the full [README.md](README.md) for detailed documentation
- Check the [AWS RDS encryption guide](https://repost.aws/knowledge-center/rds-encrypt-instance-mysql-mariadb)
- Set up automated alerts (integrate with your monitoring system)
- Consider adding authentication to the web interface for production use

## Support

For issues or questions, please refer to the main README or open an issue in the project repository.
