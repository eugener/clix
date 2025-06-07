package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// ProgressBar represents a progress bar for long-running operations
type ProgressBar struct {
	writer      io.Writer
	title       string
	total       int
	current     int
	width       int
	showPercent bool
	showCount   bool
	started     time.Time
	finished    bool
}

// ProgressBarOption configures a progress bar
type ProgressBarOption func(*ProgressBar)

// NewProgressBar creates a new progress bar
func NewProgressBar(title string, total int, opts ...ProgressBarOption) *ProgressBar {
	pb := &ProgressBar{
		writer:      os.Stderr,
		title:       title,
		total:       total,
		current:     0,
		width:       50,
		showPercent: true,
		showCount:   true,
		started:     time.Now(),
		finished:    false,
	}

	for _, opt := range opts {
		opt(pb)
	}

	return pb
}

// WithWriter sets the output writer
func WithWriter(w io.Writer) ProgressBarOption {
	return func(pb *ProgressBar) {
		pb.writer = w
	}
}

// WithWidth sets the progress bar width
func WithWidth(width int) ProgressBarOption {
	return func(pb *ProgressBar) {
		pb.width = width
	}
}

// WithoutPercent disables percentage display
func WithoutPercent() ProgressBarOption {
	return func(pb *ProgressBar) {
		pb.showPercent = false
	}
}

// WithoutCount disables count display
func WithoutCount() ProgressBarOption {
	return func(pb *ProgressBar) {
		pb.showCount = false
	}
}

// Start begins the progress bar display
func (pb *ProgressBar) Start() {
	pb.started = time.Now()
	pb.render()
}

// Update updates the progress bar to the specified value
func (pb *ProgressBar) Update(current int) {
	if current > pb.total {
		current = pb.total
	}
	pb.current = current
	pb.render()
}

// UpdateTitle updates the progress bar title
func (pb *ProgressBar) UpdateTitle(title string) {
	pb.title = title
	pb.render()
}

// Increment increases the progress by 1
func (pb *ProgressBar) Increment() {
	pb.Update(pb.current + 1)
}

// Finish completes the progress bar
func (pb *ProgressBar) Finish() {
	if !pb.finished {
		pb.current = pb.total
		pb.finished = true
		pb.render()
		fmt.Fprintf(pb.writer, "\n")
	}
}

// render draws the current progress bar state
func (pb *ProgressBar) render() {
	if pb.total <= 0 {
		return
	}

	// Calculate progress
	progress := float64(pb.current) / float64(pb.total)
	if progress > 1.0 {
		progress = 1.0
	}

	// Build progress bar
	filled := int(progress * float64(pb.width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", pb.width-filled)

	// Build status line
	var parts []string
	parts = append(parts, pb.title)
	parts = append(parts, fmt.Sprintf("[%s]", bar))

	if pb.showPercent {
		parts = append(parts, fmt.Sprintf("%.1f%%", progress*100))
	}

	if pb.showCount {
		parts = append(parts, fmt.Sprintf("(%d/%d)", pb.current, pb.total))
	}

	// Add elapsed time and ETA
	elapsed := time.Since(pb.started)
	if pb.current > 0 && !pb.finished {
		eta := time.Duration(float64(elapsed) * (float64(pb.total) / float64(pb.current)) - float64(elapsed))
		parts = append(parts, fmt.Sprintf("ETA: %v", eta.Truncate(time.Second)))
	} else if pb.finished {
		parts = append(parts, fmt.Sprintf("Done in %v", elapsed.Truncate(time.Millisecond)))
	}

	// Render with carriage return to overwrite previous line
	fmt.Fprintf(pb.writer, "\r%s", strings.Join(parts, " "))
}

// Spinner represents a spinning progress indicator
type Spinner struct {
	writer   io.Writer
	title    string
	frames   []string
	interval time.Duration
	current  int
	started  time.Time
	done     chan bool
	finished bool
}

// SpinnerOption configures a spinner
type SpinnerOption func(*Spinner)

// NewSpinner creates a new spinner
func NewSpinner(title string, opts ...SpinnerOption) *Spinner {
	s := &Spinner{
		writer:   os.Stderr,
		title:    title,
		frames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		interval: 100 * time.Millisecond,
		current:  0,
		done:     make(chan bool),
		finished: false,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// WithSpinnerWriter sets the output writer
func WithSpinnerWriter(w io.Writer) SpinnerOption {
	return func(s *Spinner) {
		s.writer = w
	}
}

// WithSpinnerFrames sets custom spinner frames
func WithSpinnerFrames(frames []string) SpinnerOption {
	return func(s *Spinner) {
		s.frames = frames
	}
}

// WithSpinnerInterval sets the spinner animation interval
func WithSpinnerInterval(interval time.Duration) SpinnerOption {
	return func(s *Spinner) {
		s.interval = interval
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.started = time.Now()
	go s.animate()
}

// Stop stops the spinner animation
func (s *Spinner) Stop() {
	if !s.finished {
		s.finished = true
		s.done <- true
		fmt.Fprintf(s.writer, "\r%s ✓ Done in %v\n", s.title, time.Since(s.started).Truncate(time.Millisecond))
	}
}

// UpdateTitle updates the spinner title
func (s *Spinner) UpdateTitle(title string) {
	s.title = title
}

// animate runs the spinner animation loop
func (s *Spinner) animate() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			if !s.finished {
				s.current = (s.current + 1) % len(s.frames)
				fmt.Fprintf(s.writer, "\r%s %s", s.frames[s.current], s.title)
			}
		}
	}
}

// Predefined spinner styles
var (
	SpinnerDots     = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	SpinnerLine     = []string{"|", "/", "-", "\\"}
	SpinnerArrows   = []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"}
	SpinnerBounce   = []string{"⠁", "⠂", "⠄", "⠂"}
	SpinnerCircle   = []string{"◐", "◓", "◑", "◒"}
	SpinnerSquare   = []string{"◰", "◳", "◲", "◱"}
)