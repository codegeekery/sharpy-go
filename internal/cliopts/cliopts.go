// Package cliopts parsea los argumentos de línea de comandos de sharpy
// y produce las Options que consume el resto de la aplicación.
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

const helpTemplate = `Uso:
  sharpy <formato> [opciones]

<formato>:
  %s

Opciones:
  --dir <ruta>            Carpeta a procesar (por defecto, carpeta actual)
  -r, --recursive         Buscar imágenes en subcarpetas
  -q, --quality <n>       Calidad (0-100) para formatos con pérdida (jpeg/webp/avif/tiff)
  -f, --force             Sobrescribir si el destino ya existe
  -n, --dry-run           Simular la corrida sin escribir ni borrar nada
  --rm, --remove-original Borrar el archivo original si la conversión fue exitosa
  --rename <strategy>     Estrategia de nombre para el archivo de salida:
                            original   Mantiene el nombre del archivo (default)
                            snowflake  ID único basado en timestamp
                            uuid       UUID v4 aleatorio
  -h, --help              Mostrar ayuda

Ejemplos:
  sharpy webp
  sharpy avif -q 70 -r --rm
  sharpy jpeg --dir ./fotos -f -q 85
  sharpy webp --rename snowflake
  sharpy avif --rename uuid -r
  sharpy webp --dry-run --rm -r
`

// ShowHelpAndExit imprime la ayuda (y opcionalmente un mensaje de error)
// y termina el proceso. exit code 1 si hubo mensaje de error, 0 si no.
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

// Parse interpreta os.Args y devuelve el formato de salida elegido junto
// con las opciones. Ante cualquier argumento inválido, termina el proceso
// vía ShowHelpAndExit (mismo comportamiento que la versión original).
func Parse(argv []string) (format.OutputFormat, Options) {
	if len(argv) == 0 {
		ShowHelpAndExit("Debes indicar un formato de salida...")
	}
	if argv[0] == "--help" || argv[0] == "-h" {
		ShowHelpAndExit("")
	}

	formatArg := format.OutputFormat(strings.ToLower(argv[0]))
	if !format.Contains(format.Supported, formatArg) {
		ShowHelpAndExit(fmt.Sprintf("Formato no soportado: %s. Soportados: %s", formatArg, format.Names(format.Supported)))
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "No se pudo obtener el directorio actual:", err)
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
				ShowHelpAndExit("Falta valor para --dir")
			}
			i++
			abs, err := filepath.Abs(argv[i])
			if err != nil {
				ShowHelpAndExit("Ruta inválida para --dir")
			}
			options.Dir = abs
		case arg == "--quality" || arg == "-q":
			if i+1 >= len(argv) {
				ShowHelpAndExit("Calidad inválida para --quality")
			}
			i++
			n, err := strconv.Atoi(argv[i])
			if err != nil {
				ShowHelpAndExit("Calidad inválida para --quality")
			}
			options.Quality = n
			options.HasQuality = true
		case arg == "--rename":
			if i+1 >= len(argv) {
				ShowHelpAndExit(fmt.Sprintf("Estrategia de renombrado no válida: \"\". Opciones: %s", naming.Names(naming.Supported)))
			}
			i++
			next := naming.Strategy(strings.ToLower(argv[i]))
			if !naming.Contains(naming.Supported, next) {
				ShowHelpAndExit(fmt.Sprintf("Estrategia de renombrado no válida: \"%s\". Opciones: %s", next, naming.Names(naming.Supported)))
			}
			options.Rename = next
		case arg == "--help" || arg == "-h":
			ShowHelpAndExit("")
		default:
			ShowHelpAndExit(fmt.Sprintf("Argumento desconocido: %s", arg))
		}
	}

	return format.Normalize(formatArg), options
}
