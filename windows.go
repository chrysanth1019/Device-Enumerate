//go:build windows
// +build windows

package main

import (
	"fmt"
	"strings"

	"github.com/StackExchange/wmi"
)

type DeviceInfo struct {
	Name       string
	DeviceID   string
	Status     string
	DeviceType string
}

func ListAllDevices() ([]DeviceInfo, error) {
	var dst []struct {
		Name     string
		DeviceID string
		Status   string
		PNPClass string
	}
	err := wmi.Query("SELECT Name, DeviceID, Status, PNPClass FROM Win32_PnPEntity", &dst)
	if err != nil {
		return nil, err
	}

	var devices []DeviceInfo
	for _, d := range dst {
		deviceType := d.PNPClass

		// Detect SSDs based on common naming patterns
		if strings.Contains(strings.ToUpper(d.Name), "SSD") ||
			strings.Contains(strings.ToUpper(d.DeviceID), "NVME") {
			deviceType = "SSD"
		}

		devices = append(devices, DeviceInfo{
			Name:       d.Name,
			DeviceID:   d.DeviceID,
			Status:     d.Status,
			DeviceType: deviceType,
		})
	}
	return devices, nil
}

func enumerateForWindows() {
	devices, err := ListAllDevices()
	if err != nil || len(devices) == 0 {
		fmt.Println("No devices found or error:", err)
		return
	}

	for _, d := range devices {
		fmt.Printf("Name       : %s\n", d.Name)
		fmt.Printf("DeviceID   : %s\n", d.DeviceID)
		fmt.Printf("Status     : %s\n", d.Status)
		fmt.Printf("DeviceType : %s\n", d.DeviceType)
		fmt.Println("-----------------------------------")
	}
}

func enumerateForMAC() {
}