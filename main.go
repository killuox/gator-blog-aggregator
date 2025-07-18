package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	commands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commands.register("feeds", handlerFeeds)
	commands.register("follow", middlewareLoggedIn(handlerFollow))
	commands.register("following", middlewareLoggedIn(handlerFollowing))
	commands.register("unfollow", middlewareLoggedIn(handlerUnFollow))
	commands.register("browse", middlewareLoggedIn(handlerBrowse))

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
		fmt.Printf("You can't login to an account that doesn't exist!\n")
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
	if len(cmd.args) < 1 {
		return fmt.Errorf("A time between request like (1s, 1m, 1h) is required\n")
	}
	time_between_reqs := cmd.args[0]

	timeBetweenRequests, err := time.ParseDuration(time_between_reqs)
	if err != nil {
		return fmt.Errorf("Error parsing time duration\n")
	}
	fmt.Printf("Collecting feeds every %v\n", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
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
		UserID:    user.ID,
	})
	if err != nil {
		return err
	}

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", feed)

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

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("A url is required\n")
	}

	url := cmd.args[0]
	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return err
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("%s is now following %s feed\n", feedFollow.UserName, feedFollow.FeedName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	for _, feedFollow := range feedFollows {
		fmt.Printf("- %s\n", feedFollow.FeedName)
	}
	return nil
}

func handlerUnFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("A url is required\n")
	}

	url := cmd.args[0]
	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return err
	}
	err = s.db.DeleteFeedFollowsForUser(context.Background(), database.DeleteFeedFollowsForUserParams{
		UserID: user.ID,
		Url:    feed.Url,
	})
	if err != nil {
		return err
	}
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := int32(10)
	if len(cmd.args) > 1 {
		parsedInt64, err := strconv.ParseInt(cmd.args[0], 10, 32)
		if err != nil {
			return fmt.Errorf("Could not parse limit arguments")
		}

		num := int32(parsedInt64)
		limit = num
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  limit,
	})
	if err != nil {
		return err
	}

	for _, post := range posts {
		fmt.Print(post)
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

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}

	err = s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		LastFetchedAt: time.Now(),
		ID:            feed.ID,
	})
	if err != nil {
		return err
	}

	feed, err = s.db.GetFeedByUrl(context.Background(), feed.Url)
	if err != nil {
		return err
	}

	res, err := rss.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Printf("Error occured while fetching url: %s\n", err)
		os.Exit(1)
	}

	for _, post := range res.Channel.Item {

		pubDate, err := time.Parse(time.RFC1123, post.PubDate)
		if err != nil {
			fmt.Printf("Could not parse post %s pub date, skipping. Error: %v\n", post.Title, err) // <--- IMPROVE ERROR MESSAGE
			continue
		}
		_, err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       post.Title,
			Url:         post.Link,
			Description: post.Description,
			PublishedAt: pubDate,
			FeedID:      feed.ID,
		})

		if err != nil {
			if strings.Contains(err.Error(), "posts_url_key") {
				fmt.Printf("Skipping post '%s', already exists\n", post.Title)
				continue
			}
			fmt.Printf("Could not create post %s:  %v\n", post.Title, err)
		}
	}

	return nil
}

// Middlewares
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		currUser, err := s.db.GetUserByName(context.Background(), s.config.CurrentUserName)
		if err != nil {
			return err
		}

		err = handler(s, cmd, currUser)
		if err != nil {
			return err
		}
		return nil
	}
}
