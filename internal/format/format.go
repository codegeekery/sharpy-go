// Package format define los formatos de salida soportados por sharpy
// y las conversiones necesarias hacia govips (tipo de imagen, extensión, etc).
package format

import (
	"strings"
	"github.com/davidbyttow/govips/v2/vips"
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

var Supported = []OutputFormat{JPEG, JPG, PNG, WebP, AVIF, TIFF}

var SupportedInputs = map[string]bool{
	"jpg": true, "jpeg": true, "png": true, "webp": true, "avif": true, "tif": true, "tiff": true,
}

func Contains(list []OutputFormat, v OutputFormat) bool {
	for _, f := range list {
		if f == v {
			return true
		}
	}
	return false
}

func Names(list []OutputFormat) string {
	out := make([]string, len(list))
	for i, f := range list {
		out[i] = string(f)
	}
	return strings.Join(out, ", ")
}

func Normalize(f OutputFormat) OutputFormat {
	if f == JPG {
		return JPEG
	}
	return f
}

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

func ExcludeExt(f OutputFormat) string {
	if f == JPEG {
		return "jpg"
	}
	return string(f)
}

// IsSupported indica si f tiene una ruta de exportación implementada.
// Reemplaza el "ok" que antes devolvía BimgOptions.
func IsSupported(f OutputFormat) bool {
	switch f {
	case JPEG, WebP, AVIF, PNG, TIFF:
		return true
	default:
		return false
	}
}

// Export codifica img al formato de salida f, aplicando la calidad indicada
// (o un default sensato si no fue especificada). Devuelve ok=false si el
// formato no está manejado (mismo contrato que antes tenía BimgOptions).
func Export(img *vips.ImageRef, f OutputFormat, quality int, hasQuality bool) (buf []byte, ok bool, err error) {
	qualityOr := func(def int) int {
		if hasQuality {
			return quality
		}
		return def
	}

	switch f {
	case JPEG:
		buf, _, err = img.ExportJpeg(&vips.JpegExportParams{
			Quality: qualityOr(80),
			StripMetadata: true,
		})
	case WebP:
		buf, _, err = img.ExportWebp(&vips.WebpExportParams{
			Quality: qualityOr(80),
			StripMetadata: true,
		})
	case AVIF:
		buf, _, err = img.ExportAvif(&vips.AvifExportParams{
			Quality: qualityOr(50),
                        Effort:  0, // 0 = más rápido, 9 = más lento pero mejor compresión
                        Bitdepth: 8,
			Lossless: false,
                        StripMetadata: true,
		})
	case PNG:
		buf, _, err = img.ExportPng(vips.NewPngExportParams())
	case TIFF:
		buf, _, err = img.ExportTiff(&vips.TiffExportParams{
			Quality: qualityOr(80),
			StripMetadata: true,
		})
	default:
		return nil, false, nil
	}

	if err != nil {
		return nil, true, err
	}
	return buf, true, nil
}
