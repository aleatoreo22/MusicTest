package main

import (
	"fmt"
	"music-unified-backend/pkg/utils"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Sinal handler: se o processo receber SIGINT/SIGTERM/SIGHUP (Ctrl+C
	// do terminal em modo cooked, kill de fora, fechamento de terminal),
	// o Go runtime entrega o sinal. Se NÃO registrarmos um handler, o
	// processo morre na hora e o cleanup não roda. Aqui interceptamos
	// pra fazer cleanup síncrono antes de Bubble Tea encerrar.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sigCh
		m.FinalCleanup()
		// Não chamamos p.Quit() — o signal handler rodou cleanup,
		// deixamos main() terminar naturalmente depois.
	}()

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	// Safety net redundante: se main() voltar por qualquer caminho
	// (Ctrl+C como key event virou tea.Quit, Esc, etc.), garante cleanup.
	// FinalCleanup é idempotente.
	m.FinalCleanup()
	signal.Stop(sigCh)

	// Defere de novo por paranoia: se algo entrar em panic, ainda tenta.
	defer utils.StopActive()
}
