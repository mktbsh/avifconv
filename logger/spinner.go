package logger

import (
	"fmt"
	"time"
)

type Spinner struct {
	Frames  []string
	Message string
	Console *Console
	Done    chan bool
}

func (s *Spinner) Start() {
	go func() {
		i := 0
		for {
			select {
			case <-s.Done:
				fmt.Print("\r")
				return
			default:
				frame := s.Frames[i%len(s.Frames)]
				fmt.Printf("\r%s %s ", frame, s.Message)
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (s *Spinner) Stop(success bool, message string) {
	s.Done <- true

	if success {
		s.Console.Success("%s", message)
	} else {
		s.Console.Error("%s", message)
	}
}
