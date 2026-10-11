ปรับปรุง `LiveMonitorPage` จากโค้ดปัจจุบันที่ให้ไป โดยมีเป้าหมายให้หน้าเว็บสามารถเลือกแหล่งกล้องได้ 2 แบบ:

1. Local Camera — กล้อง Mac / iPhone / USB Camera ที่ Browser มองเห็น
2. Tapo Camera (RTSP) — กล้อง Tapo ที่ส่งภาพผ่าน RTSP

**สำคัญ:** ห้ามรื้อหรือเปลี่ยน logic ที่มีอยู่ของ MediaPipe, Canvas overlay, Face Recognition, WebSocket และการส่ง `imageBase64` ไป Backend โดยไม่จำเป็น ให้เปลี่ยนเฉพาะส่วน "แหล่งที่มาของ `<video>`" เท่านั้น

### Architecture ที่ต้องการ

```text
                    Camera Source
                         │
              ┌──────────┴──────────┐
              │                     │
        Local Camera            Tapo RTSP
              │                     │
      getUserMedia()             RTSP
              │                     │
              │                 MediaMTX
              │                     │
              │                  WebRTC
              │                     │
              └──────────┬──────────┘
                         ↓
                      <video>
                         │
              ┌──────────┴──────────┐
              ↓                     ↓
          MediaPipe              Canvas
                                    │
                                    ↓
                                WebSocket
                                    │
                                    ↓
                            Face Recognition
```

### 1. เพิ่ม Camera Source selector

เพิ่ม state:

```js
const [cameraSource, setCameraSource] = useState("local");
```

เพิ่ม UI:

```jsx
<label>
  Camera Source:
  <select
    value={cameraSource}
    onChange={(e) => setCameraSource(e.target.value)}
  >
    <option value="local">Local Camera</option>
    <option value="tapo">Tapo Camera (RTSP)</option>
  </select>
</label>
```

เมื่อเปลี่ยน option ต้องหยุด stream / connection เดิมก่อน แล้ว initialize source ใหม่อย่างถูกต้อง เพื่อไม่ให้เกิด camera stream หรือ WebSocket ค้าง

### 2. Local Camera ต้องทำงานเหมือนเดิม

เมื่อ:

```js
cameraSource === "local"
```

ให้ใช้ logic เดิม:

```js
navigator.mediaDevices.getUserMedia(...)
```

และยังรองรับ `selectedDeviceId` สำหรับเลือก Mac Camera / iPhone Camera / USB Camera เหมือนเดิม

อย่าลบ logic `enumerateDevices()` ถ้ายังจำเป็นสำหรับ Local Camera

### 3. เพิ่ม Tapo RTSP Camera

เมื่อ:

```js
cameraSource === "tapo"
```

ห้ามทำ:

```js
video.src = "rtsp://..."
```

และห้ามพยายามใช้ `<video>` เล่น RTSP โดยตรง เพราะ Browser ไม่รองรับ RTSP โดยตรง

ให้ใช้ architecture:

```text
Tapo Camera
    ↓ RTSP
MediaMTX / RTSP Gateway
    ↓ WebRTC
Browser
    ↓
videoRef.current
```

ให้แยกส่วน WebRTC connection เป็น function/hook ที่ชัดเจน เช่น:

```js
connectTapoCamera()
disconnectTapoCamera()
```

โดย URL ของ RTSP และ WebRTC endpoint ต้องกำหนดเป็น configuration/environment variable ไม่ hard-code username/password ลงใน React source code

ตัวอย่างแนวคิด:

```env
VITE_TAPO_WEBRTC_URL=...
```

หมายเหตุ: Frontend ไม่ควรเปิด RTSP URL ที่มี username/password โดยตรง

### 4. `<video>` ต้องเป็นจุดกลางของทั้งสอง source

ไม่ว่าเลือก Camera Source ไหน สุดท้ายต้องได้:

```js
videoRef.current
```

เป็น HTMLVideoElement ที่กำลังแสดงภาพกล้อง

ดังนั้น logic ต่อไปนี้ต้องยังใช้ `<video>` เดิม:

```js
holisticRef.current.detectForVideo(video, now);
```

และ:

```js
ctx.drawImage(
  video,
  0,
  0,
  CAPTURE_WIDTH,
  CAPTURE_HEIGHT
);
```

### 5. ห้ามเปลี่ยน MediaPipe logic

