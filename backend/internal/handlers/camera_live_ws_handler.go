package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1 << 20, // 1MB - base64 frames can be a few hundred KB
	WriteBufferSize: 1 << 20,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: restrict to your actual frontend origin before production
		return true
	},
}

type liveFrameMessage struct {
	ImageBase64 string `json:"imageBase64"`
}

// CameraLiveWS handles GET /ws/camera/live.
// The iPhone opens ONE WebSocket connection and streams frames continuously
// as {"imageBase64": "..."} JSON messages. Each frame gets a detection
// response back over the same connection: {"faces": [...]}.
//
// If frames arrive faster than the python service can process them, older
// pending frames are dropped in favor of the newest one - this keeps the
// overlay feeling "live" instead of drifting behind by several seconds.
func (h *Handler) CameraLiveWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	frames := make(chan string, 1)
	done := make(chan struct{})

	// reader: pulls frames off the socket, always keeping only the latest one
	go func() {
		defer close(done)
		for {
			_, raw, err := conn.ReadMessage()
			if err != nil {
				return // client disconnected or network error
			}

			var msg liveFrameMessage
			if err := json.Unmarshal(raw, &msg); err != nil || msg.ImageBase64 == "" {
				continue
			}

			select {
			case frames <- msg.ImageBase64:
			default:
				// processor is still busy - replace the pending frame with the newest
				select {
				case <-frames:
				default:
				}
				frames <- msg.ImageBase64
			}
		}
	}()

	// processor: runs detection on whatever the latest frame is
	for {
		select {
		case <-done:
			return
		case imageBase64 := <-frames:
			faces, err := h.service.DetectFaces(imageBase64)
			if err != nil {
				_ = conn.WriteJSON(map[string]string{"error": err.Error()})
				continue
			}
			if err := conn.WriteJSON(map[string]any{"faces": faces}); err != nil {
				return // write failed - connection is likely dead
			}
		}
	}
}