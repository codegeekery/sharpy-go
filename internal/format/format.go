// Package format define los formatos de salida soportados por sharpy
// y las conversiones necesarias hacia bimg (tipo de imagen, extensión, etc).
package format

import (
	"strings"

	"github.com/h2non/bimg"
)

type OutputFormat string

const (
	JPEG OutputFormat = "jpeg"
	JPG  OutputFormat = "jpg"
	PNG  OutputFormat = "png"
	WebP OutputFormat = "webp"
	AVIF OutputFormat = "avif"
	TIFF OutputFormat = "tiff"
)

// Supported es la lista canónica de formatos de salida aceptados por CLI.
var Supported = []OutputFormat{JPEG, JPG, PNG, WebP, AVIF, TIFF}

// SupportedInputs son las extensiones de archivo que sharpy sabe leer como entrada.
var SupportedInputs = map[string]bool{
	"jpg": true, "jpeg": true, "png": true, "webp": true, "avif": true, "tif": true, "tiff": true,
}

// Contains indica si v está en la lista de formatos dada.
func Contains(list []OutputFormat, v OutputFormat) bool {
	for _, f := range list {
		if f == v {
			return true
		}
	}
	return false
}

// Names arma "jpeg, jpg, png, ..." para mensajes de ayuda/error.
func Names(list []OutputFormat) string {
	out := make([]string, len(list))
	for i, f := range list {
		out[i] = string(f)
	}
	return strings.Join(out, ", ")
}

// Normalize colapsa alias (jpg -> jpeg) a la forma canónica.
func Normalize(f OutputFormat) OutputFormat {
	if f == JPG {
		return JPEG
	}
	return f
}

// Ext devuelve la extensión de archivo (con punto) para un formato de salida.
func Ext(f OutputFormat) string {
	switch f {
	case JPEG:
		return ".jpg"
	case PNG:
		return ".png"
	case WebP:
		return ".webp"
	case AVIF:
		return ".avif"
	case TIFF:
		return ".tiff"
	default:
		return ".tiff"
	}
}

// ExcludeExt devuelve la extensión de archivo de ENTRADA a excluir del listado,
// para no volver a convertir imágenes que ya están en el formato destino.
func ExcludeExt(f OutputFormat) string {
	if f == JPEG {
		return "jpg"
	}
	return string(f)
}

// BimgOptions arma las bimg.Options correspondientes a un formato de salida,
// aplicando la calidad indicada (o un default sensato si no fue especificada).
// ok=false si el formato no está manejado.
func BimgOptions(f OutputFormat, quality int, hasQuality bool) (bimg.Options, bool) {
	opts := bimg.Options{}

	qualityOr := func(def int) int {
		if hasQuality {
			return quality
		}
		return def
	}

	switch f {
	case JPEG:
		opts.Type = bimg.JPEG
		opts.Quality = qualityOr(80)
	case WebP:
		opts.Type = bimg.WEBP
		opts.Quality = qualityOr(80)
	case AVIF:
		opts.Type = bimg.AVIF
		opts.Quality = qualityOr(50)
	case PNG:
		opts.Type = bimg.PNG
	case TIFF:
		opts.Type = bimg.TIFF
		opts.Quality = qualityOr(80)
	default:
		return bimg.Options{}, false
	}

	return opts, true
}
