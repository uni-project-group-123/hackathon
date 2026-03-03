# Sentinel - Campus Safety & Security System

## Project Overview

Sentinel is a comprehensive campus safety and security system that provides multi-factor authentication, real-time access monitoring, and user management for educational institutions.

### Mission

To create a secure, intelligent campus access control system that verifies identity through multiple methods (NFC cards, facial recognition), maintains real-time occupancy tracking, and provides security staff with actionable insights through an intuitive admin dashboard.

---

## Project History & Evolution

### Development Journey (from git log)

```
f27e2d8  Plan + Brief
8b3476e  Frontend and backend added
6c7516d  Added more api endpoints + extras
d3e2207  Added windows exe
fab2689  Python & Arduino
1a933f8  Smaller binary :)
a5fbf8c  Add server integration to face-check and rfid-operation
145aa93  Add API key authentication for access controller endpoints
```

### Iteration Breakdown

| Commit | Phase | Description |
|--------|-------|-------------|
| f27e2d8 | Planning | Project brief and plan created |
| 8b3476e | Foundation | Flask backend + HTML frontend |
| 6c7516d | Core Features | API endpoints, user management |
| d3e2207 | Cross-Platform | Windows executable support |
| fab2689 | Hardware | Python/Arduino integration |
| 1a933f8 | Optimization | Binary size reduction |
| a5fbf8c | Integration | Server communication for auth |
| 145aa93 | Security | API key authentication |

---

## System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              SENTINEL SYSTEM                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐  │
│  │   ADMIN PANEL   │     │   MOBILE APP    │     │  ACCESS POINTS   │  │
│  │   (Web UI)      │     │   (Future)      │     │   (Pi + Arduino) │  │
│  └────────┬─────────┘     └────────┬─────────┘     └────────┬─────────┘  │
│           │                       │                       │               │
│           └───────────────────────┼───────────────────────┘               │
│                                   │                                        │
│                                   ▼                                        │
│                    ┌──────────────────────────────┐                       │
│                    │      BACKEND SERVER           │                       │
│                    │         (Go/Gin)             │                       │
│                    │         Port 5000            │                       │
│                    └──────────────┬───────────────┘                       │
│                                   │                                        │
│                    ┌──────────────┴───────────────┐                       │
│                    │      DATABASE                │                       │
│                    │     (SQLite/PostgreSQL)     │                       │
│                    │   sentinel.db                │                       │
│                    └─────────────────────────────┘                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Component Details

#### 1. Backend Server (Go)

```
┌─────────────────────────────────────────────────────────────┐
│                     BACKEND SERVER                            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                 GIN ROUTER                          │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │  Admin Routes (Session Auth)                        │    │
│  │    ├─ GET  /login, POST /login                   │    │
│  │    ├─ GET  /logout                                │    │
│  │    ├─ GET  / (dashboard)                          │    │
│  │    └─ API /api/* (protected)                      │    │
│  │                                                      │    │
│  │  Access Controller Routes (API Key Auth)            │    │
│  │    ├─ POST /api/authenticate                      │    │
│  │    ├─ POST /api/access-log                        │    │
│  │    ├─ POST /api/door/:id/:action                 │    │
│  │    ├─ GET  /api/occupancy                        │    │
│  │    └─ GET  /api/health                           │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │               DATABASE MODELS                        │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │  Admin    │  User  │  Zone  │  AccessLog           │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

#### 2. Access Controller (Raspberry Pi + Arduino)

```
┌─────────────────────────────────────────────────────────────┐
│                  ACCESS CONTROLLER NODE                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    │
│  │  Arduino    │    │   Pi        │    │   Camera    │    │
│  │  Nano/Mega  │◄──►│  (Python)   │◄──►│  (USB)      │    │
│  └──────┬──────┘    └──────┬──────┘    └─────────────┘    │
│         │                  │                               │
│    ┌────┴────┐        ┌────┴────┐                         │
│    │ NFC    │        │ Server  │                         │
│    │ RC522  │        │  API   │                         │
│    └─────────┘        └────────┘                         │
│                                                              │
│  GPIO Outputs:                                              │
│    ├─ Green LED  (Access Granted)                           │
│    ├─ Red LED   (Access Denied)                            │
│    └─ Relay    (Door Lock)                                │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

