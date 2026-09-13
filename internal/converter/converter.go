// Package converter aplica la conversión de una imagen individual usando bimg.
package converter

import (
	"fmt"
	"path/filepath"

	"github.com/h2non/bimg"

	"sharpy/internal/format"
	"sharpy/internal/naming"
	"sharpy/internal/scanner"
)

type Result struct {
	Src    string
	Dest   string
	OK     bool
	Reason string
	// DryRun indica que este resultado es una simulación: no se leyó,
	// escribió ni borró ningún archivo real.
	DryRun bool
}

// Params agrupa las opciones de conversión relevantes para una sola imagen,
// desacoplado de cliopts.Options para no atar este paquete al parser de CLI.
type Params struct {
	Force      bool
	Quality    int
	HasQuality bool
	Rename     naming.Strategy
	// DryRun, si es true, calcula el destino y valida colisiones/formato
	// pero no lee la imagen ni escribe nada en disco.
	DryRun bool
}

// DestPath calcula la ruta de salida para src, dado el formato de salida
// y la estrategia de renombrado.
func DestPath(src string, outFmt format.OutputFormat, rename naming.Strategy) string {
	dir := filepath.Dir(src)
	base := naming.Resolve(src, rename)
	return filepath.Join(dir, base+format.Ext(outFmt))
}

// ConvertOne lee, convierte y escribe una sola imagen. No borra el original:
// eso es responsabilidad del llamador (ver runner), para mantener esta función
// enfocada solo en la conversión en sí.
func ConvertOne(src string, outFmt format.OutputFormat, p Params) Result {
	dest := DestPath(src, outFmt, p.Rename)

	if !p.Force && scanner.FileExists(dest) {
		return Result{Src: src, Dest: dest, OK: false, Reason: "destino ya existe (usa --force para sobrescribir)"}
	}

	// Validamos que el formato esté soportado incluso en dry-run, para que
	// la simulación detecte el mismo tipo de errores que una corrida real.
	if _, ok := format.BimgOptions(outFmt, p.Quality, p.HasQuality); !ok {
		return Result{Src: src, Dest: dest, OK: false, Reason: fmt.Sprintf("Formato de salida no manejado: %s", outFmt)}
	}

	if p.DryRun {
		return Result{Src: src, Dest: dest, OK: true, DryRun: true, Reason: "(dry-run, no se escribió nada)"}
	}

	buf, err := bimg.Read(src)
	if err != nil {
		return Result{Src: src, Dest: dest, OK: false, Reason: err.Error()}
	}

	// Ya validamos arriba que el formato es soportado, así que el ok=true acá.
	bimgOpts, _ := format.BimgOptions(outFmt, p.Quality, p.HasQuality)

	out, err := bimg.NewImage(buf).Process(bimgOpts)
	if err != nil {
		return Result{Src: src, Dest: dest, OK: false, Reason: err.Error()}
	}

	if err := bimg.Write(dest, out); err != nil {
		return Result{Src: src, Dest: dest, OK: false, Reason: err.Error()}
	}

	return Result{Src: src, Dest: dest, OK: true}
}
