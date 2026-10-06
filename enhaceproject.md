ได้ครับ ถ้า AI อีกตัว **ยังไม่ได้เริ่มทำอะไรเลย** ควรให้เป็น prompt เดียวที่บอกทั้ง **บริบท → architecture ที่ต้องการ → สิ่งที่ต้องตรวจ → สิ่งที่ต้องแก้ → วิธีทดสอบ** เพื่อให้มัน inspect โปรเจกต์ก่อนลงมือ ไม่ใช่กระโดดไปแก้ไฟล์ใดไฟล์หนึ่ง

เอาอันนี้ส่งได้เลย:

คุณได้รับ task ให้ตรวจสอบและแก้ไขระบบ Face Recognition ของโปรเจกต์ DentistChair ให้ทำงาน End-to-End จริง

**สำคัญ: ตอนนี้ยังไม่มีการแก้ code ใด ๆ ให้เริ่มจากการ inspect โปรเจกต์และทำความเข้าใจ architecture ปัจจุบันก่อน แล้วจึงค่อยแก้**

ห้ามเดาโครงสร้างจาก prompt นี้เพียงอย่างเดียว ให้ตรวจสอบ source code, database migration, Docker Compose และ frontend จริงก่อนลงมือ

---

# เป้าหมายหลัก

ต้องทำให้ระบบสามารถ:

```text
Register Dentist
    ↓
Upload รูปใบหน้าไป MinIO
    ↓
สร้าง Face Embedding
    ↓
บันทึก Embedding ใน PostgreSQL
    ↓
Live Camera ส่งภาพผ่าน WebSocket
    ↓
Python ตรวจใบหน้า + สร้าง Embedding
    ↓
เปรียบเทียบกับ Embedding ที่ลงทะเบียนไว้
    ↓
ระบุ Dentist ID / Dentist Name
    ↓
ส่งผลกลับ Frontend
    ↓
แสดงชื่อ Dentist บนกรอบใบหน้า
```

ต้องเป็นระบบจริง ไม่ใช่ mock หรือ hard-code

---

# Architecture ที่ต้องการ

## 1. MinIO

**MinIO ใช้เก็บรูปต้นฉบับเท่านั้น**

ตัวอย่าง:

```text
dentist-images/
└── dentists/
    ├── 1/
    │   └── face.jpg
    └── 2/
        └── face.jpg
```

PostgreSQL จะเก็บเพียง object path/key เช่น:

```text
dentists/1/face.jpg
```

ผ่าน column:

```text
dentists.image_path
```

**ห้ามเปลี่ยนไปเก็บรูปใหม่เป็น Base64 ใน PostgreSQL**

---

# 2. PostgreSQL

ใช้ table `dentists` ที่มีอยู่แล้วเป็น Dentist table หลัก

Current schema:

```sql
CREATE TABLE dentists (
    dentist_id SERIAL PRIMARY KEY,
    image_base64 TEXT,
    chair_height DOUBLE PRECISION NOT NULL,
    dentist_name VARCHAR(255),
    body_composition_data TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

และ migration ปัจจุบันเพิ่ม:

```sql
ALTER TABLE dentists
    ADD COLUMN IF NOT EXISTS image_path TEXT,
    ADD COLUMN IF NOT EXISTS pose_landmarks JSONB,
    ADD COLUMN IF NOT EXISTS ergonomics_sit_height NUMERIC,
    ADD COLUMN IF NOT EXISTS update_at TIMESTAMPTZ NOT NULL DEFAULT now();
