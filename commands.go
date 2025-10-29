package main

import (
	"errors"
	"fmt"
	"context"
	"time"
	"html"
	"database/sql"
	"strconv"

	"github.com/lib/pq"
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

	name := cmd.arguments[0]

	user, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		fmt.Println("User isn't registered")
		return err
	}

	err = s.config.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Printf("User %v has logged in\n", user.Name)
	return nil

}

func handlerRegister(s *state, cmd command) error {
	current_time := time.Now()
	name := cmd.arguments[0]

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
	
	err = s.config.SetUser(name)
	if err != nil {
		return err
	}
	fmt.Printf("User %v has been set\n", user.Name)
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
		if user.Name == s.config.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Println("*", user.Name)
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return errors.New("Must enter a time interval to fetch feeds.\n1m, 3m, 5m, 10m, 1h")
	}	
	time_between_requests, err := time.ParseDuration(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Println("Collecting feeds every", cmd.arguments[0])
	ticker := time.NewTicker(time_between_requests)
	for ; ; <-ticker.C {
		fmt.Println("ScrapingFeeds!")
		err = scrapeFeeds(s)
		if err != nil {
		return err
		}
	}
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.arguments) < 2 {
		return errors.New("Must have specify blog name and blog url")
	}

	current_time := time.Now()
	user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return err
	}
	
	feed, err := s.db.AddFeed(context.Background(), 
		database.AddFeedParams{
				ID: uuid.New(),
				CreatedAt: current_time,
				UpdatedAt: current_time,
				Name: cmd.arguments[0],
				Url: cmd.arguments[1],
				UserID: user.ID,
			},
	)
	if err != nil {
		return err
	}
	fmt.Printf("Feed added: %s\nUrl: %s\n",
				feed.Name,
				feed.Url,
			)

	_, err = s.db.CreateFeedFollow(context.Background(), 
				database.CreateFeedFollowParams {
					ID: uuid.New(),
					CreatedAt: current_time,
					UpdatedAt: current_time,
					UserID:	user.ID,
					FeedID: feed.ID,
				},
	)
	if err != nil {
		return err
	}
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		user, err := s.db.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
		return err
		}
		fmt.Printf("Feed: %s\nUrl: %s\nCreated by: %s\n",
					feed.Name,
					feed.Url,
					user.Name)
	}
	return nil
}

func handlerFollow(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return errors.New("Must enter a url to the feed you want to fullow")
	}

	current_time := time.Now()
	user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
	feed, err := s.db.GetFeedByURL(context.Background(), cmd.arguments[0])

	resp, err := s.db.CreateFeedFollow(context.Background(), 
				database.CreateFeedFollowParams {
					ID: uuid.New(),
					CreatedAt: current_time,
					UpdatedAt: current_time,
					UserID:	user.ID,
					FeedID: feed.ID,
				},
	)
	if err != nil {
		return err
	}

	fmt.Printf("Feed: %s\nUser: %s\n", resp.FeedName, resp.UserName)
	return nil
}

func handlerFollowing(s *state, cmd command) error {
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		fmt.Println(feed.FeedsName)
	}
	return nil
}

func handlerUnfollowFeed(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return errors.New("Must specify feed url to unfollow")
	}
	
	user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return err
	}

	feed, err := s.db.GetFeedByURL(context.Background(), cmd.arguments[0])
	if err != nil {
		return err
	}
	
	err = s.db.UnfollowFeed(context.Background(), 
		database.UnfollowFeedParams{
			UserID: user.ID,
			FeedID: feed.ID,	
	})
	if err != nil {
		return err
	}
	fmt.Printf("Feed %s has been unfollowed\n", feed.Name)
	return nil
}

func handlerBrowse (s *state, cmd command) error {
	var post_limit int
	var err error	
	if len(cmd.arguments) < 1 {
		fmt.Println("No post amount limit spcified, defaulting to 2 posts")
		post_limit = 2
	} else {
		post_limit, err = strconv.Atoi(cmd.arguments[0])
		if err != nil {
			fmt.Println("Failed to parse spcified limit, defaulting to 2")
			post_limit = 2
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), int32(post_limit))
	if err != nil {
		return err
	}
	for _, post := range posts {
		fmt.Printf("%s\n%s\n%s\n%s\n\n",
					html.UnescapeString(post.Title),
					post.Url,
					post.PublishedAt,
					html.UnescapeString(post.Description),
				)
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

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	current_time := time.Now()
	err = s.db.MarkFeedFetched(
		context.Background(), 
		database.MarkFeedFetchedParams{
			LastFetchedAt: sql.NullTime{
				Time: current_time,
				Valid: true,
				},
			ID: feed.ID,
			},
	)
	if err != nil {
		return err
	}

	rss_feed, err := rssStuff.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}

	for _,item := range rss_feed.Channel.Item {
		published_at, err := time.Parse(time.RFC1123Z, item.PubDate)
		current_time := time.Now()
		if err != nil {
			fmt.Println("Failed to parse published time")
			return err
		}
		err = s.db.CreatePost(context.Background(),
							  database.CreatePostParams{
								ID: uuid.New(),
								CreatedAt: current_time,
								UpdatedAt: current_time,
								Title: html.UnescapeString(item.Title),
								Url: item.Link,
								Description: html.UnescapeString(item.Description),
								PublishedAt: published_at,
								FeedID: feed.ID,
							  },
		)
		if err != nil {
			var e *pq.Error
			if errors.As(err, &e) && e.Code == "23505" {
				continue
			}
			return err
		}
	}
	return nil
}