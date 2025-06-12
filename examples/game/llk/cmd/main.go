package main

import (
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

	// err = bot.EnterGame(context.Background())
	// require.NoError(t, err, "Failed to enter game")

	for {
		err = bot.Play()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to play game")
		}
		time.Sleep(1 * time.Second)
	}
}
