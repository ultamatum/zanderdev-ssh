package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ultamatum/zanderdev-ssh/internal/ui/views"

	tea "charm.land/bubbletea/v2"
	"charm.land/ssh"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	"charm.land/wish/v2/bubbletea"
	"charm.land/wish/v2/logging"
	"charm.land/wish/v2/ratelimiter"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/time/rate"
)

func main() {
	addr := os.Getenv("SSH_ADDR")
	if addr == "" {
		addr = "localhost:23234"
	}

	s, err := wish.NewServer(
		wish.WithAddress(addr),
		wish.WithHostKeyPath(".ssh/host_ed25519"),
		wish.WithPublicKeyAuth(func(ssh.Context, ssh.PublicKey) bool {
			return true
		}),
		wish.WithKeyboardInteractiveAuth(func(ssh.Context, gossh.KeyboardInteractiveChallenge) bool {
			return true
		}),
		wish.WithIdleTimeout(10*time.Minute),
		wish.WithMaxTimeout(1*time.Hour),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			ratelimiter.Middleware(ratelimiter.NewRateLimiter(rate.Every(time.Second), 5, 1000)),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Listening on: ", addr)
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	visitor := "anonymous"
	if key := s.PublicKey(); key != nil {
		visitor = gossh.FingerprintSHA256(key)
	}

	pty, _, _ := s.Pty()
	return views.NewBoard(visitor, pty.Window.Width, pty.Window.Height), nil
}
