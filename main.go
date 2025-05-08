package main

import (
	"fmt"
	"runtime"
)

func main() {
	// Check the current operating system
	switch runtime.GOOS {
	case "darwin":
		enumerateForMAC()
	case "windows":
		enumerateForWindows()
	default:
		fmt.Println("Unsupported OS")
	}
}
