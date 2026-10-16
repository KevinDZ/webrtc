package webrtc

import (
	"io"
	"time"

	"webrtc/internal/redroid"

	"github.com/pion/webrtc/v4/pkg/media"
)

func StartScreenPusherToCaptureH264Frame(session *Session, containerName string, stop <-chan struct{}) {
	ticker := time.NewTicker(33 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
		}

		frame, err := redroid.CaptureH264Frame(containerName)
		if err != nil || len(frame) == 0 {
			continue
		}

		if err := session.VideoTrack().WriteSample(media.Sample{
			Data:     frame,
			Duration: 33 * time.Millisecond,
		}); err != nil {
			return
		}
	}
}

func StartScreenPusherToScrcpyStream(session *Session, adbSerial string, stop <-chan struct{}) {
	stream, err := redroid.NewScrcpyStream(adbSerial, "h264", 4_000_000, 1280, 30)
	if err != nil {
		return
	}
	defer stream.Close()

	for {
		select {
		case <-stop:
			return
		default:
		}

		frame, err := stream.ReadFrame()
		if err != nil || err == io.EOF {
			return
		}
		if len(frame) == 0 {
			continue
		}

		if err := session.VideoTrack().WriteSample(media.Sample{
			Data:     frame,
			Duration: 33 * time.Millisecond,
		}); err != nil {
			return
		}
	}
}
