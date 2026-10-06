import React, { useEffect, useRef, useState } from 'react';
import {
    FilesetResolver,
    HolisticLandmarker,
    DrawingUtils
} from "@mediapipe/tasks-vision";

// FIX (part 1): HAND_CONNECTIONS / POSE_CONNECTIONS / FACE_LANDMARKS_TESSELATION
// do NOT exist in "@mediapipe/drawing_utils" (that legacy package only exports
// the drawing functions themselves, not these constants). They now live as
// static members on the tasks-vision landmarker classes.
//
// FIX (part 2, the actual reason it was STILL not showing after part 1):
// @mediapipe/drawing_utils's drawConnectors() reads each connection as an
// array tuple `conn[0]` / `conn[1]` - the OLD MediaPipe solutions format
// (e.g. [[11,13],[13,15]]). But tasks-vision's connection constants are
// arrays of {start, end} OBJECTS. `{start:0,end:1}[0]` is `undefined`, so
// every connection was silently skipped - no error, just nothing drawn.
// The fix is to stop using @mediapipe/drawing_utils entirely and use the
// DrawingUtils class that ships inside @mediapipe/tasks-vision itself, whose
// drawConnectors() correctly reads `.start` / `.end`.
const POSE_CONNECTIONS = HolisticLandmarker.POSE_CONNECTIONS;
const HAND_CONNECTIONS = HolisticLandmarker.HAND_CONNECTIONS;
const FACE_LANDMARKS_TESSELATION = HolisticLandmarker.FACE_LANDMARKS_TESSELATION;

// Adjust if your backend is served from a different host than the frontend.
const WS_PATH = '/ws/camera/live';

// Resolution the frame is captured/sent at. Bounding boxes from the backend
// are in these pixel coordinates, then scaled up to however large the
// <video> is actually displayed on screen.
const CAPTURE_WIDTH = 640;
const CAPTURE_HEIGHT = 480;
const SEND_INTERVAL_MS = 350; // ~3fps - enough for face matching, light on the pipeline
const DETECT_INTERVAL_MS = 33; // cap local holistic inference to ~30fps to avoid saturating rAF

function statusLabel(status) {
  switch (status) {
    case 'connecting':
      return 'Connecting...';
    case 'live':
      return 'Live';
    case 'error':
      return 'Camera or connection error';
    default:
      return 'Idle';
  }
}
function drawBoundingBoxes(
  ctx,
  faces,
  scaleX,
  scaleY
) {
  if (!faces?.length) return;

  faces.forEach((face) => {
    const [x, y, w, h] = face.bbox;

    const boxX = x * scaleX;
    const boxY = y * scaleY;
    const boxW = w * scaleX;
    const boxH = h * scaleY;

    ctx.strokeStyle = face.matched
      ? "#1D9E75"
      : "#E24B4A";

    ctx.lineWidth = 2;

    ctx.strokeRect(
      boxX,
      boxY,
      boxW,
      boxH
    );

    const label = face.label || "unknown";

    ctx.font = "14px sans-serif";

    const textWidth =
      ctx.measureText(label).width;

    ctx.fillStyle = face.matched
      ? "#1D9E75"
      : "#E24B4A";

    ctx.fillRect(
      boxX,
      boxY - 20,
      textWidth + 12,
      20
    );

    ctx.fillStyle = "#fff";

    ctx.fillText(
      label,
      boxX + 6,
      boxY - 5
    );
  });
}

