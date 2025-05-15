//go:build linux
// +build linux

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DeviceInfo struct {
	Name         string `json:"Name"`
	ID           string `json:"DeviceID"`
	SerialNumber string `json:"Serial"`
	Status       string `json:"Status"`
	DeviceType   string `json:"DeviceType"`
	VendorID     string `json:"VID"`
	ProductID    string `json:"PID"`
	Enumerator   string `json:"EnumID"`
}

func readFirstLine(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func printDeviceAsJSON(device DeviceInfo) {
	output, err := json.MarshalIndent(device, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling to JSON:", err)
		return
	}
	fmt.Println(string(output))
}

func listUSBDevices() {
	basePath := "/sys/bus/usb/devices"
	files, err := os.ReadDir(basePath)
	if err != nil {
		fmt.Println("Error reading USB devices:", err)
		return
	}

	for _, f := range files {
		id := f.Name()
		path := filepath.Join(basePath, id)

		vid := readFirstLine(filepath.Join(path, "idVendor"))
		pid := readFirstLine(filepath.Join(path, "idProduct"))
		serial := readFirstLine(filepath.Join(path, "serial"))

		if vid == "" || pid == "" {
			continue
		}

		product := readFirstLine(filepath.Join(path, "product"))
		if product == "" {
			product = "Unknown USB Device"
		}

		auth := readFirstLine(filepath.Join(path, "authorized"))
		status := "unknown"
		if auth == "1" {
			status = "connected"
		} else if auth == "0" {
			status = "not connected"
		}

		device := DeviceInfo{
			DeviceType:   "usb",
			ID:           id,
			Name:         product,
			Status:       status,
			Enumerator:   "usb",
			VendorID:     vid,
			ProductID:    pid,
			SerialNumber: serial,
		}
		printDeviceAsJSON(device)
	}
}

func listPCIDevices() {
	basePath := "/sys/bus/pci/devices"
	files, err := os.ReadDir(basePath)
	if err != nil {
		return
	}

	for _, f := range files {
		id := f.Name()
		path := filepath.Join(basePath, id)
		vendor := readFirstLine(filepath.Join(path, "vendor"))
		device := readFirstLine(filepath.Join(path, "device"))
		serial := readFirstLine(filepath.Join(path, "serial"))

		if serial == "" {
			serial = "N/A"
		}

		name := fmt.Sprintf("Vendor: %s, Device: %s", vendor, device)

		deviceInfo := DeviceInfo{
			DeviceType:   "pci",
			ID:           id,
			Name:         name,
			Status:       "connected",
			Enumerator:   "pci",
			VendorID:     vendor,
			ProductID:    device,
			SerialNumber: serial,
		}
		printDeviceAsJSON(deviceInfo)
	}
}

func listStorageDevices() {
	basePath := "/sys/block"
	files, err := os.ReadDir(basePath)
	if err != nil {
		return
	}

	for _, f := range files {
		id := f.Name()
		devicePath := filepath.Join(basePath, id)

		model := readFirstLine(filepath.Join(devicePath, "device/model"))
		if model == "" {
			model = "Unknown Storage Device"
		}

		serial := readFirstLine(filepath.Join(devicePath, "device/serial"))
		if serial == "" {
			serial = "N/A"
		}

		deviceInfo := DeviceInfo{
			DeviceType:   "storage",
			ID:           id,
			Name:         model,
			Status:       "available",
			Enumerator:   "storage",
			VendorID:     "N/A",
			ProductID:    "N/A",
			SerialNumber: serial,
		}
		printDeviceAsJSON(deviceInfo)
	}
}

func listNetworkInterfaces() {
	basePath := "/sys/class/net"
	files, err := os.ReadDir(basePath)
	if err != nil {
		fmt.Println("Error reading network interfaces:", err)
		return
	}

	for _, f := range files {
		id := f.Name()
		interfacePath := filepath.Join(basePath, id)
		status := readFirstLine(filepath.Join(interfacePath, "operstate"))
		if status == "" {
			status = "unknown"
		}

		wirelessPath := filepath.Join(interfacePath, "wireless")
		deviceType := "network"
		if _, err := os.Stat(wirelessPath); err == nil {
			deviceType = "wifi"
		}

		deviceInfo := DeviceInfo{
			DeviceType:   deviceType,
			ID:           id,
			Name:         id,
			Status:       status,
			Enumerator:   "network",
			VendorID:     "N/A",
			ProductID:    "N/A",
			SerialNumber: "N/A",
		}
		printDeviceAsJSON(deviceInfo)
	}
}

func listWebcams() {
	basePath := "/sys/class/video4linux"
	files, err := os.ReadDir(basePath)
	if err != nil {
		return
	}

	for _, f := range files {
		id := f.Name()
		name := readFirstLine(filepath.Join(basePath, id, "name"))
		if name == "" {
			name = "Unknown Webcam"
		}

		deviceInfo := DeviceInfo{
			DeviceType:   "webcam",
			ID:           id,
			Name:         name,
			Status:       "connected",
			Enumerator:   "webcam",
			VendorID:     "N/A",
			ProductID:    "N/A",
			SerialNumber: "N/A",
		}
		printDeviceAsJSON(deviceInfo)
	}
}

func listAllDevices() {
	listNetworkInterfaces()
	listUSBDevices()
	listPCIDevices()
	listStorageDevices()
	listWebcams()
}

func enumerateForLinux() {
	if _, err := os.Stat("/sys"); os.IsNotExist(err) {
		fmt.Println("This program must run on a Linux system with /sys available.")
		return
	}
	listAllDevices()
}
func enumerateForMAC() {
}

func enumerateForWindows() {
}