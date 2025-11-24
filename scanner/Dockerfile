FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY scan.py .

# reports will live in /reports (mounted volume)
ENV REPORT_PATH=/reports/report.json

# Unbuffered Python output for real-time logs
ENV PYTHONUNBUFFERED=1

CMD ["python", "-u", "scan.py"]