// Package converter applies the conversion of a single image using govips.
package converter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidbyttow/govips/v2/vips"

	"sharpy/internal/format"
	"sharpy/internal/naming"
	"sharpy/internal/scanner"
)

type Result struct {
	Src    string
	Dest   string
	OK     bool
	Reason string
	DryRun bool
}

type Params struct {
	Force      bool
	Quality    int
	HasQuality bool
	Rename     naming.Strategy
	DryRun     bool
}

func DestPath(src string, outFmt format.OutputFormat, rename naming.Strategy) string {
	dir := filepath.Dir(src)
	base := naming.Resolve(src, rename)
	return filepath.Join(dir, base+format.Ext(outFmt))
}

// ConvertOne reads, converts, and writes a single image. It does not delete
// the original: that's the caller's responsibility (see runner).
func ConvertOne(src string, outFmt format.OutputFormat, p Params) Result {
	dest := DestPath(src, outFmt, p.Rename)

	if !p.Force && scanner.FileExists(dest) {
		return Result{Src: src, Dest: dest, OK: false, Reason: "destination already exists (use --force to overwrite)"}
	}

	// We validate the format even in dry-run, so the simulation
	// detects the same kind of errors as a real run.
	if !format.IsSupported(outFmt) {
		return Result{Src: src, Dest: dest, OK: false, Reason: fmt.Sprintf("Unsupported output format: %s", outFmt)}
	}

	if p.DryRun {
		return Result{Src: src, Dest: dest, OK: true, DryRun: true, Reason: "(dry-run, nothing was written)"}
	}

	img, err := vips.NewImageFromFile(src)
	if err != nil {
		return Result{Src: src, Dest: dest, OK: false, Reason: err.Error()}
	}

	out, ok, err := format.Export(img, outFmt, p.Quality, p.HasQuality)
	// Close the decoded image buffer as soon as encoding is done, instead
	// of waiting for a deferred Close at the end of the function — this
	// frees the (usually large, uncompressed) decode buffer before we even
	// get to the disk write below.
	img.Close()

	if !ok {
		return Result{Src: src, Dest: dest, OK: false, Reason: fmt.Sprintf("Unsupported output format: %s", outFmt)}
	}
	if err != nil {
		return Result{Src: src, Dest: dest, OK: false, Reason: err.Error()}
	}

	if err := os.WriteFile(dest, out, 0644); err != nil {
		return Result{Src: src, Dest: dest, OK: false, Reason: err.Error()}
	}

	return Result{Src: src, Dest: dest, OK: true}
}
