# Multi-Database Monitoring Guide

The MariaDB Encryption Migration Monitor supports monitoring multiple database pairs simultaneously. This guide explains how to configure and use this feature.

## Overview

You can monitor multiple source-target database pairs in a single monitor instance. This is useful when:

- Migrating multiple databases to encryption simultaneously
- Monitoring databases across different regions
- Managing multiple applications with separate databases
- Consolidating monitoring for your entire infrastructure

## Configuration Format

### Basic Structure

```yaml
# Global settings (apply to all database pairs)
monitoring_interval: "30s"
replica_lag_threshold: "60s"
web_server_port: 8080
log_level: "info"

# Define multiple database pairs
database_pairs:
  - name: "database-1"
    source_db:
      host: "source1.example.com"
      port: 3306
      username: "user"
      password: "pass"
      database: "db1"
    target_db:
      host: "target1.example.com"
      port: 3306
      username: "user"
      password: "pass"
      database: "db1"
    tables_to_monitor:
      - "table1"
      - "table2"

  - name: "database-2"
    source_db:
      host: "source2.example.com"
      port: 3306
      username: "user"
      password: "pass"
      database: "db2"
    target_db:
      host: "target2.example.com"
      port: 3306
      username: "user"
      password: "pass"
      database: "db2"
    tables_to_monitor:
      - "table3"
      - "table4"
```

### Configuration Fields

#### Global Settings

- **monitoring_interval**: How often to check all database pairs (minimum 10 seconds)
- **replica_lag_threshold**: Alert threshold for replica lag (applies to all pairs)
- **web_server_port**: Port for the web interface
- **log_level**: Logging verbosity (debug, info, warn, error)

#### Database Pair Settings

Each database pair requires:

- **name**: Unique identifier for the database pair (used in logs and alerts)
- **source_db**: Connection details for the source (unencrypted) database
  - host, port, username, password, database
- **target_db**: Connection details for the target (encrypted) database
  - host, port, username, password, database
- **tables_to_monitor**: List of tables to check for checksums and consistency

## Examples

### Example 1: Multiple Production Databases

```yaml
monitoring_interval: "30s"
replica_lag_threshold: "60s"
web_server_port: 8080
log_level: "info"

database_pairs:
  - name: "prod-us-east"
    source_db:
      host: "prod-source-use1.rds.amazonaws.com"
      port: 3306
      username: "monitor"
      password: "password1"
      database: "production"
    target_db:
      host: "prod-target-use1.rds.amazonaws.com"
      port: 3306
      username: "monitor"
      password: "password1"
      database: "production"
    tables_to_monitor:
      - "users"
      - "orders"
      - "products"

  - name: "prod-eu-west"
    source_db:
      host: "prod-source-euw1.rds.amazonaws.com"
      port: 3306
      username: "monitor"
      password: "password2"
      database: "production"
    target_db:
      host: "prod-target-euw1.rds.amazonaws.com"
      port: 3306
      username: "monitor"
      password: "password2"
      database: "production"
    tables_to_monitor:
      - "users"
      - "orders"
      - "products"
```

### Example 2: Different Applications

```yaml
monitoring_interval: "30s"
replica_lag_threshold: "60s"
web_server_port: 8080
log_level: "info"

database_pairs:
  - name: "ecommerce-app"
    source_db:
      host: "ecommerce-source.example.com"
      port: 3306
      username: "monitor"
      password: "pass1"
      database: "ecommerce"
    target_db:
      host: "ecommerce-target.example.com"
      port: 3306
      username: "monitor"
      password: "pass1"
      database: "ecommerce"
    tables_to_monitor:
      - "products"
      - "orders"
      - "customers"

  - name: "analytics-app"
    source_db:
      host: "analytics-source.example.com"
      port: 3306
      username: "monitor"
      password: "pass2"
      database: "analytics"
    target_db:
      host: "analytics-target.example.com"
      port: 3306
      username: "monitor"
      password: "pass2"
      database: "analytics"
    tables_to_monitor:
      - "events"
      - "metrics"

  - name: "auth-app"
    source_db:
      host: "auth-source.example.com"
      port: 3306
      username: "monitor"
      password: "pass3"
      database: "authentication"
    target_db:
      host: "auth-target.example.com"
      port: 3306
      username: "monitor"
      password: "pass3"
      database: "authentication"
    tables_to_monitor:
      - "users"
      - "sessions"
      - "tokens"
```

### Example 3: Many Tables Per Database

```yaml
monitoring_interval: "30s"
replica_lag_threshold: "60s"
web_server_port: 8080
log_level: "info"

database_pairs:
  - name: "main-database"
    source_db:
      host: "main-source.example.com"
      port: 3306
      username: "monitor"
      password: "password"
      database: "maindb"
    target_db:
      host: "main-target.example.com"
      port: 3306
      username: "monitor"
      password: "password"
      database: "maindb"
    tables_to_monitor:
      - "users"
      - "user_profiles"
      - "user_preferences"
      - "orders"
      - "order_items"
      - "order_history"
      - "products"
      - "product_categories"
      - "product_reviews"
      - "inventory"
      - "inventory_transactions"
      - "payments"
      - "payment_methods"
      - "shipping_addresses"
      - "notifications"
```

## Web Interface

The web interface displays metrics for all configured database pairs:

### Connection Status
- Shows connection status for each database pair
- Green indicator: Both source and target connected
- Red indicator: Connection issue

### Replica Lag
- Displays current lag for each database pair
- Color-coded by severity (green < 10s, yellow < 60s, red > 60s)
- Separate metrics per database pair

### Checksum Validation
- Shows checksum results grouped by database pair
- Table-level granularity: `[database-pair] table_name`
- Match/mismatch indicators

