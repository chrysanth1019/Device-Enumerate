//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unsafe"

	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows"
)

type DeviceInfo struct {
	Name         string
	DeviceID     string
	Serial string
	Status       string
	DeviceType   string
	VID     string
	PID    string
	EnumID   string
}


type SPDevInfoData struct {
	CbSize        uint32
	DevInst       uint32
	ParentDevInst uint32
	ClassGuid     windows.GUID
	DevEnumID uint16
	Reserved      uint32
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

		if strings.Contains(strings.ToUpper(d.Name), "SSD") ||
			strings.Contains(strings.ToUpper(d.DeviceID), "NVME") {
			deviceType = "SSD"
		}

		devices = append(devices, DeviceInfo{
			Name:       d.Name,
			DeviceID:   d.DeviceID,
			EnumID: 	extractEnumID(d.DeviceID),
			Status:     d.Status,
			DeviceType: deviceType,
		})
	}
	return devices, nil
}

func extractVIDPID(deviceID string) (string, string, error) {
	re := regexp.MustCompile(`VID_(\w{4})&PID_(\w{4})`)
	matches := re.FindStringSubmatch(deviceID)
	if len(matches) < 3 {
		return "", "", fmt.Errorf("VID and PID not found in device ID: %s", deviceID)
	}
	return matches[1], matches[2], nil
}

func extractSerialFromDeviceID(deviceID string) string {
	parts := strings.Split(deviceID, "\\")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

func extractEnumID(deviceID string) string {
	parts := strings.Split(deviceID, "\\")
	if len(parts) > 0 {
		return parts[0]
	}
	return "Unknown"
}

func getParentDeviceID(deviceID string) (string, error) {
	setupAPI := windows.NewLazySystemDLL("setupapi.dll")
	setupDiGetClassDevs := setupAPI.NewProc("SetupDiGetClassDevsW")
	setupDiEnumDeviceInfo := setupAPI.NewProc("SetupDiEnumDeviceInfo")
	setupDiGetDeviceInstanceId := setupAPI.NewProc("SetupDiGetDeviceInstanceIdW")

	var devInfoSet windows.Handle
	ret, _, _ := setupDiGetClassDevs.Call(0, 0, 0, 2) 
	if ret == 0 {
		return "", fmt.Errorf("failed to get device information set")
	}
	devInfoSet = windows.Handle(ret)
	defer windows.CloseHandle(devInfoSet)

	var devIndex uint32
	for {
		var deviceInfoData SPDevInfoData
		deviceInfoData.CbSize = uint32(unsafe.Sizeof(deviceInfoData))

		ret, _, _ := setupDiEnumDeviceInfo.Call(uintptr(devInfoSet), uintptr(devIndex), uintptr(unsafe.Pointer(&deviceInfoData)))
		if ret == 0 {
			break
		}

		var instanceID [256]uint16
		ret, _, _ = setupDiGetDeviceInstanceId.Call(uintptr(devInfoSet), uintptr(unsafe.Pointer(&deviceInfoData)), uintptr(unsafe.Pointer(&instanceID[0])), uintptr(len(instanceID)), 0)
		if ret == 0 {
			continue
		}

		instanceIDStr := windows.UTF16ToString(instanceID[:])

		if strings.Contains(instanceIDStr, deviceID) {
			vid, pid, err := extractVIDPID(instanceIDStr)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("VID_%s&PID_%s", vid, pid), nil
		}

		devIndex++
	}

	return "", fmt.Errorf("parent device ID not found")
}

func enumerateForWindows() {
	devices, err := ListAllDevices()
	if err != nil || len(devices) == 0 {
		fmt.Println("No devices found or error:", err)
		return
	}

	for i, d := range devices {
		vid, pid, err := extractVIDPID(d.DeviceID)
		if err == nil {
			devices[i].VID = vid
			devices[i].PID = pid
		}

		serial := extractSerialFromDeviceID(d.DeviceID)
		if serial != "" {
			devices[i].Serial = serial
		}

		devices[i].EnumID = extractEnumID(d.DeviceID)

		if strings.HasPrefix(d.DeviceID, "USBSTOR") {
			parentDeviceID, err := getParentDeviceID(d.DeviceID)
			if err == nil {
				devices[i].DeviceID += " (Parent: " + parentDeviceID + ")"
			}
		}
	}

	output, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling to JSON:", err)
		return
	}

	fmt.Println(string(output))
}

func enumerateForMAC() {
}

func enumerateForLinux() {
}