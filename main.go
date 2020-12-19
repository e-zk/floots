package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/mattn/go-mastodon"
)

var (
	excludeStr []string
	dryRun     bool
)

func isExcluded(toot *mastodon.Status) bool {
	for _, exStr := range excludeStr {
		if strings.Contains(toot.Content, exStr) || string(toot.ID) == exStr {
			return true
		}
	}

	return false
}

func removeToots(c *mastodon.Client, ageLimit time.Duration) error {
	var (
		err error
	)

	a, err := c.GetAccountCurrentUser(context.Background())
	if err != nil {
		return err
	}

	toots, err := c.GetAccountStatuses(context.Background(), a.ID, nil)
	if err != nil {
		return err
	}

	for _, toot := range toots {
		if time.Since(toot.CreatedAt) > ageLimit && !isExcluded(toot) {
			fmt.Printf("deleting toot:\n\t%s\n\t- %s\n", toot.Content, toot.CreatedAt)
			if !dryRun {
				err = c.DeleteStatus(context.Background(), toot.ID)
			}
		}
	}

	return nil
}

func main() {
	var (
		err    error
		maxAge time.Duration
	)

	if envPath, ok := os.LookupEnv("FLOOTS_ENV"); ok {
		err = godotenv.Load(envPath)
	} else {
		err = godotenv.Load()
	}

	// env file is not mandatory since envirnoment vars could be defined normally
	if err != nil {
		log.Println("error loading .env file; continuing anyway")
	}

	dryRun, err = strconv.ParseBool(os.Getenv("FLOOTS_DRY_RUN"))
	if err != nil {
		log.Fatal("invalid value of FLOOTS_DRY_RUN specified")
	}

	excludeStr = strings.Split(os.Getenv("FLOOTS_EXCLUDE"), ":")
	if len(excludeStr) == 0 {
		log.Println("no exclude tags defined")
	}

	maxAge, err = time.ParseDuration(os.Getenv("FLOOTS_MAX_AGE"))
	if err != nil {
		log.Fatal("invalid value of FLOOTS_MAX_AGE specified")
	}

	c := mastodon.NewClient(&mastodon.Config{
		Server:       os.Getenv("FLOOTS_INSTANCE"),
		ClientID:     os.Getenv("FLOOTS_CLIENT_ID"),
		ClientSecret: os.Getenv("FLOOTS_CLIENT_SECRET"),
		AccessToken:  os.Getenv("FLOOTS_ACCESS_TOKEN"),
	})

	err = removeToots(c, maxAge)
	if err != nil {
		log.Fatalf("error when attempting to remove toots:\n%v\n", err)
	}
}
