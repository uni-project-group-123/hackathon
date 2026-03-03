# Sentinel - Campus Safety & Security System

## Overview

Sentinel is a campus safety and security system that provides:
- Multi-factor authentication (Face Recognition, ID Card, PIN)
- Real-time access monitoring and logging
- Admin dashboard for security staff
- User management for students/faculty/staff

## Project Structure

```
sentinel/
├── backend-go/                  # Go API server
│   ├── main.go                 # Main application
│   ├── go.mod                 # Go module
│   ├── go.sum                 # Dependencies
│   ├── sentinel               # Compiled binary
│   ├── sentinel.db            # SQLite database
│   ├── build.sh               # Build script
│   ├── start.sh               # Start script
│   └── access-controller/     # Pi access controller
│       └── main.go           # Access controller code
├── frontend/                   # Web frontend
│   └── templates/            # HTML templates
│       ├── index.html        # Admin dashboard
│       └── login.html        # Login page
├── .github/workflows/          # CI/CD
│   └── build.yml             # Build workflow
├── brief.md                   # Project brief
├── plan.md                   # Development plan
└── documentation.md          # This file
```

## Quick Start

### 1. Install Go

```bash
# On Raspberry Pi
sudo apt install golang

# Or download from https://go.dev/dl/
```

### 2. Run the Server

```bash
cd backend-go
go mod download
go run main.go
```

Or run the compiled binary:

```bash
./sentinel
```

The server will:
- Initialize the SQLite database (`sentinel.db`)
- Create default admin user
- Seed default zones and sample users
- Start on `http://0.0.0.0:5000`

### 3. Access Admin Panel

Open browser: `http://<server-ip>:5000`

Default credentials:
- Username: `admin`
- Password: `admin123`

### 4. Default Data

The server creates sample users with card IDs:

| User ID | Name | Role | Card ID | Can Access |
|---------|------|------|---------|------------|
| STU001 | John Doe | student | STU001 | Computer Lab, Library |
| STU002 | Jane Smith | student | STU002 | Computer Lab, Library |
| FAC001 | Dr. Emily Brown | faculty | FAC001 | All zones |

## API Endpoints

### Admin Endpoints (require login)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/login` | GET/POST | Admin login |
| `/logout` | GET | Logout |
| `/` | GET | Admin dashboard |
| `/api/stats` | GET | Dashboard statistics |
| `/api/access-logs` | GET | Access logs (50 most recent) |
| `/api/alerts` | GET | System alerts |
| `/api/alerts/:id` | POST | Update alert status |
| `/api/users` | GET/POST | List/Create users |
| `/api/users/:id` | PUT/DELETE | Update/Delete user |
| `/api/zones` | GET/POST | List/Create zones |
| `/api/occupancy` | GET | Live occupancy per zone |

### Access Controller Endpoints (no auth)

These endpoints are used by Pi/Arduino devices at access points.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/authenticate` | POST | Validate card access |
| `/api/access-log` | POST | Log entry/exit/denied |
| `/api/door/:zone_id/:action` | POST | Lock/unlock/status door |
| `/api/health` | GET | Health check |

---

## Access Controller API Examples

### 1. Authenticate a Card

Used by Arduino/Pi to verify if a card has access.

**Request:**
```bash
curl -X POST http://localhost:5000/api/authenticate \
  -H "Content-Type: application/json" \
  -d '{
    "card_id": "STU001",
    "zone_id": 1,
    "method": "card"
  }'
```

**Success Response:**
```json
{
  "success": true,
  "user_id": "STU001",
  "user_name": "John Doe",
  "message": "Access granted",
  "role": "student"
}
```

**Denied Response:**
```json
{
  "success": false,
  "user_id": "STU001",
  "user_name": "John Doe",
  "message": "Access denied: Restricted area"
}
```

### 2. Log Access Event

Log when someone enters or exits.

**Request:**
```bash
curl -X POST http://localhost:5000/api/access-log \
  -H "Content-Type: application/json" \
  -d '{
    "card_id": "STU001",
    "zone_id": 1,
    "method": "card",
    "action": "entry",
    "success": true
  }'