### Data Consistency
- Row count comparison grouped by database pair
- Format: `[database-pair] table_name`
- Consistent/inconsistent indicators

### Alerts
- All alerts include the database pair name in brackets
- Example: `[production-db] Replica lag exceeds threshold`
- Helps identify which database pair has issues

## API Endpoints

### Get Metrics
```bash
curl http://localhost:8080/api/metrics | jq
```

Response includes metrics for all database pairs:
```json
{
  "ReplicaLag": {
    "production-db": {
      "DatabasePair": "production-db",
      "LagSeconds": 2.5,
      "Status": "ok"
    },
    "analytics-db": {
      "DatabasePair": "analytics-db",
      "LagSeconds": 45.2,
      "Status": "ok"
    }
  },
  "ChecksumResults": {
    "production-db:users": {
      "DatabasePair": "production-db",
      "TableName": "users",
      "Match": true
    },
    "analytics-db:events": {
      "DatabasePair": "analytics-db",
      "TableName": "events",
      "Match": true
    }
  },
  "ConnectionStatus": {
    "production-db": {
      "SourceConnected": true,
      "TargetConnected": true
    },
    "analytics-db": {
      "SourceConnected": true,
      "TargetConnected": true
    }
  }
}
```

### Health Check
```bash
curl http://localhost:8080/api/health | jq
```

Response:
```json
{
  "status": "ok",
  "total_pairs": 3,
  "connected_pairs": 2,
  "connection_status": {
    "production-db": {
      "SourceConnected": true,
      "TargetConnected": true
    },
    "analytics-db": {
      "SourceConnected": true,
      "TargetConnected": false
    },
    "customer-db": {
      "SourceConnected": true,
      "TargetConnected": true
    }
  }
}
```

## Logging

Logs include the database pair name for easy identification:

```
2024/11/04 10:00:00 Starting monitoring engine for 3 database pair(s)...
2024/11/04 10:00:00 Connecting to database pair: production-db
2024/11/04 10:00:00 Successfully connected to source[production-db] database
2024/11/04 10:00:00 Successfully connected to target[production-db] database
2024/11/04 10:00:00 Connecting to database pair: analytics-db
2024/11/04 10:00:00 Successfully connected to source[analytics-db] database
2024/11/04 10:00:00 Successfully connected to target[analytics-db] database
2024/11/04 10:00:05 Running monitoring cycle...
2024/11/04 10:00:05 [production-db] Replica lag: 2.5 seconds
2024/11/04 10:00:05 [analytics-db] Replica lag: 45.2 seconds
```

## Performance Considerations

### Resource Usage

- **Memory**: Each database pair adds minimal memory overhead (~1-2 MB)
- **CPU**: Monitoring cycles run concurrently for all pairs
- **Network**: Each pair requires 2 database connections (source + target)
- **Disk**: No disk usage (all metrics stored in memory)

### Recommendations

- **Small deployments** (1-5 pairs): No special considerations
- **Medium deployments** (6-20 pairs): Consider increasing monitoring interval to 60s
- **Large deployments** (20+ pairs): 
  - Use monitoring interval of 60-120s
  - Monitor fewer tables per database
  - Consider running multiple monitor instances

### Connection Limits

Each database pair requires 2 connections (source + target). Ensure your database servers can handle the connection load:

- 10 database pairs = 20 connections
- 50 database pairs = 100 connections
- 100 database pairs = 200 connections

## Troubleshooting

### Issue: Some database pairs not connecting

Check logs for specific error messages:
```
2024/11/04 10:00:00 Warning: Failed to connect to source database for pair 'analytics-db': dial tcp: connection refused
```

Solutions:
- Verify host and port are correct
- Check network connectivity
- Verify credentials
- Check firewall rules

### Issue: High memory usage

If monitoring many database pairs with long history:

1. Reduce monitoring interval
2. Monitor fewer tables
3. The application keeps 24 hours of replica lag history

### Issue: Slow monitoring cycles

If cycles take too long:

1. Increase monitoring interval
2. Reduce number of tables monitored
3. Check database query performance
4. Verify network latency

## Migration Workflow

### Before Migration

1. Configure all database pairs in config.yaml
2. Start the monitor
3. Verify all connections are successful
4. Establish baseline metrics

### During Migration

1. Monitor replica lag for all pairs
2. Watch for checksum mismatches
3. Check data consistency
4. Review alerts regularly

### Before Cutover

For each database pair:
- ✓ Replica lag < 1 second
- ✓ All checksums match
- ✓ Row counts consistent
- ✓ No critical alerts

### After Cutover

- Continue monitoring for 24-48 hours
- Verify application connectivity
- Monitor for any delayed issues

## Best Practices

1. **Naming Convention**: Use descriptive names for database pairs
   - Good: `production-us-east`, `analytics-eu-west`
   - Bad: `db1`, `db2`, `test`

2. **Table Selection**: Monitor critical tables first
   - Start with high-value tables
   - Add more tables gradually
   - Don't monitor all tables if not necessary

3. **Monitoring Interval**: Balance freshness vs. load
   - Critical migrations: 10-30 seconds
   - Standard migrations: 30-60 seconds
   - Low-priority: 60-120 seconds

4. **Alert Thresholds**: Adjust based on your requirements
   - High-traffic databases: Lower threshold (30s)
   - Low-traffic databases: Higher threshold (120s)

5. **Credentials**: Use environment variables for sensitive data
   - Store passwords in secrets management
   - Use IAM authentication where possible
   - Rotate credentials regularly

## Support

For issues or questions about multi-database monitoring, refer to the main README or open an issue in the project repository.
