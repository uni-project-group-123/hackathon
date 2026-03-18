# Sentinel - Access Control System Setup

A biometric access control system with face recognition, RFID support, and real-time monitoring.

## Prerequisites

- Go 1.21+
- Python 3.10+
- Git

## Project Structure

```
hackathon/
├── backend-go/           # Go backend (main server)
│   ├── access-controller/  # ESP32 access controller code
│   └── main.go
├── face-check/          # Python face recognition service
│   └── main.py
├── frontend/            # Web interface templates
└── rfid-operation/     # RFID card enrollment Arduino code
```

## Setup Steps

### 1. Clone & Install Dependencies

```bash
# Backend Go dependencies
cd backend-go
go mod download

# Python dependencies
cd ../face-check
pip install -r requirements.txt
```

### 2. Build the Backend

```bash
cd backend-go

# Linux/macOS
./build.sh

# Windows
build.bat
```

This creates the `sentinel` executable.

### 3. Configure Environment Variables

Create a `.env` file in `backend-go/`:

```env
SESSION_SECRET=your-secure-random-secret-key-here
```

Or export directly:
```bash
export SESSION_SECRET="your-secure-random-secret-key-here"
```

### 4. Configure Access Controller

Edit `backend-go/access-controller/config.json`:

```json
{
  "backend_url": "http://localhost:5000",
  "zone_id": 1,
  "access_point": "main-door",
  "api_key": "your-api-key-here"
}
```

### 5. Add Face Recognition Models

Download the required ONNX models to `face-check/models/`:

- [YuNet face detection](https://github.com/opencv/opencv_zoo/tree/master/models/face_detection_yunet) - `face_detection_yunet_2023mar.onnx`
- [SFace recognition](https://github.com/opencv/opencv_zoo/tree/master/models/face_recognition_sface) - `face_recognition_sface_2021dec.onnx`

```bash
mkdir -p face-check/models
# Download models to this directory
```

### 6. Add User Face Images

Place registered user photos in `face-check/users/`:

```
face-check/users/
├── user1.jpg
└── user2.jpg
```

Images are matched to users via card ID naming convention (check `face-check/main.py` for details).

### 7. Run the System

**Backend (terminal 1):**
```bash
cd backend-go
./sentinel
```

**Face recognition service (terminal 2):**
```bash
cd face-check
python main.py
```

**Access controller (ESP32):**
Flash `rfid-operation/main.ino` to your ESP32 device.

## First-Time Setup

On first run, create an admin account by accessing `/login` and registering the first user.

## Default Ports

| Service | Port |
|---------|------|
| Backend API | 5000 |
| Face Check Service | 5001 |

## Hardware Requirements

- ESP32 development board
- RC522 RFID reader
- Camera module (USB or CSI)
- Relay module for door lock
