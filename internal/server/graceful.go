package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func RunWithGracefulShutdown(srv *Server) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on port %s", srv.config.Server.Port)
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
			os.Exit(1)
		}
	}()

	<-quit
	log.Println("Received shutdown signal")

	// TODO: Make this configurable
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	log.Println("Server exited cleanly")
	return nil
}
