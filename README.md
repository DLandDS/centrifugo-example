# Centrifugo Messaging Application

Aplikasi messaging real-time menggunakan Centrifugo dengan backend Go dan frontend Svelte. Aplikasi ini mendukung messaging berdasarkan topic/channel.

## Arsitektur

- **Backend**: Go server dengan API HTTP untuk mengirim pesan
- **Frontend**: Aplikasi Svelte dengan koneksi WebSocket real-time ke Centrifugo  
- **Messaging**: Sistem messaging berbasis topic menggunakan channel Centrifugo
- **Deployment**: Docker Compose untuk setup yang mudah

## Fitur

- ✅ Real-time messaging menggunakan WebSocket
- ✅ Multiple topic/channel (general, tech, random, announcements)
- ✅ UI yang responsive dan user-friendly
- ✅ Auto-scroll untuk pesan baru
- ✅ Indikator status koneksi
- ✅ Highlight pesan dari user sendiri
- ✅ Timestamp untuk setiap pesan
- ✅ CORS support untuk development

## Prerequisites

- Docker dan Docker Compose
- Node.js 18+ (untuk development)
- Go 1.21+ (untuk development)

## Quick Start

### 1. Clone Repository

```bash
git clone <repository-url>
cd centrifugo-example
```

### 2. Start dengan Docker Compose

```bash
cd docker
docker-compose up -d
```

Ini akan menjalankan:
- Centrifugo server di port 8000
- Go backend di port 8080

### 3. Setup Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend akan berjalan di http://localhost:5173

## Development Setup

### Backend Development

```bash
cd backend
go run main.go
```

Backend API akan tersedia di:
- `GET /health` - Health check
- `GET /api/topics` - Daftar topic yang tersedia
- `POST /api/messages` - Kirim pesan ke topic

### Frontend Development

```bash
cd frontend
npm install
npm run dev
```

### Centrifugo Development

Jalankan Centrifugo standalone:

```bash
docker run --rm -p 8000:8000 -v $PWD/docker/config.json:/centrifugo/config.json centrifugo/centrifugo:v5 centrifugo --config=config.json
```

## Cara Menggunakan

1. **Buka aplikasi** di browser: http://localhost:5173
2. **Ganti username** sesuai keinginan (default: User{random})
3. **Pilih topic** dengan klik tombol topic (general, tech, random, announcements)
4. **Kirim pesan** dengan mengetik di input field dan tekan Enter atau klik Send
5. **Lihat pesan real-time** dari user lain yang bergabung di topic yang sama

## API Documentation

### POST /api/messages

Kirim pesan ke topic tertentu.

**Request Body:**
```json
{
  "topic": "general",
  "content": "Hello, world!",
  "author": "username"
}
```

**Response:**
```json
{
  "success": true,
  "message": {
    "id": "1234567890123456789",
    "topic": "general", 
    "content": "Hello, world!",
    "author": "username",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### GET /api/topics

Mendapatkan daftar topic yang tersedia.

**Response:**
```json
{
  "topics": ["general", "tech", "random", "announcements"]
}
```

## Konfigurasi

### Centrifugo Configuration (docker/config.json)

```json
{
  "token_hmac_secret_key": "token_hmac_secret_key",
  "admin_password": "password", 
  "admin_secret": "admin_secret",
  "api_key": "api_key",
  "allowed_origins": ["http://localhost:5173", "http://localhost:3000"],
  "allow_anonymous": true,
  "log_level": "info",
  "health": true
}
```

### Environment Variables

Backend mendukung environment variables berikut:

- `CENTRIFUGO_URL` - URL Centrifugo server (default: http://localhost:8000)
- `CENTRIFUGO_API_KEY` - API key untuk Centrifugo (default: api_key)
- `PORT` - Port untuk backend server (default: 8080)

## Struktur Project

```
centrifugo-example/
├── backend/
│   ├── main.go          # Go backend server
│   ├── go.mod           # Go dependencies  
│   ├── go.sum
│   └── Dockerfile       # Docker image untuk backend
├── frontend/
│   ├── src/
│   │   └── routes/
│   │       └── +page.svelte  # Main messaging UI
│   ├── package.json     # Node.js dependencies
│   └── ...             # Svelte project files
├── docker/
│   ├── docker-compose.yml  # Docker Compose configuration
│   └── config.json      # Centrifugo configuration
└── README.md
```

## Troubleshooting

### Backend tidak bisa connect ke Centrifugo

Pastikan Centrifugo berjalan dan API key sesuai:
```bash
curl http://localhost:8000/health
```

### Frontend tidak bisa connect ke backend

Pastikan backend berjalan dan CORS dikonfigurasi dengan benar:
```bash
curl http://localhost:8080/health
```

### Pesan tidak muncul real-time

Periksa koneksi WebSocket ke Centrifugo:
```bash
# Periksa log browser console untuk error
# Pastikan Centrifugo config allow_anonymous: true
```

## Development Tips

- Gunakan browser developer tools untuk debug WebSocket connection
- Centrifugo admin panel tersedia di http://localhost:8000 (password: password)
- Backend logs akan menampilkan error detail jika ada masalah dengan Centrifugo API
- Frontend akan menampilkan fallback topics jika backend tidak tersedia

## Contributing

1. Fork repository
2. Buat feature branch
3. Commit perubahan
4. Push ke branch
5. Buat Pull Request

## License

MIT License