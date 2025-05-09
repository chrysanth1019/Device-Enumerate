package main

import (
	"fmt"
	"runtime"
)

func main() {
	switch runtime.GOOS {
	case "darwin":
		enumerateForMAC()
	case "windows":
		enumerateForWindows()
	case "linux":
		enumerateForLinux()
	default:
		fmt.Println("Unsupported OS")
	}
}
