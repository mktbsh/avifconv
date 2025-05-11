package main

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"avifconv/logger"

	"github.com/gen2brain/avif"
	_ "golang.org/x/image/webp"
)

var supportedFormats = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

type Processor struct {
	Options    avif.Options
	Console    *logger.Console
	NumWorkers int
	QueueSize  int
}

type ProcessStats struct {
	mu                  sync.Mutex
	TotalOriginalSize   int64
	TotalCompressedSize int64
	TotalFiles          int
	ProcessedFiles      int
	SuccessfulFiles     int
	FailedFiles         int
}

type workerStatus struct {
	StartTime   time.Time
	CurrentFile string
	Busy        bool
}

func NewProcessor(cfg *Config, console *logger.Console) *Processor {
	return &Processor{
		Options:    cfg.GetEncodingOptions(),
		NumWorkers: cfg.Workers,
		QueueSize:  cfg.QueueSize,
		Console:    console,
	}
}

func (p *Processor) ProcessPath(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path validation error: %w", err)
	}

	if fileInfo.IsDir() {
		return p.ProcessDirectory(path)
	}

	return p.ProcessSingleFile(path)
}

func (p *Processor) ProcessDirectory(dirPath string) error {
	p.Console.Info("Processing directory: %s (workers: %d, quality: %d, speed: %d)",
		dirPath, p.NumWorkers, p.Options.Quality, p.Options.Speed)

	filesToProcess, err := p.collectFiles(dirPath)
	if err != nil {
		return fmt.Errorf("file collection error: %w", err)
	}

	totalFiles := len(filesToProcess)
	if totalFiles == 0 {
		p.Console.Warn("No files found to process")
		return nil
	}

	p.Console.Info("Starting batch processing of %d files", totalFiles)

	// Start parallel processing
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stats := &ProcessStats{TotalFiles: totalFiles}
	p.processFilesParallel(ctx, filesToProcess, stats)

	// Display results
	p.displayResults(stats)

	return nil
}

func (p *Processor) collectFiles(dirPath string) ([]string, error) {
	var filesToProcess []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if supportedFormats[ext] {
			filesToProcess = append(filesToProcess, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error while exploring directory: %w", err)
	}

	return filesToProcess, nil
}

func (p *Processor) processFilesParallel(ctx context.Context, files []string, stats *ProcessStats) {
	queueSize := p.QueueSize
	if queueSize > len(files) {
		queueSize = len(files)
	}

	jobs := make(chan string, queueSize)

	workerStatuses := make([]workerStatus, p.NumWorkers)

	bar := p.Console.NewProgressBar(int64(len(files)), "Converting images")

	var wg sync.WaitGroup

	for w := 0; w < p.NumWorkers; w++ {
		wg.Add(1)
		go p.worker(ctx, w, jobs, stats, &workerStatuses[w], &wg, bar)
	}

	go func() {
		for _, file := range files {
			select {
			case <-ctx.Done():
				return
			case jobs <- file:
			}
		}
		close(jobs)
	}()

	wg.Wait()
	bar.Complete()
}

func (p *Processor) worker(ctx context.Context, id int, jobs <-chan string, stats *ProcessStats,
	status *workerStatus, wg *sync.WaitGroup, bar *logger.ProgressBar) {
	defer wg.Done()

	for filePath := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			stats.mu.Lock()
			status.Busy = true
			status.CurrentFile = filePath
			status.StartTime = time.Now()
			stats.mu.Unlock()

			origSize, compSize, err := p.processFileWithStats(filePath)

			stats.mu.Lock()
			stats.ProcessedFiles++
			progress := float64(stats.ProcessedFiles) / float64(stats.TotalFiles) * 100

			status.Busy = false
			status.CurrentFile = ""

			if err != nil {
				stats.FailedFiles++
				p.Console.Error("Worker %d: Error processing %s: %v (%.1f%% complete)",
					id+1, filepath.Base(filePath), err, progress)
			} else {
				stats.SuccessfulFiles++
				stats.TotalOriginalSize += origSize
				stats.TotalCompressedSize += compSize
			}

			bar.Increment(1)

			stats.mu.Unlock()
		}
	}
}

