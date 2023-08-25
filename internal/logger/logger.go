package logger

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

type Logger struct {
	*zerolog.Logger
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func NewLogger() Logger {
	zerolog.CallerMarshalFunc = customCallerMarshal

	logger := log.With().
		Caller().
		Int("pid", os.Getpid()).
		Str("app", "gophermart").
		Logger()

	return Logger{&logger}
}

func (logger *Logger) RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		next.ServeHTTP(&lw, r)

		duration := time.Since(start)

		logger.Info().
			Str("uri", r.RequestURI).
			Str("method", r.Method).
			Dur("duration", duration).
			Int("response_status", responseData.status).
			Int("response_size", responseData.size).
			Send()
	})
}

func customCallerMarshal(pc uintptr, file string, line int) string {
	root, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err)
	}

	filePath := strings.ReplaceAll(file, root, "")

	return filePath + ":" + strconv.Itoa(line)
}
