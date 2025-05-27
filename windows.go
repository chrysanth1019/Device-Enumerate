//go:build windows
// +build windows

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"unsafe"

	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows"
)

type BSSIDInfo struct {
	BSSID string
	Signal string
	RadioType string
	Channel	string
	BasicRates_Mbps string
	OtherRates_Mbps string
}

type SSIDInfo struct {
	SSID 			string
	NetworkType 	string
	Authentication 	string
	Encryption 		string
	BSSID		[]BSSIDInfo
}

type DeviceInfo struct {
	Name          string
	DeviceID      string
	Serial        string
	Status        string
	DeviceType    string
	VID           string
	PID           string
	EnumID        string
	Parent        string
	ParentVID     string
	ParentPID     string
	SSID          []SSIDInfo
}

type WLANInterface struct {
    Name        string   
    Description string   
    SSIDs       []SSIDInfo 
}


const (
	MAX_DEVICE_ID_LEN = 200
	DIGCF_PRESENT     = 0x00000002
	DIGCF_ALLCLASSES  = 0x00000004
)

var (
	setupapi           = syscall.NewLazyDLL("setupapi.dll")
	cfgmgr32           = syscall.NewLazyDLL("cfgmgr32.dll")

	procSetupDiGetClassDevs         = setupapi.NewProc("SetupDiGetClassDevsW")
	procSetupDiEnumDeviceInfo       = setupapi.NewProc("SetupDiEnumDeviceInfo")
	procSetupDiGetDeviceInstanceId  = setupapi.NewProc("SetupDiGetDeviceInstanceIdW")

	procCMGetParent   = cfgmgr32.NewProc("CM_Get_Parent")
	procCMGetDeviceID = cfgmgr32.NewProc("CM_Get_Device_IDW")
)

type SP_DEVINFO_DATA struct {
	cbSize    uint32
	ClassGuid windows.GUID
	DevInst   uint32
	Reserved  uintptr
}

