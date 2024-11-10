package utils

import (
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	COLOR_RESET  = "\033[0m"
	COLOR_BLACK  = "\033[0;30m"
	COLOR_RED    = "\033[0;31m"
	COLOR_GREEN  = "\033[0;32m"
	COLOR_YELLOW = "\033[0;33m"
	COLOR_BLUE   = "\033[0;34m"
	COLOR_PURPLE = "\033[0;35m"
	COLOR_CYAN   = "\033[0;36m"
	COLOR_WHITE  = "\033[0;37m"

	COLOR_BOLD_BLACK  = "\033[1;30m"
	COLOR_BOLD_RED    = "\033[1;31m"
	COLOR_BOLD_GREEN  = "\033[1;32m"
	COLOR_BOLD_YELLOW = "\033[1;33m"
	COLOR_BOLD_BLUE   = "\033[1;34m"
	COLOR_BOLD_PURPLE = "\033[1;35m"
	COLOR_BOLD_CYAN   = "\033[1;36m"
	COLOR_BOLD_WHITE  = "\033[1;37m"

	COLOR_BACKGROUND_BLACK  = "\033[40m"
	COLOR_BACKGROUND_RED    = "\033[41m"
	COLOR_BACKGROUND_GREEN  = "\033[42m"
	COLOR_BACKGROUND_YELLOW = "\033[43m"
	COLOR_BACKGROUND_BLUE   = "\033[44m"
	COLOR_BACKGROUND_PURPLE = "\033[45m"
	COLOR_BACKGROUND_CYAN   = "\033[46m"
	COLOR_BACKGROUND_WHITE  = "\033[47m"
)

var colors = []string{
	COLOR_RESET,
	COLOR_BLACK,
	COLOR_RED,
	COLOR_GREEN,
	COLOR_YELLOW,
	COLOR_BLUE,
	COLOR_PURPLE,
	COLOR_CYAN,
	COLOR_WHITE,

	COLOR_BOLD_BLACK,
	COLOR_BOLD_RED,
	COLOR_BOLD_GREEN,
	COLOR_BOLD_YELLOW,
	COLOR_BOLD_BLUE,
	COLOR_BOLD_PURPLE,
	COLOR_BOLD_CYAN,
	COLOR_BOLD_WHITE,

	COLOR_BACKGROUND_BLACK,
	COLOR_BACKGROUND_RED,
	COLOR_BACKGROUND_GREEN,
	COLOR_BACKGROUND_YELLOW,
	COLOR_BACKGROUND_BLUE,
	COLOR_BACKGROUND_PURPLE,
	COLOR_BACKGROUND_CYAN,
	COLOR_BACKGROUND_WHITE,
}

var Log zerolog.Logger

var once sync.Once

func init() {
	once.Do(func() {
		// Set up external file for logging as we use OsStout
		logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			panic(err)
		}
		output := zerolog.ConsoleWriter{Out: logFile, TimeFormat: time.RFC3339}
		Log = log.Output(output)

		// Define if log should be set
		debugMode := os.Getenv("DEBUG")

		if debugMode == "true" {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			Log.Info().Msg("Debug mode enabled")
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
			Log.Info().Msg("Debug mode disabled, logging at Info level")
		}
	})
}