#### 3. Admin Dashboard

```
┌─────────────────────────────────────────────────────────────┐
│                     ADMIN DASHBOARD                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  STATS CARDS                                        │   │
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐      │   │
│  │  │Access  │ │Alerts  │ │Zones   │ │Users   │      │   │
│  │  │Today   │ │Active  │ │Occupied│ │Total   │      │   │
│  │  └────────┘ └────────┘ └────────┘ └────────┘      │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌─────────────────────┐  ┌─────────────────────────┐    │
│  │  ALERTS PANEL       │  │  ACCESS LOGS TABLE       │    │
│  │  - Warning          │  │  Time | User | Zone     │    │
│  │  - Info             │  │  ─────────────────────  │    │
│  │  - Success          │  │  09:15 | John | Lab 1   │    │
│  └─────────────────────┘  └─────────────────────────┘    │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  TABS: Dashboard | Users | Zones                    │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Data Flow

### Authentication Flow

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Student    │     │   Arduino    │     │     Pi       │
│  (Tap Card)  │     │  (NFC Read)  │     │  (Python)    │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                     │                      │
       │  UID:A1B2C3D4      │                      │
       │────────────────────►│                      │
       │                     │                      │
       │                     │  UID,A1B2C3D4        │
       │                     │────────────────────►│
       │                     │                      │
       │                     │         ┌───────────┴───────────┐
       │                     │         │  1. Validate Card    │
       │                     │         │  2. Check Zone       │
       │                     │         │  3. Check Capacity   │
       │                     │         └───────────┬───────────┘
       │                     │                      │
       │                     │    POST /api/authenticate
       │                     │ ────────────────────►│ Server
       │                     │                      │
       │                     │         ┌───────────┴───────────┐
       │                     │         │  Validate API Key   │
       │                     │         │  Check User Status  │
       │                     │         │  Check Zone Access │
       │                     │         └───────────┬───────────┘
       │                     │                      │
       │                     │   {success: true}   │
       │                     │◄─────────────────────│
       │                     │                      │
       │                     │  VERIFIED/NOT_AUTH   │
       │                     │◄─────────────────────│
       │                     │                      │
       │         ┌──────────┴──────────┐            │
       │         │  If verified:      │            │
       │         │  - Face Check     │            │
       │         │  - Log Access     │            │
       │         │  - Update Count   │            │
       │         └───────────────────┘            │
       │                                          │
       ▼                                          ▼
┌──────────────┐                         ┌──────────────┐
│  Door Opens │                         │  Server DB   │
│  (Relay)    │                         │  Updated    │
└──────────────┘                         └──────────────┘
```

### Access Log Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    ACCESS LOGGING FLOW                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  User Access Event                                              │
│       │                                                         │
│       ▼                                                         │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Pi sends POST /api/access-log                         │    │
│  │  {card_id, zone_id, action, method, success}           │    │
│  └─────────────────────────┬───────────────────────────────┘    │
│                            │                                    │
│                            ▼                                    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Server Processing:                                    │    │
│  │  1. Validate API Key                                  │    │
│  │  2. Find user by card_id                              │    │
│  │  3. Create AccessLog record                           │    │
│  │  4. Update zone occupancy (+1 entry, -1 exit)        │    │
│  └─────────────────────────┬───────────────────────────────┘    │
│                            │                                    │
│                            ▼                                    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  DATABASE                                             │    │
│  │  Table: access_logs                                   │    │
│  │  ┌────┬───────┬────────┬────────┬────────────────┐  │    │
│  │  │ id │user_id│ zone_id│ action  │ timestamp      │  │    │
│  │  ├───┼───────┼────────┼────────┼────────────────┤  │    │
│  │  │ 1 │ STU001 │   1    │ entry  │ 2026-03-03 09 │  │    │
│  │  │ 2 │ STU001 │   1    │ exit   │ 2026-03-03 10 │  │    │
│  │  └────┴───────┴────────┴────────┴────────────────┘  │    │
│  │                                                       │    │
│  │  Table: zones                                        │    │
│  │  ┌────┬────────────┬────────────┬──────────────┐   │    │
│  │  │ id │    name    │max_capacity│current_occ   │   │    │
│  │  ├───┼────────────┼────────────┼──────────────┤   │    │
│  │  │ 1 │ Computer 1 │    30     │     5        │   │    │
│  │  └────┴────────────┴────────────┴──────────────┘   │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Database Schema

