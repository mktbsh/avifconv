package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"runtime"
	"strings"

	"avifconv/logger"

	"github.com/gen2brain/avif"
)

type Config struct {
	InputPath    string
	Version      string
	Workers      int
	Quality      int
	QualityAlpha int
	Speed        int
	QueueSize    int
}

var (
	Version    = "dev"
	BuildDate  = "unknown"
	GitCommit  = "unknown"
	QueueRatio = 3
)

func ParseConfig(console *logger.Console) (*Config, error) {
	cpu := runtime.NumCPU()

	cfg := &Config{
		Version:   Version,
		QueueSize: cpu * QueueRatio,
	}

	flag.IntVar(&cfg.Workers, "workers", runtime.NumCPU(), "Number of concurrent workers")
	flag.IntVar(&cfg.Quality, "quality", 80, "Image quality (0-100, higher is better)")
	flag.IntVar(&cfg.QualityAlpha, "quality-alpha", 80, "Alpha channel quality (0-100)")
	flag.IntVar(&cfg.Speed, "speed", 6, "Encoding speed (0-10, lower is better quality but slower)")

	showVersion := flag.Bool("version", false, "Show version information")

	flag.Parse()

	if *showVersion {
		versionInfo := fmt.Sprintf(
			"Version: %s\nBuild date: %s\nGit commit: %s",
			cfg.Version, BuildDate, GitCommit,
		)
		console.Box("avifconv version information", versionInfo)
		os.Exit(0)
	}

	args := flag.Args()

	if len(args) == 0 {
		console.Info("Usage: avifconv [options] <file or directory path>")
		console.Info("Options:")

		old := flag.CommandLine.Output()

		r, w, _ := os.Pipe()
		flag.CommandLine.SetOutput(w)

		flag.PrintDefaults()

		w.Close()
		flag.CommandLine.SetOutput(old)

		var buf [8192]byte
		n, _ := r.Read(buf[:])
		r.Close()

		for _, line := range strings.Split(string(buf[:n]), "\n") {
			if line != "" {
				console.Log("  %s", line)
			}
		}

		return nil, fmt.Errorf("no input path specified")
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	cfg.InputPath = args[0]

	if _, err := os.Stat(cfg.InputPath); err != nil {
		return nil, fmt.Errorf("error: %v", err)
	}

	return cfg, nil
}

func (cfg *Config) validate() error {
	if cfg.Quality < 0 || cfg.Quality > 100 {
		return fmt.Errorf("error: quality must be in range 0-100")
	}
	if cfg.QualityAlpha < 0 || cfg.QualityAlpha > 100 {
		return fmt.Errorf("error: alpha quality must be in range 0-100")
	}
	if cfg.Speed < 0 || cfg.Speed > 10 {
		return fmt.Errorf("error: encoding speed must be in range 0-10")
	}
	return nil
}

func (cfg *Config) GetEncodingOptions() avif.Options {
	return avif.Options{
		Quality:           cfg.Quality,
		QualityAlpha:      cfg.QualityAlpha,
		Speed:             cfg.Speed,
		ChromaSubsampling: image.YCbCrSubsampleRatio420,
	}
}
