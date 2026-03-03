import os
import time
import serial
import serial.tools.list_ports
import cv2
import numpy as np
import requests

BAUD = 115200
CAMERA_INDEX = 0
USERS_DIR = "users"
MODELS_DIR = "models"

BACKEND_URL = os.environ.get("BACKEND_URL", "http://localhost:5000")
ZONE_ID = int(os.environ.get("ZONE_ID", "1"))

DET_MODEL = os.path.join(MODELS_DIR, "face_detection_yunet_2023mar.onnx")
REC_MODEL = os.path.join(MODELS_DIR, "face_recognition_sface_2021dec.onnx")

SIM_THRESHOLD = 0.45

WINDOW_NAME = "Attendance Verifier"
WINDOW_W = 570
WINDOW_H = 720
INSET_W = 140
INSET_H = 180
PADDING = 12


def find_arduino_port():
    ports = list(serial.tools.list_ports.comports())
    for p in ports:
        desc = (p.description or "").lower()
        if (
            "arduino" in desc
            or "ch340" in desc
            or "usb serial" in desc
            or "wch" in desc
        ):
            return p.device
    if ports:
        return ports[0].device
    raise RuntimeError("Arduino port not found.")


def uid_to_path(uid: str) -> str:
    return os.path.join(USERS_DIR, uid.replace(":", "_") + ".jpg")


def assert_model_ok(path: str, min_bytes: int):
    if not os.path.exists(path):
        raise RuntimeError(f"Missing model file: {path}")
    size = os.path.getsize(path)
    if size < min_bytes:
        raise RuntimeError(f"Model appears corrupted/empty: {path} size={size}")


def resize_to_window(img_bgr):
    return cv2.resize(img_bgr, (WINDOW_W, WINDOW_H))


def largest_face(faces):
    if faces is None or len(faces) == 0:
        return None
    faces = sorted(faces, key=lambda f: f[2] * f[3], reverse=True)
    return faces[0]


def cosine_similarity(a, b) -> float:
    a = a.reshape(-1).astype(np.float32)
    b = b.reshape(-1).astype(np.float32)
    denom = (np.linalg.norm(a) * np.linalg.norm(b)) + 1e-8
    return float(np.dot(a, b) / denom)


def get_embedding_from_face(img_bgr, face, face_recognizer):
    aligned = face_recognizer.alignCrop(img_bgr, face)
    emb = face_recognizer.feature(aligned)
    return emb
def backend_authenticate(card_id: str) -> dict:
    """Call backend /api/authenticate to verify card access."""
    try:
        resp = requests.post(
            f"{BACKEND_URL}/api/authenticate",
            json={
                "card_id": card_id,
                "zone_id": ZONE_ID,
                "method": "Face Recognition",
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
            },
            timeout=5,
        )
        data = resp.json()
        print(f"[BACKEND] authenticate: {data}")
        return data
    except Exception as e:
        print(f"[BACKEND] authenticate error: {e}")
        return {"success": False, "message": str(e)}


def backend_log_access(card_id: str, action: str, success: bool, method: str = "Face Recognition"):
    """Call backend /api/access-log to record the event."""
    try:
        resp = requests.post(
            f"{BACKEND_URL}/api/access-log",
            json={
                "card_id": card_id,
                "zone_id": ZONE_ID,
                "method": method,
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
                "action": action,
                "success": success,
            },
            timeout=5,
        )
        data = resp.json()
        print(f"[BACKEND] access-log: {data}")
    except Exception as e:
        print(f"[BACKEND] access-log error: {e}")


def backend_get_suggested_action(card_id: str) -> str:
    """Query backend for the user's last action to determine entry/exit."""
    try:
        resp = requests.get(
            f"{BACKEND_URL}/api/last-action",
            params={"card_id": card_id, "zone_id": ZONE_ID},
            timeout=5,
        )
        data = resp.json()
        suggested = data.get("suggested_action", "entry")
        print(f"[BACKEND] last-action: {data} -> suggested: {suggested}")
        return suggested
    except Exception as e:
        print(f"[BACKEND] last-action error: {e}, defaulting to entry")
        return "entry"


def overlay_inset(frame, inset_img, uid_text, status_text=None):
    h, w = frame.shape[:2]
    x2 = w - PADDING
    y2 = h - PADDING
    x1 = x2 - INSET_W
    y1 = y2 - INSET_H

    cv2.rectangle(frame, (x1 - 6, y1 - 60), (x2 + 6, y2 + 6), (0, 0, 0), -1)

    if inset_img is not None:
        thumb = cv2.resize(inset_img, (INSET_W, INSET_H))
        frame[y1:y2, x1:x2] = thumb
        cv2.rectangle(frame, (x1, y1), (x2, y2), (255, 255, 255), 2)

    cv2.putText(
        frame,
        f"UID: {uid_text}",
        (x1, y1 - 28),
        cv2.FONT_HERSHEY_SIMPLEX,
        0.6,
        (255, 255, 255),
        2,
        cv2.LINE_AA,
    )

    if status_text:
        cv2.putText(
            frame,
            status_text,
            (x1, y1 - 8),
            cv2.FONT_HERSHEY_SIMPLEX,
            0.6,
            (0, 255, 0) if "VERIFIED" in status_text else (0, 0, 255),
            2,
            cv2.LINE_AA,
        )