export default function LiveMonitorPage() {
  const videoRef = useRef(null);
  const captureCanvasRef = useRef(null); // hidden - used only to encode frames
  const overlayCanvasRef = useRef(null); // visible - draws boxes on top of the video
  const wsRef = useRef(null);
  const sendTimerRef = useRef(null);
  const reconnectTimerRef = useRef(null);
  const isUnmountedRef = useRef(false);
  const holisticRef = useRef(null);
  const lastDetectTimeRef = useRef(0);

  const [status, setStatus] = useState('idle'); // idle | connecting | live | error
  const [recognitionError, setRecognitionError] = useState('');
  const [faces, setFaces] = useState([]);
  const [devices, setDevices] = useState([]);
  const [selectedDeviceId, setSelectedDeviceId] = useState('');

  const [poseLandmarks, setPoseLandmarks] = useState([]);
  const [faceLandmarks, setFaceLandmarks] = useState([]);
  const [leftHandLandmarks, setLeftHandLandmarks] = useState([]);
  const [rightHandLandmarks, setRightHandLandmarks] = useState([]);

  // list available cameras once permission is granted, so "iPhone Camera"
  // (via Continuity Camera) can be picked instead of the Mac's built-in one
  useEffect(() => {
    let permissionStream;
    let isActive = true;
    let permissionRequested = false;

    async function loadDevices() {
      try {
        // request permission first so device labels are populated
        if (!permissionRequested) {
          permissionRequested = true;
          permissionStream = await navigator.mediaDevices.getUserMedia({ video: true });
        }

        const list = await navigator.mediaDevices.enumerateDevices();
        const cams = list.filter((d) => d.kind === "videoinput");
        if (!isActive) return;
        setDevices(cams);

        const iphone = cams.find((d) =>
          d.label.toLowerCase().includes("iphone")
        );
        if (iphone && !selectedDeviceId) {
          setSelectedDeviceId(iphone.deviceId);
        }
      } catch (err) {
        console.error("Cannot enumerate cameras:", err);
      } finally {
        // This stream only unlocks device labels; the capture effect owns its
        // separate stream. Release it after the initial enumeration.
        permissionStream?.getTracks().forEach((track) => track.stop());
        permissionStream = null;
      }
    }

    loadDevices();
    navigator.mediaDevices?.addEventListener?.("devicechange", loadDevices);

    return () => {
      isActive = false;
      navigator.mediaDevices?.removeEventListener?.("devicechange", loadDevices);
      permissionStream?.getTracks().forEach((track) => track.stop());
    };
  }, []);

  useEffect(() => {
    async function initHolistic() {
      const vision = await FilesetResolver.forVisionTasks(
        "https://cdn.jsdelivr.net/npm/@mediapipe/tasks-vision/wasm"
      );

      holisticRef.current = await HolisticLandmarker.createFromOptions(
        vision,
        {
          baseOptions: {
            modelAssetPath:
              "https://storage.googleapis.com/mediapipe-models/holistic_landmarker/holistic_landmarker/float16/latest/holistic_landmarker.task"
          },
          runningMode: "VIDEO"
          // NOTE: `numPoses` isn't a HolisticLandmarker option (that belongs
          // to PoseLandmarker); Holistic always tracks a single person.
        }
      );
    }

    initHolistic();

    // FIX: dispose the model on unmount, otherwise the WASM/GPU-backed
    // landmarker is leaked (very noticeable during dev hot reloads).
    return () => {
      holisticRef.current?.close();
      holisticRef.current = null;
    };
  }, []);

  useEffect(() => {
    let animationId;

    function detect() {
      const video = videoRef.current;
      const now = performance.now();

      // FIX: Holistic is a heavy model; running it on every rAF tick (~60fps)
      // can saturate the pipeline and cause frame backlog. Cap to ~30fps.
      if (
        holisticRef.current &&
        video &&
        video.readyState >= 2 &&
        now - lastDetectTimeRef.current >= DETECT_INTERVAL_MS
      ) {
        lastDetectTimeRef.current = now;

        const result = holisticRef.current.detectForVideo(video, now);

        setPoseLandmarks(result.poseLandmarks?.[0] ?? []);
        setFaceLandmarks(result.faceLandmarks?.[0] ?? []);
        setLeftHandLandmarks(result.leftHandLandmarks?.[0] ?? []);
        setRightHandLandmarks(result.rightHandLandmarks?.[0] ?? []);
      }

      animationId = requestAnimationFrame(detect);
    }

    detect();

    return () => cancelAnimationFrame(animationId);
  }, []);

  useEffect(() => {
    let stream;

    async function start() {
      try {
        stream = await navigator.mediaDevices.getUserMedia({
          video: selectedDeviceId
            ? { deviceId: { exact: selectedDeviceId }, width: CAPTURE_WIDTH, height: CAPTURE_HEIGHT }
            : { width: CAPTURE_WIDTH, height: CAPTURE_HEIGHT },
          audio: false,
        });
        if (videoRef.current) {
          videoRef.current.srcObject = stream;
        }
        connectSocket();
      } catch (err) {
        console.error('camera access failed', err);
        setStatus('error');
      }
    }

    function connectSocket() {
      setStatus('connecting');
      const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
      const ws = new WebSocket(`${protocol}://${window.location.host}${WS_PATH}`);
      wsRef.current = ws;

      ws.onopen = () => {
        setStatus('live');
        sendTimerRef.current = setInterval(sendFrame, SEND_INTERVAL_MS);
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.error) {
            setRecognitionError(data.error);
            setFaces([]);
          } else if (data.faces) {
            setRecognitionError('');
            setFaces(data.faces);
          }
        } catch (err) {
          console.error('bad message from server', err);
        }
      };

      ws.onerror = () => setStatus('error');

      ws.onclose = () => {
        clearInterval(sendTimerRef.current);
        // FIX: without this guard, a reconnect could still fire ~1.5s after
        // the component unmounted (or the effect re-ran for a device change),
        // opening a dangling socket nothing is listening to anymore.
        if (isUnmountedRef.current) return;
        setStatus('idle');
        reconnectTimerRef.current = setTimeout(connectSocket, 1500);
      };
    }

    function sendFrame() {
      const ws = wsRef.current;
      const video = videoRef.current;
      const canvas = captureCanvasRef.current;
      if (!ws || ws.readyState !== WebSocket.OPEN || !video || video.readyState < 2 || !canvas) return;

      const ctx = canvas.getContext('2d');
      ctx.drawImage(video, 0, 0, CAPTURE_WIDTH, CAPTURE_HEIGHT);
      const imageBase64 = canvas.toDataURL('image/jpeg', 0.7);
      ws.send(JSON.stringify({ imageBase64 }));
    }

    isUnmountedRef.current = false;
    start();

    return () => {
      isUnmountedRef.current = true;
      clearInterval(sendTimerRef.current);
      clearTimeout(reconnectTimerRef.current);
      wsRef.current?.close();
      stream?.getTracks().forEach((t) => t.stop());
    };
  }, [selectedDeviceId]);

  useEffect(() => {
    const video = videoRef.current;
    const overlay = overlayCanvasRef.current;

    if (!video || !overlay) return;

    const displayWidth = video.clientWidth;
    const displayHeight = video.clientHeight;

    overlay.width = displayWidth;
    overlay.height = displayHeight;

    const scaleX = displayWidth / CAPTURE_WIDTH;
    const scaleY = displayHeight / CAPTURE_HEIGHT;

    const ctx = overlay.getContext("2d");

    ctx.clearRect(0, 0, displayWidth, displayHeight);

    //-------------------------
    // MediaPipe
    //-------------------------
    // FIX: draw using tasks-vision's own DrawingUtils (see import comment
    // above) instead of the incompatible @mediapipe/drawing_utils functions.
    const drawingUtils = new DrawingUtils(ctx);

    // drawingUtils.drawConnectors(faceLandmarks, FACE_LANDMARKS_TESSELATION, {
    //   color: "#AAAAAA",
    //   lineWidth: 0.5
    // });

    // drawingUtils.drawLandmarks(faceLandmarks, {
    //   color: "#00FF00",
    //   radius: 1
    // });

    // drawingUtils.drawConnectors(leftHandLandmarks, HAND_CONNECTIONS, {
    //   color: "#00FFFF",
    //   lineWidth: 3
    // });

    // drawingUtils.drawLandmarks(leftHandLandmarks, {
    //   color: "#0000FF",
    //   radius: 3
    // });

    // // FIX: right hand was computed and counted in the UI but never actually
    // // drawn on the overlay canvas.
    // drawingUtils.drawConnectors(rightHandLandmarks, HAND_CONNECTIONS, {
    //   color: "#FFAA00",
    //   lineWidth: 3
    // });

    // drawingUtils.drawLandmarks(rightHandLandmarks, {
    //   color: "#FF00FF",
    //   radius: 3
    // });

    drawingUtils.drawConnectors(poseLandmarks, POSE_CONNECTIONS, {
      color: "#00FF00",
      lineWidth: 4
    });

    drawingUtils.drawLandmarks(poseLandmarks, {
      color: "#FF0000",
      radius: 1
    });

    drawingUtils.close();

    //-------------------------
    // Face Recognition
    //-------------------------

    drawBoundingBoxes(ctx, faces, scaleX, scaleY);
  }, [
    faces,
    poseLandmarks,
    faceLandmarks,
    leftHandLandmarks,
    rightHandLandmarks
  ]);

  return (
    <section className="panel">
      <h2>Live Monitor</h2>
      <p>Status: {statusLabel(status)}</p>
      {recognitionError && <p role="alert">Recognition error: {recognitionError}</p>}
      <p>
        {faces.length > 0
          ? faces.map((face, index) => face.matched
            ? `Recognized: ${face.label}`
            : `Face ${index + 1}: unknown`).join(' · ')
          : status === 'live' ? 'No face detected in the current frame' : 'Waiting for camera connection'}
      </p>

      <label style={{ display: 'block', marginBottom: 8 }}>
        Camera:{' '}
        <select value={selectedDeviceId} onChange={(e) => setSelectedDeviceId(e.target.value)}>
          <option value="">Default camera</option>
          {devices.map((d, index) => (
            <option key={d.deviceId || `camera-${index}`} value={d.deviceId}>
              {d.label || `Camera ${index + 1}`}
            </option>
          ))}
        </select>
      </label>

      <div className="live-frame" style={{ position: 'relative', width: CAPTURE_WIDTH, maxWidth: '100%' }}>
        <video
          ref={videoRef}
          autoPlay
          playsInline
          muted
          style={{ width: '100%', display: 'block' }}
        />
        <canvas
          ref={overlayCanvasRef}
          style={{ position: 'absolute', top: 0, left: 0, width: '100%', height: '100%' }}
        />
        {/* hidden canvas used only to encode frames before sending - never displayed */}
        <canvas
          ref={captureCanvasRef}
          width={CAPTURE_WIDTH}
          height={CAPTURE_HEIGHT}
          style={{ display: 'none' }}
        />
      </div>
      <div
        style={{
          marginTop: 20,
          padding: 12,
          border: "1px solid #ddd",
          borderRadius: 8
        }}
      >
        <h3>Realtime Output</h3>

        <div>Pose : {poseLandmarks.length}</div>
      </div>
    </section>
  );
}
