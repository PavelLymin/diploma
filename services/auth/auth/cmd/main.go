package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/PavelLymin/diploma/services/auth/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("services/auth/.env"); err != nil && !os.IsNotExist(err) {
		slog.Error("failed to load .env", "error", err)
		os.Exit(1)
	}

	if err := run(); err != nil {
		slog.Error("auth service stopped with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	application, err := app.New()
	if err != nil {
		return fmt.Errorf("initialize application: %w", err)
	}

	defer application.Stop()

	if err := application.Run(); err != nil {
		return fmt.Errorf("run application: %w", err)
	}

	return nil
}