def verify_live_with_gui(
    uid,
    ref_img_bgr,
    ref_emb,
    face_detector,
    face_recognizer,
    cap,
    seconds=5.0,
    close_delay=3.0,
):
    if not cap.isOpened():
        raise RuntimeError("Cannot open camera.")

    start = time.time()
    best_sim = -1.0
    verified = False

    cv2.namedWindow(WINDOW_NAME, cv2.WINDOW_NORMAL)
    cv2.resizeWindow(WINDOW_NAME, WINDOW_W, WINDOW_H)

    while time.time() - start < seconds:
        ret, frame = cap.read()
        if not ret:
            continue

        small = resize_to_window(frame)
        sh, sw = small.shape[:2]
        face_detector.setInputSize((sw, sh))
        faces = face_detector.detect(small)[1]

        status_line = None

        f = largest_face(faces)
        if f is not None:
            x, y, bw, bh = map(int, f[:4])
            x = max(0, x)
            y = max(0, y)
            bw = max(1, bw)
            bh = max(1, bh)

            cv2.rectangle(small, (x, y), (x + bw, y + bh), (0, 255, 0), 2)

            cam_emb = get_embedding_from_face(small, f, face_recognizer)
            sim = cosine_similarity(ref_emb, cam_emb)
            best_sim = max(best_sim, sim)

            status_line = f"SIM: {sim:.3f}  TH: {SIM_THRESHOLD:.2f}"
            if sim >= SIM_THRESHOLD:
                verified = True
                overlay_inset(small, ref_img_bgr, uid, status_text="VERIFIED")
                cv2.putText(
                    small,
                    status_line,
                    (PADDING, 30),
                    cv2.FONT_HERSHEY_SIMPLEX,
                    0.8,
                    (0, 255, 0),
                    2,
                    cv2.LINE_AA,
                )
                cv2.imshow(WINDOW_NAME, small)
                cv2.waitKey(1)
                _show_live_result(
                    cap, face_detector, ref_img_bgr, uid, "VERIFIED", close_delay
                )
                return True, best_sim
        else:
            status_line = "No face detected"

        cv2.putText(
            small,
            status_line,
            (PADDING, 30),
            cv2.FONT_HERSHEY_SIMPLEX,
            0.8,
            (255, 255, 255),
            2,
            cv2.LINE_AA,
        )

        cv2.putText(
            small,
            "Press Q to quit",
            (PADDING, sh - 12),
            cv2.FONT_HERSHEY_SIMPLEX,
            0.6,
            (200, 200, 200),
            2,
            cv2.LINE_AA,
        )

        cv2.imshow(WINDOW_NAME, small)
        key = cv2.waitKey(10) & 0xFF
        if key in (ord("q"), ord("Q")):
            cv2.destroyWindow(WINDOW_NAME)
            cv2.waitKey(1)
            raise KeyboardInterrupt

    _show_live_result(cap, face_detector, ref_img_bgr, uid, "UNVERIFIED", close_delay)
    return False, best_sim


def _show_live_result(cap, face_detector, ref_img_bgr, uid, status, delay_sec):
    color = (0, 255, 0) if status == "VERIFIED" else (0, 0, 255)
    t0 = time.time()

    while time.time() - t0 < delay_sec:
        ret, frame = cap.read()
        if not ret:
            key = cv2.waitKey(10) & 0xFF
            if key in (ord("q"), ord("Q")):
                break
            continue

        small = resize_to_window(frame)
        sh, sw = small.shape[:2]

        face_detector.setInputSize((sw, sh))
        faces = face_detector.detect(small)[1]
        f = largest_face(faces)
        if f is not None:
            x, y, bw, bh = map(int, f[:4])
            x = max(0, x)
            y = max(0, y)
            bw = max(1, bw)
            bh = max(1, bh)
            cv2.rectangle(small, (x, y), (x + bw, y + bh), color, 2)

        if status == "VERIFIED":
            overlay_inset(small, ref_img_bgr, uid, status_text=status)

        cv2.putText(
            small,
            status,
            (PADDING, 30),
            cv2.FONT_HERSHEY_SIMPLEX,
            0.9,
            color,
            2,
            cv2.LINE_AA,
        )

        remaining = max(0.0, delay_sec - (time.time() - t0))
        cv2.putText(
            small,
            f"Closing in {remaining:.1f}s",
            (PADDING, 60),
            cv2.FONT_HERSHEY_SIMPLEX,
            0.6,
            (200, 200, 200),
            2,
            cv2.LINE_AA,
        )

        cv2.imshow(WINDOW_NAME, small)
        key = cv2.waitKey(10) & 0xFF
        if key in (ord("q"), ord("Q")):
            break

    cv2.destroyWindow(WINDOW_NAME)
    cv2.waitKey(1)


