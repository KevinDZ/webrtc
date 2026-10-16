package redroid

import (
	"os/exec"
)

func CaptureH264Frame(containerName string) ([]byte, error) {
	cmd := exec.Command(
		"docker", "exec", containerName,
		"screenrecord", "--output-format=h264", "--time-limit=1", "/sdcard/frame.h264",
	)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	pull := exec.Command("docker", "exec", containerName, "cat", "/sdcard/frame.h264")
	return pull.Output()
}