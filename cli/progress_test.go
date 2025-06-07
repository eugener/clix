package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestNewProgressBar(t *testing.T) {
	pb := NewProgressBar("Test Progress", 100)
	if pb == nil {
		t.Fatal("NewProgressBar returned nil")
	}
}

func TestNewSpinner(t *testing.T) {
	spinner := NewSpinner("Loading...")
	if spinner == nil {
		t.Fatal("NewSpinner returned nil")
	}
}

func TestProgressWithOptions(t *testing.T) {
	var buf bytes.Buffer
	pb := NewProgressBar("Test", 10,
		WithProgressWriter(&buf),
		WithProgressWidth(20),
		WithoutPercent(),
		WithoutCount(),
	)

	pb.Start()
	output := buf.String()

	if !strings.Contains(output, "Test") {
		t.Error("Progress bar should contain title")
	}
}

func TestSpinnerWithOptions(t *testing.T) {
	var buf bytes.Buffer
	spinner := NewSpinner("Processing",
		WithSpinnerWriter(&buf),
		WithSpinnerFrames(SpinnerLine),
		WithSpinnerInterval(50*time.Millisecond),
	)

	spinner.Start()
	time.Sleep(100 * time.Millisecond)
	spinner.Stop()

	output := buf.String()
	if !strings.Contains(output, "Processing") {
		t.Error("Spinner should contain title")
	}
}

func TestWithProgress(t *testing.T) {
	var buf bytes.Buffer
	
	handler := WithProgress("Processing items", 5, func(config struct{}, pb *ProgressBar) error {
		// Override writer for testing
		pb = NewProgressBar("Processing items", 5, WithProgressWriter(&buf))
		pb.Start()
		defer pb.Finish()
		
		for i := 0; i < 5; i++ {
			pb.Update(i + 1)
		}
		return nil
	})

	err := handler(struct{}{})
	if err != nil {
		t.Errorf("WithProgress handler failed: %v", err)
	}
}

