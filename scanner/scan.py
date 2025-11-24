import asyncio
import json
import os
from typing import List, Dict, Any
from datetime import datetime
import aiohttp
from yarl import URL
import sys
import time

# You can change this list or load from env/JSON file
DEFAULT_TARGETS = [
    "https://example.com",
    "https://httpbin.org/get",
]

CONCURRENCY = int(os.getenv("SCANNER_CONCURRENCY", "5"))
REQUEST_TIMEOUT = int(os.getenv("SCANNER_TIMEOUT", "10"))

AUTH_HEADER = os.getenv("SCANNER_AUTH_HEADER")  # e.g. "Bearer abc123"


def build_headers() -> Dict[str, str]:
    headers = {
        "User-Agent": "ApiSecScanner/1.0",
        "Accept": "*/*",
    }
    if AUTH_HEADER:
        headers["Authorization"] = AUTH_HEADER
    return headers


def analyze_result(url: str, status: int, headers: Dict[str, str]) -> Dict[str, Any]:
    issues = []

    parsed = URL(url)
    if parsed.scheme != "https":
        issues.append("Connection is not HTTPS.")

    # Standard headers
    checks = {
        "X-Frame-Options": "Missing X-Frame-Options header.",
        "Content-Security-Policy": "Missing Content-Security-Policy header.",
        "X-Content-Type-Options": "Missing X-Content-Type-Options header.",
        "Strict-Transport-Security": "Missing HSTS header.",
    }

    for h, msg in checks.items():
        if h not in headers:
            issues.append(msg)

    # CORS
    if headers.get("Access-Control-Allow-Origin") == "*":
        issues.append("CORS allows any origin (*) — risky.")

    # Missing content-type on JSON endpoints
    if "/api" in url or "/v1" in url:
        ct = headers.get("Content-Type", "")
        if "json" not in ct.lower():
            issues.append("API missing JSON Content-Type header.")

    # Cache-Control
    if "Cache-Control" not in headers:
        issues.append("Missing Cache-Control header (security/caching risk).")

    # Server fingerprinting
    for fp in ["Server", "X-Powered-By"]:
        if fp in headers:
            issues.append(f"Leaking server fingerprint via {fp} header.")

    # Status checks
    if status >= 500:
        issues.append(f"Server error {status}.")
    elif status >= 400:
        issues.append(f"Client error {status}.")

    # Severity logic
    if any("500" in i or "not HTTPS" in i for i in issues):
        severity = "high"
    elif issues:
        severity = "medium"
    else:
        severity = "info"

    return {
        "url": url,
        "status": status,
        "issues": issues,
        "severity": severity,
    }


async def scan_one(session: aiohttp.ClientSession, url: str, semaphore: asyncio.Semaphore):
    async with semaphore:
        start_time = time.time()
        try:
            async with session.get(url, timeout=aiohttp.ClientTimeout(total=REQUEST_TIMEOUT)) as resp:
                response_time = round((time.time() - start_time) * 1000, 2)  # ms
                headers = dict(resp.headers)
                analysis = analyze_result(url, resp.status, headers)
                analysis["security_headers"] = {
                    "X-Frame-Options": headers.get("X-Frame-Options"),
                    "Content-Security-Policy": headers.get("Content-Security-Policy"),
                    "X-Content-Type-Options": headers.get("X-Content-Type-Options"),
                    "Strict-Transport-Security": headers.get("Strict-Transport-Security"),
                }
                analysis["response_time_ms"] = response_time
                analysis["timestamp"] = datetime.utcnow().isoformat() + "Z"
                
                print(f"✓ {url} [{resp.status}] {response_time}ms", flush=True)
                return analysis
        except Exception as e:
            response_time = round((time.time() - start_time) * 1000, 2)
            print(f"✗ {url} FAILED: {str(e)}", flush=True)
            return {
                "url": url,
                "status": 0,
                "security_headers": {},
                "error": str(e),
                "severity": "error",
                "issues": [f"Request failed: {e}"],
                "response_time_ms": response_time,
                "timestamp": datetime.utcnow().isoformat() + "Z",
            }


def load_targets() -> List[str]:
    """
    Load targets from env or fall back to DEFAULT_TARGETS.
    You can also add a 'targets.json' later.
    """
    env_value = os.getenv("SCANNER_TARGETS")
    if env_value:
        return [t.strip() for t in env_value.split(",") if t.strip()]
    return DEFAULT_TARGETS


async def main():
    targets = load_targets()
    scan_id = datetime.utcnow().strftime("%Y%m%d_%H%M%S")
    
    print(f"\n{'='*60}", flush=True)
    print(f"🔍 SCAN #{scan_id}", flush=True)
    print(f"{'='*60}", flush=True)
    print(f"Targets: {len(targets)}", flush=True)
    for i, t in enumerate(targets, 1):
        print(f"  {i}. {t}", flush=True)
    print(flush=True)

    semaphore = asyncio.Semaphore(CONCURRENCY)
    start = time.time()
    
    async with aiohttp.ClientSession(headers=build_headers()) as session:
        tasks = [scan_one(session, url, semaphore) for url in targets]
        results = await asyncio.gather(*tasks)

    duration = round(time.time() - start, 2)
    
    # Add metadata
    report = {
        "scan_id": scan_id,
        "timestamp": datetime.utcnow().isoformat() + "Z",
        "duration_seconds": duration,
        "total_targets": len(targets),
        "results": results
    }

    # Save to shared volume path
    out_path = os.getenv("REPORT_PATH", "./reports/report.json")
    os.makedirs(os.path.dirname(out_path), exist_ok=True)
    with open(out_path, "w", encoding="utf-8") as f:
        json.dump(report, f, indent=2)

    # Summary
    high = sum(1 for r in results if r.get("severity") == "high")
    medium = sum(1 for r in results if r.get("severity") == "medium")
    info_count = sum(1 for r in results if r.get("severity") == "info")
    errors = sum(1 for r in results if r.get("severity") == "error")
    
    print(f"\n{'='*60}", flush=True)
    print(f"📊 SUMMARY", flush=True)
    print(f"{'='*60}", flush=True)
    print(f"Duration: {duration}s", flush=True)
    print(f"🔴 High:   {high}", flush=True)
    print(f"🟡 Medium: {medium}", flush=True)
    print(f"🔵 Info:   {info_count}", flush=True)
    print(f"⚠️  Errors: {errors}", flush=True)
    print(f"{'='*60}\n", flush=True)


if sys.platform.startswith("win"):
    asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy())

SCAN_INTERVAL = int(os.getenv("SCAN_INTERVAL", "30"))   # seconds

if __name__ == "__main__":
    print("\n🚀 API Security Scanner - Real-Time Mode", flush=True)
    print(f"⏱️  Scan interval: {SCAN_INTERVAL} seconds\n", flush=True)
    
    while True:
        try:
            asyncio.run(main())
            print(f"⏸️  Sleeping {SCAN_INTERVAL} seconds...\n", flush=True)
            time.sleep(SCAN_INTERVAL)
        except KeyboardInterrupt:
            print("\n🛑 Scanner stopped by user", flush=True)
            break
        except Exception as e:
            print(f"❌ Error: {e}", flush=True)
            time.sleep(5)