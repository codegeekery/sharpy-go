// Package runner orchestrates the concurrent conversion of a batch of images:
// distributes work among workers, deletes originals when applicable, and prints
// progress and the final summary.
package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"sharpy/internal/cliopts"
	"sharpy/internal/converter"
	"sharpy/internal/format"
)

const (
	removeRetries   = 3
	removeRetryWait = 1 * time.Second
)

// Summary is the aggregated result of running the whole batch.
type Summary struct {
	OK   int
	Fail int
}

// Run converts files to the given format using a worker pool,
// respecting the options (force, remove-original, rename, quality, dry-run).
// It shows a live progress bar on stdout; details of each failed file
// are accumulated and listed at the end so as not to interrupt
// the bar while it's running.
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
	var wg sync.WaitGroup

	bar := newProgressBar(total)

	// One worker per 2 CPU cores: image encoding is memory-heavy (each
	// worker holds a full decoded buffer plus encoder working buffers), so
	// scaling workers 1:1 with cores caused RAM usage to balloon without a
	// proportional speed gain. This ratio keeps CPU well fed without
	// over-committing memory.
	
        workerCount := max(1, min(total, runtime.NumCPU()/4))

	worker := func() {
		defer wg.Done()
		for {
			i := atomic.AddInt64(&idx, 1)
			if int(i) >= total {
				return
			}
			
			current := files[i]
			res := converter.ConvertOne(current, outFmt, params)
			
			results[i] = res

			if !res.OK {
				bar.logLine(formatFailureLine(current, res, opts.Dir))
			}

			if res.OK && opts.RemoveOriginal {
				if opts.DryRun {
					relSrc, _ := filepath.Rel(opts.Dir, res.Src)
					bar.logLine(fmt.Sprintf("🧹 (dry-run) would delete: %s", relSrc))
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
	return fmt.Sprintf("[FAILED] %s -> %s | %s", relSrc, relDest, res.Reason)
}

// removeOriginal attempts to delete the original file with retries, since
// on some filesystems/network shares the file may remain briefly locked
// after reading. The first attempt happens immediately; only retries after
// a failure wait, to avoid a fixed delay on the common case where deletion
// succeeds right away. Messages go through bar.logLine so as not to
// overwrite the progress bar line.
func removeOriginal(src, baseDir string, bar *progressBar) {
	var lastErr error
	for tries := 0; tries < removeRetries; tries++ {
		if tries > 0 {
			time.Sleep(removeRetryWait)
		}
		if err := os.Remove(src); err != nil {
			lastErr = err
			continue
		}
		relSrc, _ := filepath.Rel(baseDir, src)
		bar.logLine(fmt.Sprintf("🧹 Deleted original: %s", relSrc))
		return
	}
	bar.logLine(fmt.Sprintf("⚠️ Could not delete original: %s (%v)", src, lastErr))
}
