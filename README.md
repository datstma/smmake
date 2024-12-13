# smmake - Simple Multi-platform Make

smmake is a cross-platform alternative to the traditional Unix `make` utility. It's designed to provide `make`-like functionality for Windows and other operating systems that don't have native `make` support, without requiring the installation of Win32 GNU tools.

## Features

- Parse and execute Makefiles
- Support for targets, dependencies, and commands
- Variable expansion
- Pattern rules
- Parallel execution of dependencies
- Cross-platform compatibility

## Installation

To install smmake, you need to have Go installed on your system. Then, you can use the following command:

```bash
go get github.com/datstma/smmake