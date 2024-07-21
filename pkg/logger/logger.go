package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"golang.org/x/term"
)

const (
	defaultLogLevel = zerolog.DebugLevel
	logLevelEnvName = "LOG_LEVEL"
)

var (
	defaultLogger *zerolog.Logger

	initialized bool
	once        sync.Once
)

func Init(serviceName string, serviceVersion string) {
	if initialized {
		defaultLogger.Warn().Msgf("logger has been already initialized")
	}

	once.Do(func() {
		defaultLogger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Dict("serviceContext", zerolog.Dict().
				Str("service", serviceName).
				Str("version", serviceVersion))
		})

		initialized = true
	})
}

func Default() *zerolog.Logger {
	// create a child logger so that the callers cannot override the default logger context
	l := defaultLogger.With().Logger()

	return &l
}

func init() {
	setDefaultLogger()
	setLogLevel()
}

func setLogLevel() {
	desiredLogLevel := os.Getenv(logLevelEnvName)
	fmt.Println("loglevel", desiredLogLevel)
	if desiredLogLevel == "" {
		defaultLogger.Info().
			Msgf("setting log level to default level: %v", defaultLogLevel)
		zerolog.SetGlobalLevel(defaultLogLevel)
		return
	}

	level, err := zerolog.ParseLevel(desiredLogLevel)
	if err != nil {
		defaultLogger.Info().Msgf(
			"invalid log level %v, setting to default level: %v", desiredLogLevel, defaultLogLevel)
		zerolog.SetGlobalLevel(defaultLogLevel)
	} else {
		defaultLogger.Info().Msgf("setting log level to %v", level)
		zerolog.SetGlobalLevel(level)
	}
}

var severityLevels = map[zerolog.Level]string{
	zerolog.TraceLevel: "DEBUG",
	zerolog.DebugLevel: "DEBUG",
	zerolog.InfoLevel:  "INFO",
	zerolog.WarnLevel:  "WARNING",
	zerolog.ErrorLevel: "ERROR",
	zerolog.FatalLevel: "CRITICAL",
	zerolog.PanicLevel: "ALERT",
}

func setDefaultLogger() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.TimestampFieldName = "eventTime"
	zerolog.MessageFieldName = "msg"
	zerolog.ErrorFieldName = "message"
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	l := zerolog.New(os.Stderr).With().Timestamp().Logger()
	if isTerminal() {
		l = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	}

	defaultLogger = &l
}

func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
