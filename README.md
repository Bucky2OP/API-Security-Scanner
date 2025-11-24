🚀 API Security Scanner

A fully automated, real-time API security scanning system that analyzes endpoints for common vulnerabilities, missing security headers, CORS risks, server fingerprinting leaks, and misconfigurations.
The system includes:

Python-based async scanner (aiohttp)

Go-based real-time dashboard with live WebSockets

Dark mode interface

Visual analytics (charts)

Auto-scanning every X seconds

Instant UI updates on new scan results

JSON report generation

Docker-ready architecture

📸 Dashboard Preview

✔ Dark Mode
✔ Charts (Pie + Bar)
✔ Auto-refresh via WebSockets
✔ Severity color coding
✔ Live results from Python scanner

(You can add screenshots here once your dashboard is running.)

📂 Project Structure
API-Security-Scanner/
│
├── scanner/                # Python async scanner
│   ├── scan.py
│   ├── requirements.txt
│   └── Dockerfile
│
├── worker-go/              # Go dashboard (WebSockets + Charts)
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
│
├── reports/                # Shared folder for report.json
│
└── docker-compose.yml      # Orchestration of scanner + dashboard

⚡ Features
🕵️‍♂️ 1. Automated Security Scanning

The Python scanner checks each target for:

Missing security headers

X-Frame-Options

Content-Security-Policy

X-Content-Type-Options

Strict-Transport-Security

CORS misconfigurations

Missing Content-Type for APIs

Missing Cache-Control

Server fingerprinting (Server, X-Powered-By)

HTTP/HTTPS validation

4xx/5xx response validation

Results are exported to:

reports/report.json

🔄 2. Automatic Scheduled Scanning

The scanner re-runs every X seconds (default 30):

SCAN_INTERVAL = 30


You can override with env:

SCAN_INTERVAL=10

🌐 3. Real-Time Dashboard (Go + WebSockets)

The dashboard:

Loads latest report.json

Displays stylish dark UI

Auto-refreshes when new results are available

Uses WebSockets to update instantly without reloading

Renders:

Pie chart: Severity distribution

Bar chart: Issues per endpoint

Table: Detailed findings & headers

Built with:

Go (mux router)

Gorilla/WebSocket

Chart.js

HTML templates

🎨 4. Dark Mode UI

The interface includes:

GitHub-style colors

Glow/shadow boxes

Highlighted severity rows

Easy-to-read preformatted headers

📡 5. WebSocket Live Updates

The Go service watches report.json for changes:

When scanner updates the file → dashboard gets a "reload" WebSocket push

Browser refreshes instantly

No manual refresh needed.
No polling.
Instant UI updates.

▶️ Getting Started
1. Install Python dependencies
cd scanner
pip install -r requirements.txt

2. Install Go modules
cd worker-go
go mod tidy

3. Run the Scanner (auto loops)
cd scanner
$env:REPORT_PATH="../reports/report.json"
py scan.py

4. Run the Dashboard
cd worker-go
$env:REPORT_PATH="../reports/report.json"
go run main.go


Open:

👉 http://localhost:8081/

🐳 Docker Support

You can run everything via Docker Compose:

docker-compose up --build


The scanner container will auto-scan every X seconds and the Go worker will update in real time.

⚙ Configuration

Environment variables:

Variable	Description	Default
SCAN_INTERVAL	Seconds between scans	30
SCANNER_TARGETS	Comma-separated list of URLs	example.com, httpbin
REPORT_PATH	Path to output JSON	/reports/report.json
📈 Roadmap / Future Improvements

 Export PDF reports

 Email alerts for high severity findings

 Add OAuth/JWT tests

 Rate limiting detection

 OpenAPI/Swagger scanning

 Multi-user dashboard

 Deploy to AWS (EC2, ECS, Lambda)

🤝 Contributing

Pull requests welcome!
Feel free to add more scanners, rules, or dashboards.
