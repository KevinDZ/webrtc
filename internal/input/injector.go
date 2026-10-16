package input

import (
	"os/exec"
	"strconv"
)

func InjectTap(containerName string, x, y int) error {
	return runADB(containerName, "shell", "input", "tap", strconv.Itoa(x), strconv.Itoa(y))
}

func runADB(containerName string, args ...string) error {
	full := append([]string{"exec", containerName, "adb"}, args...)
	cmd := exec.Command("docker", full...)
	return cmd.Run()
}