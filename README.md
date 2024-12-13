# smmake - Simple Multi-platform Make

[![Go Report Card](https://goreportcard.com/badge/github.com/datstma/smmake)](https://goreportcard.com/report/github.com/datstma/smmake)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

smmake is a powerful, cross-platform alternative to the traditional Unix `make` utility. It brings the familiar functionality of `make` to Windows and other operating systems without requiring the installation of odd tools such as Win32 GNU tools.

The primary target audience in Windows x86_64/arm64 developers since there's no native make support built in. But it works on Linux and Mac just as well as a bonus should you prefer smmake over GNU make (for some unkonwn reason :) )

### - It's currently a Golang project that you can build on your own, but I will be creating installable (MSI) packages for Windows. 

## 📚 Parses your Makefile in the way you'd expect

smmake supports most common Makefile features you'd expect, making it familiar for Make users while remaining approachable for newcomers:

### Basic Features

- **Targets and Commands**: Define build targets and their associated commands
  ```makefile
  hello:
      echo "Hello, World!"

## 🚀 Features

- **Native Cross-Platform Support**: Works seamlessly on Windows, macOS, and Linux
- **Make-Compatible**: Parse and execute standard Makefiles
- **Rich Feature Set**:
    - Targets and Dependencies Management
    - Variable Expansion
    - Pattern Rules Support
    - Command Execution
- **Performance Optimized**:
    - Parallel Execution of Dependencies
    - Smart Dependency Resolution

## 🔧 Installation

Build it From source on your own

```bash
git clone https://github.com/datstma/smmake.git
cd smmake
go build -o smmake.exe cmd/main.go cmd/parser.go
```
### Usage

Then run smmake (instead of make)
```bash
smmake build       # Build the project
smmake test         # Run tests
smmake clean        # Clean build artifacts
smmake --help | -h  # Shows you the help documentation
```