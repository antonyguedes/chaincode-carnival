package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/antonioforte/chaincode-carnival/agents/arena"
)

type ChaincodeFile struct {
	Path    string
	Label   string // human-friendly name shown in the menu
	ModTime int64
}

func main() {
	if len(os.Args) < 3 || os.Args[1] != "carnival" || os.Args[2] != "start" {
		fmt.Println("Usage: hfchance carnival start [path/to/chaincode.go]")
		return
	}

	// If a path was provided directly, skip the menu
	if len(os.Args) >= 4 {
		arena.Run(os.Args[3])
		return
	}

	fmt.Println("\033[32m[CHAINCODE CARNIVAL!!]\033[0m")
	fmt.Println("Welcome to the Ultimate Adversarial Arena Selection.")
	fmt.Println()

	files, err := discoverChaincodes("./chaincodes")
	if err != nil {
		fmt.Printf("Error scanning chaincodes: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No chaincode files found in ./chaincodes directory.")
		return
	}

	// Sort by most recent modification
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime > files[j].ModTime
	})

	fmt.Println("Select the Chaincode to drop into the Arena:")
	fmt.Println()
	for i, f := range files {
		fmt.Printf("  [%d] %-20s  %s\n", i+1, f.Label, f.Path)
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Selection > ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(files) {
		fmt.Println("Invalid selection. Aborting Carnival.")
		return
	}

	selectedFile := files[choice-1].Path
	arena.Run(selectedFile)
}

// discoverChaincodes walks ONLY the immediate subdirectories of root.
// For each subdirectory it finds the single primary .go file — the one
// directly inside the folder (not nested under vendor/ or any sub-package).
// This means vendor dependencies and helper packages are completely invisible
// in the selection menu.
func discoverChaincodes(root string) ([]ChaincodeFile, error) {
	var files []ChaincodeFile

	// List immediate subdirectories of root (e.g. chaincodes/hardcore)
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		chainDir := filepath.Join(root, entry.Name())

		// Look only at .go files directly inside this folder (depth 1)
		goFiles, err := filepath.Glob(filepath.Join(chainDir, "*.go"))
		if err != nil || len(goFiles) == 0 {
			continue
		}

		// Prefer main.go; otherwise take the first .go file found
		primary := ""
		for _, f := range goFiles {
			if filepath.Base(f) == "main.go" {
				primary = f
				break
			}
		}
		if primary == "" {
			primary = goFiles[0]
		}

		info, err := os.Stat(primary)
		if err != nil {
			continue
		}

		files = append(files, ChaincodeFile{
			Path:    primary,
			Label:   entry.Name(), // folder name = chaincode name
			ModTime: info.ModTime().Unix(),
		})
	}

	return files, nil
}

// StripComments removes all Go comments (// and /* */) from source code
// using the official go/parser so it handles edge cases like URLs in strings.
// The Analyzer receives clean code and must find vulnerabilities on its own.
func StripComments(src string) (string, error) {
	fset := token.NewFileSet()
	// parser.ParseComments is NOT set — comments are ignored during parsing
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		// If the source won't parse (e.g. build-tag-only file), fall back to
		// a simple line-by-line strip that handles // comments
		return stripCommentsSimple(src), nil
	}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, f); err != nil {
		return src, err
	}
	return buf.String(), nil
}

// stripCommentsSimple is the fallback: removes // line comments only.
func stripCommentsSimple(src string) string {
	var out strings.Builder
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue // drop full-line comment
		}
		// Inline comment: find // that's not inside a string
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = line[:idx]
		}
		out.WriteString(line + "\n")
	}
	return out.String()
}

