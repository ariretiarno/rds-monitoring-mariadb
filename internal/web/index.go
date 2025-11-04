package web

const indexHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MariaDB Encryption Migration Monitor</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: #f5f7fa;
            color: #333;
            padding: 20px;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
        }

        h1 {
            color: #2c3e50;
            margin-bottom: 10px;
        }

        .subtitle {
            color: #7f8c8d;
            margin-bottom: 30px;
        }

        .status-bar {
            background: white;
            padding: 15px 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            flex-wrap: wrap;
        }

        .connection-status {
            display: flex;
            gap: 20px;
            flex-wrap: wrap;
        }

        .status-item {
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .status-dot {
            width: 12px;
            height: 12px;
            border-radius: 50%;
            background: #95a5a6;
        }

        .status-dot.connected {
            background: #27ae60;
        }

        .status-dot.disconnected {
            background: #e74c3c;
        }

        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }

        .card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .card h2 {
            font-size: 18px;
            color: #2c3e50;
            margin-bottom: 15px;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
        }

        .metric {
            margin-bottom: 15px;
        }

        .metric-label {
            font-size: 14px;
            color: #7f8c8d;
            margin-bottom: 5px;
        }

        .metric-value {
            font-size: 28px;
            font-weight: bold;
            color: #2c3e50;
        }

        .metric-value.good {
            color: #27ae60;
        }

        .metric-value.warning {
            color: #f39c12;
        }

        .metric-value.critical {
            color: #e74c3c;
        }

        table {
            width: 100%;
            border-collapse: collapse;
        }

        th, td {
            padding: 10px;
            text-align: left;
            border-bottom: 1px solid #ecf0f1;
        }

        th {
            background: #f8f9fa;
            font-weight: 600;
            color: #2c3e50;
        }

        .badge {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: 600;
        }

        .badge.success {
            background: #d4edda;
            color: #155724;
        }

        .badge.danger {
            background: #f8d7da;
            color: #721c24;
        }

        .badge.warning {
            background: #fff3cd;
            color: #856404;
        }

        .badge.info {
            background: #d1ecf1;
            color: #0c5460;
        }

        .alert-item {
            padding: 12px;
            margin-bottom: 10px;
            border-radius: 6px;
            border-left: 4px solid;
        }

        .alert-item.CRITICAL {
            background: #f8d7da;
            border-color: #e74c3c;
        }

        .alert-item.WARNING {
            background: #fff3cd;
            border-color: #f39c12;
        }

        .alert-item.INFO {
            background: #d1ecf1;
            border-color: #3498db;
        }

        .alert-time {
            font-size: 12px;
            color: #7f8c8d;
        }

        .no-data {
            text-align: center;
            color: #95a5a6;
            padding: 20px;
        }

        .last-updated {
            font-size: 12px;
            color: #95a5a6;
        }

        .db-pair-title {
            margin-top: 30px;
            margin-bottom: 15px;
            color: #2c3e50;
            font-size: 24px;
            border-bottom: 3px solid #3498db;
            padding-bottom: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔒 MariaDB Encryption Migration Monitor</h1>
        <p class="subtitle">Real-time monitoring of database encryption migration</p>

        <div class="status-bar">
            <div class="connection-status" id="connection-status">
                <div class="no-data">Loading...</div>
            </div>
            <div class="last-updated" id="last-updated">Last updated: Never</div>
        </div>

        <div id="database-pairs-container">
            <div class="no-data">Loading database pairs...</div>
        </div>

        <div class="card">
            <h2>🚨 Active Alerts</h2>
            <div id="alerts">
                <div class="no-data">No active alerts</div>
            </div>
        </div>
    </div>

    <script>
        let ws;
        let reconnectInterval = 5000;

        function connectWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            ws = new WebSocket(protocol + '//' + window.location.host + '/ws');

            ws.onopen = function() {
                console.log('WebSocket connected');
            };

            ws.onmessage = function(event) {
                const message = JSON.parse(event.data);
                if (message.type === 'metrics_update') {
                    updateMetrics(message.data);
                }
            };

            ws.onclose = function() {
                console.log('WebSocket disconnected, reconnecting...');
                setTimeout(connectWebSocket, reconnectInterval);
            };

            ws.onerror = function(error) {
                console.error('WebSocket error:', error);
            };
        }

        function updateMetrics(data) {
            // Update connection status for all database pairs
            if (data.ConnectionStatus) {
                const statusDiv = document.getElementById('connection-status');
                const pairs = Object.keys(data.ConnectionStatus);
                
                if (pairs.length === 0) {
                    statusDiv.innerHTML = '<div class="no-data">No database pairs configured</div>';
                } else {
                    let html = '';
                    pairs.forEach(pairName => {
                        const status = data.ConnectionStatus[pairName];
                        const sourceClass = status.SourceConnected ? 'connected' : 'disconnected';
                        const targetClass = status.TargetConnected ? 'connected' : 'disconnected';
                        html += '<div class="status-item">';
                        html += '<div class="status-dot ' + sourceClass + '"></div>';
                        html += '<div class="status-dot ' + targetClass + '"></div>';
                        html += '<span>' + pairName + '</span>';
                        html += '</div>';
                    });
                    statusDiv.innerHTML = html;
                }
            }

            // Group data by database pair
            const databasePairs = {};
            
            // Collect all database pair names
            if (data.ReplicaLag) {
                Object.keys(data.ReplicaLag).forEach(pair => {
                    if (!databasePairs[pair]) databasePairs[pair] = {};
                    databasePairs[pair].replicaLag = data.ReplicaLag[pair];
                });
            }
            
            if (data.ChecksumResults) {
                Object.keys(data.ChecksumResults).forEach(key => {
                    const parts = key.split(':');
                    const pair = parts[0];
                    if (!databasePairs[pair]) databasePairs[pair] = {};
                    if (!databasePairs[pair].checksums) databasePairs[pair].checksums = {};
                    databasePairs[pair].checksums[parts[1]] = data.ChecksumResults[key];
                });
            }
            
            if (data.ConsistencyResults) {
                Object.keys(data.ConsistencyResults).forEach(key => {
                    const parts = key.split(':');
                    const pair = parts[0];
                    if (!databasePairs[pair]) databasePairs[pair] = {};
                    if (!databasePairs[pair].consistency) databasePairs[pair].consistency = {};
                    databasePairs[pair].consistency[parts[1]] = data.ConsistencyResults[key];
                });
            }

            // Render each database pair
            const container = document.getElementById('database-pairs-container');
            const pairNames = Object.keys(databasePairs);
            
            if (pairNames.length === 0) {
                container.innerHTML = '<div class="no-data">No data available</div>';
            } else {
                let html = '';
                pairNames.forEach(pairName => {
                    const pairData = databasePairs[pairName];
                    html += '<h2 class="db-pair-title">📦 ' + pairName + '</h2>';
                    html += '<div class="grid">';
                    
                    // Replica Lag Card
                    html += '<div class="card"><h2>📊 Replica Lag</h2>';
                    if (pairData.replicaLag) {
                        const lag = pairData.replicaLag;
                        let lagClass = 'metric-value';
                        if (lag.LagSeconds < 10) lagClass += ' good';
                        else if (lag.LagSeconds < 60) lagClass += ' warning';
                        else lagClass += ' critical';
                        
                        html += '<div class="metric">';
                        html += '<div class="metric-label">Current Lag</div>';
                        html += '<div class="' + lagClass + '">' + (lag.LagSeconds || 0).toFixed(2) + 's</div>';
                        html += '</div>';
                        html += '<div class="metric-label">Status: <span>' + (lag.Status || 'unknown') + '</span></div>';
                    } else {
                        html += '<div class="no-data">No data</div>';
                    }
                    html += '</div>';
                    
                    // Checksum Card
                    html += '<div class="card"><h2>🔍 Checksum Validation</h2>';
                    if (pairData.checksums && Object.keys(pairData.checksums).length > 0) {
                        html += '<table><tr><th>Table</th><th>Status</th></tr>';
                        Object.keys(pairData.checksums).forEach(table => {
                            const result = pairData.checksums[table];
                            const badge = result.Match ? 
                                '<span class="badge success">✓ Match</span>' : 
                                '<span class="badge danger">✗ Mismatch</span>';
                            html += '<tr><td>' + table + '</td><td>' + badge + '</td></tr>';
                        });
                        html += '</table>';
                    } else {
                        html += '<div class="no-data">No data</div>';
                    }
                    html += '</div>';
                    
                    // Consistency Card
                    html += '<div class="card"><h2>✓ Data Consistency</h2>';
                    if (pairData.consistency && Object.keys(pairData.consistency).length > 0) {
                        html += '<table><tr><th>Table</th><th>Source</th><th>Target</th><th>Status</th></tr>';
                        Object.keys(pairData.consistency).forEach(table => {
                            const result = pairData.consistency[table];
                            const badge = result.Consistent ? 
                                '<span class="badge success">✓ Consistent</span>' : 
                                '<span class="badge danger">✗ Inconsistent</span>';
                            html += '<tr><td>' + table + '</td><td>' + result.SourceRowCount + '</td><td>' + result.TargetRowCount + '</td><td>' + badge + '</td></tr>';
                        });
                        html += '</table>';
                    } else {
                        html += '<div class="no-data">No data</div>';
                    }
                    html += '</div>';
                    
                    html += '</div>'; // Close grid
                });
                container.innerHTML = html;
            }

            // Update last updated time
            document.getElementById('last-updated').textContent = 'Last updated: ' + new Date().toLocaleTimeString();

            // Fetch and update alerts
            fetchAlerts();
        }

        function fetchAlerts() {
            fetch('/api/alerts')
                .then(response => response.json())
                .then(alerts => {
                    const alertsDiv = document.getElementById('alerts');
                    const activeAlerts = alerts.filter(a => !a.Resolved);
                    
                    if (activeAlerts.length === 0) {
                        alertsDiv.innerHTML = '<div class="no-data">No active alerts</div>';
                    } else {
                        let html = '';
                        activeAlerts.forEach(alert => {
                            const time = new Date(alert.Timestamp).toLocaleString();
                            html += '<div class="alert-item ' + alert.Severity + '">';
                            html += '<strong>' + alert.Severity + '</strong>: ' + alert.Message;
                            html += '<div class="alert-time">' + time + '</div>';
                            html += '</div>';
                        });
                        alertsDiv.innerHTML = html;
                    }
                })
                .catch(error => console.error('Error fetching alerts:', error));
        }

        // Connect on page load
        connectWebSocket();
    </script>
</body>
</html>
`
