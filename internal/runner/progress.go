package runner

import (
	"fmt"
	"sync"
)

// progressBar prints an "N/Total" counter that updates in place
// (using carriage return \r) as workers finish files.
// Messages that need to remain as permanent history (failures,
// deleted originals) are emitted via logLine, which first clears
// the bar's line, prints the message on a new line, and redraws
// the bar below it.
//
// It is safe to call concurrently from multiple goroutines.
type progressBar struct {
	mu    sync.Mutex
	total int
	done  int
}

func newProgressBar(total int) *progressBar {
	return &progressBar{total: total}
}

func (b *progressBar) increment(done int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.done = int(done)
	b.render()
}

func (b *progressBar) logLine(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	fmt.Print("\r\033[K")
	fmt.Println(msg)
	b.render()
}

func (b *progressBar) finish() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.render()
	fmt.Println()
}

func (b *progressBar) render() {
	if b.total == 0 {
		return
	}
	const width = 30
	filled := width * b.done / b.total
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	pct := 100 * b.done / b.total
	fmt.Printf("\r[%s] %3d%% (%d/%d)", bar, pct, b.done, b.total)
}
