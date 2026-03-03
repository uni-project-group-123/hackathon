# Sentinel - Campus Safety & Security System
## Presentation

---

## The Problem

> *"Currently to access your labs, you are using your ID cards, create a system to ensure the right people are entering the labs"*

**Challenges:**
- ID cards can be shared or stolen
- No real-time tracking of building occupancy
- Security staff lack visibility into access patterns
- No verification beyond card swipe

---

## Our Solution: Sentinel

A multi-factor authentication system that combines:
- **NFC Card Authentication** - Primary identification
- **Facial Recognition** - Biometric verification
- **Real-time Monitoring** - Live occupancy tracking
- **Admin Dashboard** - Centralized security management

---

## System Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Admin      │     │   Access    │     │   Server     │
│   Dashboard  │◄────│   Points    │◄────│   (Pi)       │
│   (Web)      │     │ (Pi+Arduino)│     │              │
└──────────────┘     └──────────────┘     └──────┬───────┘
                                                  │
                                         ┌────────┴────────┐
                                         │   Database    │
                                         │  (SQLite)      │
                                         └────────────────┘
```

---

## Development Journey

### From Concept to Production

| Phase | Description | Key Files |
|-------|-------------|-----------|
| **Planning** | Brief & architecture design | `brief.md`, `plan.md` |
| **Foundation** | Backend + Frontend | `backend-go/main.go` |
| **Core Features** | API, User Management | `documentation.md` |
| **Hardware** | Arduino + Face Check | `rfid-operation/`, `face-check/` |
| **Security** | API Key Authentication | `main.go` |
| **Production** | Cross-platform builds | `sentinel.exe`, `sentinel` |

---

## Key Features

### 1. Multi-Factor Authentication

```
User taps card → Server validates → 
  ✓ Check user exists
  ✓ Check user is active  
  ✓ Check zone permissions
  → Face verification (optional)
  → Door unlocks
```

### 2. Real-Time Occupancy

```
┌─────────────────────────────────────┐
│  Zone        │ Capacity │ Current   │
├──────────────┼──────────┼──────────┤
│ Computer Lab │    30    │    12    │
│ Chemistry    │    20    │     5    │
│ Library      │   100    │    45    │
└─────────────────────────────────────┘
```

### 3. Access Logging

Every entry/exit/denial is logged with:
- User ID
- Zone
- Timestamp
- Method (card, face, etc.)
- Result (granted/denied)

---

## Hardware Components

### Access Point Setup

```
┌────────────────────────────────────┐
│          Raspberry Pi              │
│  ┌──────────────────────────────┐  │
│  │  Python Controller          │  │
│  │  - Serial communication     │  │
│  │  - Face verification        │  │
│  │  - Server API calls        │  │
│  └──────────────────────────────┘  │
│              │ USB                  │
│              ▼                      │
│  ┌──────────────────────────────┐  │
│  │  Arduino Nano                │  │
│  │  - NFC Reader (RC522)       │  │
│  │  - LCD Display              │  │
│  │  - Status LEDs              │  │
│  └──────────────────────────────┘  │
│                                      │
│  GPIO → Relay → Door Lock          │
└────────────────────────────────────┘
```

### Materials Required

| Component | Approx Cost | Purpose |
|-----------|-------------|---------|
| Raspberry Pi 4 | $55 | Main controller |
| Arduino Nano | $10 | NFC reading |
| NFC RC522 | $5 | Card reading |
| USB Camera | $20 | Face verification |
| LCD 16x2 | $10 | Display status |
| Relay Module | $5 | Door control |
| **Total** | **~$105** | |

---

## Security Implementation

### Authentication Layers

```
Layer 1: Admin Dashboard
  └─ Session-based auth
     - Login/password
     - Secure session cookies

Layer 2: Access Controllers  
  └─ API Key authentication
     - X-API-Key header required
     - Validated on every request
```

### Data Protection

| Data | Protection |
|------|------------|
| Passwords | SHA-256 hashed |
| Sessions | Secure cookie store |
| API Keys | Environment variables |
| Network | HTTPS recommended |

---

## Demo: Authentication Flow

```
1. User taps NFC card
   └─ Arduino reads UID

2. Pi receives UID via serial
   └─ Sends to server:
      POST /api/authenticate
      {card_id: "A1B2", zone_id: 1}

3. Server validates
   └─ Checks user exists
   └─ Checks permissions
   └─ Checks zone capacity

4. Response to Pi
   └─ {success: true, user: "John Doe"}

5. Face verification (if enabled)
   └─ Camera captures face
   └─ Compare against stored template

6. Door unlocks
   └─ Relay activates
   └─ Access logged
```

---

## Tech Stack

### Backend
- **Language**: Go 1.21
- **Framework**: Gin
- **Database**: SQLite (PostgreSQL ready)
- **ORM**: GORM

### Frontend
- **Type**: Server-side rendered HTML
- **Styling**: Custom CSS (dark theme)

### Hardware
- **Controller**: Raspberry Pi
- **MCU**: Arduino Nano
- **Language**: Python 3, C++

---

## Performance

### Binary Sizes

| Platform | Size | Notes |
|----------|------|-------|
| Linux x64 | 16.7 MB | Default build |
| Windows x64 | 17.1 MB | No CGO required |
| Linux ARM64 | ~17 MB | Pi 3/4/5 |

### Response Times

| Operation | Typical Latency |
|-----------|-----------------|
| API Auth | <10ms |
| Access Log | <15ms |
| Occupancy | <20ms |

---

## Future Enhancements

### Phase 3: SOC Dashboard
- Real-time camera feeds
- Alert notifications
- Incident logging
- Access pattern visualization

### Phase 4: AI Features
- Anomaly detection
- Suspicious behavior
- Predictive analytics

### Phase 5: Emergency Response
- Panic buttons
- Lockdown modes
- Emergency alerts

---

## References

- Documentation: [documentation.md](documentation.md)
- Source Code: [backend-go/](backend-go/)
- Hardware: [face-check/](face-check/), [rfid-operation/](rfid-operation/)
- GitHub Actions: [.github/workflows/](.github/workflows/)

---

## Questions?

### Contact
- Project Repository
- Documentation: See `documentation.md`

### License
MIT License

---

## Appendix: Git History

```
145aa93 Add API key authentication for access controller endpoints
a5fbf8c Add server integration to face-check and rfid-operation
1a933f8 smaller binary :)
fab2689 Python & Arduino
d3e2207 added windows exe
6c7516d Added more api endpoints + extras
8b3476e frontend and backend added
f27e2d8 Plan + Breif
```

---

*Thank you for your attention!*
