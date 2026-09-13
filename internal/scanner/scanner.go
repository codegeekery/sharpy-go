// Package scanner recorre el filesystem buscando imágenes de entrada soportadas.
package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"sharpy/internal/format"
)

// ListImages devuelve las rutas de todas las imágenes soportadas dentro de dir.
// Si recursive es true, también recorre subcarpetas. excludeExt (sin punto,
// ej. "jpg") permite omitir archivos que ya están en el formato destino.
func ListImages(dir string, recursive bool, excludeExt string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, ent := range entries {
		full := filepath.Join(dir, ent.Name())
		if ent.IsDir() {
			if recursive {
				sub, err := ListImages(full, true, excludeExt)
				if err != nil {
					return nil, err
				}
				files = append(files, sub...)
			}
			continue
		}
		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(ent.Name()), "."))
		if format.SupportedInputs[ext] && ext != excludeExt {
			files = append(files, full)
		}
	}
	return files, nil
}

// FileExists indica si path existe en el filesystem (sin distinguir motivo de error).
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