```

### 3. Get Live Occupancy

Get current occupancy for all zones.

**Request:**
```bash
curl http://localhost:5000/api/occupancy
```

**Response:**
```json
[
  {
    "zone_id": 1,
    "zone_name": "Computer Lab 1",
    "current_occupancy": 5,
    "max_capacity": 30,
    "percentage": 16.67
  },
  {
    "zone_id": 2,
    "zone_name": "Chemistry Lab",
    "current_occupancy": 2,
    "max_capacity": 20,
    "percentage": 10
  }
]
```

### 4. Control Door

Lock/unlock a door or get status.

```bash
# Unlock door for zone 1
curl -X POST http://localhost:5000/api/door/1/unlock

# Lock door for zone 1
curl -X POST http://localhost:5000/api/door/1/lock

# Get door status
curl http://localhost:5000/api/door/1/status
```

### 5. Health Check

```bash
curl http://localhost:5000/api/health
```

---

## Hardware Integration

### Architecture

```
┌─────────────┐         ┌─────────────┐
│   Arduino   │────────▶│   Server    │
│  + NFC RC522│  HTTP   │  (Pi/Cloud) │
│  + Camera   │         │             │
└─────────────┘         └─────────────┘
       │                        │
       │ LED/Buzzer            │ Web Dashboard
       ▼                        ▼
   Door Lock               Admin Panel
```

### Arduino Setup

Connect Arduino to Raspberry Pi via USB serial or Ethernet.

**Arduino Code (sketch):**

```cpp
#include <SPI.h>
#include <MFRC522.h>
#include <SoftwareSerial.h>

#define RST_PIN 9
#define SS_PIN 10

MFRC522 rfid(SS_PIN, RST_PIN);
SoftwareSerial espSerial(2, 3); // RX, TX - connect to ESP01

String serverIP = "192.168.1.100"; // Your Pi IP
String zoneId = "1";

void setup() {
  Serial.begin(9600);
  SPI.begin();
  rfid.PCD_Init();
  espSerial.begin(9600);
  
  pinMode(LED_BUILTIN, OUTPUT);
  pinMode(7, OUTPUT); // Green LED
  pinMode(6, OUTPUT); // Red LED
}

void loop() {
  if (!rfid.PICC_IsNewCardPresent()) return;
  if (!rfid.PICC_ReadCardSerial()) return;
  
  String cardUID = "";
  for (byte i = 0; i < rfid.uid.size; i++) {
    cardUID += String(rfid.uid.uidByte[i], HEX);
  }
  
  Serial.println("Card: " + cardUID);
  authenticateCard(cardUID);
  
  rfid.PICC_HaltA();
}

void authenticateCard(String cardUID) {
  // Send to server via Serial (Pi receives this)
  Serial.println("AUTH:" + cardUID + ":" + zoneId);
  
  // Wait for response from Pi
  delay(1000);
  
  if (Serial.available()) {
    String response = Serial.readString();
    if (response.indexOf("GRANTED") >= 0) {
      digitalWrite(7, HIGH); // Green LED
      // Open door relay
      delay(3000);
      digitalWrite(7, LOW);
    } else {
      digitalWrite(6, HIGH); // Red LED
      delay(1000);
      digitalWrite(6, LOW);
    }
  }
}
```

### Raspberry Pi (Python listener on Arduino)

```python
#!/usr/bin/env python3
import serial
import requests

SERIAL_PORT = '/dev/ttyUSB0'  # Arduino connected here
SERVER_URL = 'http://localhost:5000'

ser = serial.Serial(SERIAL_PORT, 9600)

def send_auth(card_id, zone_id):
    resp = requests.post(f'{SERVER_URL}/api/authenticate', json={
        'card_id': card_id,
        'zone_id': int(zone_id),
        'method': 'card'
    })
    data = resp.json()
    
    if data.get('success'):
        ser.write(b'GRANTED\n')
        log_access(card_id, zone_id, 'entry', True)
    else:
        ser.write(b'DENIED\n')
        log_access(card_id, zone_id, 'denied', False)

def log_access(card_id, zone_id, action, success):
    requests.post(f'{SERVER_URL}/api/access-log', json={
        'card_id': card_id,
        'zone_id': int(zone_id),
        'method': 'card',
        'action': action,
        'success': success
    })

while True:
    if ser.in_waiting:
        line = ser.readline().decode().strip()
        if line.startswith('AUTH:'):
            parts = line.split(':')
            if len(parts) == 3:
                send_auth(parts[1], parts[2])
