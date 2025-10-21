package main

import (
	"errors"
	"fmt"
	"context"
	"time"
	"database/sql"
	"html"

	"github.com/google/uuid"
	"github.com/Kaniniz/blog_gator/internal/database"
	"github.com/Kaniniz/blog_gator/internal/rssStuff"
)

type command struct {
	name string
	arguments []string
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return errors.New("Must enter a username to login")
	}

	name := sql.NullString{
		String: cmd.arguments[0],
		Valid: true,
	}

	user, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		fmt.Println("User isn't registered")
		return err
	}

	err = s.config.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Printf("User %v has logged in\n", user.Name.String)
	return nil

}

func handlerRegister(s *state, cmd command) error {
	current_time := time.Now()
	name := sql.NullString{
		String: cmd.arguments[0],
		Valid: true,
	}

	user, err := s.db.CreateUser(context.Background(),
		database.CreateUserParams{
			ID: uuid.New(),
			CreatedAt: current_time,
			UpdatedAt: current_time,
			Name: name,
		},
	)
	if err != nil {
		return err
	}
	
	err = s.config.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Printf("User %v has been set\n", user.Name.String)
	return nil
}

func handlerResetUsers(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("User table has been reset!")
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Name.String == s.config.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name.String)
		} else {
			fmt.Println("*", user.Name.String)
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	feedUrl := "https://www.wagslane.dev/index.xml"
	rss_feed, err := rssStuff.FetchFeed(context.Background(), feedUrl)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n%s\n%s\n\n", 
	html.UnescapeString(rss_feed.Channel.Title),
	rss_feed.Channel.Link,
	html.UnescapeString(rss_feed.Channel.Description))

	for _, item := range rss_feed.Channel.Item {
		fmt.Printf("%s\n%s\n%s\n\n", 
		html.UnescapeString(item.Title),
		item.Link,
		html.UnescapeString(item.Description))
	}

	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.arguments) < 2 {
		return errors.New("Must have specify blog name and blog url")
	}

	current_time := time.Now()
	name := sql.NullString{
		String: s.config.CurrentUserName,
		Valid: true,
	}
	
	user, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return err
	}

	//why the hells does the sqlc generate require sql.Nullstrings? This is getting annoying
	feed_name := sql.NullString{
		String: cmd.arguments[0],
		Valid: true,
	}
	feed_url := sql.NullString{
		String: cmd.arguments[1],
		Valid: true,
	}
	null_uuid := uuid.NullUUID {
		UUID: user.ID,
		Valid: true,
	}
	feed, err := s.db.AddFeed(context.Background(), 
		database.AddFeedParams{
				ID: uuid.New(),
				CreatedAt: current_time,
				UpdatedAt: current_time,
				Name: feed_name,
				Url: feed_url,
				UserID: null_uuid,
			},
	)
	if err != nil {
		return err
	}
	fmt.Printf("Feed added: %s\nUrl: %s\n",
				feed.Name.String,
				feed.Url.String,
			)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		user, err := s.db.GetUserByID(context.Background(), feed.UserID.UUID)
		if err != nil {
		return err
		}
		fmt.Printf("Feed: %s\nUrl: %s\nCreated by: %s\n",
					feed.Name.String,
					feed.Url.String,
					user.Name.String)
	}
	return nil
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	cmd_name := cmd.name
	_, ok := c.handlers[cmd_name]
	if !ok {
		err := fmt.Errorf("Unknown command: %s", cmd.name)
		return err
	}
	err := c.handlers[cmd_name](s, cmd)
	if err != nil {
		return err
	}

	return nil
}

func (c *commands) register(name string, f func(*state, command) error) error {
	c.handlers[name] = f
	return nil
}