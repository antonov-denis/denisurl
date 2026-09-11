package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/antonov-denis/denisurl/internal/store"
	"github.com/antonov-denis/denisurl/internal/web"
)

func run() error {
	myStore, err := store.New(os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("database connection failed", "err", err)
		return err
	}
	defer myStore.Close()

	myServer := web.New(os.Getenv("BASE_URL"), myStore)

	server := http.Server{
		Addr:         ":8000",
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      myServer,
	}

	err = server.ListenAndServe()
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}