```

### Face Recognition Option

For face recognition, use a separate camera module:

```
┌─────────────┐         ┌─────────────┐
│   Pi +      │────────▶│   Server    │
│   Camera    │  HTTP   │             │
│  (faceapi)  │         │             │
└─────────────┘         └─────────────┘
       │                        │
       │ Permission            │ Dashboard
       ▼                        ▼
   NFC Card                 Admin Panel
```

**Flow:**
1. User taps NFC card
2. Camera captures face
3. Server verifies card + face match
4. Door opens if both valid

### GPIO Door Control

Connect relay module to Pi GPIO:

```python
import RPi.GPIO as GPIO

GPIO.setmode(GPIO.BCM)
GPIO.setup(17, GPIO.OUT)  # Door relay

def unlock_door():
    GPIO.output(17, GPIO.HIGH)
    time.sleep(3)
    GPIO.output(17, GPIO.LOW)
```

---

## Configuration

### Database

Currently using SQLite. To use PostgreSQL in production:

```go
// In main.go, change:
db, err = gorm.Open(sqlite.Open("sentinel.db"), &gorm.Config{})
// To:
db, err = gorm.Open(postgres.Open("host=user dbname=sentinel sslmode=disable"), &gorm.Config{})
```

Then install PostgreSQL driver:
```bash
go get gorm.io/driver/postgres
```

### Secret Key

Change the secret key in `main.go`:

```go
store := cookie.NewStore([]byte("your-secure-secret-key"))
```

Generate a random key:
```bash
openssl rand -hex 32
```

### Add Users with Card IDs

Users need a `card_id` to authenticate. When creating users via API:

```bash
curl -X POST http://localhost:5000/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "STU003",
    "name": "New Student",
    "email": "new@campus.edu",
    "role": "student",
    "card_id": "A1B2C3D4"
  }'
```

---

## Production Checklist

### 1. Security

- [ ] Change secret key to a secure random value
- [ ] Change default admin password
- [ ] Use HTTPS (reverse proxy with nginx + let's encrypt)
- [ ] Enable firewall: `sudo ufw allow 5000/tcp`
- [ ] Add API key authentication for access controller endpoints

### 2. Database

- [ ] Migrate from SQLite to PostgreSQL
- [ ] Set up regular database backups
- [ ] Configure connection pooling

### 3. Deployment

- [ ] Build release binary:
  ```bash
  CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o sentinel .
  ```
- [ ] Set up reverse proxy with nginx
- [ ] Configure systemd service for auto-start

Example systemd service (`/etc/systemd/system/sentinel.service`):
```ini
[Unit]
Description=Sentinel Security System
After=network.target

[Service]
User=pi
WorkingDirectory=/home/pi/sentinel/backend-go
ExecStart=/home/pi/sentinel/backend-go/sentinel
Restart=always

[Install]
WantedBy=multi-user.target
```

Then enable:
```bash
sudo systemctl daemon-reload
sudo systemctl enable sentinel
sudo systemctl start sentinel
```

### 4. Hardware

- [ ] Install NFC readers at access points
- [ ] Connect door locks to relays
- [ ] Set up cameras for face recognition (optional)
- [ ] Configure UPS for power outages

### 5. Monitoring

- [ ] Set up health check endpoint
- [ ] Configure logging to file
- [ ] Set up monitoring (Prometheus, etc.)

---

## Development

### Run Development

```bash
cd backend-go
go run main.go
```

### Access Controller Testing

The access controller has a simulation mode for testing without hardware:

```bash
cd backend-go/access-controller
go run main.go
```

Enter card IDs to test authentication:
- STU001, STU002 - students
- FAC001 - faculty

---

## Troubleshooting

### Database Issues

Reset database:
```bash
rm backend-go/sentinel.db
cd backend-go
go run main.go
```

### Port Already in Use

```bash
# Find process using port 5000
lsof -i :5000
# Kill it
kill <PID>
```

### Build Errors

Ensure Go is properly installed:
```bash
go version
go mod tidy
```

### NFC Reader Not Working

1. Check serial connection: `ls -l /dev/ttyUSB0`
2. Check permissions: `sudo usermod -a -G dialout $USER`
3. Log out and back in

---

## License

MIT
