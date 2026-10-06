package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/armbdevelop/testgomg/config"
	"github.com/armbdevelop/testgomg/internal/initx"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	app, err := initx.NewApp(cfg)
	if err != nil {
		log.Fatalf("init: %v", err)
	}

	go func() {
		if err = app.Run(); err != nil {
			log.Fatalf("run: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	if err = app.Shutdown(); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}
