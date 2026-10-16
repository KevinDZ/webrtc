package webrtc

import (
	"encoding/json"
	"sync"

	"github.com/pion/webrtc/v4"
)

type Session struct {
	id          string
	pc          *webrtc.PeerConnection
	videoTrack  *webrtc.TrackLocalStaticSample
	dataChannel *webrtc.DataChannel
	mu          sync.Mutex
	onICE       func(candidate interface{})
	onControl   func(data []byte)
}

func NewSession(id string) (*Session, error) {
	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))

	pc, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	})
	if err != nil {
		return nil, err
	}

	videoTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264},
		"video",
		"screen",
	)
	if err != nil {
		_ = pc.Close()
		return nil, err
	}

	if _, err := pc.AddTrack(videoTrack); err != nil {
		_ = pc.Close()
		return nil, err
	}

	s := &Session{
		id:         id,
		pc:         pc,
		videoTrack: videoTrack,
	}

	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		s.mu.Lock()
		cb := s.onICE
		s.mu.Unlock()

		if cb != nil {
			cb(c.ToJSON())
		}
	})

	pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		s.mu.Lock()
		s.dataChannel = dc
		s.mu.Unlock()

		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			s.mu.Lock()
			cb := s.onControl
			s.mu.Unlock()

			if cb != nil {
				buf := make([]byte, len(msg.Data))
				copy(buf, msg.Data)
				cb(buf)
			}
		})
	})

	return s, nil
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) VideoTrack() *webrtc.TrackLocalStaticSample {
	return s.videoTrack
}

func (s *Session) OnICECandidate(fn func(candidate interface{})) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onICE = fn
}

func (s *Session) OnControlMessage(fn func(data []byte)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onControl = fn
}

func (s *Session) SetOffer(offerJSON json.RawMessage) (json.RawMessage, error) {
	var offer webrtc.SessionDescription
	if err := json.Unmarshal(offerJSON, &offer); err != nil {
		return nil, err
	}

	if err := s.pc.SetRemoteDescription(offer); err != nil {
		return nil, err
	}

	answer, err := s.pc.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	if err := s.pc.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	return json.Marshal(answer)
}

func (s *Session) AddICECandidate(candidateJSON json.RawMessage) error {
	var c webrtc.ICECandidateInit
	if err := json.Unmarshal(candidateJSON, &c); err != nil {
		return err
	}
	return s.pc.AddICECandidate(c)
}

func (s *Session) Close() {
	_ = s.pc.Close()
}