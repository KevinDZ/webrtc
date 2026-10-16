package redroid

import (
	"bufio"
	"os/exec"
)

type ScrcpyStream struct {
	cmd    *exec.Cmd
	reader *bufio.Reader
}

func NewScrcpyStream(adbSerial, codec string, bitrate int, maxSize int, maxFPS int) (*ScrcpyStream, error) {
	if codec == "" {
		codec = "h264"
	}
	if bitrate == 0 {
		bitrate = 4_000_000
	}
	if maxSize == 0 {
		maxSize = 1280
	}
	if maxFPS == 0 {
		maxFPS = 30
	}

	args := []string{
		"-s", adbSerial,
		"shell",
		"CLASSPATH=/data/local/tmp/scrcpy-server",
		"app_process",
		"/",
		"com.genymobile.scrcpy.Server",
		"2.0",
		"tunnel_forward=true",
		"video=true",
		"audio=false",
		"control=false",
		"display_id=0",
		"video_codec=" + codec,
		"video_bit_rate=" + itoa(bitrate),
		"max_size=" + itoa(maxSize),
		"max_fps=" + itoa(maxFPS),
		"lock_video_orientation=-1",
		"send_frame_meta=true",
	}

	cmd := exec.Command("adb", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &ScrcpyStream{
		cmd:    cmd,
		reader: bufio.NewReader(stdout),
	}, nil
}

func (s *ScrcpyStream) ReadFrame() ([]byte, error) {
	// scrcpy 裸流前几字节是帧长度和元数据，正式实现要按协议解析
	return s.reader.ReadBytes('\n')
}

func (s *ScrcpyStream) Close() error {
	return s.cmd.Process.Kill()
}

func itoa(i int) string {
	return string(rune('0' + i))
}
