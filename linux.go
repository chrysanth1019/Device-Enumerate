//go:build linux
// +build linux

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func readFirstLine(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func printDevice(deviceType, id, name, status string) {
	fmt.Printf("Name       : %s\n", name)
	fmt.Printf("DeviceID   : %s\n", id)
	fmt.Printf("Status     : %s\n", status)
	fmt.Printf("DeviceType : %s\n", deviceType)
	fmt.Println("-----------------------------------")
}

func listNetworkInterfaces() {
	basePath := "/sys/class/net"
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		fmt.Println("Error reading network interfaces:", err)
		return
	}

	for _, f := range files {
		id := f.Name()
		operstatePath := filepath.Join(basePath, id, "operstate")
		status := readFirstLine(operstatePath)
		if status == "" {
			status = "unknown"
		}
		printDevice("network", id, id, status)
	}
}

func listUSBDevices() {
	basePath := "/sys/bus/usb/devices"
	files, _ := ioutil.ReadDir(basePath)

	for _, f := range files {
		id := f.Name()
		path := filepath.Join(basePath, id)
		product := readFirstLine(filepath.Join(path, "product"))
		if product == "" {
			continue
		}
		vendor := readFirstLine(filepath.Join(path, "idVendor"))
		auth := readFirstLine(filepath.Join(path, "authorized"))
		status := "unknown"
		if auth == "1" {
			status = "connected"
		} else if auth == "0" {
			status = "not connected"
		}
		printDevice("usb", id, product+" (Vendor: "+vendor+")", status)
	}
}

func listPCIDevices() {
	basePath := "/sys/bus/pci/devices"
	files, _ := ioutil.ReadDir(basePath)

	for _, f := range files {
		id := f.Name()
		path := filepath.Join(basePath, id)
		vendor := readFirstLine(filepath.Join(path, "vendor"))
		device := readFirstLine(filepath.Join(path, "device"))
		name := fmt.Sprintf("Vendor: %s, Device: %s", vendor, device)
		printDevice("pci", id, name, "connected")
	}
}

func listStorageDevices() {
	basePath := "/sys/block"
	files, _ := ioutil.ReadDir(basePath)

	for _, f := range files {
		id := f.Name()
		model := readFirstLine(filepath.Join(basePath, id, "device/model"))
		if model == "" {
			model = "N/A"
		}
		printDevice("storage", id, model, "available")
	}
}

func listWebcams() {
	basePath := "/sys/class/video4linux"
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return
	}

	for _, f := range files {
		id := f.Name()
		name := readFirstLine(filepath.Join(basePath, id, "name"))
		printDevice("webcam", id, name, "connected")
	}
}

func listInputDevices() {
	basePath := "/sys/class/input"
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return
	}

	for _, f := range files {
		id := f.Name()
		name := readFirstLine(filepath.Join(basePath, id, "name"))
		printDevice("input", id, name, "available")
	}
}

func listTTYDevices() {
	basePath := "/sys/class/tty"
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return
	}

	for _, f := range files {
		id := f.Name()
		name := readFirstLine(filepath.Join(basePath, id, "device"))
		printDevice("tty", id, name, "available")
	}
}

func listSoundDevices() {
	basePath := "/sys/class/sound"
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return
	}

	for _, f := range files {
		id := f.Name()
		name := readFirstLine(filepath.Join(basePath, id, "id"))
		printDevice("sound", id, name, "available")
	}
}

func listBluetoothDevices() {
	basePath := "/sys/class/bluetooth"
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return
	}

	for _, f := range files {
		id := f.Name()
		devicePath := filepath.Join(basePath, id)
		deviceName := readFirstLine(filepath.Join(devicePath, "name"))
		deviceAddr := readFirstLine(filepath.Join(devicePath, "address"))
		if deviceName == "" {
			deviceName = "Unknown Bluetooth Device"
		}
		status := "available"
		printDevice("bluetooth", deviceAddr, deviceName, status)
	}
}

func listAllDevices() {
	listNetworkInterfaces()
	listUSBDevices()
	listPCIDevices()
	listStorageDevices()
	listWebcams()
	listInputDevices()
	// listTTYDevices()
	listSoundDevices()
	listBluetoothDevices()
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