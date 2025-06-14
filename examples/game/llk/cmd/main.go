package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/examples/game/llk"
	"github.com/rs/zerolog/log"
)

func main() {
	hrp.InitLogger("INFO", false, false)

	// Create game bot with real device
	bot, err := llk.NewLLKGameBot("android", "")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create game bot")
	}
	defer bot.Close()

	err = bot.EnterGame(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to enter game")
	}
	// Handle graceful shutdown and report generation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create channel to handle OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start goroutine to handle signals
	go func() {
		<-sigChan
		log.Info().Msg("Received shutdown signal, generating report...")
		if err := bot.GenerateReport(); err != nil {
			log.Error().Err(err).Msg("Failed to generate report")
		}
		cancel()
	}()

	// Start goroutine to handle context cancellation
	go func() {
		<-ctx.Done()
		log.Info().Msg("Context cancelled, generating report...")
		if err := bot.GenerateReport(); err != nil {
			log.Error().Err(err).Msg("Failed to generate report")
		}
	}()

	for {
		err = bot.Play()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to play game")
		}
		time.Sleep(1 * time.Second)
	}
}
