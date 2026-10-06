import base64
import io
import binascii
from typing import List, Optional

import face_recognition
import mediapipe as mp
import numpy as np
from fastapi import FastAPI, HTTPException, Request
from pydantic import BaseModel
from PIL import Image

app = FastAPI()

mp_holistic = mp.solutions.holistic


# ---------- schemas ----------

class RegisterRequest(BaseModel):
    imageBase64: str
    dentistName: str


class RegisterResponse(BaseModel):
    faceEmbedding: List[float]
    chairHeight: float
    modelVersion: str


class DetectRequest(BaseModel):
    imageBase64: str


class DetectedFace(BaseModel):
    bbox: List[int]  # [x, y, w, h] in pixels
    embedding: List[float]


class DetectResponse(BaseModel):
    faces: List[DetectedFace]


class EmbeddingResponse(BaseModel):
    embedding: List[float]
    modelVersion: str


MODEL_VERSION = "face_recognition_dlib_128d_v1"


# ---------- helpers ----------

def decode_base64_image(image_base64: str) -> np.ndarray:
    if "," in image_base64:
        # strip a "data:image/jpeg;base64," prefix if the frontend sent one
        image_base64 = image_base64.split(",", 1)[1]
    try:
        image_bytes = base64.b64decode(image_base64, validate=True)
        return decode_image_bytes(image_bytes)
    except (binascii.Error, ValueError) as exc:
        raise HTTPException(status_code=400, detail="invalid image data") from exc


def decode_image_bytes(image_bytes: bytes) -> np.ndarray:
    try:
        return np.array(Image.open(io.BytesIO(image_bytes)).convert("RGB"))
    except Exception as exc:
        raise HTTPException(status_code=400, detail="invalid image") from exc


def extract_face_embedding(image: np.ndarray) -> List[float]:
    locations = face_recognition.face_locations(image)
    if not locations:
        raise HTTPException(status_code=422, detail="no face detected")
    if len(locations) != 1:
        raise HTTPException(status_code=422, detail="multiple faces detected")
    encodings = face_recognition.face_encodings(image, locations)
    if len(encodings) != 1 or len(encodings[0]) != 128:
        raise HTTPException(status_code=422, detail="could not create a valid 128D face embedding")
    return encodings[0].tolist()


def estimate_chair_height(pose_landmarks) -> float:
    """Rough ergonomic estimate from shoulder-to-hip distance (normalized coords).
    TODO: replace with a properly calibrated formula once real measurements
    are available - this is a placeholder ratio, not a validated ergonomic model.
    """
    if pose_landmarks is None:
        return 75.0
    landmarks = pose_landmarks.landmark
    shoulder_y = landmarks[11].y  # left shoulder
    hip_y = landmarks[23].y       # left hip
    torso_ratio = abs(hip_y - shoulder_y)
    return round(65.0 + torso_ratio * 40.0, 1)


# ---------- routes ----------

@app.post("/register", response_model=RegisterResponse)
def register(req: RegisterRequest):
    image = decode_base64_image(req.imageBase64)
    embedding = extract_face_embedding(image)

    chair_height = 75.0
    with mp_holistic.Holistic(static_image_mode=True) as holistic:
        result = holistic.process(image)
        chair_height = estimate_chair_height(result.pose_landmarks)

    return RegisterResponse(faceEmbedding=embedding, chairHeight=chair_height, modelVersion=MODEL_VERSION)


@app.post("/embed", response_model=EmbeddingResponse)
async def embed(request: Request):
    image = decode_image_bytes(await request.body())
    return EmbeddingResponse(embedding=extract_face_embedding(image), modelVersion=MODEL_VERSION)


@app.post("/detect", response_model=DetectResponse)
def detect(req: DetectRequest):
    image = decode_base64_image(req.imageBase64)

    face_locations = face_recognition.face_locations(image)
    face_encodings = face_recognition.face_encodings(image, face_locations)

    faces = []
    for (top, right, bottom, left), encoding in zip(face_locations, face_encodings):
        faces.append(DetectedFace(
            bbox=[left, top, right - left, bottom - top],
            embedding=encoding.tolist(),
        ))

    return DetectResponse(faces=faces)


@app.get("/health")
def health():
    return {"status": "ok"}
