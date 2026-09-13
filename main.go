// sharpy - Conversor de imágenes CLI (puerto a Go del proyecto original en Node.js/Sharp)
//
// Usa bimg (bindings de libvips) como motor de conversión, el mismo motor que
// usa la librería "sharp" de Node.js, para mantener paridad de resultados.
package main

import (
	"fmt"
	"os"
	"strings"

	"sharpy/internal/cliopts"
	"sharpy/internal/format"
	"sharpy/internal/runner"
	"sharpy/internal/scanner"
)

func main() {
	outFmt, opts := cliopts.Parse(os.Args[1:])

	stat, err := os.Stat(opts.Dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Carpeta no encontrada: %s\n", opts.Dir)
		os.Exit(1)
	}
	if !stat.IsDir() {
		fmt.Fprintf(os.Stderr, "La ruta no es carpeta: %s\n", opts.Dir)
		os.Exit(1)
	}

	files, err := scanner.ListImages(opts.Dir, opts.Recursive, format.ExcludeExt(outFmt))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error listando imágenes:", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Println("No se encontraron imágenes soportadas que no estén ya en el formato de destino.")
		return
	}

	fmt.Printf("Encontradas %d imagen(es). Convirtiendo a %s... [rename: %s]\n",
		len(files), strings.ToUpper(string(outFmt)), opts.Rename)

	summary := runner.Run(files, outFmt, opts)

	fmt.Printf("\nCompletado: %d convertido(s), %d con error.\n", summary.OK, summary.Fail)
}
