package logger

import (
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SetupLogger(debug *bool) {
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	log.Logger = log.With().Caller().Logger()
	log.Debug().Msg("Debug logging enabled.")
}
