// Package runner orquesta la conversión concurrente de un lote de imágenes:
// reparte trabajo entre workers, borra originales si corresponde, e imprime
// el progreso y el resumen final.
package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"sharpy/internal/cliopts"
	"sharpy/internal/converter"
	"sharpy/internal/format"
)

const (
	concurrency     = 4
	removeRetries   = 3
	removeRetryWait = 1 * time.Second
)

// Summary es el resultado agregado de correr todo el lote.
type Summary struct {
	OK   int
	Fail int
}

// Run convierte files al formato indicado usando un pool de workers,
// respetando las opciones (force, remove-original, rename, quality, dry-run).
// Muestra una barra de progreso en vivo en stdout; los detalles de cada
// archivo con fallo se acumulan y se listan al final para no interrumpir
// la barra mientras corre.
func Run(files []string, outFmt format.OutputFormat, opts cliopts.Options) Summary {
	params := converter.Params{
		Force:      opts.Force,
		Quality:    opts.Quality,
		HasQuality: opts.HasQuality,
		Rename:     opts.Rename,
		DryRun:     opts.DryRun,
	}

	var idx int64 = -1
	var done int64
	total := len(files)
	results := make([]converter.Result, total)
	var resultsMu sync.Mutex
	var wg sync.WaitGroup

	bar := newProgressBar(total)

	workerCount := concurrency
	if total < workerCount {
		workerCount = total
	}

	worker := func() {
		defer wg.Done()
		for {
			i := atomic.AddInt64(&idx, 1)
			if int(i) >= total {
				return
			}
			current := files[i]
			res := converter.ConvertOne(current, outFmt, params)

			resultsMu.Lock()
			results[i] = res
			resultsMu.Unlock()

			if !res.OK {
				bar.logLine(formatFailureLine(current, res, opts.Dir))
			}

			if res.OK && opts.RemoveOriginal {
				if opts.DryRun {
					relSrc, _ := filepath.Rel(opts.Dir, res.Src)
					bar.logLine(fmt.Sprintf("🧹 (dry-run) se borraría: %s", relSrc))
				} else {
					removeOriginal(res.Src, opts.Dir, bar)
				}
			}

			bar.increment(atomic.AddInt64(&done, 1))
		}
	}

	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go worker()
	}
	wg.Wait()

	bar.finish()

	summary := Summary{}
	for _, r := range results {
		if r.OK {
			summary.OK++
		} else {
			summary.Fail++
		}
	}
	return summary
}

func formatFailureLine(src string, res converter.Result, baseDir string) string {
	relSrc, _ := filepath.Rel(baseDir, src)
	relDest, _ := filepath.Rel(baseDir, res.Dest)
	return fmt.Sprintf("[FALLO] %s -> %s | %s", relSrc, relDest, res.Reason)
}

// removeOriginal intenta borrar el archivo original con reintentos, ya que
// en algunos filesystems/red el archivo puede quedar bloqueado brevemente
// tras la lectura. Los mensajes van a través de bar.logLine para no pisar
// la línea de la barra de progreso.
func removeOriginal(src, baseDir string, bar *progressBar) {
	var lastErr error
	for tries := 0; tries < removeRetries; tries++ {
		time.Sleep(removeRetryWait)
		if err := os.Remove(src); err != nil {
			lastErr = err
			continue
		}
		relSrc, _ := filepath.Rel(baseDir, src)
		bar.logLine(fmt.Sprintf("🧹 Borrado original: %s", relSrc))
		return
	}
	bar.logLine(fmt.Sprintf("⚠️ No se pudo borrar original: %s (%v)", src, lastErr))
}