### Entity Relationship Diagram

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│    Admin    │         │    User     │         │    Zone     │
├─────────────┤         ├─────────────┤         ├─────────────┤
│ id (PK)     │         │ id (PK)     │         │ id (PK)     │
│ username    │         │ user_id     │         │ name        │
│ password_hash│         │ name        │         │ description │
└─────────────┘         │ email       │         │ max_capacity│
                        │ role        │         │ is_restricted
                        │ status      │         └──────┬──────┘
                        │ card_id     │                │
                        │ face_encoding│               │
                        └──────┬──────┘                │
                               │                       │
                               │ 1:N                   │
                               ▼                       │
                        ┌─────────────┐                │
                        │ AccessLog  │                │
                        ├─────────────┤                │
                        │ id (PK)     │◄──────────────┘
                        │ user_id (FK)│
                        │ zone_id (FK)│
                        │ action      │
                        │ method      │
                        │ timestamp   │
                        │ notes       │
                        └─────────────┘
```

### Table Definitions

```sql
-- Admins (system users)
CREATE TABLE admins (
    id INTEGER PRIMARY KEY,
    username VARCHAR(80) UNIQUE NOT NULL,
    password_hash VARCHAR(128) NOT NULL
);

-- Users (students, faculty, staff)
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    user_id VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(120),
    role VARCHAR(20) DEFAULT 'student',
    status VARCHAR(20) DEFAULT 'active',
    card_id VARCHAR(50),
    face_encoding TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Zones (buildings, labs, rooms)
CREATE TABLE zones (
    id INTEGER PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(200),
    max_capacity INTEGER DEFAULT 50,
    current_occupancy INTEGER DEFAULT 0,
    is_restricted BOOLEAN DEFAULT FALSE
);

-- Access Logs
CREATE TABLE access_logs (
    id INTEGER PRIMARY KEY,
    user_id VARCHAR(20) NOT NULL,
    zone_id INTEGER NOT NULL,
    action VARCHAR(20) NOT NULL,
    method VARCHAR(50),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (zone_id) REFERENCES zones(id)
);
```

---

## Security Considerations

### Authentication & Authorization

#### 1. Admin Access Control

```
┌─────────────────────────────────────────────────────────────┐
│              ADMIN AUTHENTICATION FLOW                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────┐     ┌──────────┐     ┌──────────┐           │
│  │ Browser  │────►│  Login   │────►│ Session  │           │
│  │          │     │   Form   │     │  Store  │           │
│  └──────────┘     └──────────┘     └──────────┘           │
│       │                                   │                │
│       │  GET /                           │                │
│       │──────────────────────────────────►                │
│       │              ┌──────────┐                           │
│       │              │ Check    │                           │
│       │              │ Session  │                           │
│       │              └──────────┘                           │
│       │                  │                                  │
│       │           ┌──────┴──────┐                           │
│       │           │ Valid?       │                           │
│       │           └──────┬──────┘                           │
│       │                  │                                  │
│       ▼                  ▼                                  │
│  ┌──────────┐     ┌──────────┐                             │
│  │ Dashboard │◄────│ 200 OK   │                             │
│  │  Render   │     │          │                             │
│  └──────────┘     └──────────┘                             │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Security Measures:**
- Session-based authentication with secure cookies
- Passwords hashed using SHA-256
- Session validation on every protected request

