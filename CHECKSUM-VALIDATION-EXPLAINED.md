# Checksum Validation - How It Works

## Overview

The checksum validation feature verifies data integrity between source and target databases by comparing cryptographic checksums of table data. This ensures that data has been replicated correctly without corruption during the encryption migration process.

## How It Works

### 1. The CHECKSUM TABLE Command

The application uses MariaDB's built-in `CHECKSUM TABLE` command:

```sql
CHECKSUM TABLE `table_name`
```

This command:
- Calculates a checksum value for the entire table
- Includes all rows and columns in the calculation
- Returns a numeric checksum value
- Is fast and efficient (uses MariaDB's internal algorithms)

### 2. Validation Process

For each table configured in `tables_to_monitor`, the application:

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Connect to Source Database                               │
│    Execute: CHECKSUM TABLE `users`                          │
│    Result: 1234567890                                        │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. Connect to Target Database                               │
│    Execute: CHECKSUM TABLE `users`                          │
│    Result: 1234567890                                        │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. Compare Checksums                                         │
│    Source: 1234567890                                        │
│    Target: 1234567890                                        │
│    Match: ✓ YES                                              │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. Store Result & Display in Web Interface                  │
│    Status: ✓ Match                                           │
└─────────────────────────────────────────────────────────────┘
```

### 3. Code Flow

Here's the step-by-step code execution:

```go
// 1. For each table in configuration
for _, table := range tables {
    
    // 2. Calculate checksum on source database
    sourceChecksum := CHECKSUM TABLE `table_name` on source
    
    // 3. Calculate checksum on target database
    targetChecksum := CHECKSUM TABLE `table_name` on target
    
    // 4. Compare the checksums
    if sourceChecksum == targetChecksum {
        result.Match = true  // ✓ Data is identical
    } else {
        result.Match = false // ✗ Data differs - ALERT!
    }
    
    // 5. Store and display result
    storage.StoreChecksumResult(result)
}
```

## What Gets Checked

The `CHECKSUM TABLE` command validates:

✅ **All Rows**: Every row in the table
✅ **All Columns**: Every column in each row
✅ **Data Values**: The actual data content
✅ **Data Types**: Implicit in the checksum calculation
✅ **NULL Values**: Properly handled in checksum

It does NOT check:
❌ Table structure (column definitions)
❌ Indexes
❌ Triggers
❌ Constraints
❌ Auto-increment values

## Example Scenarios

### Scenario 1: Perfect Match ✓

**Source Database:**
```
users table:
id | name  | email
1  | Alice | alice@example.com
2  | Bob   | bob@example.com
```

**Target Database:**
```
users table:
id | name  | email
1  | Alice | alice@example.com
2  | Bob   | bob@example.com
```

**Result:**
- Source Checksum: `1234567890`
- Target Checksum: `1234567890`
- **Match: ✓ YES**
- Status: Data is identical

---

### Scenario 2: Data Mismatch ✗

**Source Database:**
```
users table:
id | name  | email
1  | Alice | alice@example.com
2  | Bob   | bob@example.com
```

**Target Database:**
```
users table:
id | name  | email
1  | Alice | alice@example.com
2  | Bob   | bob@different.com  ← Different!
```

**Result:**
- Source Checksum: `1234567890`
- Target Checksum: `9876543210`
- **Match: ✗ NO**
- Alert: `[production-db] Checksum mismatch for table users`

---

### Scenario 3: Missing Row ✗

**Source Database:**
```
users table:
id | name  | email
1  | Alice | alice@example.com
2  | Bob   | bob@example.com
3  | Carol | carol@example.com
```

**Target Database:**
```
users table:
id | name  | email
1  | Alice | alice@example.com
2  | Bob   | bob@example.com
← Row 3 missing!
```

**Result:**
- Source Checksum: `1234567890`
- Target Checksum: `5555555555`
- **Match: ✗ NO**
- Alert: `[production-db] Checksum mismatch for table users`

## When Checksums Are Calculated

Checksums are calculated:

1. **On Schedule**: Every monitoring interval (e.g., every 30 seconds)
2. **For All Tables**: All tables listed in `tables_to_monitor`
3. **Concurrently**: All database pairs checked in parallel
4. **Continuously**: As long as the monitor is running

## Performance Considerations

### Speed
- **Fast**: MariaDB's `CHECKSUM TABLE` is optimized
- **Table Size Impact**: Larger tables take longer
- **Typical Speed**: 
  - Small tables (< 1M rows): < 1 second
  - Medium tables (1-10M rows): 1-5 seconds
  - Large tables (> 10M rows): 5-30 seconds

### Database Load
- **Read-Only**: No writes to the database
- **Table Lock**: Brief read lock during checksum
- **CPU Usage**: Minimal on database server
- **I/O Impact**: Reads table data from disk/cache

### Recommendations

**For Small Tables (< 100K rows):**
- Monitor every 10-30 seconds
- No performance concerns

**For Medium Tables (100K - 10M rows):**
- Monitor every 30-60 seconds
- Acceptable performance impact

**For Large Tables (> 10M rows):**
- Monitor every 60-120 seconds
- Consider monitoring only critical tables
- May want to monitor during off-peak hours

**For Very Large Tables (> 100M rows):**
- Monitor every 5-10 minutes
- Monitor only the most critical tables
- Consider using sampling or alternative methods

## Limitations

### 1. Doesn't Identify Specific Differences

If checksums don't match, you know data differs but not:
- Which rows are different
- Which columns are different
- What the actual differences are

**Solution**: Use the consistency checker (row counts) and manual investigation.

### 2. Timing Sensitivity

If data is actively being written:
- Source and target may have different data at the moment of check
- This is expected during active replication
- Wait for replication to catch up (check replica lag first)

### 3. Table Must Exist

If a table doesn't exist in either database:
- Checksum returns NULL
- Error is logged
- Alert is generated

## Integration with Other Checks

The checksum validation works alongside:

### 1. Replica Lag Monitoring
```
Check replica lag first → If lag is low → Then check checksums
```
- If replica lag is high, checksums may not match (expected)
- Only trust checksum results when replica lag is low (< 5 seconds)

### 2. Consistency Checking (Row Counts)
```
Checksum mismatch → Check row counts → Identify scope of issue
```
- Row count check is faster but less thorough
- Use together for complete picture

### 3. Alert System
```
Checksum mismatch → Generate CRITICAL alert → Notify admin
```
- Checksum mismatches are treated as CRITICAL
- Requires immediate investigation

## Monitoring Workflow

### Ideal Scenario
```
1. Replica Lag: 0.5 seconds ✓
2. Row Counts: Match ✓
3. Checksums: Match ✓
→ Safe to cutover!
```

### Problem Scenario
```
1. Replica Lag: 0.5 seconds ✓
2. Row Counts: Match ✓
3. Checksums: MISMATCH ✗
→ Data corruption detected!
→ Investigate before cutover
```

## Troubleshooting

### Issue: Checksums Always Mismatch

**Possible Causes:**
1. **Active Writes**: Data is being written to source
   - Solution: Check replica lag, wait for it to catch up

2. **Replication Not Working**: Target not receiving updates
   - Solution: Check `SHOW SLAVE STATUS` on target

3. **Different Data**: Actual data corruption or replication issue
   - Solution: Investigate specific tables, check logs

4. **Table Structure Differs**: Columns added/removed
   - Solution: Verify table structures match

### Issue: Checksum Returns NULL

**Possible Causes:**
1. **Table Doesn't Exist**: Table not in target database
   - Solution: Verify table exists in both databases

2. **Permission Issue**: User lacks SELECT permission
   - Solution: Grant SELECT permission to monitor user

3. **Corrupted Table**: Table is corrupted
   - Solution: Run `CHECK TABLE` and repair if needed

### Issue: Checksum Takes Too Long

**Possible Causes:**
1. **Large Table**: Table has many rows
   - Solution: Increase monitoring interval

2. **Slow Disk**: Database on slow storage
   - Solution: Optimize database server

3. **High Load**: Database server is busy
   - Solution: Monitor during off-peak hours

## Best Practices

1. **Start with Critical Tables**
   - Monitor your most important tables first
   - Add more tables gradually

2. **Check Replica Lag First**
   - Always verify replica lag is low before trusting checksums
   - Recommended: < 5 seconds

3. **Monitor During Stable Periods**
   - Best results when writes are minimal
   - Consider monitoring during maintenance windows

4. **Use Appropriate Intervals**
   - Balance freshness vs. database load
   - Adjust based on table sizes

5. **Investigate Mismatches Immediately**
   - Checksum mismatches are serious
   - Don't proceed with cutover until resolved

6. **Combine with Other Checks**
   - Use replica lag + row counts + checksums together
   - All three should agree before cutover

## Configuration Example

```yaml
database_pairs:
  - name: "production-db"
    source_db:
      host: "source.example.com"
      port: 3306
      username: "monitor_user"
      password: "password"
      database: "production"
    target_db:
      host: "target.example.com"
      port: 3306
      username: "monitor_user"
      password: "password"
      database: "production"
    tables_to_monitor:
      # Critical tables - check frequently
      - "users"
      - "orders"
      - "payments"
      
      # Important tables
      - "products"
      - "inventory"
      
      # Less critical - consider removing if performance is an issue
      # - "logs"
      # - "audit_trail"

# Check every 30 seconds
monitoring_interval: "30s"
```

## Summary

The checksum validation:
- ✅ Uses MariaDB's built-in `CHECKSUM TABLE` command
- ✅ Compares entire table contents between source and target
- ✅ Detects any data differences or corruption
- ✅ Runs automatically on configured tables
- ✅ Generates alerts on mismatches
- ✅ Works alongside replica lag and consistency checks
- ✅ Essential for verifying safe migration cutover

It's a critical component of ensuring data integrity during your MariaDB encryption migration!
