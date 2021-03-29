package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/lindebergue/desafio-go-stone/database"
	"github.com/lindebergue/desafio-go-stone/router"
)

func main() {
	jwtSecret := os.Getenv("APP_JWT_SECRET")
	databaseURL := os.Getenv("APP_DATABASE_URL")

	db, err := database.NewPostgresDB(databaseURL)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	srv := &http.Server{
		Addr: ":9999",
		Handler: router.New(router.Options{
			DB:        db,
			JWTSecret: []byte(jwtSecret),
		}),
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error starting server: %v", err)
		}
	}()

	log.Printf("server listening for connections on %s", srv.Addr)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	log.Println("interrupt signal received; shutting down the server...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("error shutting down the server gracefully: %v", err)
	}
}
