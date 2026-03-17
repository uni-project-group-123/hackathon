# Sentinel - Campus Safety & Security System

## Overview

Sentinel is a multi-factor authentication system for campus security that combines:
- **NFC Card Authentication** - Primary identification
- **Facial Recognition** - Biometric verification
- **Real-time Monitoring** - Live occupancy tracking
- **Admin Dashboard** - Centralized security management

## Problem Being Solved

Currently, campus labs rely on ID cards for access, which can be shared or stolen. There's no real-time tracking of building occupancy and security staff lack visibility into access patterns.

## Project Structure

```
sentinel/
├── backend-go/                  # Go API server
│   ├── main.go                 # Main application
│   ├── go.mod                  # Go module
│   ├── go.sum                  # Dependencies
│   ├── sentinel                # Compiled Linux binary
│   ├── sentinel.exe            # Compiled Windows binary
│   ├── sentinel.db             # SQLite database
│   ├── build.sh                # Build script
│   ├── build.bat               # Windows build script
│   └── access-controller/      # Access controller
│       ├── main.go            # Controller code
│       └── config.json        # Configuration
├── frontend/                   # Web frontend (server-side rendered)
│   └── templates/            # HTML templates
│       ├── index.html        # Admin dashboard
│       └── login.html        # Login page
├── face-check/                # Face recognition module
│   ├── main.py               # Face verification script
│   ├── test.py               # Test script
│   ├── requirements.txt      # Python dependencies
│   ├── models/               # ONNX models
│   │   ├── face_detection_yunet_2023mar.onnx
│   │   └── face_recognition_sface_2021dec.onnx
│   └── users/                # Registered face images
├── rfid-operation/            # Arduino sketch
│   └── main.ino              # Arduino NFC reader code
├── .github/workflows/         # CI/CD
│   └── build.yml             # Build workflow
├── brief.md                   # Project brief
├── plan.md                   # Development plan
├── presentation.md           # Presentation slides
├── documentation.md          # Full documentation
└── README.md                 # This file
```

## Quick Start

### 1. Run the Server

```bash
cd backend-go
go run main.go
```

Or run the compiled binary:

```bash
# Linux
./backend-go/sentinel

# Windows
.\backend-go\sentinel.exe
```

The server will:
- Initialize the SQLite database (`sentinel.db`)
- Create default admin user
- Seed default zones and sample users
- Start on `http://0.0.0.0:5000`

### 2. Access Admin Panel

Open browser: `http://localhost:5000`

Default credentials:
- Username: `admin`
- Password: `admin123`

### 3. Default Data

The server creates sample users:

| User ID | Name | Role | Card ID | Can Access |
|---------|------|------|---------|------------|
| STU001 | John Doe | student | STU001 | Computer Lab, Library |
| STU002 | Jane Smith | student | STU002 | Computer Lab, Library |
| FAC001 | Dr. Emily Brown | faculty | FAC001 | All zones |

### 4. Build from Source

```bash
# Linux/macOS
cd backend-go
./build.sh

# Windows
cd backend-go
build.bat
```

Or manually:
```bash
go build -ldflags="-s -w" -o sentinel .
```

## API Endpoints

### Admin Endpoints (require login)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/login` | GET/POST | Admin login |
| `/logout` | GET | Logout |
| `/` | GET | Admin dashboard |
| `/api/stats` | GET | Dashboard statistics |
| `/api/access-logs` | GET | Access logs (50 most recent) |
| `/api/users` | GET/POST | List/Create users |
| `/api/users/:id` | PUT/DELETE | Update/Delete user |
| `/api/zones` | GET/POST | List/Create zones |
| `/api/occupancy` | GET | Live occupancy per zone |

### Access Controller Endpoints

These endpoints are used by access point devices (laptop + Arduino).

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/authenticate` | POST | Validate card access |
| `/api/access-log` | POST | Log entry/exit/denied |
| `/api/door/:zone_id/:action` | POST | Lock/unlock/status door |
| `/api/health` | GET | Health check |

> **Note:** Access controller endpoints require API key authentication via `X-API-Key` header.

### API Examples

**Authenticate a Card:**
```bash
curl -X POST http://localhost:5000/api/authenticate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"card_id": "STU001", "zone_id": 1, "method": "card"}'
```

**Get Live Occupancy:**
```bash
curl http://localhost:5000/api/occupancy
```

## Hardware Components

### Access Point Setup

```
┌────────────────────────────────────┐
│          Laptop (any OS)           │
│  ┌──────────────────────────────┐  │
│  │  Python Controller            │  │
│  │  - Serial communication       │  │
│  │  - Face verification         │  │
│  │  - Server API calls          │  │
│  └──────────────────────────────┘  │
│              │ USB                   │
│              ▼                      │
│  ┌──────────────────────────────┐  │
│  │  Arduino Nano                 │  │
│  │  - NFC Reader (RC522)        │  │
│  │  - LCD Display                │  │
│  │  - Status LEDs                │  │
│  └──────────────────────────────┘  │
│                                    │
│  GPIO → Relay → Door Lock         │
└────────────────────────────────────┘
```

### Materials Required

| Component | Approx Cost | Purpose |
|-----------|-------------|---------|
| Laptop (any OS) | $0 | Main controller |
| Arduino Nano | $10 | NFC reading |
| NFC RC522 | $5 | Card reading |
| USB Camera | $20 | Face verification |
| LCD 16x2 | $10 | Display status |
| Relay Module | $5 | Door control |
| **Total** | **~$105** | |

### Face Recognition Module

The `face-check/` directory contains the facial recognition system:

```bash
cd face-check
pip install -r requirements.txt
python main.py
```

This module:
- Detects faces using YuNet model
- Verifies identity using SFace model
- Can be integrated with the main authentication flow

### Arduino Code

The `rfid-operation/main.ino` contains the Arduino sketch for:
- Reading NFC cards via RC522
- Displaying status on LCD
- Communicating with the Python controller via serial

## Security Implementation

### Authentication Layers

1. **Admin Dashboard**: Session-based auth with secure cookies
2. **Access Controllers**: API key authentication via `X-API-Key` header

### Data Protection

| Data | Protection |
|------|------------|
| Passwords | SHA-256 hashed |
| Sessions | Secure cookie store |
| API Keys | Environment variables |

## Performance

### Binary Sizes

| Platform | Size |
|----------|------|
| Linux x64 | 16.7 MB |
| Windows x64 | 17.1 MB |

### Response Times

| Operation | Typical Latency |
|-----------|-----------------|
| API Auth | <10ms |
| Access Log | <15ms |
| Occupancy | <20ms |

## Future Enhancements

- **Phase 3**: SOC Dashboard with real-time camera feeds
- **Phase 4**: AI features (anomaly detection, predictive analytics)
- **Phase 5**: Emergency response (panic buttons, lockdown modes)

## License

MIT