package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ga28299/rssagg/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	feed, err := urlToFeed("https://wagslane.dev/index.xml")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(feed)

	godotenv.Load(".env")

	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("No port set. Check env file")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("No Database URL found. Check env file")
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Can't connect to DB!", err)
	}

	if err != nil {
		log.Fatal("Can't create db connection!", err)
	}
	db := database.New(conn)
	apiCFG := apiConfig{
		DB: database.New(conn),
	}

	go startScrapping(db, 10, time.Minute)

	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	v1Router := chi.NewRouter()
	v1Router.Get("/healthz", handlerReadiness)
	v1Router.Get("/err", handlerErr)
	v1Router.Post("/users", apiCFG.handlerCreateUser)
	v1Router.Get("/users", apiCFG.middlewareAuth(apiCFG.handlerGetUser))
	v1Router.Post("/feeds", apiCFG.middlewareAuth(apiCFG.handlerCreateFeed))
	v1Router.Get("/feeds", apiCFG.handlerGetFeeds)
	v1Router.Post("/feed_follows", apiCFG.middlewareAuth(apiCFG.handlerCreateFeedFollow))
	v1Router.Get("/feed_follows", apiCFG.middlewareAuth(apiCFG.handlerGetFeedFollows))
	v1Router.Delete("/feed_follows/{feedFollowID}", apiCFG.middlewareAuth(apiCFG.handlerDeleteFeedFollow))

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}

	log.Printf("Server starting on Port %s", portString)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
