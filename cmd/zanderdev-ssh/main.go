package main

import (
	"log"

	"github.com/ultamatum/zanderdev-ssh/internal/ui/views"

	tea "charm.land/bubbletea/v2"
	"charm.land/ssh"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	"charm.land/wish/v2/bubbletea"
)

func main() {
	s, err := wish.NewServer(
		wish.WithAddress("localhost:23234"),
		wish.WithHostKeyPath(".ssh/host_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Starting SSH server on localhost:23234")
	log.Fatalln(s.ListenAndServe())
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()
	return views.NewBoard(pty.Window.Width, pty.Window.Height), nil
}
