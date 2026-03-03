import os

DET_MODEL = r"models\face_detection_yunet_2023mar.onnx"
REC_MODEL = r"models\face_recognition_sface_2021dec.onnx"

for p in [DET_MODEL, REC_MODEL]:
    print(p, "exists=", os.path.exists(p), "size=", os.path.getsize(p) if os.path.exists(p) else None)
    if os.path.exists(p):
        with open(p, "rb") as f:
            head = f.read(80)
        print("head:", head[:60])