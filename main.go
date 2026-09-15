// sharpy - CLI image converter (Go port of the original Node.js/Sharp project)
//
// Uses govips (libvips bindings) as the conversion engine, the same engine
// used by Node.js's "sharp" library, to keep result parity.
package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"

	"sharpy/internal/cliopts"
	"sharpy/internal/format"
	"sharpy/internal/runner"
	"sharpy/internal/scanner"
)

func main() {
	// Parse args first: -h/--help and invalid-argument cases call os.Exit
	// inside ShowHelpAndExit, so we avoid initializing libvips unless we're
	// actually going to convert something.
	outFmt, opts := cliopts.Parse(os.Args[1:])

	stat, err := os.Stat(opts.Dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Folder not found: %s\n", opts.Dir)
		os.Exit(1)
	}
	if !stat.IsDir() {
		fmt.Fprintf(os.Stderr, "Path is not a folder: %s\n", opts.Dir)
		os.Exit(1)
	}

	files, err := scanner.ListImages(opts.Dir, opts.Recursive, format.ExcludeExt(outFmt))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error listing images:", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Println("No supported images found that aren't already in the destination format.")
		return
	}

	// More aggressive GC to keep peak memory down under heavy concurrent
	// image processing (large short-lived buffers per image), at a small
	// CPU cost.
	debug.SetGCPercent(50)
        debug.SetMemoryLimit(4 << 30)

	// Silence govips/libvips internal logging (info/debug noise) so only
	// our own output is shown to the user.
	vips.LoggingSettings(nil, vips.LogLevelError)

	// govips needs to explicitly initialize libvips before use, and free
	// its resources on exit. bimg used to do this implicitly; govips doesn't.
	// ConcurrencyLevel is set to 1 because we already parallelize across
	// images ourselves (see runner); letting libvips also multithread a
	// single image would fight our own worker pool for CPU.
	vips.Startup(&vips.Config{
		ConcurrencyLevel: 4, 			// Limit concurrent threads used by vips 
		MaxCacheFiles: 100, 			// Max number of intermediate files vips can
		MaxCacheMem: 500 * 1024 * 1024, 	// 500MB max memory for vips
		MaxCacheSize: 1000, 	  		// Max number of operations to keep in cache
        	// ReportInputBufferLeaks: true, 	// Useful for debugging memory leaks
	})
	defer vips.Shutdown()

	fmt.Printf("Found %d image(s). Converting to %s... [rename: %s]\n",
		len(files), strings.ToUpper(string(outFmt)), opts.Rename)

	summary := runner.Run(files, outFmt, opts)

	fmt.Printf("\nCompleted: %d converted, %d failed.\n", summary.OK, summary.Fail)
}