func TestWithSpinner(t *testing.T) {
	var buf bytes.Buffer
	
	handler := WithSpinner("Loading data", func(config struct{}, spinner *Spinner) error {
		// Override writer for testing
		testSpinner := NewSpinner("Loading data", WithSpinnerWriter(&buf))
		testSpinner.Start()
		defer testSpinner.Stop()
		
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	err := handler(struct{}{})
	if err != nil {
		t.Errorf("WithSpinner handler failed: %v", err)
	}
}

func TestProgress(t *testing.T) {
	var buf bytes.Buffer
	pb, callback := Progress("Test", 10, WithProgressWriter(&buf))
	
	pb.Start()
	callback(5, 10, "Halfway done")
	pb.Finish()

	output := buf.String()
	if !strings.Contains(output, "Test") {
		t.Error("Progress should contain initial title")
	}
}

func TestProgressWithWriter(t *testing.T) {
	var buf bytes.Buffer
	pb := ProgressWithWriter(&buf, "Test", 10)
	
	pb.Start()
	pb.Update(5)
	pb.Finish()

	output := buf.String()
	if !strings.Contains(output, "Test") {
		t.Error("Progress should contain title")
	}
	if !strings.Contains(output, "50.0%") {
		t.Error("Progress should show 50% at halfway point")
	}
}

func TestSpinnerWithWriter(t *testing.T) {
	var buf bytes.Buffer
	spinner := SpinnerWithWriter(&buf, "Loading")
	
	spinner.Start()
	time.Sleep(100 * time.Millisecond)
	spinner.Stop()

	output := buf.String()
	if !strings.Contains(output, "Loading") {
		t.Error("Spinner should contain title")
	}
}

func TestDelayedSpinner(t *testing.T) {
	var buf bytes.Buffer
	delay := 50 * time.Millisecond
	
	spinner := DelayedSpinner("Slow operation", delay, WithSpinnerWriter(&buf))
	spinner.Start()
	
	// Stop immediately - should not have started yet
	spinner.Stop()
	
	// Wait for delay period to pass
	time.Sleep(delay + 10*time.Millisecond)
	
	// Should have minimal output since it was stopped before delay
	output := buf.String()
	// The output might be empty or minimal since we stopped it quickly
	_ = output // Just verify it doesn't crash
}

func TestDelayedSpinnerWithDelay(t *testing.T) {
	var buf bytes.Buffer
	delay := 20 * time.Millisecond
	
	spinner := DelayedSpinner("Delayed operation", delay, WithSpinnerWriter(&buf))
	spinner.Start()
	
	// Wait for delay + some animation time
	time.Sleep(delay + 50*time.Millisecond)
	
	spinner.Stop()
	
	output := buf.String()
	if !strings.Contains(output, "Delayed operation") {
		t.Error("Delayed spinner should eventually show title")
	}
}

func TestDelayedSpinnerUpdateTitle(t *testing.T) {
	spinner := DelayedSpinner("Initial", 10*time.Millisecond)
	spinner.UpdateTitle("Updated")
	
	// Should not crash and should update internal title
	if spinner.title != "Updated" {
		t.Error("DelayedSpinner should update title")
	}
}

func TestPredefinedSpinnerStyles(t *testing.T) {
	styles := [][]string{
		SpinnerDots,
		SpinnerLine,
		SpinnerArrows,
		SpinnerBounce,
		SpinnerCircle,
		SpinnerSquare,
	}

	for i, style := range styles {
		if len(style) == 0 {
			t.Errorf("Predefined spinner style %d should not be empty", i)
		}
		
		// Test that we can create a spinner with each style
		var buf bytes.Buffer
		spinner := NewSpinner("Test", WithSpinnerWriter(&buf), WithSpinnerFrames(style))
		spinner.Start()
		time.Sleep(50 * time.Millisecond)
		spinner.Stop()
	}
}

func TestSimpleProgress(t *testing.T) {
	var buf bytes.Buffer
	pb := NewProgressBar("Test", 10, WithProgressWriter(&buf))
	callback := SimpleProgress(pb)
	
	pb.Start()
	callback(3, 10, "Working...")
	callback(7, 10, "Almost done...")
	pb.Finish()

	output := buf.String()
	if !strings.Contains(output, "Test") {
		t.Error("Progress should show title")
	}
}

func TestProgressBarUpdate(t *testing.T) {
	var buf bytes.Buffer
	pb := NewProgressBar("Test", 4, WithProgressWriter(&buf))
	
	tests := []struct {
		update   int
		expected string
	}{
		{1, "25.0%"},
		{2, "50.0%"},
		{3, "75.0%"},
		{4, "100.0%"},
	}
	
	pb.Start()
	for _, test := range tests {
		buf.Reset()
		pb.Update(test.update)
		output := buf.String()
		if !strings.Contains(output, test.expected) {
			t.Errorf("Expected %s for update %d, got: %s", test.expected, test.update, output)
		}
	}
}

// Integration test for realistic usage
func TestProgressBarIntegration(t *testing.T) {
	var buf bytes.Buffer
	
	// Simulate a file processing operation
	files := []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt", "file5.txt"}
	pb := NewProgressBar("Processing files", len(files), WithProgressWriter(&buf))
	
	pb.Start()
	for i, file := range files {
		// Simulate work
		time.Sleep(1 * time.Millisecond)
		
		// Update progress
		pb.Update(i + 1)
		
		// Verify progress shows correct percentage
		if i == 2 { // 3/5 = 60%
			output := buf.String()
			if !strings.Contains(output, "60.0%") {
				t.Error("Progress should show 60% at 3/5 completion")
			}
		}
		
		_ = file // Use the variable to avoid unused variable warning
	}
	
	pb.Finish()
	
	finalOutput := buf.String()
	if !strings.Contains(finalOutput, "100.0%") {
		t.Error("Final progress should show 100%")
	}
	if !strings.Contains(finalOutput, "Done in") {
		t.Error("Finished progress should show completion time")
	}
}