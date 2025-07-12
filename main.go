package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/killuox/gator-blog-aggregator/internal/config"
	"github.com/killuox/gator-blog-aggregator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	config *config.Config
	db     *database.Queries
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

	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		fmt.Print("Error connecting to the database")
		os.Exit(1)
	}

	dbQueries := database.New(db)

	state := &state{
		config: &cfg,
		db:     dbQueries,
	}

	commands := commands{
		handlers: make(map[string]commandHandler),
	}

	// commands
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)

	if len(os.Args) < 2 {
		fmt.Print("Not enough arguments provided.\n")
		os.Exit(1)
	}

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

// Handlers
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("A username is required\n")
	}
	name := cmd.args[0]

	user, err := s.db.GetUserByName(context.Background(), name)
	if err != nil {
		fmt.Printf("You can't login to an account that doesn't exist!")
		os.Exit(1)
	}

	err = s.config.SetUser(name)
	if err != nil {
		return fmt.Errorf("Error while registering your username: %s\n", err)
	}
	fmt.Printf("Hello %s, you're now logged in", user.Name)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("A name is required\n")
	}
	name := cmd.args[0]

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	})
	if err != nil {
		fmt.Printf("Error occurred while creating user: %s\n", err)
		os.Exit(1)
	}

	err = s.config.SetUser(name)
	if err != nil {
		return fmt.Errorf("Error while registering your username: %s\n", err)
	}

	fmt.Printf("The user '%v' was created successfully\n", user)

	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		fmt.Printf("Could not reset users: %s", err)
		os.Exit(1)
	}
	return nil
}

// Utilities
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
