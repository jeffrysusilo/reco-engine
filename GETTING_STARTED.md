# Getting Started - First Time Setup

Selamat datang di Recommendation Engine! Panduan ini akan membantu Anda menjalankan sistem untuk pertama kali.

## Prasyarat

Pastikan Anda sudah menginstall:

1. **Docker Desktop** (wajib)
   - Windows: https://www.docker.com/products/docker-desktop
   - Pastikan Docker Desktop sudah berjalan

2. **Go 1.21+** (opsional, untuk development)
   - Download: https://go.dev/dl/

3. **Git** (untuk clone repository)

## Langkah 1: Setup Project

### Windows (PowerShell)

```powershell
# Clone repository (jika dari Git)
# git clone https://github.com/yourusername/reco-engine.git
# cd reco-engine

# Atau jika sudah ada folder, masuk ke folder
cd d:\projek\reco-engine

# Jalankan setup otomatis
.\setup.ps1
```

### Linux/Mac (Bash)

```bash
# Clone repository (jika dari Git)
# git clone https://github.com/yourusername/reco-engine.git
# cd reco-engine

# Beri izin eksekusi
chmod +x setup.sh
chmod +x scripts/generate_events.sh

# Jalankan setup
./setup.sh
```

## Langkah 2: Verifikasi Services

Setelah setup selesai, cek apakah semua services berjalan:

```bash
docker-compose ps
```

Anda akan melihat output seperti ini:

```
NAME                   STATUS    PORTS
reco-postgres          Up        0.0.0.0:5432->5432/tcp
reco-redis             Up        0.0.0.0:6379->6379/tcp
reco-kafka             Up        0.0.0.0:9092->9092/tcp
reco-ingest            Up        0.0.0.0:8080->8080/tcp
reco-processor         Up
reco-api               Up        0.0.0.0:8081->8081/tcp
reco-prometheus        Up        0.0.0.0:9090->9090/tcp
reco-grafana           Up        0.0.0.0:3000->3000/tcp
```

## Langkah 3: Test API

### Test Event Ingest

**Windows PowerShell:**
```powershell
$body = @{
    user_id = 1
    item_id = 5
    event_type = "VIEW"
    session_id = "test_session"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/events" -Method Post -Body $body -ContentType "application/json"
```

**Bash/Curl:**
```bash
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "item_id": 5,
    "event_type": "VIEW",
    "session_id": "test_session"
  }'
```

Response yang diharapkan:
```json
{"status":"ok"}
```

### Test Recommendation API

```bash
curl "http://localhost:8081/recommendations?user_id=1&count=10"
```

Atau buka di browser: http://localhost:8081/recommendations?user_id=1&count=10

### Test Popular Items

```bash
curl "http://localhost:8081/popular?count=20"
```

Atau buka di browser: http://localhost:8081/popular?count=20

## Langkah 4: Generate Sample Data

Untuk mendapatkan rekomendasi yang bermakna, Anda perlu data events:

**Windows:**
```powershell
.\scripts\generate_events.ps1 -NumEvents 1000
```

**Linux/Mac:**
```bash
./scripts/generate_events.sh http://localhost:8080 1000
```

Script ini akan generate 1000 events random dengan variasi:
- User IDs: 1-100
- Item IDs: 1-50
- Event types: VIEW, CLICK, CART, PURCHASE
- Sessions: 50 unique sessions

## Langkah 5: Monitor System

### Prometheus Metrics

1. Buka browser: http://localhost:9090
2. Coba query berikut:
   ```
   rate(events_ingested_total[1m])
   ```
   ```
   histogram_quantile(0.95, rate(recommendation_latency_seconds_bucket[5m]))
   ```

### Grafana Dashboard

1. Buka browser: http://localhost:3000
2. Login dengan:
   - Username: `admin`
   - Password: `admin`
3. Skip jika diminta ganti password (atau ganti sesuai keinginan)
4. Add data source:
   - Pilih Prometheus
   - URL: `http://prometheus:9090`
   - Click "Save & Test"

### Check Logs

**Semua services:**
```bash
docker-compose logs -f
```

**Service tertentu:**
```bash
docker-compose logs -f ingest
docker-compose logs -f processor
docker-compose logs -f api
```

### Check Redis Data