func (p *Processor) displayResults(stats *ProcessStats) {
	var overallCompressionRatio float64
	if stats.TotalOriginalSize > 0 {
		overallCompressionRatio = float64(stats.TotalCompressedSize) / float64(stats.TotalOriginalSize) * 100
	}

	table := p.Console.NewTable([]string{"Metric", "Value"})
	table.AddRow("Processed files", fmt.Sprintf("%d/%d", stats.SuccessfulFiles, stats.TotalFiles))
	table.AddRow("Failed files", fmt.Sprintf("%d", stats.FailedFiles))
	table.AddRow("Original size", fmt.Sprintf("%.2f MB", float64(stats.TotalOriginalSize)/1024/1024))
	table.AddRow("Compressed size", fmt.Sprintf("%.2f MB", float64(stats.TotalCompressedSize)/1024/1024))
	table.AddRow("Compression ratio", fmt.Sprintf("%.1f%%", overallCompressionRatio))

	if overallCompressionRatio > 0 && stats.TotalOriginalSize > stats.TotalCompressedSize {
		savedSpace := stats.TotalOriginalSize - stats.TotalCompressedSize
		table.AddRow("Space saved", fmt.Sprintf("%.2f MB", float64(savedSpace)/1024/1024))
	}

	p.Console.Info("\nProcessing Summary:")
	table.Print()
}

func (p *Processor) processFileWithStats(filePath string) (int64, int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get file info: %w", err)
	}
	originalSize := fileInfo.Size()

	f, err := os.Open(filePath)
	if err != nil {
		return originalSize, 0, fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return originalSize, 0, fmt.Errorf("error decoding image: %w", err)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(filePath), "*.avif")
	if err != nil {
		return originalSize, 0, fmt.Errorf("error creating temporary file: %w", err)
	}
	tempPath := tempFile.Name()

	tempFileClosed := false
	defer func() {
		if !tempFileClosed {
			tempFile.Close()
		}
		if err != nil {
			if _, statErr := os.Stat(tempPath); statErr == nil {
				os.Remove(tempPath)
			}
		}
	}()

	err = avif.Encode(tempFile, img, p.Options)
	if err != nil {
		return originalSize, 0, fmt.Errorf("error encoding to AVIF: %w", err)
	}

	tempFile.Close()
	tempFileClosed = true

	compressedFileInfo, err := os.Stat(tempPath)
	if err != nil {
		return originalSize, 0, fmt.Errorf("failed to get compressed file info: %w", err)
	}
	compressedSize := compressedFileInfo.Size()

	outputPath := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".avif"

	err = os.Remove(filePath)
	if err != nil {
		return originalSize, compressedSize, fmt.Errorf("error deleting original file: %w", err)
	}

	err = os.Rename(tempPath, outputPath)
	if err != nil {
		return originalSize, compressedSize, fmt.Errorf("error renaming file: %w", err)
	}

	return originalSize, compressedSize, nil
}

func (p *Processor) ProcessSingleFile(filePath string) error {
	p.Console.Info("Processing file: %s", filePath)

	timer := p.Console.StartTimer("File conversion")

	origSize, compSize, err := p.processFileWithStats(filePath)
	if err != nil {
		p.Console.Error("Processing failed: %v", err)
		return fmt.Errorf("file processing error: %w", err)
	}

	duration := timer.End()

	var compressionRatio float64
	if origSize > 0 {
		compressionRatio = float64(compSize) / float64(origSize) * 100
	}

	p.Console.Success("Successfully converted to AVIF: %s", filePath)
	p.Console.Info("Compression ratio: %.1f%% (%d KB → %d KB) in %v",
		compressionRatio, origSize/1024, compSize/1024, duration)

	return nil
}
