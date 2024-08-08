package main

import (
	"errors"
	"log/slog"
	"os"

	"github.com/xsbin7/console-slog"
)

func main() {
	user := struct {
		Name string
		Age  int
	}{Name: `a`, Age: 10}

	logger := slog.New(
		console.NewHandler(os.Stderr, &console.HandlerOptions{Level: slog.LevelDebug, AddSource: true}),
	)
	slog.SetDefault(logger)
	slog.Info("Hello world!", "foo", "bar")
	slog.Debug("Debug message")
	slog.Warn("Warning message")
	slog.Error("Error message", "err", errors.New("the error"))

	slog.Info(`user info`, `user`, user)

	logger = logger.With("foo", "bar").
		WithGroup("the-group").
		With("bar", "baz")

	logger.Info("group info", "attr", "value")
}
