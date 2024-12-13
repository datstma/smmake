package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

var (
	DEBUG   bool
	VERSION = "v0.1.2"
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

		if DEBUG {
			fmt.Printf("Target '%s' not found in Makefile\n", targetName) //DEDUG
		}
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
	fmt.Println("  --debug        Enable debug mode")
	fmt.Println("\nExamples:")
	fmt.Println("  smmake         # Run the default target")
	fmt.Println("  smmake test    # Run the 'test' target")
	fmt.Println("  smmake -f custom.mk build  # Use 'custom.mk' file and run 'build' target")
	fmt.Println("  smmake --debug build  # Run 'build' target with debug output")
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {

	args := parseArgs(os.Args[1:])

	if args.showHelp {
		printHelp()
		return nil
	}

	if args.showVersion {
		fmt.Println("smmake version", VERSION)
		return nil
	}

	if DEBUG {
		fmt.Println("Debug mode enabled")
	}

	fmt.Printf("Attempting to parse Makefile: %s\n", args.makefilePath)
	makefile, err := ParseMakefile(args.makefilePath)
	if err != nil {
		return fmt.Errorf("error parsing Makefile: %w", err)
	}
	fmt.Println("Makefile parsed successfully")

	if args.targetName == "" {
		args.targetName = "all" // Default target
	}

	fmt.Printf("Attempting to execute target: %s\n", args.targetName)
	if err := makefile.ExecuteTarget(args.targetName); err != nil {
		return fmt.Errorf("error executing target: %w", err)
	}

	fmt.Println("Target execution completed")
	return nil
}

type arguments struct {
	showHelp     bool
	showVersion  bool
	makefilePath string
	targetName   string
}

func parseArgs(args []string) arguments {
	result := arguments{
		makefilePath: "Makefile",
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			result.showHelp = true
			return result
		case "-v", "--version":
			result.showVersion = true
			return result
		case "--debug":
			DEBUG = true
		case "-f", "--file":
			if i+1 < len(args) {
				result.makefilePath = args[i+1]
				i++
			} else {
				log.Fatal("Error: -f or --file option requires a filename")
			}
		default:
			if result.targetName == "" {
				result.targetName = args[i]
			}
		}
	}

	return result
}
