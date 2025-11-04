# Multi-Database Support - Update Summary

## What Changed

The MariaDB Encryption Migration Monitor now supports monitoring **multiple database pairs** simultaneously. This is a major enhancement that allows you to monitor many databases from a single monitor instance.

## Key Features

✅ **Multiple Database Pairs**: Monitor unlimited source-target database pairs
✅ **Per-Pair Configuration**: Each pair has its own tables to monitor
✅ **Grouped Display**: Web interface groups metrics by database pair
✅ **Labeled Alerts**: All alerts include the database pair name
✅ **Concurrent Monitoring**: All pairs monitored in parallel
✅ **Backward Compatible**: Old single-database configs still work

## Configuration Changes

### Old Format (Still Supported)
```yaml
source_db:
  host: "source.example.com"
  ...
target_db:
  host: "target.example.com"
  ...
tables_to_monitor:
  - "users"
```

### New Format (Recommended)
```yaml
database_pairs:
  - name: "production-db"
    source_db:
      host: "source1.example.com"
      ...
    target_db:
      host: "target1.example.com"
      ...
    tables_to_monitor:
      - "users"
      - "orders"
  
  - name: "analytics-db"
    source_db:
      host: "source2.example.com"
      ...
    target_db:
      host: "target2.example.com"
      ...
    tables_to_monitor:
      - "events"
```

## Web Interface Changes

### Before
- Single connection status indicator
- One replica lag metric
- Tables listed without context

### After
- Connection status for each database pair
- Replica lag per database pair
- Tables grouped by database pair
- Clear visual separation between pairs

## API Response Changes

### Metrics Endpoint (`/api/metrics`)

**Before:**
```json
{
  "ReplicaLag": {
    "LagSeconds": 2.5,
    "Status": "ok"
  },
  "ChecksumResults": {
    "users": { "Match": true }
  }
}
```

**After:**
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
  }
}
```

### Health Endpoint (`/api/health`)

**Before:**
```json
{
  "status": "ok",
  "source_connected": true,
  "target_connected": true
}
```

**After:**
```json
{
  "status": "ok",
  "total_pairs": 2,
  "connected_pairs": 2,
  "connection_status": {
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

## Log Format Changes

**Before:**
```
2024/11/04 10:00:00 Successfully connected to source database
2024/11/04 10:00:00 Replica lag: 2.5 seconds
```

**After:**
```
2024/11/04 10:00:00 Starting monitoring engine for 2 database pair(s)...
2024/11/04 10:00:00 Connecting to database pair: production-db
2024/11/04 10:00:00 Successfully connected to source[production-db] database
2024/11/04 10:00:05 [production-db] Replica lag: 2.5 seconds
2024/11/04 10:00:05 [analytics-db] Replica lag: 45.2 seconds
```

## Alert Message Changes

**Before:**
```
Replica lag (65.2 seconds) exceeds threshold (60.0 seconds)
Checksum mismatch for table users
```

**After:**
```
[production-db] Replica lag (65.2 seconds) exceeds threshold (60.0 seconds)
[analytics-db] Checksum mismatch for table users
```

## Migration Guide

### If You Have a Single Database

**Option 1: Keep using old format (automatic conversion)**
Your existing config will work as-is. It will be automatically converted to the new format internally.

**Option 2: Update to new format**
```yaml
# Change from:
source_db:
  host: "source.example.com"
  ...
target_db:
  host: "target.example.com"
  ...
tables_to_monitor:
  - "users"

# To:
database_pairs:
  - name: "my-database"
    source_db:
      host: "source.example.com"
      ...
    target_db:
      host: "target.example.com"
      ...
    tables_to_monitor:
      - "users"
```

### If You Want Multiple Databases

1. Copy `config-multi-db.example.yaml` to `config.yaml`
2. Add your database pairs
3. Restart the monitor
4. Access the web interface to see all pairs

## Example Configurations

See these files for examples:
- `config-single-db.example.yaml` - Single database pair
- `config-multi-db.example.yaml` - Multiple database pairs
- `MULTI-DATABASE-GUIDE.md` - Comprehensive guide

## Performance Impact

- **Memory**: ~1-2 MB per database pair
- **CPU**: Minimal (concurrent monitoring)
- **Network**: 2 connections per pair (source + target)
- **Monitoring Cycle**: All pairs monitored in parallel

## Troubleshooting

### Issue: "Cannot read properties of undefined"

This was a JavaScript error in the old web interface. It's now fixed in the updated version. If you see this:

1. Rebuild the application: `go build -o monitor ./cmd/monitor`
2. Restart the monitor
3. Hard refresh your browser (Ctrl+F5 or Cmd+Shift+R)

### Issue: Old config not working

If your old config format isn't working:

1. Check that you have at least one database pair configured
2. The automatic conversion only works if you have `source_db` and `target_db` defined
3. Consider updating to the new `database_pairs` format

### Issue: Can't see all database pairs

1. Check the logs for connection errors
2. Verify each database pair has a unique name
3. Check that all required fields are filled in
4. Use `/api/health` to see connection status

## Benefits of Multi-Database Support

1. **Consolidated Monitoring**: One instance for all databases
2. **Cost Effective**: Fewer resources than multiple instances
3. **Unified View**: See all migrations in one place
4. **Easier Management**: Single configuration file
5. **Better Visibility**: Compare metrics across databases

## Next Steps

1. Review the `MULTI-DATABASE-GUIDE.md` for detailed documentation
2. Check example configurations in `config-multi-db.example.yaml`
3. Update your configuration to add more database pairs
4. Monitor multiple databases simultaneously!

## Questions?

Refer to:
- `README.md` - Main documentation
- `MULTI-DATABASE-GUIDE.md` - Multi-database specific guide
- `QUICKSTART.md` - Getting started guide