คง:

```js
HolisticLandmarker
FilesetResolver
DrawingUtils
```

และ logic:

```js
detectForVideo(video, now)
```

เอาไว้เหมือนเดิม

ไม่ต้องเปลี่ยน model หรือ detection architecture

### 6. ห้ามเปลี่ยน Face Recognition WebSocket logic

หลังจากได้ภาพจาก `<video>` แล้ว ให้ใช้ flow เดิม:

```text
video
 ↓
captureCanvas
 ↓
canvas.toDataURL("image/jpeg", 0.7)
 ↓
imageBase64
 ↓
WebSocket
 ↓
Backend
```

โค้ดลักษณะนี้ต้องยังทำงาน:

```js
ws.send(JSON.stringify({
  imageBase64,
  poseLandmarks: poseLandmarksRef.current
}));
```

Backend ไม่ควรต้องเปลี่ยนเพียงเพราะเปลี่ยน camera source

### 7. UI

เมื่อเลือก Local Camera:

```text
Camera Source: Local Camera
Camera: [Mac Camera ▼]
```

ให้แสดง device selector เดิม

เมื่อเลือก Tapo:

```text
Camera Source: Tapo Camera (RTSP)
```

ให้ซ่อน Local Camera device selector เพราะไม่จำเป็น

### 8. Status

ใช้ status เดิม:

```text
Idle
Connecting...
Live
Camera or connection error
```

แต่ต้องทำให้ครอบคลุมทั้ง Local Camera และ Tapo/WebRTC connection

ถ้า Tapo WebRTC connection ล้มเหลว ให้แสดง:

```text
Camera or connection error
```

และ log error ที่ console เพื่อ debug ได้

### 9. Cleanup สำคัญมาก

เมื่อ:

- เปลี่ยน Local → Tapo
- เปลี่ยน Tapo → Local
- component unmount
- WebRTC connection ปิด
- WebSocket ปิด

ต้อง cleanup ให้ถูกต้อง:

- stop local MediaStream tracks
- close WebRTC PeerConnection
- clear timers
- close WebSocket
- clear reconnect timers
- ไม่ให้ connection เก่ากลับมา reconnect หลัง component ถูก unmount

ห้ามเกิด camera / WebRTC / WebSocket connection ซ้อนกัน

### 10. ห้ามแก้ส่วนที่ไม่เกี่ยวข้อง

อย่าเปลี่ยน:

- Face Recognition backend API
- WebSocket message format
- `imageBase64`
- `poseLandmarks`
- MediaPipe model
- Canvas bounding boxes
- Pose landmarks
- Face recognition result format
- Database
- Backend

เว้นแต่จำเป็นจริง ๆ เพื่อรองรับ camera source

### 11. ผลลัพธ์ที่ต้องการ

สุดท้ายผู้ใช้สามารถเปิดหน้า Live Monitor แล้วเลือก:

```text
Camera Source:
[ Local Camera ▼ ]
```

หรือ:

```text
Camera Source:
[ Tapo Camera (RTSP) ▼ ]
```

ถ้าเลือก Local:

```text
Local Camera
    ↓
getUserMedia
    ↓
<video>
    ↓
MediaPipe + Face Recognition
```

ถ้าเลือก Tapo:

```text
Tapo
    ↓
RTSP
    ↓
MediaMTX
    ↓
WebRTC
    ↓
<video>
    ↓
MediaPipe + Face Recognition
```

และทั้งสองกรณีต้องใช้ pipeline หลัง `<video>` ร่วมกันทั้งหมด

### สำคัญที่สุด

ก่อนเขียน code ให้ตรวจสอบโครงสร้าง project ก่อนว่า frontend ใช้ Vite/React และ backend / Docker / MediaMTX มีอยู่แล้วหรือไม่

ถ้ายังไม่มี MediaMTX ให้ระบุไฟล์ที่ต้องเพิ่ม เช่น:

```text
docker-compose.yml
mediamtx.yml
```

และบอกวิธีตั้งค่า RTSP source ของ Tapo

อย่าเดา URL ของ Tapo และอย่า hard-code password

ให้ทำ implementation ให้สามารถสลับ Local Camera ↔ Tapo RTSP ได้จริง และแสดง code diff/ไฟล์ที่แก้ไขทั้งหมด พร้อมอธิบายว่าแต่ละไฟล์เปลี่ยนอะไร