#### 2. Access Controller API Security

```
┌─────────────────────────────────────────────────────────────┐
│           ACCESS CONTROLLER API SECURITY                    │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Before processing any request:                              │
│                                                              │
│  1. API KEY VALIDATION                                      │
│     ┌──────────────────────────────────────────┐           │
│     │ Header: X-API-Key: <secret>              │           │
│     │ or Form: api_key=<secret>               │           │
│     └────────────────────┬─────────────────────┘           │
│                         │                                  │
│                         ▼                                  │
│     ┌──────────────────────────────────────────┐           │
│     │ Compare against API_KEY env variable     │           │
│     │ Default: dev-api-key-change-in-production│           │
│     └────────────────────┬─────────────────────┘           │
│                         │                                  │
│              ┌──────────┴──────────┐                     │
│              │ Match?               │                     │
│              └──────────┬──────────┘                     │
│                         │                                  │
│            ┌────────────┴────────────┐                    │
│            │                         │                    │
│            ▼                         ▼                    │
│     ┌─────────────┐          ┌─────────────┐             │
│     │  200 OK    │          │ 401 Error   │             │
│     │ Process    │          │ Unauthorized │             │
│     └─────────────┘          └─────────────┘             │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Security Measures:**
- API key required for all access controller endpoints
- Keys stored in environment variables, not in code
- Separate authentication from admin and hardware

### Data Protection

| Data Type | Protection | Notes |
|-----------|------------|-------|
| Admin Password | SHA-256 Hash | Never stored in plaintext |
| User Card IDs | Plaintext | Needed for NFC matching |
| Face Templates | Optional | Can be stored as embeddings |
| Session Cookies | Encrypted | Gin sessions with secure key |
| API Keys | Environment | Never committed to git |

### Network Security

```
┌─────────────────────────────────────────────────────────────┐
│                    NETWORK SECURITY                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Production Deployment (Recommended):                         │
│                                                              │
│  ┌─────────┐     ┌─────────┐     ┌──────────────┐        │
│  │  User   │     │   Nginx │     │   Sentinel   │        │
│  │ Browser │────►│ (HTTPS) │────►│   Server    │        │
│  └─────────┘     └─────────┘     └──────────────┘        │
│                         │                  │               │
│                    Let's Encrypt      Internal             │
│                    Auto-renew         Network              │
│                                                              │
│  Firewall Rules:                                             │
│    - Allow 443 (HTTPS) from external                        │
│    - Allow 5000 from internal only                         │
│    - Block all other inbound                               │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Vulnerabilities & Mitigations

| Vulnerability | Risk Level | Mitigation |
|--------------|-------------|------------|
| SQL Injection | Low | ORM prevents this (GORM) |
| XSS | Low | Template auto-escaping |
| CSRF | Medium | Use gin-csrf middleware in production |
| Weak API Key | Medium | Require strong key in production |
| No HTTPS | High | Use reverse proxy with TLS |
| Default Credentials | High | Change admin password immediately |
| Unrestricted Access | High | Firewall rules |

---

## Ethical Considerations

### Privacy

