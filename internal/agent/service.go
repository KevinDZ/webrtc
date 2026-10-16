package agent

import (
	"encoding/json"
	"sync"

	"webrtc/internal/input"
	"webrtc/internal/webrtc"
)

type Service struct {
	mu        sync.Mutex
	sessions  map[string]*webrtc.Session
	containers map[string]string
}

func NewService() *Service {
	return &Service{
		sessions:   make(map[string]*webrtc.Session),
		containers: make(map[string]string),
	}
}

func (s *Service) CreateSession(sessionID string) (*webrtc.Session, error) {
	session, err := webrtc.NewSession(sessionID)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.sessions[sessionID] = session
	containerName := s.containers[sessionID]
	s.mu.Unlock()

	if containerName == "" {
		containerName = "redroid-default"
	}

	stop := make(chan struct{})

	session.OnControlMessage(func(data []byte) {
		var event struct {
			Type string  `json:"type"`
			X    float64 `json:"x"`
			Y    float64 `json:"y"`
		}

		if err := json.Unmarshal(data, &event); err != nil {
			return
		}

		if event.Type == "tap" {
			_ = input.InjectTap(containerName, int(event.X), int(event.Y))
		}
	})

	go func() {
		webrtc.StartScreenPusherToScrcpyStream(session, containerName, stop)
		s.mu.Lock()
		delete(s.sessions, sessionID)
		s.mu.Unlock()
	}()

	return session, nil
}