package logging

import (
    "os"
    "time"

    "github.com/rs/zerolog"
)

func New(environment string) zerolog.Logger {
    level := zerolog.InfoLevel
    if environment != "production" {
        level = zerolog.DebugLevel
    }

    logger := zerolog.New(os.Stdout).
        With().
        Timestamp().
        Str("env", environment).
        Logger().
        Level(level)

    zerolog.TimeFieldFormat = time.RFC3339Nano
    return logger
}
