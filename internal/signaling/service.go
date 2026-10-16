package signaling

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"webrtc/internal/webrtc"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type SignalingServer struct {
	service Service
}

type Service interface {
	CreateSession(sessionID string) (*webrtc.Session, error)
}

func NewServer(service Service) *SignalingServer {
	return &SignalingServer{service: service}
}

type SignalMessage struct {
	Type      string          `json:"type"`
	SessionID string          `json:"sessionId"`
	Payload   json.RawMessage `json:"payload"`
}

func (s *SignalingServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("websocket upgrade error:", err)
		return
	}
	defer conn.Close()

	var session *webrtc.Session

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if session != nil {
				session.Close()
			}
			break
		}

		var msg SignalMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "create-session":
			sess, err := s.service.CreateSession(msg.SessionID)
			if err != nil {
				log.Println("create session error:", err)
				continue
			}

			session = sess

			session.OnICECandidate(func(candidate interface{}) {
				data, _ := json.Marshal(SignalMessage{
					Type:      "ice-candidate",
					SessionID: session.ID(),
					Payload:   mustJSON(candidate),
				})
				_ = conn.WriteMessage(websocket.TextMessage, data)
			})

			session.OnControlMessage(func(data []byte) {
				// 触控事件交给 agent 处理
			})

		case "offer":
			if session == nil {
				continue
			}

			answer, err := session.SetOffer(msg.Payload)
			if err != nil {
				log.Println("set offer error:", err)
				continue
			}

			resp, _ := json.Marshal(SignalMessage{
				Type:      "answer",
				SessionID: session.ID(),
				Payload:   answer,
			})
			_ = conn.WriteMessage(websocket.TextMessage, resp)

		case "ice-candidate":
			if session == nil {
				continue
			}
			_ = session.AddICECandidate(msg.Payload)
		}
	}
}

func mustJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}