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
	for _, deviceType := range deviceTypes {
		data, err := getDeviceInfo(deviceType)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		switch deviceType {
		case "SPUSBDataType":
			printUSBData(data)
		case "SPStorageDataType":
			printStorageData(data)
		case "SPAirPortDataType":
			printWiFiData(data)
		case "SPBluetoothDataType":
			printBluetoothData(data)
		case "SPNetworkDataType":
			printNetworkData(data)
		case "SPCameraDataType":
			printWebcamData(data)
		case "SPPCIDataType":
			printPCIDevicesData(data)
		}
	}
}

func getDeviceInfo(deviceType string) (string, error) {
	cmd := exec.Command("system_profiler", deviceType, "-json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run system_profiler: %v", err)
	}
	return removeErrorMessages(string(output)), nil
}

func removeErrorMessages(input string) string {
	lines := strings.Split(input, "\n")
	var cleanedLines []string
	for _, line := range lines {
		if strings.Contains(line, "error") || strings.Contains(line, "failed") {
			continue
		}
		cleanedLines = append(cleanedLines, line)
	}
	return strings.Join(cleanedLines, "\n")
}

func printUSBData(data string) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}
	for _, devices := range result {
		if deviceList, ok := devices.([]interface{}); ok {
			for _, device := range deviceList {
				if deviceMap, ok := device.(map[string]interface{}); ok {
					printDeviceInfoExcluding(deviceMap, "_items")
					fmt.Println("----------------------------------------------------------------------")
					if items, ok := deviceMap["_items"].([]interface{}); ok {
						for _, item := range items {
							if itemMap, ok := item.(map[string]interface{}); ok {
								printDeviceInfo(itemMap)
								fmt.Println("----------------------------------------------------------------------")
							}
						}
					}
				}
			}
		}
	}
}

func printStorageData(data string) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}
	for _, devices := range result {
		if deviceList, ok := devices.([]interface{}); ok {
			for _, device := range deviceList {
				if deviceMap, ok := device.(map[string]interface{}); ok {
					printDeviceInfo(deviceMap)
					fmt.Println("----------------------------------------------------------------------")
				}
			}
		}
	}
}

func printWiFiData(data string) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}
	for _, devices := range result {
		if deviceList, ok := devices.([]interface{}); ok {
			for _, device := range deviceList {
				if deviceMap, ok := device.(map[string]interface{}); ok {
					printDeviceInfo(deviceMap)
					fmt.Println("----------------------------------------------------------------------")
				}
			}
		}
	}
}

func printBluetoothData(data string) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}
	for _, devices := range result {
		if deviceList, ok := devices.([]interface{}); ok {
			for _, device := range deviceList {
				if deviceMap, ok := device.(map[string]interface{}); ok {
					printDeviceInfo(deviceMap)
					fmt.Println("----------------------------------------------------------------------")
				}
			}
		}
	}
}

func printNetworkData(data string) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}
	for _, devices := range result {
		if deviceList, ok := devices.([]interface{}); ok {
			for _, device := range deviceList {
				if deviceMap, ok := device.(map[string]interface{}); ok {
					printDeviceInfo(deviceMap)
					fmt.Println("----------------------------------------------------------------------")
				}
			}
		}
	}
}

func printWebcamData(data string) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}
	for _, devices := range result {
		if deviceList, ok := devices.([]interface{}); ok {
			for _, device := range deviceList {
				if deviceMap, ok := device.(map[string]interface{}); ok {
					printDeviceInfo(deviceMap)
					fmt.Println("----------------------------------------------------------------------")
				}
			}
		}
	}
}

func printPCIDevicesData(data string) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}
	for _, devices := range result {
		if deviceList, ok := devices.([]interface{}); ok {
			for _, device := range deviceList {
				if deviceMap, ok := device.(map[string]interface{}); ok {
					printDeviceInfo(deviceMap)
					fmt.Println("----------------------------------------------------------------------")
				}
			}
		}
	}
}

func printDeviceInfoExcluding(deviceMap map[string]interface{}, excludeKey string) {
	for key, value := range deviceMap {
		if key == excludeKey || isEmpty(value) {
			continue
		}
		printKeyValue(key, value)
	}
}

func printDeviceInfo(deviceMap map[string]interface{}) {
	if name, exists := deviceMap["_name"]; exists {
		printKeyValue("Name", name)
		delete(deviceMap, "_name")
	}

	for key, value := range deviceMap {
		if isEmpty(value) {
			continue
		}
		key = capitalizeFirstLetter(key)

		printKeyValue(key, value)
	}
}

func isEmpty(value interface{}) bool {
	switch v := value.(type) {
	case nil:
		return true
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	}
	return false
}

func capitalizeFirstLetter(key string) string {
	if len(key) > 0 {
		return strings.ToUpper(string(key[0])) + key[1:]
	}
	return key
}

func printKeyValue(key string, value interface{}) {
	switch v := value.(type) {
	case string:
		fmt.Printf("%-30s: %s\n", key, v)
	case float64:
		fmt.Printf("%-30s: %v\n", key, v)
	case map[string]interface{}:
		for subKey, subVal := range v {
			printKeyValue(subKey, subVal)
		}
	case []interface{}:
		for i, item := range v {
			itemKey := fmt.Sprintf("%s[%d]", key, i)
			if _, ok := item.(map[string]interface{}); ok {
				printKeyValue(itemKey, item)
			} else {
				printKeyValue(itemKey, item)
			}
		}
	}
}

func enumerateForWindows() {	
}


func enumerateForLinux() {
}