```

มี table:

```sql
CREATE TABLE IF NOT EXISTS tb_face_embedding (
    id SERIAL PRIMARY KEY,
    dentist_id INTEGER NOT NULL REFERENCES dentists(dentist_id) ON DELETE CASCADE,
    embedding JSONB NOT NULL,
    model_version TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

และ:

```sql
CREATE INDEX IF NOT EXISTS idx_face_embedding_dentist_id
ON tb_face_embedding(dentist_id);
```

## ห้าม

- สร้าง `tb_dentist`
- rename `dentists`
- สร้าง Dentist table ใหม่
- เปลี่ยน schema โดยไม่จำเป็น
- เปลี่ยน `camera_history_logs` หรือ `adjustable_items` เพราะไม่เกี่ยวกับ task นี้

ใช้:

```text
dentists.dentist_id
```

เป็น primary key ของ Dentist

และ:

```text
tb_face_embedding.dentist_id
    →
dentists.dentist_id
```

เป็น relationship

---

# 3. Face Embedding

ต้องใช้ embedding model เดียวกันทั้ง Registration และ Live Camera

ให้ใช้:

```python
face_recognition.face_encodings()
```

เป็น implementation หลัก

ดังนั้น:

```text
Registration
    ↓
face_recognition
    ↓
128D embedding
```

และ:

```text
Live Camera
    ↓
face_recognition
    ↓
128D embedding
```

ต้องเป็น model / implementation เดียวกัน

**ห้าม** ใช้ Hugging Face embedding ใน Registration แล้วเอาไปเทียบกับ `face_recognition` embedding จาก Live Camera

ตรวจสอบ code ปัจจุบันก่อนว่า Hugging Face ถูกใช้ตรงไหน และแก้ให้ registration ใช้ Python `face_recognition` flow เดียวกับ live camera

---

# 4. Registration Flow

ตรวจสอบ route ปัจจุบันที่ใช้สร้าง Dentist โดยเฉพาะ:

```text
POST /api/dentist
```

และ service/repository ที่ route นี้เรียก

Registration ที่ถูกต้องควรเป็น:

```text
1. รับ Dentist data
2. รับ face image
3. Validate image
4. ส่ง image ไป Python service
5. Detect face
6. ต้องพบ face เพียง 1 หน้า
7. Generate 128D embedding
8. Validate embedding dimension = 128
9. Create dentist record ใน PostgreSQL
10. Upload original image ไป MinIO
11. Update dentists.image_path
12. Insert embedding ลง tb_face_embedding
13. Commit transaction
14. Return success
```

ถ้าไม่พบหน้า:

```text
no face detected
```

ถ้าพบหลายหน้า:

```text
multiple faces detected
```

ต้อง return error ที่เหมาะสม

ห้ามเลือกหน้าแบบสุ่ม

---

# 5. Database Transaction

ตรวจสอบให้ registration ไม่เกิดข้อมูลค้างครึ่ง ๆ กลาง ๆ

ตัวอย่าง flow ที่เหมาะสม:

```text
Create Dentist
    ↓
Get dentist_id
    ↓
Generate embedding
    ↓
Upload image to MinIO
    ↓
Save image_path
    ↓
Save embedding
    ↓
Commit
```

ถ้า step ใดล้มเหลว ต้อง handle error อย่างถูกต้อง

เนื่องจาก MinIO ไม่ใช่ PostgreSQL transaction หาก upload สำเร็จแต่ DB transaction fail ให้พิจารณา cleanup object ที่ orphan ด้วย

อย่าให้ registration สำเร็จโดยมี:

```text
Dentist ไม่มี embedding
```

หรือ:

```text
Dentist มี image_path แต่ไม่มี image จริง
```

โดยไม่แจ้ง error

---

# 6. `image_base64`

Column นี้มีอยู่ใน schema เดิม:

```text
dentists.image_base64
```

**อย่าเพิ่ง DROP column**

ก่อนอื่นค้นหาทั้ง repository ว่ามี code ไหนใช้งานอยู่หรือไม่

ถ้าเป็น legacy code ให้เลิกใช้ใน registration ใหม่ และใช้:

```text
MinIO
    +
dentists.image_path
```

แทน

ถ้าจะลบ column ต้องทำเป็น migration แยก และต้องมั่นใจว่าไม่มี code อื่นใช้งาน

---

# 7. Face Embedding Database

Embedding ต้องบันทึกลง:

```text
tb_face_embedding
```

เช่น:

```text
dentist_id = 1
embedding = [128 values]
model_version = "face_recognition_dlib_128d_v1"
```

ต้อง validate:

```text
len(embedding) == 128
```

ทั้งตอน save และตอน load

`model_version` ต้องใช้จริง เพื่อป้องกันการเอา embedding จากคนละ model มา compare กัน

---

# 8. `ListDentists()`

ตรวจสอบ implementation ปัจจุบันของ:

```text
ListDentists()
```

โดยเฉพาะปัญหา SELECT และ `rows.Scan()` ที่อาจมีจำนวน arguments ไม่ตรงกัน

แก้ให้ตรงกัน 100%

และสำคัญที่สุด:

`ListDentists()` หรือ repository function ที่ matcher ใช้ ต้องสามารถโหลด Dentist พร้อม Face Embedding ได้

ข้อมูลที่ matcher ต้องได้อย่างน้อย:

```text
dentist_id
dentist_name
image_path
embedding
model_version
```

สามารถใช้ JOIN:

```sql
dentists
LEFT JOIN tb_face_embedding
ON tb_face_embedding.dentist_id = dentists.dentist_id
```

หรือวิธีอื่นที่เหมาะสมกับ codebase

ถ้า Dentist 1 คนมีหลาย embeddings ให้ handle ความสัมพันธ์ 1:N อย่างถูกต้อง ไม่ให้ duplicate Dentist โดยไม่ตั้งใจ

---

# 9. Matcher

ตรวจสอบ:

```text
backend/internal/service/matcher.go
```

Matcher ต้องทำ:

```text
Live embedding
      ↓
Validate 128D
      ↓
Load registered embeddings
      ↓
Validate model compatibility
      ↓
Calculate Euclidean distance
      ↓
Find lowest distance
      ↓
Compare threshold
```

Threshold ปัจจุบันใช้:

```text
0.6
```

ไปก่อน

ห้ามเปลี่ยน threshold เพื่อแก้ปัญหาที่เกิดจาก embedding model ไม่ตรงกัน

ผลลัพธ์:

ถ้า match:

```json
{
  "label": "Dentist Name",
  "matched": true,
  "distance": 0.42
}
```

ถ้าไม่ match:

```json
{
  "label": "unknown",
  "matched": false,
  "distance": 0.83
}
```

---

# 10. Python Service

Current Docker Compose มี:

```text
python-service
```

ให้ตรวจสอบ:

```text
backend/python_service/main.py
```

ต้องทำให้ Registration และ Live Camera ใช้ implementation สำหรับ embedding เดียวกัน

ควรมี shared function แนวคิด:

```python
def extract_face_embedding(image):
    ...
```

แล้วทั้งสอง endpoint ใช้ function เดียวกัน

ต้องตรวจสอบ:

- face detection
- face encoding
- 128D output
- multiple faces
- no face
- invalid image
- error response

---

# 11. Docker Compose

Current services:

```text
minio
minio-init
db
backend
python-service
frontend
```

คง architecture นี้ไว้

Backend มี:

```yaml
PYTHON_SERVICE_URL: http://python-service:8000
MINIO_ENDPOINT: minio:9000
MINIO_BUCKET: dentist-images
```

ตรวจสอบว่าทั้งหมดถูกใช้จริง

มี environment variable:

```yaml
FACE_EMBEDDING_ENDPOINT: http://face-embedding:8000/embed
```

แต่ปัจจุบันไม่มี service ชื่อ:

```text
face-embedding
```

ดังนั้นต้อง inspect code ก่อน

ถ้าไม่มี code path ใช้ตัวแปรนี้:

```text
remove it
```

ถ้ามี code path ใช้:

```text
เปลี่ยนให้ใช้ python-service
```

ไม่ต้องสร้าง service `face-embedding` เพิ่ม ถ้า `python-service` รองรับ embedding อยู่แล้ว

เป้าหมายคือ:

```text
python-service
├── Face Detection
└── Face Embedding
```

---

# 12. MinIO Init

ตรวจสอบ `minio-init`

Current bucket:

```text
dentist-images
```

ทำให้ bucket creation สามารถ run ซ้ำได้โดยไม่ fail หาก bucket มีอยู่แล้ว

เช่นใช้ equivalent ของ:

```bash
mc mb --ignore-existing
```

หรือ implementation ที่เหมาะสม

นอกจากนี้ตรวจสอบ:

```bash
mc anonymous set download
```

เนื่องจากรูปใบหน้าเป็นข้อมูลที่ควรระวัง

อย่าเปิด bucket public โดยไม่จำเป็น

แต่ **อย่าเปลี่ยน authentication architecture ใหญ่ ๆ** ถ้าไม่จำเป็นต่อ task นี้

---

# 13. Live Camera Flow

ตรวจสอบ:

```text
camera_live_ws_handler.go
camera_live_service.go
pythonclient.go
```

ต้องได้:

```text
Frontend
   ↓
WebSocket /ws/camera/live
   ↓
Go Backend
   ↓
Python Service
   ↓
Face Detection
   ↓
Face Embedding
   ↓
Matcher
   ↓
Result
   ↓
WebSocket
   ↓
Frontend
```

Frontend ไม่ควรต้องโหลดรูปจาก MinIO ทุก frame

---

# 14. Frontend

ตรวจสอบ:

```text
LiveMonitorPage.jsx
```

ไม่ต้อง rewrite ถ้า implementation ปัจจุบันใช้ได้อยู่แล้ว

ต้องรองรับผลลัพธ์:

```json
{
  "faces": [
    {
      "bbox": {},
      "label": "Dentist Name",
      "matched": true,
      "distance": 0.42
    }
  ]
}
```

และแสดง:

```text
Dentist Name
```

บน bbox

ถ้า unknown:

```text
unknown
```

---

# 15. Other Database Tables

อย่าแก้โดยไม่จำเป็น:

```text
camera_history_logs
adjustable_items
```

แต่ตรวจสอบ compatibility หากหลัง Face Recognition มีการส่ง `dentist_id` ไปบันทึก history

ปัจจุบัน:

```text
dentists.dentist_id = INTEGER
camera_history_logs.dentist_id = VARCHAR(100)
```

ห้ามเปลี่ยน schema เพียงเพราะ type ต่างกันใน task นี้

ถ้าต้องใช้ ให้ handle conversion ใน application layer อย่างปลอดภัย

---

# 16. Do NOT use pgvector

ปัจจุบัน:

```text
tb_face_embedding.embedding = JSONB
```

ให้ใช้ JSONB ต่อไป

ไม่ต้องติดตั้ง pgvector ใน task นี้

ไม่ต้องเปลี่ยนเป็น:

```text
vector(512)
```

เพราะ Face Recognition ที่กำหนดให้ใช้ได้ 128 dimensions

สำหรับจำนวน Dentist ใน MVP ให้:

```text
PostgreSQL
    ↓
Load embeddings
    ↓
Go matcher
    ↓
Euclidean distance
```

เพียงพอแล้ว

---

# 17. Important Existing Bug

จาก previous inspection พบปัญหาที่ต้องตรวจสอบและแก้:

### `CreateDentist()`

มี INSERT ที่จำนวน columns / placeholders / arguments ไม่ตรงกัน

ต้องแก้ให้ตรงกับ schema จริง

### `ListDentists()`

SELECT และ `rows.Scan()` ไม่ตรงกัน

ต้องแก้

### Registration

ปัจจุบัน registration และ live camera อาจใช้ embedding คนละ implementation

ต้องทำให้ใช้ `face_recognition` เดียวกัน

### `tb_face_embedding`

Registration บันทึก embedding ที่นี่ แต่ matcher เดิมอาจไม่ได้โหลด table นี้

ต้องเชื่อมให้เรียบร้อย

### Migration

ตรวจสอบว่า `tb_face_embedding` ถูกสร้างจริงตอน application startup

ไม่ใช่แค่มี SQL file แต่ไม่ได้ execute

---

# 18. Testing หลังแก้

หลังแก้ code ต้องทดสอบจริง

เริ่มจาก:

```bash
docker compose down -v
docker compose up --build
```

จากนั้นตรวจ:

### Database

ต้องมี:

```text
dentists
tb_face_embedding
camera_history_logs
adjustable_items
```

### Registration

Register Dentist 1 คน

ตรวจว่า:

```text
MinIO
└── face image exists

PostgreSQL
├── dentists
│   └── dentist record
│       └── image_path
│
└── tb_face_embedding
    └── embedding length = 128
```

### Load

เรียก repository / service ที่ matcher ใช้

ต้องได้:

```text
dentist_id
dentist_name
embedding[128]
model_version
```

ไม่มี Scan error

### Matching

ใช้รูปเดียวกับตอน register

ต้อง:

```text
matched = true
label = Dentist name
distance < 0.6
```

### Unknown

ใช้รูปของคนที่ไม่ได้ register

ต้อง:

```text
matched = false
label = unknown
```

### Live Camera

เปิด Live Monitor และตรวจจริงว่า:

```text
Camera
 ↓
WebSocket
 ↓
Python
 ↓
Embedding
 ↓
Matcher
 ↓
Dentist Name
 ↓
Frontend
```

ทำงานครบ

---

# 19. Definition of Done

งานถือว่าเสร็จต่อเมื่อ flow นี้ทำงานจริง:

```text
                 REGISTER
                     │
                     ▼
                Face Image
                  /     \
                 /       \
                ▼         ▼
             MinIO      Python
              │           │
              │        128D embedding
              │           │
              ▼           ▼
        image_path    tb_face_embedding
              │           │
              └─────┬─────┘
                    ▼
                 dentists


                  LIVE
                    │
                    ▼
                 Camera
                    │
                    ▼
             WebSocket
                    │
                    ▼
             Python Service
                    │
                    ▼
             128D embedding
                    │
                    ▼
          Registered Embeddings
                    │
                    ▼
              Face Matcher
                 /      \
                /        \
          matched       unknown
             │
             ▼
       Dentist Name
             │
             ▼
          Frontend
```

---

# 20. Constraints

ห้าม:

- สร้าง `tb_dentist`
- เก็บรูปใหม่ใน PostgreSQL เป็น Base64
- ใช้ Hugging Face embedding เทียบกับ `face_recognition`
- โหลดรูปจาก MinIO ทุก frame
- hard-code Dentist name
- hard-code embedding
- fake matching
- bypass database
- เพิ่ม embedding service ซ้ำ
- เพิ่ม pgvector โดยไม่จำเป็น
- rewrite Frontend โดยไม่จำเป็น
- เปลี่ยน WebSocket protocol โดยไม่จำเป็น
- แก้ unrelated tables
- แก้ threshold เพื่อซ่อน model mismatch

---

# 21. วิธีการทำงานที่ต้องการจากคุณ

ทำตามลำดับนี้:

### Phase 1 — Inspect

อ่าน source code และ schema ทั้งหมดที่เกี่ยวข้องก่อน

### Phase 2 — Report

สรุป:

```text
Current Architecture
Current Registration Flow
Current Live Camera Flow
Current Database Flow
Current Problems
```

### Phase 3 — Implement

แก้ code ให้ตรงกับ architecture นี้

### Phase 4 — Test

Build และ test จริง

### Phase 5 — Verify

ตรวจสอบ End-to-End ตั้งแต่:

```text
MinIO
→ PostgreSQL
→ Python
→ Matcher
→ WebSocket
→ Frontend
```

### Phase 6 — Final Report

รายงาน:

| Component | Status | Evidence |
|---|---|---|
| Database migration | PASS/FAIL | |
| MinIO | PASS/FAIL | |
| Dentist registration | PASS/FAIL | |
| Image upload | PASS/FAIL | |
| Embedding generation | PASS/FAIL | |
| 128D validation | PASS/FAIL | |
| Embedding storage | PASS/FAIL | |
| Embedding loading | PASS/FAIL | |
| Matcher | PASS/FAIL | |
| Live Camera | PASS/FAIL | |
| Frontend identification | PASS/FAIL | |

ถ้ายังมีปัญหา ให้ระบุ:

```text
File:
Line:
Error:
Root Cause:
Fix:
Test Result:
```

**ห้ามบอกว่า PASS หรือ Done หากยังไม่ได้ทดสอบจริง**

เริ่มจาก **Phase 1: Inspect project ก่อน ห้ามแก้ code ทันที**