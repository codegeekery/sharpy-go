// Package cliopts parses sharpy's command-line arguments
// and produces the Options consumed by the rest of the application.
package cliopts

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sharpy/internal/format"
	"sharpy/internal/naming"
)

type Options struct {
	Dir            string
	Recursive      bool
	Quality        int
	HasQuality     bool
	Force          bool
	RemoveOriginal bool
	Rename         naming.Strategy
	DryRun         bool
}

const helpTemplate = `Usage:
  sharpy <format> [options]

<format>:
  %s

Options:
  --dir <path>            Folder to process (default: current folder)
  -r, --recursive         Look for images in subfolders
  -q, --quality <n>       Quality (0-100) for lossy formats (jpeg/webp/avif/tiff)
  -f, --force             Overwrite if the destination already exists
  -n, --dry-run           Simulate the run without writing or deleting anything
  --rm, --remove-original Delete the original file after a successful conversion
  --rename <strategy>     Naming strategy for the output file:
                            original   Keeps the original filename (default)
                            snowflake  Unique ID based on timestamp
                            uuid       Random UUID v4
  -h, --help              Show help

Examples:
  sharpy webp
  sharpy avif -q 70 -r --rm
  sharpy jpeg --dir ./photos -f -q 85
  sharpy webp --rename snowflake
  sharpy avif --rename uuid -r
  sharpy webp --dry-run --rm -r
`

// ShowHelpAndExit prints the help (and optionally an error message)
// and terminates the process. Exit code 1 if there was an error message, 0 otherwise.
func ShowHelpAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, "\nError: %s\n\n", msg)
	}
	fmt.Printf(helpTemplate, format.Names(format.Supported))
	if msg != "" {
		os.Exit(1)
	}
	os.Exit(0)
}

// Parse parses os.Args and returns the chosen output format along
// with the options. On any invalid argument, it terminates the process
// via ShowHelpAndExit (same behavior as the original version).
func Parse(argv []string) (format.OutputFormat, Options) {
	if len(argv) == 0 {
		ShowHelpAndExit("You must specify an output format...")
	}
	if argv[0] == "--help" || argv[0] == "-h" {
		ShowHelpAndExit("")
	}

	formatArg := format.OutputFormat(strings.ToLower(argv[0]))
	if !format.Contains(format.Supported, formatArg) {
		ShowHelpAndExit(fmt.Sprintf("Unsupported format: %s. Supported: %s", formatArg, format.Names(format.Supported)))
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not get current directory:", err)
		os.Exit(1)
	}

	options := Options{
		Dir:       cwd,
		Recursive: false,
		Force:     false,
		Rename:    naming.Original,
	}

	for i := 1; i < len(argv); i++ {
		arg := argv[i]
		switch {
		case arg == "--recursive" || arg == "-r":
			options.Recursive = true
		case arg == "--force" || arg == "-f":
			options.Force = true
		case arg == "--dry-run" || arg == "-n":
			options.DryRun = true
		case arg == "--remove-original" || arg == "--rm":
			options.RemoveOriginal = true
		case arg == "--dir":
			if i+1 >= len(argv) {
				ShowHelpAndExit("Missing value for --dir")
			}
			i++
			abs, err := filepath.Abs(argv[i])
			if err != nil {
				ShowHelpAndExit("Invalid path for --dir")
			}
			options.Dir = abs
		case arg == "--quality" || arg == "-q":
			if i+1 >= len(argv) {
				ShowHelpAndExit("Invalid quality for --quality")
			}
			i++
			n, err := strconv.Atoi(argv[i])
			if err != nil {
				ShowHelpAndExit("Invalid quality for --quality")
			}
			options.Quality = n
			options.HasQuality = true
		case arg == "--rename":
			if i+1 >= len(argv) {
				ShowHelpAndExit(fmt.Sprintf("Invalid naming strategy: \"\". Options: %s", naming.Names(naming.Supported)))
			}
			i++
			next := naming.Strategy(strings.ToLower(argv[i]))
			if !naming.Contains(naming.Supported, next) {
				ShowHelpAndExit(fmt.Sprintf("Invalid naming strategy: \"%s\". Options: %s", next, naming.Names(naming.Supported)))
			}
			options.Rename = next
		case arg == "--help" || arg == "-h":
			ShowHelpAndExit("")
		default:
			ShowHelpAndExit(fmt.Sprintf("Unknown argument: %s", arg))
		}
	}

	return format.Normalize(formatArg), options
}
