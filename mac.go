//go:build darwin
// +build darwin

package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

var deviceTypes = []string{
	"SPUSBDataType",
	"SPStorageDataType",
	"SPAirPortDataType",
	"SPBluetoothDataType",
	"SPNetworkDataType",
	"SPCameraDataType",
	"SPPCIDataType",
}


func enumerateForMAC() {
	output := make(map[string]interface{})

	for _, deviceType := range deviceTypes {
		data, err := getDeviceInfo(deviceType)
		if err != nil {
			fmt.Println("Error collecting", deviceType, ":", err)
			continue
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(data), &parsed); err != nil {
			fmt.Println("Error parsing JSON for", deviceType, ":", err)
			continue
		}

		for key, value := range parsed {
			output[key] = value
		}
	}

	prettyJSON, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling output:", err)
		return
	}

	fmt.Println(string(prettyJSON))
}

func getDeviceInfo(deviceType string) (string, error) {
	cmd := exec.Command("system_profiler", deviceType, "-json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("system_profiler failed: %v", err)
	}
	return removeErrorMessages(string(output)), nil
}

func removeErrorMessages(input string) string {
	lines := strings.Split(input, "\n")
	var cleaned []string
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "error") || strings.Contains(lower, "failed") {
			continue
		}
		cleaned = append(cleaned, line)
	}
	return strings.Join(cleaned, "\n")
}

func enumerateForWindows() {}

func enumerateForLinux() {}