package cli

import (
	"io"
	"sync"
	"time"

	"github.com/eugener/clix/internal/ui"
)

// ProgressBar provides a visual progress indicator for long-running operations
type ProgressBar = ui.ProgressBar

// ProgressBarOption configures a progress bar
type ProgressBarOption = ui.ProgressBarOption

// NewProgressBar creates a new progress bar with the given title and total steps
func NewProgressBar(title string, total int, opts ...ProgressBarOption) *ProgressBar {
	return ui.NewProgressBar(title, total, opts...)
}

// Spinner provides a spinning progress indicator for operations with unknown duration
type Spinner = ui.Spinner

// SpinnerOption configures a spinner
type SpinnerOption = ui.SpinnerOption

// NewSpinner creates a new spinner with the given title
func NewSpinner(title string, opts ...SpinnerOption) *Spinner {
	return ui.NewSpinner(title, opts...)
}

// Progress bar options - re-exported for convenience
var (
	WithProgressWriter = ui.WithWriter
	WithProgressWidth  = ui.WithWidth
	WithoutPercent     = ui.WithoutPercent
	WithoutCount       = ui.WithoutCount
)

// Spinner options - re-exported for convenience
var (
	WithSpinnerWriter   = ui.WithSpinnerWriter
	WithSpinnerFrames   = ui.WithSpinnerFrames
	WithSpinnerInterval = ui.WithSpinnerInterval
)

// Predefined spinner styles - re-exported for convenience
var (
	SpinnerDots   = ui.SpinnerDots
	SpinnerLine   = ui.SpinnerLine
	SpinnerArrows = ui.SpinnerArrows
	SpinnerBounce = ui.SpinnerBounce
	SpinnerCircle = ui.SpinnerCircle
	SpinnerSquare = ui.SpinnerSquare
)

// ProgressCmd creates a command that automatically shows progress for long-running operations
func ProgressCmd[T any](name, description string, handler func(config T, progress *ProgressBar) error) any {
	return func(config T) error {
		// Create a default progress bar (will be updated by the handler)
		pb := NewProgressBar(description, 100)
		pb.Start()
		defer pb.Finish()

		return handler(config, pb)
	}
}

// SpinnerCmd creates a command that automatically shows a spinner for operations with unknown duration
func SpinnerCmd[T any](name, description string, handler func(config T, spinner *Spinner) error) any {
	return func(config T) error {
		// Create a spinner
		spinner := NewSpinner(description)
		spinner.Start()
		defer spinner.Stop()

		return handler(config, spinner)
	}
}

// WithProgress wraps a handler to automatically show progress
func WithProgress[T any](title string, total int, handler func(config T, progress *ProgressBar) error) func(T) error {
	return func(config T) error {
		pb := NewProgressBar(title, total)
		pb.Start()
		defer pb.Finish()

		return handler(config, pb)
	}
}

// WithSpinner wraps a handler to automatically show a spinner
func WithSpinner[T any](title string, handler func(config T, spinner *Spinner) error) func(T) error {
	return func(config T) error {
		spinner := NewSpinner(title)
		spinner.Start()
		defer spinner.Stop()

		return handler(config, spinner)
	}
}

// ProgressCallback is a function type for reporting progress
type ProgressCallback func(current, total int, message string)

// SimpleProgress creates a simple progress callback that updates a progress bar
func SimpleProgress(pb *ProgressBar) ProgressCallback {
	return func(current, total int, message string) {
		if message != "" {
			// Update the title if message is provided
			pb.UpdateTitle(message)
		}
		pb.Update(current)
	}
}

// Progress creates a new progress bar and returns it along with a simple callback
func Progress(title string, total int, opts ...ProgressBarOption) (*ProgressBar, ProgressCallback) {
	pb := NewProgressBar(title, total, opts...)
	callback := SimpleProgress(pb)
	return pb, callback
}

// ProgressWithWriter is a convenience function to create a progress bar with a custom writer
func ProgressWithWriter(writer io.Writer, title string, total int) *ProgressBar {
	return NewProgressBar(title, total, WithProgressWriter(writer))
}

// SpinnerWithWriter is a convenience function to create a spinner with a custom writer
func SpinnerWithWriter(writer io.Writer, title string) *Spinner {
	return NewSpinner(title, WithSpinnerWriter(writer))
}

// DelayedSpinner creates a spinner that only shows after a delay (useful for operations that might be quick)
func DelayedSpinner(title string, delay time.Duration, opts ...SpinnerOption) *DelayedSpinnerWrapper {
	return &DelayedSpinnerWrapper{
		title:   title,
		delay:   delay,
		opts:    opts,
		started: time.Now(),
	}
}

// DelayedSpinnerWrapper wraps a spinner with delayed activation
type DelayedSpinnerWrapper struct {
	title   string
	delay   time.Duration
	opts    []SpinnerOption
	spinner *Spinner
	started time.Time
	stopped bool
	mu      sync.RWMutex
}

// Start begins the delayed spinner
func (d *DelayedSpinnerWrapper) Start() {
	go func() {
		time.Sleep(d.delay)
		
		d.mu.Lock()
		defer d.mu.Unlock()
		
		if !d.stopped && time.Since(d.started) >= d.delay {
			d.spinner = NewSpinner(d.title, d.opts...)
			d.spinner.Start()
		}
	}()
}

// Stop stops the delayed spinner
func (d *DelayedSpinnerWrapper) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.stopped = true
	if d.spinner != nil {
		d.spinner.Stop()
	}
}

// UpdateTitle updates the spinner title
func (d *DelayedSpinnerWrapper) UpdateTitle(title string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.title = title
	if d.spinner != nil {
		d.spinner.UpdateTitle(title)
	}
}
