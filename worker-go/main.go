package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type ScanResult struct {
	URL             string                 `json:"url"`
	Status          int                    `json:"status"`
	SecurityHeaders map[string]interface{} `json:"security_headers"`
	Issues          []string               `json:"issues"`
	Severity        string                 `json:"severity"`
	Error           string                 `json:"error"`
	ResponseTimeMs  float64                `json:"response_time_ms"`
	Timestamp       string                 `json:"timestamp"`
}

type Report struct {
	ScanID          string       `json:"scan_id"`
	Timestamp       string       `json:"timestamp"`
	DurationSeconds float64      `json:"duration_seconds"`
	TotalTargets    int          `json:"total_targets"`
	Results         []ScanResult `json:"results"`
}

func loadReport(path string) (*Report, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var report Report
	if err := json.Unmarshal(b, &report); err != nil {
		return nil, err
	}
	return &report, nil
}

func countSeverity(results []ScanResult, sev string) int {
	c := 0
	for _, r := range results {
		if r.Severity == sev {
			c++
		}
	}
	return c
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

func addClient(c *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	clients[c] = true
	log.Printf("[ws] Client connected (total: %d)", len(clients))
}

func removeClient(c *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	delete(clients, c)
	c.Close()
	log.Printf("[ws] Client disconnected (total: %d)", len(clients))
}

func broadcast(msg string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for c := range clients {
		if err := c.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Println("[ws] write error:", err)
			delete(clients, c)
			c.Close()
		}
	}
}

