# smmake - Simple Multi-platform Make

[![Go Report Card](https://goreportcard.com/badge/github.com/datstma/smmake)](https://goreportcard.com/report/github.com/datstma/smmake)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

smmake is a powerful, cross-platform alternative to the traditional Unix `make` utility. It brings the familiar functionality of `make` to Windows and other operating systems without requiring the installation of odd tools such as Win32 GNU tools.

#### - The primary target audience is Windows x86_64/arm64 developers since there's no native make support built in. But it works on Linux and Mac just as well as a bonus should you prefer smmake over GNU make (for some unkonwn reason :) )

## 📚 Parses your Makefile in the way you'd expect

smmake supports most common Makefile features you'd expect, has the same syntax as regular Unix Make and some extras:

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
### Windows installer, latest release

https://github.com/datstma/smmake/releases/latest

Or,build it from source on your own

Requires that you have Golang(Go) installed locally. Follow the installation instructions on https://go.dev/

Assuming that you have a working Go development environment installed, just execute the following:

Windows
```bash
git clone https://github.com/datstma/smmake.git
cd smmake/cmd
go build -o smmake.exe cmd/main.go cmd/parser.go
```
Will create a statically linked binary named smmake.exe

Linux/ Mac/ *nix
```bash
git clone https://github.com/datstma/smmake.git
cd smmake/cmd
go build -o smmake cmd/main.go cmd/parser.go
```
Will create a statically linked binary named smmake

### Usage

Then run smmake (instead of make)
```bash
smmake build       # Build the project
smmake test         # Run tests
smmake clean        # Clean build artifacts
smmake --help | -h  # Shows you the help documentation
```
Pro-tip, copy your smmake binary into a suitable folder and then add that folder into your $PATH environment variable. 
If you use the Windows installer provided in the releases section, the PATH with will be automatically added for you. 

## Author
Stefan Månsby
stefan@mansby.se