package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ParseMakefile reads and parses a Makefile.
//
// It processes the file line by line, identifying targets, dependencies, commands,
// and variable definitions. It creates a Makefile struct that represents the
// parsed content of the Makefile.
//
// Parameters:
//   - filename: A string representing the path to the Makefile to be parsed.
//
// Returns:
//   - *Makefile: A pointer to a Makefile struct containing the parsed information.
//   - error: An error if any occurred during the parsing process, nil otherwise.
func ParseMakefile(filename string) (*Makefile, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening makefile: %v", err)
	}
	defer file.Close()

	makefile := NewMakefile()
	scanner := bufio.NewScanner(file)
	var currentTarget *Target

	for scanner.Scan() {
		line := scanner.Text()
		if DEBUG {
			fmt.Printf("Parsing line: %s\n", line) //DEBUG
		}
		// Skip empty lines and comments
		if len(strings.TrimSpace(line)) == 0 || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		// Handle variable definitions
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				varName := strings.TrimSpace(parts[0])
				varValue := strings.TrimSpace(parts[1])
				makefile.Variables[varName] = varValue
				continue
			}
		}

		// Check if this is a target definition
		if !strings.HasPrefix(line, "\t") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			targetName := strings.TrimSpace(parts[0])

			// Handle pattern rules
			if strings.Contains(targetName, "%") {
				pattern := strings.Split(targetName, "%")
				if len(pattern) == 2 {
					currentTarget = &Target{
						Name:        targetName,
						Commands:    make([]Command, 0),
						Pattern:     true,
						PatternFrom: pattern[0],
						PatternTo:   pattern[1],
					}
				}
			} else {
				currentTarget = &Target{
					Name:         targetName,
					Commands:     make([]Command, 0),
					Dependencies: make([]string, 0),
				}
			}

			// Parse dependencies
			if len(parts) > 1 {
				deps := strings.Fields(parts[1])
				currentTarget.Dependencies = deps
			}

			makefile.Targets[targetName] = currentTarget
			continue
		}

		// If line starts with a tab and we have a current target, it's a command
		if strings.HasPrefix(line, "\t") {
			if currentTarget != nil {
				command := strings.TrimPrefix(line, "\t")
				silent := false
				if strings.HasPrefix(command, "@") {
					silent = true
					command = strings.TrimPrefix(command, "@")
				}
				command = strings.TrimSpace(command)
				// Expand variables in command
				command = makefile.expandVariables(command)
				currentTarget.Commands = append(currentTarget.Commands, Command{
					Cmd:    command,
					Silent: silent,
				})
			}
		}
	}

	// At the end of the function, print out the parsed targets //DEBUG
	if DEBUG {
		for targetName, target := range makefile.Targets {
			fmt.Printf("Parsed target: %s\n", targetName)
			fmt.Printf("  Commands:\n")
			for _, cmd := range target.Commands {
				silentStr := ""
				if cmd.Silent {
					silentStr = "(silent) "
				}
				fmt.Printf("    %s%s\n", silentStr, cmd.Cmd)
			}
			fmt.Printf("  Dependencies: %v\n", target.Dependencies)
		}
	}

	return makefile, nil
}

// expandVariables replaces $(VAR) or ${VAR} with their values
func (m *Makefile) expandVariables(str string) string {
	re := regexp.MustCompile(`\$[\(\{]([^\)\}]+)[\)\}]`)
	return re.ReplaceAllStringFunc(str, func(match string) string {
		varName := match[2 : len(match)-1]
		if val, ok := m.Variables[varName]; ok {
			return val
		}
		return match
	})
}
