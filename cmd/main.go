package main

import (
	"bufio"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

// Target represents a make target and its commands
type Target struct {
	Name         string
	Commands     []Command
	Dependencies []string
	Pattern      bool
	PatternFrom  string
	PatternTo    string
}

type Command struct {
	Cmd    string
	Silent bool
}

// Makefile represents the parsed makefile
type Makefile struct {
	Targets    map[string]*Target
	Variables  map[string]string
	mutex      sync.Mutex
	executed   map[string]bool
	processing map[string]bool
}

// NewMakefile creates a new Makefile instance
func NewMakefile() *Makefile {
	return &Makefile{
		Targets:    make(map[string]*Target),
		Variables:  make(map[string]string),
		executed:   make(map[string]bool),
		processing: make(map[string]bool),
	}
}

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
//
// ParseMakefile reads and parses a Makefile
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
		line := scanner.Text()                 // Don't trim here
		fmt.Printf("Parsing line: %s\n", line) //DEBUG

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

// findMatchingPatternRule finds a pattern rule that matches the target
func (m *Makefile) findMatchingPatternRule(target string) *Target {
	for _, t := range m.Targets {
		if !t.Pattern {
			continue
		}
		pattern := fmt.Sprintf("%s(.*)%s", t.PatternFrom, t.PatternTo)
		if matched, _ := regexp.MatchString(pattern, target); matched {
			return t
		}
	}
	return nil
}

// ExecuteTarget runs the commands for a specified target
func (m *Makefile) ExecuteTarget(targetName string) error {
	m.mutex.Lock()
	if m.processing[targetName] {
		m.mutex.Unlock()
		return fmt.Errorf("circular dependency detected for target '%s'", targetName)
	}
	if m.executed[targetName] {
		m.mutex.Unlock()
		return nil
	}
	m.processing[targetName] = true
	m.mutex.Unlock()

	target := m.Targets[targetName]
	if target == nil {
		fmt.Printf("Target '%s' not found in Makefile\n", targetName) //DEDUG
		// Check for pattern rules
		if patternTarget := m.findMatchingPatternRule(targetName); patternTarget != nil {
			target = patternTarget
		} else {
			// Check if it's a file
			if _, err := os.Stat(targetName); err == nil {
				m.mutex.Lock()
				m.processing[targetName] = false
				m.executed[targetName] = true
				m.mutex.Unlock()
				return nil
			}
			return fmt.Errorf("target '%s' not found", targetName)
		}
	}

	// Execute dependencies in parallel
	var wg sync.WaitGroup
	errChan := make(chan error, len(target.Dependencies))

	for _, dep := range target.Dependencies {
		wg.Add(1)
		go func(dep string) {
			defer wg.Done()
			if err := m.ExecuteTarget(dep); err != nil {
				errChan <- fmt.Errorf("error in dependency '%s': %v", dep, err)
			}
		}(dep)
	}

	// Wait for all dependencies to complete
	wg.Wait()
	close(errChan)

	// Check for dependency errors
	for err := range errChan {
		return err
	}

	// Execute commands for this target
	for _, cmd := range target.Commands {
		fmt.Printf("Executing: %s\n", cmd.Cmd)

		parts := strings.Fields(cmd.Cmd)
		if len(parts) == 0 {
			continue
		}

		command := exec.Command(parts[0], parts[1:]...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			return fmt.Errorf("error executing command '%s': %v", cmd.Cmd, err)
		}
	}

	m.mutex.Lock()
	m.processing[targetName] = false
	m.executed[targetName] = true
	m.mutex.Unlock()

	return nil
}

func printHelp() {
	fmt.Println("smmake - Simple Multi-platform Make")
	fmt.Println("\nUsage:")
	fmt.Println("  smmake [options] [target]")
	fmt.Println("\nOptions:")
	fmt.Println("  -h, --help     Show this help message")
	fmt.Println("  -f, --file     Specify a Makefile (default is 'Makefile')")
	fmt.Println("  -v, --version  Show version information")
	fmt.Println("\nExamples:")
	fmt.Println("  smmake         # Run the default target")
	fmt.Println("  smmake test    # Run the 'test' target")
	fmt.Println("  smmake -f custom.mk build  # Use 'custom.mk' file and run 'build' target")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: smmake <target>")
		fmt.Println("Run 'smmake --help' for more information.")
		os.Exit(1)
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Check for help or version flags
	switch os.Args[1] {
	case "-h", "--help":
		printHelp()
		os.Exit(0)
	case "-v", "--version":
		version := os.Getenv("VERSION")
		fmt.Println("smmake version " + version)
		os.Exit(0)
	}

	makefilePath := "Makefile"
	targetName := ""

	// Parse command-line arguments
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "-f" || arg == "--file" {
			if i+1 < len(os.Args) {
				makefilePath = os.Args[i+1]
				i++
			} else {
				fmt.Println("Error: -f or --file option requires a filename")
				os.Exit(1)
			}
		} else {
			targetName = arg
			break
		}
	}

	fmt.Printf("Attempting to parse Makefile: %s\n", makefilePath)
	makefile, err := ParseMakefile(makefilePath)
	if err != nil {
		log.Fatalf("Error parsing Makefile: %v", err)
	}
	fmt.Println("Makefile parsed successfully")

	if targetName == "" {
		targetName = "all" // Default target
	}

	fmt.Printf("Attempting to execute target: %s\n", targetName)
	if err := makefile.ExecuteTarget(targetName); err != nil {
		log.Fatalf("Error executing target: %v", err)
	}
	fmt.Println("Target execution completed")
}