func watchReport(path string) {
	var last time.Time
	for {
		info, err := os.Stat(path)
		if err == nil {
			if info.ModTime().After(last) {
				last = info.ModTime()
				log.Println("[watch] Report updated, broadcasting reload")
				broadcast("reload")
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func main() {

	reportPath := os.Getenv("REPORT_PATH")
	if reportPath == "" {
		reportPath = "./reports/report.json"
	}

	go watchReport(reportPath)

	r := mux.NewRouter()

	r.HandleFunc("/ws", func(w http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Println("[ws] upgrade:", err)
			return
		}
		addClient(conn)

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				removeClient(conn)
				break
			}
		}
	})

	r.HandleFunc("/api/report", func(w http.ResponseWriter, req *http.Request) {
		report, err := loadReport(reportPath)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(report)
	})

	tmpl := template.Must(template.New("dash").Funcs(template.FuncMap{
		"countSeverity": countSeverity,
	}).Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>API Security Scanner - Real-Time</title>
<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
<style>
* { margin:0; padding:0; box-sizing:border-box; }
body { background:#0d1117; color:#c9d1d9; font-family:'Segoe UI',Arial,sans-serif; padding:20px; }
h1 { color:#58a6ff; margin-bottom:10px; display:flex; align-items:center; gap:10px; }
.status-indicator { 
	width:12px; height:12px; border-radius:50%; 
	animation: pulse 2s infinite;
}
.status-live { background:#22c55e; }
.status-waiting { background:#eab308; }
@keyframes pulse {
	0%, 100% { opacity:1; }
	50% { opacity:0.5; }
}

.header-bar {
	display:flex;
	justify-content:space-between;
	align-items:center;
	margin-bottom:20px;
	padding:15px;
	background:#161b22;
	border:1px solid #30363d;
	border-radius:8px;
}

.scan-info {
	display:flex;
	gap:30px;
	font-size:14px;
}

.scan-info-item {
	display:flex;
	flex-direction:column;
}

.scan-info-label {
	color:#8b949e;
	font-size:12px;
	margin-bottom:4px;
}

.scan-info-value {
	color:#c9d1d9;
	font-weight:600;
}

.card { 
	background:#161b22; 
	border:1px solid #30363d; 
	padding:20px; 
	margin-bottom:20px; 
	border-radius:12px; 
	box-shadow:0 4px 6px rgba(0,0,0,0.3);
}

.card h2 {
	color:#c9d1d9;
	margin-bottom:15px;
	font-size:18px;
	border-bottom:1px solid #30363d;
	padding-bottom:10px;
}

.stats-grid {
	display:grid;
	grid-template-columns:repeat(auto-fit,minmax(150px,1fr));
	gap:15px;
	margin-bottom:20px;
}

.stat-box {
	background:#0d1117;
	padding:15px;
	border-radius:8px;
	border:1px solid #30363d;
	text-align:center;
}

.stat-label {
	color:#8b949e;
	font-size:12px;
	text-transform:uppercase;
	margin-bottom:8px;
}

.stat-value {
	font-size:28px;
	font-weight:bold;
}

.stat-high { color:#f85149; }
.stat-medium { color:#d29922; }
.stat-info { color:#58a6ff; }
.stat-error { color:#f85149; }

table { 
	width:100%; 
	border-collapse:collapse; 
	background:#161b22; 
	font-size:14px;
}
th { 
	background:#21262d; 
	padding:12px; 
	text-align:left; 
	font-weight:600;
	color:#c9d1d9;
	border-bottom:2px solid #30363d;
}
td { 
	padding:12px; 
	border-top:1px solid #30363d; 
	vertical-align:top;
}
pre { 
	background:#0d1117; 
	padding:8px; 
	border:1px solid #30363d; 
	border-radius:6px; 
	font-size:12px;
	overflow-x:auto;
}

.sev-high   { background:rgba(248,81,73,0.15); border-left:3px solid #f85149; }
.sev-medium { background:rgba(210,153,34,0.15); border-left:3px solid #d29922; }
.sev-info   { background:rgba(88,166,255,0.15); border-left:3px solid #58a6ff; }
.sev-error  { background:rgba(248,81,73,0.15); border-left:3px solid #f85149; }

.badge {
	display:inline-block;
	padding:4px 8px;
	border-radius:4px;
	font-size:11px;
	font-weight:600;
	text-transform:uppercase;
}

.badge-high { background:#f85149; color:#fff; }
.badge-medium { background:#d29922; color:#000; }
.badge-info { background:#58a6ff; color:#000; }
.badge-error { background:#f85149; color:#fff; }

ul { 
	list-style:none; 
	padding-left:0;
}

ul li {
	padding:4px 0;
	padding-left:16px;
	position:relative;
}

ul li:before {
	content:"▸";
	position:absolute;
	left:0;
	color:#8b949e;
}

.no-issues {
	color:#3fb950;
	font-weight:600;
}

.response-time {
	color:#8b949e;
	font-size:12px;
}

.charts-row {
	display:grid;
	grid-template-columns:1fr 2fr;
	gap:20px;
	margin-bottom:20px;
}

@media (max-width: 768px) {
	.charts-row {
		grid-template-columns:1fr;
	}
}

.empty-state {
	text-align:center;
	padding:60px 20px;
	color:#8b949e;
}

.empty-state h2 {
	color:#c9d1d9;
	margin-bottom:15px;
	border:none;
}

.spinner {
	width:50px;
	height:50px;
	margin:0 auto 20px;
	border:4px solid #30363d;
	border-top-color:#58a6ff;
	border-radius:50%;
	animation:spin 1s linear infinite;
}

@keyframes spin {
	to { transform:rotate(360deg); }
}
</style>
</head>
<body>

<div class="header-bar">
	<div>
		<h1>
			<span class="status-indicator {{ if .Report }}status-live{{ else }}status-waiting{{ end }}"></span>
			API Security Scanner
		</h1>
	</div>
	<div class="scan-info">
		{{ if .Report }}
		<div class="scan-info-item">
			<div class="scan-info-label">Scan ID</div>
			<div class="scan-info-value">{{ .Report.ScanID }}</div>
		</div>
		<div class="scan-info-item">
			<div class="scan-info-label">Last Update</div>
			<div class="scan-info-value">{{ .Report.Timestamp }}</div>
		</div>
		<div class="scan-info-item">
			<div class="scan-info-label">Duration</div>
			<div class="scan-info-value">{{ printf "%.2f" .Report.DurationSeconds }}s</div>
		</div>
		{{ end }}
	</div>
</div>

{{ if .Report }}
{{ $results := .Report.Results }}
{{ $total := len $results }}

<div class="stats-grid">
	<div class="stat-box">
		<div class="stat-label">Total Scanned</div>
		<div class="stat-value">{{ $total }}</div>
	</div>
	<div class="stat-box">
		<div class="stat-label">High Risk</div>
		<div class="stat-value stat-high">{{ countSeverity $results "high" }}</div>
	</div>
	<div class="stat-box">
		<div class="stat-label">Medium Risk</div>
		<div class="stat-value stat-medium">{{ countSeverity $results "medium" }}</div>
	</div>
	<div class="stat-box">
		<div class="stat-label">Healthy</div>
		<div class="stat-value stat-info">{{ countSeverity $results "info" }}</div>
	</div>
	<div class="stat-box">
		<div class="stat-label">Errors</div>
		<div class="stat-value stat-error">{{ countSeverity $results "error" }}</div>
	</div>
</div>

<div class="charts-row">
	<div class="card">
		<h2>Severity Distribution</h2>
		<canvas id="sevChart" height="200"></canvas>
	</div>

	<div class="card">
		<h2>Issues Per Endpoint</h2>
		<canvas id="issueChart" height="200"></canvas>
	</div>
</div>

<div class="card">
	<h2>Scan Results</h2>
	<table>
		<thead>
			<tr>
				<th>Endpoint</th>
				<th>Status</th>
				<th>Severity</th>
				<th>Issues</th>
				<th>Security Headers</th>
			</tr>
		</thead>
		<tbody>
		{{ range $results }}
			<tr class="sev-{{ .Severity }}">
				<td>
					<strong>{{ .URL }}</strong>
					{{ if .ResponseTimeMs }}
					<div class="response-time">⚡ {{ printf "%.0f" .ResponseTimeMs }}ms</div>
					{{ end }}
				</td>
				<td>
					{{ if .Status }}
						{{ .Status }}
					{{ else }}
						<span style="color:#8b949e;">N/A</span>
					{{ end }}
				</td>
				<td>
					<span class="badge badge-{{ .Severity }}">{{ .Severity }}</span>
				</td>
				<td>
					{{ if .Issues }}
						<ul>
						{{ range .Issues }}
							<li>{{ . }}</li>
						{{ end }}
						</ul>
					{{ else }}
						<span class="no-issues">✓ No issues detected</span>
					{{ end }}
				</td>
				<td>
					<pre>{{ range $k,$v := .SecurityHeaders }}{{ $k }}: {{ if $v }}{{ $v }}{{ else }}<span style="color:#8b949e;">missing</span>{{ end }}
{{ end }}</pre>
				</td>
			</tr>
		{{ end }}
		</tbody>
	</table>
</div>

<script>
const sevData = {
	labels:["High","Medium","Info","Error"],
	datasets:[{
		data:[
			{{ countSeverity $results "high" }},
			{{ countSeverity $results "medium" }},
			{{ countSeverity $results "info" }},
			{{ countSeverity $results "error" }}
		],
		backgroundColor:["#f85149","#d29922","#58a6ff","#f85149"]
	}]
};

new Chart(document.getElementById("sevChart"), {
	type:"doughnut",
	data:sevData,
	options:{
		plugins:{
			legend:{ labels:{ color:"#c9d1d9" } }
		}
	}
});

new Chart(document.getElementById("issueChart"), {
	type:"bar",
	data:{
		labels:[ {{ range $results }}"{{ .URL }}",{{ end }} ],
		datasets:[{
			label:"Issues Found",
			data:[ {{ range $results }}{{ len .Issues }},{{ end }} ],
			backgroundColor:"#58a6ff"
		}]
	},
	options:{ 
		scales:{ 
			y:{ beginAtZero:true, ticks:{color:"#c9d1d9"}, grid:{color:"#30363d"} },
			x:{ ticks:{color:"#c9d1d9"}, grid:{color:"#30363d"} }
		},
		plugins:{
			legend:{ labels:{ color:"#c9d1d9" } }
		}
	}
});

let ws = new WebSocket((location.protocol==="https:"?"wss://":"ws://")+location.host+"/ws");
ws.onmessage = (msg) => { 
	if(msg.data==="reload"){ 
		console.log("New scan detected, reloading...");
		location.reload(); 
	} 
};
ws.onclose = () => {
	console.log("WebSocket disconnected, attempting reconnect...");
	setTimeout(() => location.reload(), 3000);
};
</script>

{{ else }}

<div class="empty-state">
	<div class="spinner"></div>
	<h2>Waiting for first scan...</h2>
	<p>The scanner is starting up. Results will appear here automatically.</p>
</div>

<script>
let ws = new WebSocket((location.protocol==="https:"?"wss://":"ws://")+location.host+"/ws");
ws.onmessage = () => location.reload();
</script>

{{ end }}

</body>
</html>
`))

	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		report, err := loadReport(reportPath)
		if err != nil {
			report = nil
		}
		tmpl.Execute(w, map[string]any{
			"Report": report,
		})
	})

	log.Println("🚀 Dashboard running on :8081")
	log.Println("📊 Real-time updates enabled via WebSocket")
	log.Println("📁 Watching:", reportPath)
	http.ListenAndServe(":8081", r)
}