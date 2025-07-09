package main

import (
	"fmt"
	"os"

	"github.com/killuox/gator-blog-aggregator/internal/config"
)

type state struct {
	config *config.Config
}

type command struct {
	name    string
	args    []string
	handler commandHandler
}

type commandHandler func(s *state, cmd command) error

type commands struct {
	handlers map[string]commandHandler
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("An error occured while reading config: %s\n", err)
		os.Exit(1)
	}

	state := &state{
		config: &cfg,
	}

	commands := commands{
		handlers: make(map[string]commandHandler),
	}

	commands.register("login", handlerLogin)
	if len(os.Args) < 2 {
		fmt.Print("Not enough arguments provided.\n")
		os.Exit(1)
	}

	// pName := os.Args[0]
	cName := os.Args[1]
	args := os.Args[2:]

	handler, ok := commands.handlers[cName]
	if !ok {
		fmt.Printf("Command name '%s' not found\n", cName)
		os.Exit(1)
	}

	cmd := command{
		name:    cName,
		args:    args,
		handler: handler,
	}

	err = commands.run(state, cmd)
	if err != nil {
		fmt.Printf("Error while running the command: %s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("A username is required\n")
	}

	err := s.config.SetUser(cmd.args[0])
	if err != nil {
		return fmt.Errorf("Error while registering your username: %s\n", err)
	}
	fmt.Print("The username has been set.\n")
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("command %s not found\n", cmd.name)
	}
	return handler(s, cmd)
}

func (c *commands) register(name string, f commandHandler) {
	c.handlers[name] = f
}
