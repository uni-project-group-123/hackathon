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
├── backend/               # Go API server
│   ├── main.go            # Main application
│   ├── go.mod             # Go module
│   ├── go.sum             # Dependencies
│   ├── sentinel           # Compiled binary
│   ├── sentinel.db        # SQLite database
│   └── templates/         # HTML templates
│       ├── index.html     # Admin dashboard
│       └── login.html     # Login page
├── frontend/              # Future: React Native app, etc.
├── brief.md               # Project brief
├── plan.md               # Development plan
└── documentation.md       # This file
```

## Quick Start

### 1. Install Go

If Go is not installed:

```bash
# On Raspberry Pi
sudo apt install golang

# Or download from https://go.dev/dl/
```

### 2. Run the Server

```bash
cd backend
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

## Building

```bash
cd backend
go build -o sentinel .
```

This creates a standalone binary that can be run on any system with the same architecture.

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/login` | GET/POST | Admin login |
| `/logout` | GET | Logout |
| `/` | GET | Admin dashboard |
| `/api/stats` | GET | Dashboard statistics |
| `/api/access-logs` | GET | Access logs |
| `/api/alerts` | GET | System alerts |
| `/api/users` | GET/POST | List/Create users |
| `/api/users/<id>` | PUT/DELETE | Update/Delete user |
| `/api/zones` | GET/POST | List/Create zones |

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

### Admin Credentials

Default admin is created on first run. To add more admins or reset:

```go
// Add to main.go temporarily
admin := Admin{
    Username:     "admin",
    PasswordHash: hashPassword("your-new-password"),
}
db.Create(&admin)
```

## Production Checklist

### 1. Security

- [ ] Change secret key to a secure random value
- [ ] Change default admin password
- [ ] Use HTTPS (reverse proxy with nginx + let's encrypt)
- [ ] Enable firewall: `sudo ufw allow 5000/tcp`
- [ ] Consider rate limiting on login endpoint

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
WorkingDirectory=/home/pi/sentinel/backend
ExecStart=/home/pi/sentinel/backend/sentinel
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

- [ ] Set up log rotation

### 4. Hardware Integration

To connect physical hardware (NFC readers, cameras):

1. **NFC Reader**: Use `github.com/eblot/go-scard` or similar library
2. **Camera**: Use OpenCV or picamera for face recognition
3. **Door Controller**: Use GPIO pins on Raspberry Pi or MQTT

### 5. Monitoring

- [ ] Set up health check endpoint
- [ ] Configure logging to file
- [ ] Set up monitoring (Prometheus, etc.)

## Development

### Add New Features

1. **Add new database model** in `main.go`:
```go
type NewModel struct {
    ID    uint   `gorm:"primaryKey" json:"id"`
    Field string `json:"field"`
}
```

2. **Add API endpoint**:
```go
func getNewModel(c *gin.Context) {
    var items []NewModel
    db.Find(&items)
    c.JSON(http.StatusOK, items)
}

// Add route
r.GET("/api/new-model", authRequired(), getNewModel)
```

### Run Development

```bash
cd backend
go run main.go
```

## Troubleshooting

### Database Issues

Reset database:
```bash
rm backend/sentinel.db
cd backend
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

## License

MIT
