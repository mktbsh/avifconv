package main

import (
	"avifconv/logger"
	"os"
)

func main() {
	console := logger.NewConsole(logger.DefaultOptions())

	cfg, err := ParseConfig(console)
	if err != nil {
		os.Stderr.WriteString("Configuration error: " + err.Error() + "\n")
		os.Exit(1)
	}

	processor := NewProcessor(cfg, console)

	if err := processor.ProcessPath(cfg.InputPath); err != nil {
		console.Error("Processing error: %v", err)
		os.Exit(1)
	}

	console.Success("All processing completed successfully")
}
