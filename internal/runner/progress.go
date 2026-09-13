package runner

import (
	"fmt"
	"sync"
)

// progressBar imprime un contador "N/Total" que se actualiza en el lugar
// (usando retorno de carro \r) a medida que los workers terminan archivos.
// Los mensajes que necesitan quedar como historial permanente (fallos,
// borrados de originales) se emiten con logLine, que primero limpia la
// línea de la barra, imprime el mensaje en una línea nueva, y vuelve a
// dibujar la barra debajo.
//
// Es seguro llamarlo concurrentemente desde múltiples goroutines.
type progressBar struct {
	mu    sync.Mutex
	total int
	done  int
}

func newProgressBar(total int) *progressBar {
	return &progressBar{total: total}
}

// increment redibuja la barra con el nuevo conteo de "done" (ya calculado
// atómicamente por el caller, para mantener un único contador de verdad).
func (b *progressBar) increment(done int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.done = int(done)
	b.render()
}

// logLine imprime un mensaje permanente sin romper la barra: borra la
// línea actual, escribe el mensaje, y vuelve a pintar la barra.
func (b *progressBar) logLine(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	fmt.Print("\r\033[K") // \r + limpiar línea hasta el final
	fmt.Println(msg)
	b.render()
}

// finish deja la barra al 100% y agrega el salto de línea final, para que
// lo que se imprima después (el resumen) empiece en línea limpia.
func (b *progressBar) finish() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.render()
	fmt.Println()
}

// render dibuja la barra actual. Debe llamarse con b.mu ya tomado.
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
