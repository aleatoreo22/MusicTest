package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type Player struct {
	cmd     *exec.Cmd
	mu      sync.Mutex
	done    chan struct{}
	backend string
}

var (
	playerMu  sync.Mutex
	active    *Player
	available []string
	detected  bool
)

func detectPlayers() []string {
	if detected {
		return available
	}
	detected = true
	candidates := []string{
		"mpv",
		"ffplay",
		"vlc",
		"cvlc",
	}
	for _, bin := range candidates {
		if p, err := exec.LookPath(bin); err == nil {
			available = append(available, p)
		}
	}
	return available
}

func AvailablePlayer() string {
	ps := detectPlayers()
	if len(ps) == 0 {
		return ""
	}
	return ps[0]
}

var ErrNoPlayer = errors.New("nenhum player de áudio disponível (instale mpv, ffplay ou vlc)")

// Play inicia a reprodução de uma URL. startSec >= 0 indica a posição (em
// segundos) onde o player deve começar — usado pra retomar de onde parou.
// Quando startSec > 0, passa --start=<sec> (ou equivalente) pro backend.
func Play(url string, startSec int) (*Player, error) {
	if url == "" {
		return nil, errors.New("URL vazia")
	}

	playerMu.Lock()
	if active != nil {
		_ = active.stopLocked()
		active = nil
	}
	playerMu.Unlock()

	backends := detectPlayers()
	if len(backends) == 0 {
		return nil, ErrNoPlayer
	}

	bin := backends[0]
	args := argsFor(bin, url, startSec)
	cmd := exec.Command(bin, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("falha ao iniciar %s: %w", bin, err)
	}

	p := &Player{cmd: cmd, backend: bin, done: make(chan struct{})}

	playerMu.Lock()
	active = p
	playerMu.Unlock()

	go p.wait()

	return p, nil
}

func argsFor(bin, url string, startSec int) []string {
	base := filepath.Base(bin)
	switch base {
	case "mpv":
		args := []string{"--no-terminal", "--quiet", "--no-video", "--really-quiet"}
		if startSec > 0 {
			args = append(args, fmt.Sprintf("--start=%d", startSec))
		}
		args = append(args, url)
		return args
	case "ffplay":
		args := []string{"-nodisp", "-loglevel", "quiet", "-hide_banner", "-autoexit"}
		if startSec > 0 {
			args = append(args, "-ss", fmt.Sprintf("%d", startSec))
		}
		args = append(args, url)
		return args
	case "vlc", "cvlc":
		args := []string{"--intf", "dummy", "--play-and-exit", "--no-video"}
		if startSec > 0 {
			args = append(args, "--start-time", fmt.Sprintf("%d", startSec))
		}
		args = append(args, url)
		return args
	}
	return []string{url}
}

func (p *Player) Backend() string {
	return p.backend
}

func (p *Player) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	playerMu.Lock()
	if active == p {
		active = nil
	}
	playerMu.Unlock()
	return p.stopLocked()
}

func (p *Player) stopLocked() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	if err := p.cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	return nil
}

func (p *Player) wait() {
	_ = p.cmd.Wait()
	close(p.done)
}

func (p *Player) Done() <-chan struct{} {
	return p.done
}

func StopActive() {
	playerMu.Lock()
	defer playerMu.Unlock()
	if active != nil {
		_ = active.stopLocked()
		active = nil
	}
}

var ErrNoYtdlp = errors.New("yt-dlp não encontrado no PATH (instale com 'pip install yt-dlp' ou pacote do sistema)")

func HasYtdlp() bool {
	_, err := exec.LookPath("yt-dlp")
	return err == nil
}

func ResolveYouTubeStream(ctx context.Context, query string) (string, error) {
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return "", ErrNoYtdlp
	}
	args := []string{
		"-f", "bestaudio/best",
		"-g",
		"--no-playlist",
		"--no-warnings",
		"--no-check-certificates",
		"--no-part",
		"ytsearch1:" + query,
	}
	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("yt-dlp falhou: %s", msg)
	}
	out := strings.TrimSpace(stdout.String())
	if out == "" {
		return "", errors.New("yt-dlp não retornou URL para: " + query)
	}
	if i := strings.IndexByte(out, '\n'); i >= 0 {
		out = out[:i]
	}
	return out, nil
}
