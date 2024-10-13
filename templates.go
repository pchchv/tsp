package main

const historyTemplateFile = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Status History</title>
    <style>
        body {
            font-family: sans-serif;
            line-height: 1.6;
            color: #e0e0e0;
            max-width: 1200px;
            margin: auto;
            padding: 20px;
            background: #181818;
            transition: background 0.3s ease, color 0.3s ease;
        }
        h1, h2 {
            color: #e0e0e0;
            text-align: center;
        }
        .history-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 40px;
        }
        .history-item {
            background: #242424;
            border-radius: 8px;
            padding: 15px;
            box-shadow: 0 2px 4px rgba(255,255,255,0.1);
            max-height: 300px;
            overflow: auto;
        }
        .history-item h2 {
            font-size: 1.2rem;
            margin: 0;
        }
        .history-entry {
            margin-bottom: 5px;
            font-size: 0.9rem;
            display: flex;
            justify-content: space-between;
        }
        .status-up { color: #27ae60; }
        .status-down { color: #e74c3c; }
        .footer {
            text-align: center;
            font-size: .9em;
            color: #a0a0a0;
            margin-top: 40px;
        }
        .footer a {
            color: #9b59b6;
            text-decoration: none;
        }
        .footer a:hover { text-decoration: underline; }
    </style>
</head>
<body>
<h1>Status History</h1>
<div class="history-grid">
    {{ range $service, $entries := .history }}
    <div class="history-item">
        <h2>{{ $service }}</h2>
        {{ range $entry := $entries }}
        <div class="history-entry">
            <span>{{ index (split $entry.Timestamp "T") 0 }} {{ slice (index (split $entry.Timestamp "T") 1) 0 8 }}</span>
            <span class="{{ if $entry.Status }}status-up{{ else }}status-down{{ end }}">
                {{ if $entry.Status }}Up{{ else }}Down{{ end }}
            </span>
        </div>
        {{ end }}
    </div>
    {{ end }}
</div>
<div class="footer">
    <p>Last updated: {{.last_updated}}</p>
    <p><a href="/">Back to Current Status</a></p>
	<p>Powered by <a href="https://github.com/pchchv/tsp">tsp</a></p>
</div>
</body>
</html>`