def main():
    os.makedirs(USERS_DIR, exist_ok=True)

    assert_model_ok(DET_MODEL, 200_000)
    assert_model_ok(REC_MODEL, 5_000_000)

    face_detector = cv2.FaceDetectorYN.create(DET_MODEL, "", (320, 320), 0.6, 0.3, 5000)
    face_recognizer = cv2.FaceRecognizerSF.create(REC_MODEL, "")

    port = find_arduino_port()
    print(f"[INFO] Arduino port: {port}")

    cap = cv2.VideoCapture(CAMERA_INDEX)
    if not cap.isOpened():
        raise RuntimeError("Cannot open camera.")
    time.sleep(0.5)
    for _ in range(5):
        cap.read()
    print("[INFO] Camera ready.")

    # Check backend connectivity
    try:
        r = requests.get(f"{BACKEND_URL}/api/health", timeout=3)
        print(f"[INFO] Backend connected: {r.json()}")
    except Exception as e:
        print(f"[WARN] Cannot reach backend at {BACKEND_URL}: {e}")

    try:
        with serial.Serial(port, BAUD, timeout=0.2) as ser:
            time.sleep(2.0)
            ser.reset_input_buffer()
            print("[INFO] Waiting for UID...")

            while True:
                cap.grab()

                line = ser.readline().decode(errors="ignore").strip()
                if not line:
                    continue

                print(f"[ARDUINO] {line}")

                if not line.startswith("UID,"):
                    continue

                uid = line.split(",", 1)[1].strip()
                print(f"[INFO] Card scanned: {uid}")

                # Check if card is registered in the backend
                auth = backend_authenticate(uid)
                if not auth.get("success") and "not registered" in auth.get("message", "").lower():
                    print(f"[WARN] Card {uid} not registered in system")
                    ser.write(b"RESULT,UNKNOWN_CARD\n")
                    continue

                ref_path = uid_to_path(uid)
                if not os.path.exists(ref_path):
                    print(f"[WARN] Missing photo: {ref_path} \u2014 logging as unverified")
                    action = backend_get_suggested_action(uid)
                    ser.write(b"RESULT,CHECKED_UNVERIFIED_NO_PHOTO\n")
                    backend_log_access(uid, action, True, method="Card Only (Unverified)")
                    print(f"[INFO] Logged as: {action} (unverified, no photo)")
                    continue

                ref_img = cv2.imread(ref_path)
                if ref_img is None:
                    print("[WARN] Cannot load reference image \u2014 logging as unverified")
                    action = backend_get_suggested_action(uid)
                    ser.write(b"RESULT,CHECKED_UNVERIFIED_NO_PHOTO\n")
                    backend_log_access(uid, action, True, method="Card Only (Unverified)")
                    print(f"[INFO] Logged as: {action} (unverified, bad image)")
                    continue

                ref_small = resize_to_window(ref_img)
                rh, rw = ref_small.shape[:2]
                face_detector.setInputSize((rw, rh))
                faces = face_detector.detect(ref_small)[1]
                f = largest_face(faces)

                if f is None:
                    print("[WARN] No face found in reference image \u2014 logging as unverified")
                    action = backend_get_suggested_action(uid)
                    ser.write(b"RESULT,CHECKED_UNVERIFIED_NO_PHOTO\n")
                    backend_log_access(uid, action, True, method="Card Only (Unverified)")
                    print(f"[INFO] Logged as: {action} (unverified, no face in ref)")
                    continue

                ref_emb = get_embedding_from_face(ref_small, f, face_recognizer)

                try:
                    ok, best_sim = verify_live_with_gui(
                        uid,
                        ref_img,
                        ref_emb,
                        face_detector,
                        face_recognizer,
                        cap,
                        seconds=6.0,
                        close_delay=3.0,
                    )
                    print(f"[INFO] best_sim={best_sim:.3f} ok={ok}")

                    if ok:
                        action = backend_get_suggested_action(uid)
                        ser.write(b"RESULT,CHECKED_VERIFIED\n")
                        backend_log_access(uid, action, True, method="Face Recognition")
                        print(f"[INFO] Logged as: {action} (verified)")
                    else:
                        action = backend_get_suggested_action(uid)
                        ser.write(b"RESULT,CHECKED_UNVERIFIED\n")
                        backend_log_access(uid, action, True, method="Face Mismatch (Unverified)")
                        print(f"[INFO] Logged as: {action} (unverified)")

                except KeyboardInterrupt:
                    print("[INFO] Stopped by user.")
                    return
                except Exception as e:
                    print(f"[ERROR] {e}")
                    ser.write(b"RESULT,ERROR\n")
    finally:
        cap.release()
        cv2.destroyAllWindows()


if __name__ == "__main__":
    main()