func ListAllDevices() ([]DeviceInfo, error) {
	var dstPnP []struct {
		Name     string
		DeviceID string
		Status   string
		PNPClass string
	}

	var dstDisk []struct {
		Model         string
		DeviceID      string
		InterfaceType string
		MediaType     string
	}

	err := wmi.Query("SELECT Name, DeviceID, Status, PNPClass FROM Win32_PnPEntity", &dstPnP)
	if err != nil {
		return nil, err
	}

	err = wmi.Query("SELECT Model, DeviceID, InterfaceType, MediaType FROM Win32_DiskDrive", &dstDisk)
	if err != nil {
		return nil, err
	}

	var devices []DeviceInfo
	for _, d := range dstDisk {		
		deviceType := "DiskDrive"
		if strings.Contains(strings.ToUpper(d.Model), "SSD") || strings.Contains(strings.ToUpper(d.DeviceID), "NVME") {
			deviceType = "SSD"
		}
		devices = append(devices, DeviceInfo{
			Name:          d.Model,
			DeviceID:      d.DeviceID,
			EnumID: 	   d.InterfaceType,
			DeviceType:    deviceType,
			Status: 	   "Connected",
			Serial: 		extractSerialFromDeviceID(d.DeviceID),
			SSID: 			[]SSIDInfo{},
		})
	}

	ssid_interfaces, ssid_err := GetWLANInterfaces()

	for _, d := range dstPnP {
		var ssids = []SSIDInfo{}
		
		deviceType := d.PNPClass
		if strings.Contains(strings.ToUpper(d.Name), "SSD") || strings.Contains(strings.ToUpper(d.DeviceID), "NVME") {
			deviceType = "SSD"
		}
		
		if ssid_err == nil {
			for _, s := range ssid_interfaces {
				if s.Description == d.Name {
					ssids = s.SSIDs
					break
				}
			}
		}
		
		devices = append(devices, DeviceInfo{
			Name:       d.Name,
			DeviceID:   d.DeviceID,
			EnumID:     extractEnumID(d.DeviceID),
			Status:     d.Status,
			DeviceType: deviceType,
			Serial: 	extractSerialFromDeviceID(d.DeviceID),
			SSID: 		ssids,
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

func ExtractParentVIDPID(deviceID string) (string, string, error) {
	venDevRegex := regexp.MustCompile(`VEN_([0-9A-Fa-f]{4})&DEV_([0-9A-Fa-f]{4})`)
	matches := venDevRegex.FindStringSubmatch(deviceID)
	if len(matches) == 3 {
		return matches[1], matches[2], nil
	}
	return "", "", fmt.Errorf("no VEN and DEV IDs found")
}

func GetParentDeviceID(targetID string) string {
	pattern := ".*"
	guidStr := "{21EC2020-3AEA-1069-A2DD-08002B30309D}"
	guid, err := windows.GUIDFromString(guidStr)
	if err != nil {
		return ""
	}

	handle, _, _ := procSetupDiGetClassDevs.Call(
		uintptr(unsafe.Pointer(&guid)), 0, 0, DIGCF_PRESENT|DIGCF_ALLCLASSES,
	)
	if handle == 0 {
		return ""
	}
	defer syscall.CloseHandle(syscall.Handle(handle))

	var i uint32
	for {
		devInfo := SP_DEVINFO_DATA{cbSize: uint32(unsafe.Sizeof(SP_DEVINFO_DATA{}))}
		ret, _, _ := procSetupDiEnumDeviceInfo.Call(handle, uintptr(i), uintptr(unsafe.Pointer(&devInfo)))
		if ret == 0 {
			break
		}

		var buf [MAX_DEVICE_ID_LEN]uint16
		procSetupDiGetDeviceInstanceId.Call(
			handle,
			uintptr(unsafe.Pointer(&devInfo)),
			uintptr(unsafe.Pointer(&buf[0])),
			MAX_DEVICE_ID_LEN,
			0,
		)

		deviceID := syscall.UTF16ToString(buf[:])
		if strings.EqualFold(deviceID, targetID) {
			current := devInfo.DevInst
			var parentBuf [MAX_DEVICE_ID_LEN]uint16

			for {
				var parentInst uint32
				r1, _, _ := procCMGetParent.Call(uintptr(unsafe.Pointer(&parentInst)), uintptr(current), 0)
				if r1 != 0 {
					break
				}

				r2, _, _ := procCMGetDeviceID.Call(
					uintptr(parentInst),
					uintptr(unsafe.Pointer(&parentBuf[0])),
					MAX_DEVICE_ID_LEN,
					0,
				)
				if r2 != 0 {
					break
				}

				parentID := syscall.UTF16ToString(parentBuf[:])
				match, _ := regexp.MatchString(pattern, parentID)
				if match {
					return parentID
				}
				current = parentInst
			}
		}
		i++
	}
	return ""
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
		
		if devices[i].VID == "" && devices[i].PID == "" {
			devices[i].VID, devices[i].PID, _ = ExtractParentVIDPID(d.DeviceID)
		}

		devices[i].Parent = GetParentDeviceID(d.DeviceID)
		devices[i].ParentVID, devices[i].ParentPID, _ = extractVIDPID(devices[i].Parent)
		if devices[i].ParentVID == "" && devices[i].ParentPID == "" {
			devices[i].ParentVID, devices[i].ParentPID, _ = ExtractParentVIDPID(devices[i].Parent)
		}
	}
	printDevices(devices)
}

func printDevices(devices interface{}) {
	var b strings.Builder

	encoder := json.NewEncoder(&b)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(devices); err != nil {
		fmt.Println("JSON encode error:", err)
		return
	}

	result := strings.ReplaceAll(b.String(), `\\`, `\`)
	fmt.Println(result)
}

func GetWLANInterfaces() ([]WLANInterface, error) {
    cmd := exec.Command("netsh", "wlan", "show", "interfaces")
    var out bytes.Buffer
    cmd.Stdout = &out

    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("failed to execute netsh: %w", err)
    }

    output := out.String()
    lines := strings.Split(output, "\n")

    var interfaces []WLANInterface
    var current WLANInterface

    for _, line := range lines {
        trimmed := strings.TrimSpace(line)

        if strings.HasPrefix(trimmed, "Name") {
            if current.Name != "" {
                interfaces = append(interfaces, current)
                current = WLANInterface{}
            }
            current.Name = extractValueWithDot(trimmed)
        } else if strings.HasPrefix(trimmed, "Description") {
            current.Description = extractValueWithDot(trimmed)
        }
    }

    if current.Name != "" {
        interfaces = append(interfaces, current)
    }

    for i, iface := range interfaces {
        ssids, err := getSSIDsForInterface(iface.Name)
        if err != nil {
            return nil, fmt.Errorf("failed to get SSIDs for %s: %w", iface.Name, err)
        }
        interfaces[i].SSIDs = ssids
    }

    return interfaces, nil
}

func getSSIDsForInterface(interfaceName string) ([]SSIDInfo, error) {
    cmd := exec.Command("netsh", "wlan", "show", "networks", "mode=bssid", "interface="+interfaceName)
    var out bytes.Buffer
    cmd.Stdout = &out

    if err := cmd.Run(); err != nil {
        return nil, err
    }

    output := out.String()
    lines := strings.Split(output, "\n")
    var ssids []SSIDInfo
	
    for _, line := range lines {
		trimmed := strings.TrimSpace(line)
        if strings.HasPrefix(trimmed, "SSID") {
			parts := strings.SplitN(trimmed, ":", 2)
            if len(parts) == 2 {
				ssid := strings.TrimSpace(parts[1])
                if ssid != "" && ssid != "0" {
					var current SSIDInfo
					current.SSID = ssid
					ssids = append(ssids, current)
                }
            }
        } else if strings.HasPrefix(trimmed, "Network type") {
            parts := strings.SplitN(trimmed, ":", 2)
            if len(parts) == 2 {
                network_type := strings.TrimSpace(parts[1])
                if network_type != "" && network_type != "0" {
					if len(ssids) > 0 {
						ssids[len(ssids) - 1].NetworkType = network_type
					}
                }
            }
        } else if strings.HasPrefix(trimmed, "Authentication") {
            parts := strings.SplitN(trimmed, ":", 2)
            if len(parts) == 2 {
                auth := strings.TrimSpace(parts[1])
                if auth != "" && auth != "0" {
					if len(ssids) > 0 {
						ssids[len(ssids) - 1].Authentication = auth
					}
                }
            }
        } else if strings.HasPrefix(trimmed, "Encryption") {
            parts := strings.SplitN(trimmed, ":", 2)
            if len(parts) == 2 {
                enc := strings.TrimSpace(parts[1])
                if enc != "" && enc != "0" {
					if len(ssids) > 0 {
						ssids[len(ssids) - 1].Encryption = enc
					}
                }
            }
        } else if strings.HasPrefix(trimmed, "BSSID") {
            parts := strings.SplitN(trimmed, ":", 2)
            if len(parts) == 2 {
                bssid := strings.TrimSpace(parts[1])
                if bssid != "" && bssid != "0" {
					if len(ssids) > 0 {
						var v BSSIDInfo
						v.BSSID = bssid
						ssids[len(ssids) - 1].BSSID = append(ssids[len(ssids)-1].BSSID, v)
					}
                }
            }
        } else if strings.HasPrefix(trimmed, "Signal") {
            parts := strings.SplitN(trimmed, ":", 2)
            if len(parts) == 2 {
                sig := strings.TrimSpace(parts[1])
                if sig != "" && sig != "0" {
					if len(ssids) > 0 {
						if len(ssids[len(ssids)-1].BSSID) > 0 {
							ssids[len(ssids)-1].BSSID[len(ssids[len(ssids)-1].BSSID) - 1].Signal = sig
						}
					}
                }
            }
        } else if strings.HasPrefix(trimmed, "Radio type") {
            parts := strings.SplitN(trimmed, ":", 2)
            if len(parts) == 2 {
                radio := strings.TrimSpace(parts[1])
                if radio != "" && radio != "0" {
					if len(ssids) > 0 {
						if len(ssids[len(ssids)-1].BSSID) > 0 {
							ssids[len(ssids)-1].BSSID[len(ssids[len(ssids)-1].BSSID) - 1].RadioType = radio
						}
					}
                }
            }
        } else if strings.HasPrefix(trimmed, "Channel") {
            parts := strings.SplitN(trimmed, ":", 2)
            if len(parts) == 2 {
                chnn := strings.TrimSpace(parts[1])
                if chnn != "" && chnn != "0" {
					if len(ssids) > 0 {
						if len(ssids[len(ssids)-1].BSSID) > 0 {
							ssids[len(ssids)-1].BSSID[len(ssids[len(ssids)-1].BSSID) - 1].Channel = chnn
						}
					}
                }
            }
        } else if strings.HasPrefix(trimmed, "Basic rates") {
            parts := strings.SplitN(trimmed, ":", 2)
            if len(parts) == 2 {
                brates := strings.TrimSpace(parts[1])
                if brates != "" && brates != "0" {
					if len(ssids) > 0 {
						if len(ssids[len(ssids)-1].BSSID) > 0 {
							ssids[len(ssids)-1].BSSID[len(ssids[len(ssids)-1].BSSID) - 1].BasicRates_Mbps = brates
						}
					}
                }
            }
        } else if strings.HasPrefix(trimmed, "Other rates") {
            parts := strings.SplitN(trimmed, ":", 2)
            if len(parts) == 2 {
                orates := strings.TrimSpace(parts[1])
                if orates != "" && orates != "0" {
					if len(ssids) > 0 {
						if len(ssids[len(ssids)-1].BSSID) > 0 {
							ssids[len(ssids)-1].BSSID[len(ssids[len(ssids)-1].BSSID) - 1].OtherRates_Mbps = orates
						}
					}
                }
            }
        }
    }

    return ssids, nil
}

func extractValueWithDot(line string) string {
    parts := strings.SplitN(line, ":", 2)
    if len(parts) == 2 {
        return strings.TrimSpace(parts[1])
    }
    return ""
}

func enumerateForMAC() {
}

func enumerateForLinux() {
}
