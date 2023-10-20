package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ga28299/rssagg/internal/database"
)

func startScrapping(
	db *database.Queries,
	concurrency int,
	timeBetweenRequest time.Duration,
) {
	log.Printf("Scrapping on %v goroutines eveery %s duration", concurrency, timeBetweenRequest)
	ticker := time.NewTicker(timeBetweenRequest)
	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(
			context.Background(), int32(concurrency),
		)
		if err != nil {
			log.Println("error getting feeds to fetch", err)
			continue
		}

		wg := &sync.WaitGroup{}
		for _, feed := range feeds {
			wg.Add(1)

			go scrapeFeed(db, wg, feed)
		}
		wg.Wait()
	}

}

func scrapeFeed(db *database.Queries, wg *sync.WaitGroup, feed database.Feed) {
	defer wg.Done()

	_, err := db.MarkFeedAsFetched(context.Background(), feed.ID)
	if err != nil {
		log.Println("error marking feed as fetched", err)
		return
	}

	rssFeed, err := urlToFeed(feed.Url)
	if err != nil {
		log.Println("error fetching feed", err)
	}
	for _, item := range rssFeed.Channel.Item {
		log.Println("found post", item.Title, "on feed: ", feed.Name)
	}
	log.Println("Feeed %s collected, %v posts found", feed.Name, len(rssFeed.Channel.Item))

}
