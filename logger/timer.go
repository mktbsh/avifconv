package logger

import "time"

type Timer struct {
	StartTime time.Time
	Name      string
	Console   *Console
}

func (t *Timer) End() time.Duration {
	duration := time.Since(t.StartTime)
	t.Console.Info("%s completed in %v", t.Name, duration)
	return duration
}
