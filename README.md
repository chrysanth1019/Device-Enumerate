# Hardware Enumerator (macOS, Windows, Linux)

This Go project is a cross-platform hardware enumerator that retrieves detailed system hardware information, such as USB devices, storage, network interfaces, Bluetooth, Wi-Fi, cameras, PCI devices, and CPU details.

## Features

- ✅ macOS: Uses `system_profiler` to fetch hardware details in JSON.
- ✅ Windows: Uses WMI (Windows Management Instrumentation) via `StackExchange/wmi` for querying system information.
- ✅ CPU information: Collected using `runtime`, `sysctl`, WMI, or `lscpu` depending on platform.
- 🌐 Clean output formatting with key name normalization.

## Supported Platforms

- macOS
- Windows
- Linux

## Requirements

### macOS

- Go 1.18+
- `system_profiler` (built-in on macOS)

### Windows

- Go 1.18+
- WMI enabled
- Required Go package:
  ```bash
  go get github.com/StackExchange/wmi

### Linux

- Go 1.18+
