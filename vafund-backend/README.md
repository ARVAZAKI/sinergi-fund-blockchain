# VaFund Backend API

Backend API untuk aplikasi VaFund menggunakan Go Fiber dan Hyperledger Fabric.

## 🚀 Fitur

- **MVC Architecture**: Struktur kode yang terorganisir dengan Models, Views (JSON responses), dan Controllers
- **REST API**: Endpoint untuk mengelola donasi
- **Hyperledger Fabric Integration**: Integrasi dengan smart contract VaFund
- **CORS Support**: Mendukung cross-origin requests

## 📋 API Endpoints

### Health Check
- **GET** `/api/health` - Cek status server

### Donation Management
- **POST** `/api/donations` - Buat donasi baru
- **GET** `/api/donations` - Ambil semua donasi
- **GET** `/api/donations/:id` - Ambil donasi berdasarkan ID
- **GET** `/api/donations/total` - Ambil total donasi

## 📝 Request/Response Format

### Create Donation
**POST** `/api/donations`

Request Body:
```json
{
  "donationId": "donation123",
  "senderName": "John Doe",
  "amount": "100000",
  "message": "Semoga bermanfaat"
}
```

Response:
```json
{
  "success": true,
  "message": "Donation created successfully",
  "data": {
    "id": "donation123",
    "senderName": "John Doe",
    "amount": 100000,
    "message": "Semoga bermanfaat",
    "timestamp": "2025-11-10T10:30:00Z",
    "txId": "tx123abc"
  }
}
```

### Get Total Donations
**GET** `/api/donations/total`

Response:
```json
{
  "success": true,
  "message": "Total donations retrieved successfully",
  "totalAmount": 350000,
  "totalCount": 3
}
```

## 🛠️ Setup dan Instalasi

### Prerequisites
- Go 1.20 atau lebih tinggi

### Instalasi

1. **Masuk ke folder backend:**
   ```bash
   cd vafund-backend
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Jalankan server:**
   ```bash
   go run main.go
   ```

   Atau build terlebih dahulu:
   ```bash
   go build -o vafund-backend
   ./vafund-backend
   ```

Server akan berjalan di `http://localhost:3000`

## 🧪 Testing

### Test dengan curl

1. **Health check:**
   ```bash
   curl http://localhost:3000/api/health
   ```

2. **Create donation:**
   ```bash
   curl -X POST http://localhost:3000/api/donations \
     -H "Content-Type: application/json" \
     -d '{
       "donationId": "donation001",
       "senderName": "John Doe",
       "amount": "100000",
       "message": "Test donation"
     }'
   ```

3. **Get total donations:**
   ```bash
   curl http://localhost:3000/api/donations/total
   ```

## 📁 Struktur Folder

```
vafund-backend/
├── config/          # Konfigurasi aplikasi
├── controllers/     # HTTP request handlers
├── models/          # Data models dan structs
├── routes/          # Route definitions
├── services/        # Business logic dan Fabric integration
├── main.go          # Entry point aplikasi
├── go.mod           # Go module dependencies
└── README.md        # Dokumentasi
```