package main

import (
	"fmt"
	"regexp"
	"strings"
	"unsafe"

	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows"
)

type DeviceInfo struct {
	Name       string
	DeviceID   string
	Status     string
	DeviceType string
	VendorID   string
	ProductID  string
}

type SPDevInfoData struct {
	CbSize     uint32
	DevInst    uint32
	ParentDevInst uint32
	ClassGuid  windows.GUID
	DevEnumerator uint16
	Reserved   uint32
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

func getParentDeviceID(deviceID string) (string, error) {
	setupAPI := windows.NewLazySystemDLL("setupapi.dll")
	setupDiGetClassDevs := setupAPI.NewProc("SetupDiGetClassDevsW")
	setupDiEnumDeviceInfo := setupAPI.NewProc("SetupDiEnumDeviceInfo")
	setupDiGetDeviceInstanceId := setupAPI.NewProc("SetupDiGetDeviceInstanceIdW")
	// setupDiGetDeviceProperty := setupAPI.NewProc("SetupDiGetDevicePropertyW")

	var devInfoSet windows.Handle
	ret, _, _ := setupDiGetClassDevs.Call(0, 0, 0, 2) // 2 = DIGCF_ALLCLASSES | DIGCF_PRESENT
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

	// Enumerate and display device details
	for _, d := range devices {
		fmt.Printf("Name       : %s\n", d.Name)
		fmt.Printf("DeviceID   : %s\n", d.DeviceID)
		fmt.Printf("Status     : %s\n", d.Status)
		fmt.Printf("DeviceType : %s\n", d.DeviceType)

		if strings.HasPrefix(d.DeviceID, "USBSTOR") {
			vid, pid, err := extractVIDPID(d.DeviceID)
			if err != nil {
				fmt.Println("Error extracting VID/PID:", err)
			} else {
				fmt.Printf("VendorID   : %s\n", vid)
				fmt.Printf("ProductID  : %s\n", pid)
			}

			fmt.Println(d.DeviceID)
			parentDeviceID, err := getParentDeviceID(d.DeviceID)
			if err != nil {
				fmt.Println("Error getting parent device ID:", err)
			} else {
				fmt.Println("Parent Device VID/PID:", parentDeviceID)
			}
		}
		fmt.Println("-----------------------------------")
	}
}

func enumerateForMAC() {
}

func enumerateForLinux() {
}