```
┌─────────────────────────────────────────────────────────────┐
│                    PRIVACY CONSIDERATIONS                    │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  DATA COLLECTED:                                            │
│  ├─ Access times and locations                              │
│  ├─ Card/ID numbers                                         │
│  └─ Facial templates (optional)                             │
│                                                              │
│  DATA NOT COLLECTED:                                        │
│  ├─ Location tracking outside campus                        │
│  ├─ Biometric raw images                                    │
│  └─ Financial information                                   │
│                                                              │
│  USER RIGHTS:                                                │
│  ├─ Right to access their data                             │
│  ├─ Right to correction                                    │
│  └─ Right to deletion (where legal)                        │
│                                                              │
│  TRANSPARENCY:                                               │
│  ├─ Clear signage at entry points                          │
│  ├─ Privacy policy available                                │
│  └─ Purpose limitation (security only)                      │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Bias & Fairness

- **Role-based access**: System distinguishes students vs faculty appropriately
- **No biometric storage by default**: Face templates are optional
- **Graceful fallbacks**: If face fails, card alone can grant access

### Consent

- Campus-wide notification of security system
- Opt-in face registration (not mandatory)
- Alternative authentication always available

---

## Deployment

### Raspberry Pi Setup

```
┌─────────────────────────────────────────────────────────────┐
│                  PI DEPLOYMENT DIAGRAM                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │            Raspberry Pi 3/4/5                       │    │
│  │  ┌─────────────────────────────────────────────┐   │    │
│  │  │  Sentinel Access Controller (Python)         │   │    │
│  │  │  - Serial communication with Arduino         │   │    │
│  │  │  - Face verification                         │   │    │
│  │  │  - HTTP calls to server                     │   │    │
│  │  └─────────────────────────────────────────────┘   │    │
│  │                      │                              │    │
│  │              USB / UART │                           │    │
│  │                      │                              │    │
│  │  ┌───────────────────┴───────────────────────┐    │    │
│  │  │  Arduino Nano/Mega                          │    │    │
│  │  │  - NFC Reader (RC522)                      │    │    │
│  │  │  - LCD Display                             │    │    │
│  │  │  - Status LEDs                             │    │    │
│  │  └───────────────────────────────────────────┘    │    │
│  │                                                      │    │
│  │  GPIO:                                               │    │
│  │    D3  → Green LED (Access)                         │    │
│  │    D5  → Red LED (Deny)                           │    │
│  │    D6  → Blue LED (Processing)                    │    │
│  │    D7  → Relay (Door lock)                         │    │
│  │                                                      │    │
│  └─────────────────────────────────────────────────────┘    │
│                           │                                  │
│                    Ethernet / WiFi                          │
│                           │                                  │
│                           ▼                                  │
│                    ┌──────────────┐                         │
│                    │  Campus      │                         │
│                    │  Network     │                         │
│                    └──────┬───────┘                         │
│                           │                                  │
│                           ▼                                  │
│                    ┌──────────────┐                         │
│                    │  Sentinel    │                         │
│                    │  Server      │                         │
│                    └──────────────┘                         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### systemd Service (Pi)

```ini
[Unit]
Description=Sentinel Access Controller
After=network.target

[Service]
Type=simple
User=pi
WorkingDirectory=/home/pi/sentinel
ExecStart=/usr/bin/python3 /home/pi/sentinel/face-check/main.py
Environment="SERVER_URL=http://192.168.1.100:5000"
Environment="ZONE_ID=1"
Environment="API_KEY=your-secure-api-key"
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

---

## API Reference

### Authentication Required (Admin)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Dashboard |
| GET/POST | `/login` | Admin login |
| GET | `/logout` | Logout |
| GET | `/api/stats` | Dashboard stats |
| GET | `/api/access-logs` | Access history |
| GET | `/api/alerts` | System alerts |
| GET/POST | `/api/users` | User CRUD |
| GET/POST | `/api/zones` | Zone CRUD |

### API Key Required (Hardware)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/authenticate` | Verify card access |
| POST | `/api/access-log` | Log access event |
| POST | `/api/door/:id/:action` | Lock/unlock door |

### Public

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/occupancy` | Zone occupancy |

---

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| NFC not read | Check serial connection, try different port |
| Face not detected | Ensure good lighting, camera position |
| Server unreachable | Check network, firewall rules |
| API key error | Verify API_KEY matches server |
| Database locked | Check file permissions |

### Log Locations

- Server: System journal or file
- Pi: Console output or systemd logs
- Arduino: Serial monitor

---

## License

MIT License - See LICENSE file for details
