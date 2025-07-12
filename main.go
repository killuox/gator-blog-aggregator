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
	"github.com/killuox/gator-blog-aggregator/internal/rss"
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
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", handlerAddFeed)
	commands.register("feeds", handlerFeeds)

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
		fmt.Printf("Could not reset users: %s\n", err)
		os.Exit(1)
	}
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	currUser := s.config.CurrentUserName
	if err != nil {
		fmt.Printf("Could not get users: %s\n", err)
		os.Exit(1)
	}

	for _, u := range users {
		if u.Name == currUser {
			fmt.Printf("* %s (current)\n", u.Name)
		} else {
			fmt.Printf("* %s\n", u.Name)
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	url := "https://www.wagslane.dev/index.xml"

	res, err := rss.FetchFeed(context.Background(), url)
	if err != nil {
		fmt.Printf("Error occured while fetching url: %s\n", err)
		os.Exit(1)
	}

	fmt.Print(res)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	currUser, err := s.db.GetUserByName(context.Background(), s.config.CurrentUserName)
	if err != nil {
		fmt.Printf("You do not have the permission to add feed, please make sure to login or register before")
		os.Exit(1)
	}

	if len(cmd.args) < 1 {
		return fmt.Errorf("A name is required\n")
	}

	if len(cmd.args) < 2 {
		return fmt.Errorf("An url is required\n")
	}

	name := cmd.args[0]
	url := cmd.args[1]

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    currUser.ID,
	})

	fmt.Printf("%s\n", feed)

	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		fmt.Printf("Could not get feeds: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s\n", feeds)

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
