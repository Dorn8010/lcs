// lcs - library command search
// Version 0.92 (Robust CSV Parser)
// © 2025 by Alexander Dorn, MIT license
// https://github.com/Dorn8010/lcs

// To compile on Linux :
// sudo apt install golang && go build -o lcs lcs.go
// To compile on Mac :
// brew install go && go build -o lcs lcs.go
// To install locally
// sudo cp lcs /usr/local/bin/

package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io" // Added for EOF handling
	"os"
	"os/exec"
	"os/signal" // Needed for signal masking
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

// Entry represents a command in the DB
type Entry struct {
	Description   string
	Command       string
	OriginalIndex int // Needed for removal/edit
}

const appVersion = "0.92"

func main() {
	// 1. Setup Flags
	var fastChoice int
	flag.IntVar(&fastChoice, "f", 0, "Fast selection of option number")
	flag.IntVar(&fastChoice, "fast", 0, "Fast selection of option number")

	helpFlag := flag.Bool("h", false, "Show help")
	verboseFlag := flag.Bool("v", false, "Verbose output")
	versionFlag := flag.Bool("version", false, "Show version")
	addFlag := flag.Bool("add", false, "Add a new entry")
	removeFlag := flag.Bool("remove", false, "Remove an entry")
	editFlag := flag.Bool("edit", false, "Edit an entry")
	printFlag := flag.Bool("print", false, "Print command only (don't execute)")
	copyFlag := flag.Bool("copy", false, "Copy command to clipboard (don't execute)")
	dbPathFlag := flag.String("db", "", "Path to custom DB file")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("lcs version %s\n", appVersion)
		return
	}

	if *helpFlag {
		fmt.Println("============================================")
		fmt.Println("Library Command Search tool for CLI commands")
		fmt.Println("Store and find long commands easily")
		fmt.Println("© 2025 by Alexander Dorn        MIT license")
		fmt.Printf("                                Version %s\n", appVersion)
		fmt.Println("============================================\n")
		fmt.Println("Usage: lcs [option] search_term_last\n")
		fmt.Println("============================================")
		fmt.Println("Searches for a command in the description")
		fmt.Println("and offers the findings for execution\n")
		fmt.Println("The DB contains an explanation and")
		fmt.Println("the command with optional variables\n")
		fmt.Println("~/.lcs-db.csv is a ; separated CSV")
		fmt.Println("Here an example of an entry :")
		fmt.Println("Echo test;echo \"Hello World\" ")
		fmt.Println("\nOptions :")
		fmt.Println("  --help, -h     Show this help")
		fmt.Println("  --version      Show version info")
		fmt.Println("  --verbose, -v  Show verbose logging")
		fmt.Println("  --fast, -f     Fast select option number (e.g. -f 2)")
		fmt.Println("  --print        Print command only")
		fmt.Println("  --copy         Copy command to clipboard")
		fmt.Println("                 (no execution)")
		fmt.Println("  --add          Add a new command")
		fmt.Println("                 Usage: lcs --add \"Desc\" \"Cmd\"")
		fmt.Println("                 or interactive mode")
		fmt.Println("  --edit         Search, remove and re-add/edit")
		fmt.Println("  --remove       Search and remove a command")
		fmt.Println("  --db           Path to custom database")
		fmt.Println("                 default: ~/.lcs-db.csv")
		fmt.Println("\nUsing Variables:")
		fmt.Println("  You can define variables in commands to be filled at runtime.")
		fmt.Println("  Syntax: {\"Label\":\"DefaultValue\"}")
		fmt.Println("  Example: ssh -i {\"KeyFile\":\"~/.ssh/id_rsa\"} user@host")
		return
	}

	// 2. Determine DB Path
	dbPath := *dbPathFlag
	if dbPath == "" {
		usr, err := user.Current()
		if err != nil {
			fmt.Println("Error getting user home:", err)
			os.Exit(1)
		}
		dbPath = filepath.Join(usr.HomeDir, ".lcs-db.csv")
	}

	if *verboseFlag {
		fmt.Println("Using DB:", dbPath)
	}

	consoleReader := bufio.NewReader(os.Stdin)

	// --- ADD FEATURE ---
	if *addFlag {
		var desc, cmdStr string
		args := flag.Args()
		if len(args) >= 2 {
			desc = args[0]
			cmdStr = strings.Join(args[1:], " ")
		} else {
			fmt.Println("--- Add New Command ---")
			fmt.Print("Description: ")
			desc, _ = consoleReader.ReadString('\n')
			desc = strings.TrimSpace(desc)
			fmt.Print("Command: ")
			cmdStr, _ = consoleReader.ReadString('\n')
			cmdStr = strings.TrimSpace(cmdStr)
		}

		if desc == "" || cmdStr == "" {
			fmt.Println("Error: Description and Command cannot be empty.")
			os.Exit(1)
		}

		f, err := os.OpenFile(dbPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Error opening DB for writing:", err)
			os.Exit(1)
		}
		defer f.Close()

		writer := csv.NewWriter(f)
		writer.Comma = ';'
		if err := writer.Write([]string{desc, cmdStr}); err != nil {
			fmt.Println("Error writing to DB:", err)
			os.Exit(1)
		}
		writer.Flush()
		fmt.Println("Entry added successfully.")
		return
	}

	// 3. Open and Parse CSV (Robust Mode)
	file, err := os.Open(dbPath)
	if err != nil {
		fmt.Printf("Error opening DB (%s): %v\n", dbPath, err)
		fmt.Println("Please create the file with format: Description;Command")
		os.Exit(1)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	reader.Comment = '#'     // Treat lines starting with # as comments
	reader.LazyQuotes = true // Allow quotes to appear in non-quoted fields

	var rawRecords [][]string

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		// If there is a parsing error, skip this line and continue
		if err != nil {
			continue
		}
		// Ensure valid record
		if len(record) < 2 {
			continue
		}
		rawRecords = append(rawRecords, record)
	}

	// 4. Search Logic
	searchTerm := strings.ToLower(strings.Join(flag.Args(), " "))
	var matches []Entry

	for idx, record := range rawRecords {
		if len(record) < 2 {
			continue
		}
		desc := record[0]
		cmd := record[1]

		if searchTerm == "" || strings.Contains(strings.ToLower(desc), searchTerm) || strings.Contains(strings.ToLower(cmd), searchTerm) {
			matches = append(matches, Entry{
				Description:   desc,
				Command:       cmd,
				OriginalIndex: idx,
			})
		}
	}

	if len(matches) == 0 {
		fmt.Println("No matches found.")
		return
	}

	// 5. Selection Logic
	var selectionIndex int

	if fastChoice > 0 {
		// --- FAST SELECTION MODE ---
		if fastChoice > len(matches) {
			fmt.Printf("Error: Fast choice %d is out of range. Only %d matches found.\n", fastChoice, len(matches))
			os.Exit(1)
		}
		selectionIndex = fastChoice
		if *verboseFlag {
			fmt.Printf("Fast selected [%d]: %s\n", selectionIndex, matches[selectionIndex-1].Description)
		}
	} else if len(matches) == 1 {
		selectionIndex = 1
		fmt.Printf("Found 1 match: %s\n", matches[0].Description)
		fmt.Printf("Cmd : %s\n", matches[0].Command)
	} else {
		if *removeFlag {
			fmt.Println("Select command to REMOVE:")
		} else if *editFlag {
			fmt.Println("Select command to EDIT:")
		} else {
			fmt.Println("Found commands:")
		}

		for i, m := range matches {
			fmt.Printf("[%d] %s \n    Cmd: %s\n", i+1, m.Description, m.Command)
		}

		fmt.Print("\nSelect a number: ")
		inputStr, _ := consoleReader.ReadString('\n')
		inputStr = strings.TrimSpace(inputStr)

		var err error
		selectionIndex, err = strconv.Atoi(inputStr)
		if err != nil || selectionIndex < 1 || selectionIndex > len(matches) {
			fmt.Println("Invalid selection.")
			os.Exit(1)
		}
	}

	selectedEntry := matches[selectionIndex-1]

	// --- REMOVE / EDIT FEATURE ---
	if *removeFlag || *editFlag {
		targetOriginalIndex := selectedEntry.OriginalIndex
		var newRecords [][]string
		for i, rec := range rawRecords {
			if i != targetOriginalIndex {
				newRecords = append(newRecords, rec)
			}
		}

		f, err := os.OpenFile(dbPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Println("Error opening DB for rewrite:", err)
			os.Exit(1)
		}

		writer := csv.NewWriter(f)
		writer.Comma = ';'
		if err := writer.WriteAll(newRecords); err != nil {
			fmt.Println("Error saving DB:", err)
			os.Exit(1)
		}
		f.Close()

		if *removeFlag {
			fmt.Println("Entry removed successfully.")
			return
		}
	}

	// --- EDIT SPECIFIC ---
	if *editFlag {
		fmt.Println("\n--- Edit Entry (Press Enter to keep current) ---")

		fmt.Printf("Description [%s]: ", selectedEntry.Description)
		newDesc, _ := consoleReader.ReadString('\n')
		newDesc = strings.TrimSpace(newDesc)
		if newDesc == "" {
			newDesc = selectedEntry.Description
		}

		fmt.Printf("Command [%s]: ", selectedEntry.Command)
		newCmd, _ := consoleReader.ReadString('\n')
		newCmd = strings.TrimSpace(newCmd)
		if newCmd == "" {
			newCmd = selectedEntry.Command
		}

		f, err := os.OpenFile(dbPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Error opening DB for writing:", err)
			os.Exit(1)
		}
		defer f.Close()

		writer := csv.NewWriter(f)
		writer.Comma = ';'
		if err := writer.Write([]string{newDesc, newCmd}); err != nil {
			fmt.Println("Error writing to DB:", err)
			os.Exit(1)
		}
		writer.Flush()
		fmt.Println("Entry edited successfully.")
		return
	}

	// 7. Parameter Parsing
	selectedCmd := selectedEntry.Command
	re := regexp.MustCompile(`\{"([^"]+)":"([^"]*)"\}`)

	finalCmd := re.ReplaceAllStringFunc(selectedCmd, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		label := submatches[1]
		defaultValue := submatches[2]

		fmt.Printf("Input for '%s' [%s]: ", label, defaultValue)
		val, _ := consoleReader.ReadString('\n')
		val = strings.TrimSpace(val)

		if val == "" {
			return defaultValue
		}
		return val
	})

	// --- FEATURE: PRINT ONLY ---
	if *printFlag {
		fmt.Println(finalCmd)
		return
	}

	// --- FEATURE: COPY TO CLIPBOARD ---
	if *copyFlag {
		var copyCmd *exec.Cmd

		if runtime.GOOS == "darwin" {
			copyCmd = exec.Command("pbcopy")
		} else if runtime.GOOS == "linux" {
			// Try xclip first
			_, err := exec.LookPath("xclip")
			if err == nil {
				copyCmd = exec.Command("xclip", "-selection", "clipboard")
			} else {
				// Try xsel
				_, err := exec.LookPath("xsel")
				if err == nil {
					copyCmd = exec.Command("xsel", "--clipboard", "--input")
				} else {
					fmt.Println("Error: Neither 'xclip' nor 'xsel' found. Please install one to use --copy on Linux.")
					os.Exit(1)
				}
			}
		} else {
			fmt.Println("Error: Clipboard copy not implemented for this OS.")
			os.Exit(1)
		}

		copyCmd.Stdin = strings.NewReader(finalCmd)
		if err := copyCmd.Run(); err != nil {
			fmt.Println("Error copying to clipboard:", err)
			os.Exit(1)
		}
		fmt.Println("Command copied to clipboard.")
		return
	}

	// 8. Execution - Robust Child Process Mode
	// This uses standard exec.Command but masks signals in Go so SSH/Vim
	// can handle them natively without Go killing the parent process.
	if *verboseFlag {
		fmt.Println("\nExecuting:", finalCmd)
	} else {
		if len(matches) > 1 {
			fmt.Println("\nExecuting...")
		}
	}

	// Setup command with TTY attached
	cmd := exec.Command("bash", "-c", finalCmd)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Create a channel to catch signals
	sigChan := make(chan os.Signal, 1)
	// Notify the channel for SIGINT (Ctrl+C) and SIGTERM
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start a goroutine that keeps the parent alive but does nothing
	// when a signal is received. This allows the child process (SSH/Bash),
	// which shares the terminal, to receive and handle the signal itself.
	go func() {
		for range sigChan {
			// Do nothing. Just swallowing the signal in the Go process
			// allows SSH to handle it naturally via the TTY.
		}
	}()

	err = cmd.Run()
	if err != nil {
		// Verify if it's a real error or just an exit code mismatch
		// (SSH often returns non-zero on disconnects, which is fine)
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		fmt.Println("Execution error:", err)
		os.Exit(1)
	}
}
