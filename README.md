# 🔍 API Security Scanner

A containerized, real-time security scanner that continuously monitors your APIs and web endpoints for security vulnerabilities. Features a beautiful live dashboard that automatically updates via WebSocket whenever new scans complete.

## ✨ Features

### 🔐 Comprehensive Security Checks
- **HTTPS Enforcement** - Detects insecure HTTP connections
- **Security Headers** - Validates presence of CSP, HSTS, X-Frame-Options, X-Content-Type-Options
- **CORS Configuration** - Identifies overly permissive CORS policies
- **Server Fingerprinting** - Detects information leakage via Server/X-Powered-By headers
- **Cache Control** - Validates proper cache header configuration
- **HTTP Status Monitoring** - Tracks 4xx/5xx errors

### ⚡ Real-Time Monitoring
- Continuous scanning at configurable intervals (default: 30 seconds)
- Live dashboard updates via WebSocket - no manual refresh needed
- Response time tracking for performance monitoring
- Visual severity indicators (High/Medium/Info/Error)

### 📊 Beautiful Dashboard
- Dark-themed, GitHub-inspired UI
- Interactive charts (pie & bar charts) for data visualization
- Color-coded severity levels with visual badges
- Detailed results table with all security findings
- Mobile-responsive design

### 🐳 Fully Containerized
- Docker Compose orchestration for easy deployment
- Python scanner service (async/aiohttp)
- Go dashboard service (Gorilla WebSocket & Mux)
- Shared volume for report persistence
- Auto-restart policies for reliability

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- No other dependencies needed!

### Installation

1. **Clone the repository**
```bash
   git clone https://github.com/yourusername/api-security-scanner.git
   cd api-security-scanner
```

2. **Configure targets** (optional)
   
   Edit `docker-compose.yml` to add your endpoints:
```yaml
   environment:
     SCANNER_TARGETS: "https://api.example.com,https://api2.example.com"
```

3. **Start the scanner**
```bash
   docker-compose up --build
```

4. **Open the dashboard**
```
   http://localhost:8081
```

That's it! The scanner will start monitoring your endpoints immediately.

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SCANNER_TARGETS` | Comma-separated list of URLs to scan | `https://example.com,https://httpbin.org/get` |
| `SCAN_INTERVAL` | Seconds between scans | `30` |
| `SCANNER_TIMEOUT` | Request timeout in seconds | `10` |
| `SCANNER_CONCURRENCY` | Max concurrent requests | `5` |
| `SCANNER_AUTH_HEADER` | Optional auth header (e.g., "Bearer token") | - |

### Example Configuration
```yaml
environment:
  SCANNER_TARGETS: "https://api.myapp.com,https://api.myapp.com/v2"
  SCAN_INTERVAL: "60"
  SCANNER_TIMEOUT: "15"
  SCANNER_AUTH_HEADER: "Bearer your-api-token-here"
```

## 📁 Project Structure
```
api-security-scanner/
├── docker-compose.yml          # Service orchestration
├── scanner/                    # Python scanner service
│   ├── Dockerfile
│   ├── requirements.txt
│   └── scan.py                 # Async security scanner
└── worker-go/                  # Go dashboard service
    ├── Dockerfile
    ├── go.mod
    ├── go.sum
    └── main.go                 # WebSocket dashboard server
```

## 🎯 Use Cases

- **Continuous Security Monitoring** - Keep track of your API security posture 24/7
- **DevOps Integration** - Run as part of your CI/CD pipeline
- **Compliance Auditing** - Generate reports for security compliance
- **Penetration Testing** - Initial reconnaissance for security assessments
- **API Gateway Monitoring** - Ensure proper security headers across all endpoints

## 📊 Dashboard Features

### Live Stats
- Total endpoints scanned
- High/Medium/Low risk issue counts
- Error tracking
- Scan duration and timestamps

### Visual Analytics
- **Pie Chart** - Severity distribution overview
- **Bar Chart** - Issues per endpoint comparison
- **Real-time Updates** - Auto-refresh on new scans

### Detailed Results
- URL and HTTP status codes
- Response times (performance metrics)
- Complete list of security issues
- Missing security headers breakdown
- Color-coded severity levels

## 🔧 Commands

### Basic Operations
```bash
# Start services
docker-compose up

# Start in background
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Complete reset
docker-compose down -v
```

### Individual Service Logs
```bash
# Scanner logs
docker-compose logs -f scanner

# Dashboard logs
docker-compose logs -f worker
```

## 🛠️ Development

### Scanner (Python)
- Built with `aiohttp` for async HTTP requests
- Uses `yarl` for URL parsing
- Configurable concurrency and timeout
- JSON report generation

### Dashboard (Go)
- Gorilla Mux for routing
- Gorilla WebSocket for real-time updates
- Chart.js for data visualization
- File system watcher for report updates

## 📈 Example Output

### Console Output
```
🚀 API Security Scanner - Real-Time Mode
⏱️  Scan interval: 30 seconds

============================================================
🔍 SCAN #20241116_143052
============================================================
Targets: 3
  1. https://example.com
  2. https://httpbin.org/get
  3. https://jsonplaceholder.typicode.com/posts/1

✓ https://example.com [200] 145.23ms
✓ https://httpbin.org/get [200] 234.56ms
✓ https://jsonplaceholder.typicode.com/posts/1 [200] 189.45ms

============================================================
📊 SUMMARY
============================================================
Duration: 0.45s
🔴 High:   0
🟡 Medium: 2
🔵 Info:   1
⚠️  Errors: 0
============================================================
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [aiohttp](https://docs.aiohttp.org/) for async HTTP
- Dashboard powered by [Gorilla WebSocket](https://github.com/gorilla/websocket)
- Charts by [Chart.js](https://www.chartjs.org/)
- Inspired by OWASP security best practices

## 🐛 Troubleshooting

### Dashboard shows "Waiting for first scan"
Wait 30 seconds for the first scan cycle to complete.

### Port 8081 already in use
Change the port mapping in `docker-compose.yml`:
```yaml
ports:
  - "8082:8081"
```

### Scanner failing to connect
- Verify URLs are accessible
- Check if authentication is required
- Increase timeout value
