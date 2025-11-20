# VaFund Smart Contract

Smart contract untuk aplikasi VaFund menggunakan Hyperledger Fabric dengan arsitektur yang terorganisir.

## 🏗️ Struktur Folder

```
vafund-smartcontract/
├── chaincode/
│   ├── entities/           # Domain entities/models
│   │   ├── donation.go     # Donation entity
│   │   └── event.go        # Event entity
│   ├── services/           # Business logic layer
│   │   ├── donation_service.go  # Donation operations
│   │   └── event_service.go     # Event operations
│   ├── main.go             # Entry point
│   ├── vafund.go           # Smart contract interface
│   └── go.mod              # Dependencies
├── test/
│   └── simulation.go       # Testing tanpa blockchain
└── scripts/
    └── restart-fabric.sh   # Fabric network setup
```

## 📋 Entities

### Donation
- **ID**: Unique identifier untuk donasi
- **SenderName**: Nama pengirim donasi
- **Amount**: Jumlah donasi (float64)
- **Message**: Pesan optional dari pengirim
- **Timestamp**: Waktu transaksi dari blockchain
- **TxID**: Transaction ID dari Hyperledger Fabric

### Event
- **Code**: Unique identifier untuk event
- **Name**: Nama event
- **StartDate**: Tanggal mulai event
- **EndDate**: Tanggal berakhir event
- **IsActive**: Status aktif event
- **Timestamp**: Waktu pembuatan event
- **TxID**: Transaction ID dari Hyperledger Fabric

## 🔧 Services

### DonationService
- `Create()` - Membuat donasi baru
- `GetByID()` - Mengambil donasi berdasarkan ID
- `GetAll()` - Mengambil semua donasi
- `GetByMinAmount()` - Filter donasi berdasarkan jumlah minimum
- `GetTotal()` - Menghitung total donasi
- `Exists()` - Mengecek keberadaan donasi

### EventService
- `Create()` - Membuat event baru
- `GetByCode()` - Mengambil event berdasarkan code
- `GetAll()` - Mengambil semua event
- `GetActive()` - Mengambil event yang sedang aktif
- `Exists()` - Mengecek keberadaan event
- `UpdateStatus()` - Update status aktif event

## 🚀 Smart Contract Functions

### Donation Functions
- `MakeDonation(donationID, senderName, amount, message)`
- `ReadDonation(id)`
- `GetAllDonations()`
- `GetDonationsByAmount(minAmount)`
- `GetTotalDonations()`
- `DonationExists(id)`

### Event Functions
- `CreateEvent(code, name, startDate, endDate, isActive)`
- `ReadEvent(code)`
- `GetAllEvents()`
- `GetActiveEvents()`
- `EventExists(code)`
- `UpdateEventStatus(code, isActive)`

## 🧪 Testing

### Build Smart Contract
```bash
cd chaincode
go build .
```

### Run Simulation Test
```bash
cd test
go run simulation.go
```

### Deploy ke Fabric Network
```bash
cd scripts
./restart-fabric.sh
```

## 📝 Format Data

### Create Donation
```json
{
  "donationID": "donation123",
  "senderName": "John Doe", 
  "amount": "100000",
  "message": "Semoga bermanfaat"
}
```

### Create Event
```json
{
  "code": "RAMADAN2025",
  "name": "Ramadan Charity Drive 2025",
  "startDate": "2025-03-01T00:00:00Z",
  "endDate": "2025-04-30T23:59:59Z", 
  "isActive": "true"
}
```

## ✨ Keunggulan Arsitektur

1. **Separation of Concerns**: Entity, Service, dan Contract layer terpisah
2. **Maintainability**: Kode mudah dipelihara dan diupdate
3. **Testability**: Setiap layer dapat ditest secara terpisah
4. **Scalability**: Mudah menambah entity atau service baru
5. **Clean Code**: Mengikuti best practices Go dan Fabric

## 🔄 Workflow Development

1. **Entities**: Definisikan model data di `entities/`
2. **Services**: Implementasi business logic di `services/`
3. **Contract**: Expose functions di `vafund.go`
4. **Testing**: Test dengan `simulation.go`
5. **Deploy**: Deploy ke Fabric network

## 📖 Best Practices

- Gunakan pointer untuk struct yang besar
- Implementasi validation di entity level
- Error handling yang comprehensive
- JSON marshaling yang konsisten
- Prefix untuk key separation (EVENT_ untuk events)