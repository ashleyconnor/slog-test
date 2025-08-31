package main

import (
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Debug(
		"executing database query",
		slog.String("query", "SELECT * FROM users"),
	)
	logger.Info("image upload successful", slog.String("image_id", "39ud88"))
	logger.Warn(
		"storage is 90% full",
		slog.String("available_space", "900.1 MB"),
	)
	logger.Error(
		"An error occurred while processing the request",
		slog.String("url", "https://example.com"),
	)
}
