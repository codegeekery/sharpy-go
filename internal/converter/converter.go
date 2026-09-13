// Package converter aplica la conversión de una imagen individual usando govips.
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

// ConvertOne lee, convierte y escribe una sola imagen. No borra el original:
// eso es responsabilidad del llamador (ver runner).
func ConvertOne(src string, outFmt format.OutputFormat, p Params) Result {
	dest := DestPath(src, outFmt, p.Rename)

	if !p.Force && scanner.FileExists(dest) {
		return Result{Src: src, Dest: dest, OK: false, Reason: "destino ya existe (usa --force para sobrescribir)"}
	}

	// Validamos el formato incluso en dry-run, para que la simulación
	// detecte el mismo tipo de errores que una corrida real.
	if !format.IsSupported(outFmt) {
		return Result{Src: src, Dest: dest, OK: false, Reason: fmt.Sprintf("Formato de salida no manejado: %s", outFmt)}
	}

	if p.DryRun {
		return Result{Src: src, Dest: dest, OK: true, DryRun: true, Reason: "(dry-run, no se escribió nada)"}
	}

	img, err := vips.NewImageFromFile(src)
	if err != nil {
		return Result{Src: src, Dest: dest, OK: false, Reason: err.Error()}
	}
	defer img.Close()

	out, ok, err := format.Export(img, outFmt, p.Quality, p.HasQuality)
	if !ok {
		return Result{Src: src, Dest: dest, OK: false, Reason: fmt.Sprintf("Formato de salida no manejado: %s", outFmt)}
	}
	if err != nil {
		return Result{Src: src, Dest: dest, OK: false, Reason: err.Error()}
	}

	if err := os.WriteFile(dest, out, 0644); err != nil {
		return Result{Src: src, Dest: dest, OK: false, Reason: err.Error()}
	}

	return Result{Src: src, Dest: dest, OK: true}
}