```bash
# Masuk ke Redis CLI
docker exec -it reco-redis redis-cli

# Lihat semua keys
KEYS *

# Lihat user recent items
LRANGE user:recent:1 0 -1

# Lihat popular items
ZREVRANGE item:popularity 0 10 WITHSCORES

# Lihat co-view untuk item 1
ZREVRANGE co_view:1 0 10 WITHSCORES

# Exit
exit
```

### Check PostgreSQL Data

```bash
# Masuk ke PostgreSQL
docker exec -it reco-postgres psql -U reco -d reco

# Lihat semua tables
\dt

# Count events
SELECT COUNT(*) FROM events;

# Lihat 10 events terakhir
SELECT * FROM events ORDER BY timestamp DESC LIMIT 10;

# Lihat items
SELECT * FROM items;

# Exit
\q
```

## Langkah 6: Testing dengan Browser/Postman

### Postman Collection

Buat collection dengan requests berikut:

**1. Health Check - Ingest**
- Method: GET
- URL: `http://localhost:8080/health`

**2. Health Check - API**
- Method: GET
- URL: `http://localhost:8081/health`

**3. Ingest Event**
- Method: POST
- URL: `http://localhost:8080/events`
- Headers: `Content-Type: application/json`
- Body (raw JSON):
  ```json
  {
    "user_id": 1,
    "item_id": 5,
    "event_type": "VIEW",
    "session_id": "session_1"
  }
  ```

**4. Get Recommendations**
- Method: GET
- URL: `http://localhost:8081/recommendations?user_id=1&count=10`

**5. Get Popular Items**
- Method: GET
- URL: `http://localhost:8081/popular?category=electronics&count=20`

## Troubleshooting

### Problem: Services tidak start

**Solution:**
```bash
# Stop semua
docker-compose down

# Hapus volumes (data akan hilang!)
docker-compose down -v

# Start lagi
docker-compose up -d
```

### Problem: Port sudah dipakai

**Solution:**
Edit `docker-compose.yml` dan ganti port yang conflict:
```yaml
ports:
  - "8082:8080"  # Ganti 8080 ke 8082
```

### Problem: Kafka error "broker not available"

**Solution:**
```bash
# Tunggu beberapa saat, Kafka butuh waktu start
# Cek logs
docker-compose logs kafka

# Restart Kafka
docker-compose restart kafka
```

### Problem: Recommendation kosong

**Solution:**
- Pastikan sudah generate events
- Tunggu beberapa detik untuk processor memproses
- Cek logs processor: `docker-compose logs processor`

### Problem: Docker out of memory

**Solution:**
- Buka Docker Desktop > Settings > Resources
- Increase memory limit (minimal 4GB recommended)

## Next Steps

Setelah sistem berjalan dengan baik:

1. **Pelajari API**: Baca `docs/API.md`
2. **Pahami Architecture**: Baca `docs/ARCHITECTURE.md`
3. **Load Testing**: 
   ```bash
   # Install k6 (https://k6.io/docs/getting-started/installation/)
   k6 run scripts/load_test_k6.js
   ```
4. **Development**: Baca `docs/QUICKSTART.md` untuk local development
5. **Production Deploy**: Baca `docs/DEPLOYMENT.md`

## Cleanup

Jika ingin stop dan hapus semua:

```bash
# Stop services
docker-compose down

# Hapus termasuk data (HATI-HATI!)
docker-compose down -v

# Hapus images
docker-compose down --rmi all
```

## Support

Jika ada masalah:
1. Cek logs: `docker-compose logs -f`
2. Cek issues di GitHub
3. Buat issue baru dengan detail error

## Frequently Asked Questions

**Q: Berapa lama waktu yang dibutuhkan untuk start pertama kali?**  
A: Sekitar 1-2 menit untuk download images dan start semua services.

**Q: Berapa resource yang dibutuhkan?**  
A: Minimal 4GB RAM, 10GB disk space.

**Q: Apakah data akan hilang jika restart?**  
A: Tidak, data disimpan di Docker volumes. Gunakan `docker-compose down -v` jika ingin hapus data.

**Q: Bagaimana cara update code?**  
A: Rebuild images dengan `docker-compose build` lalu `docker-compose up -d`.

**Q: Bisa dijalankan di production?**  
A: Bisa, tapi perlu konfigurasi tambahan (lihat `docs/DEPLOYMENT.md`).

---

Selamat mencoba! 🚀
