package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestNewProgressBar(t *testing.T) {
	pb := NewProgressBar("Test", 100)

	if pb.title != "Test" {
		t.Errorf("Expected title 'Test', got '%s'", pb.title)
	}

	if pb.total != 100 {
		t.Errorf("Expected total 100, got %d", pb.total)
	}

	if pb.current != 0 {
		t.Errorf("Expected current 0, got %d", pb.current)
	}

	if pb.width != 50 {
		t.Errorf("Expected width 50, got %d", pb.width)
	}
}

func TestProgressBarOptions(t *testing.T) {
	var buf bytes.Buffer

	pb := NewProgressBar("Test", 100,
		WithWriter(&buf),
		WithWidth(20),
		WithoutPercent(),
		WithoutCount(),
	)

	if pb.writer != &buf {
		t.Error("Writer option not applied")
	}

	if pb.width != 20 {
		t.Errorf("Expected width 20, got %d", pb.width)
	}

	if pb.showPercent {
		t.Error("Expected showPercent to be false")
	}

	if pb.showCount {
		t.Error("Expected showCount to be false")
	}
}

func TestProgressBarUpdate(t *testing.T) {
	var buf bytes.Buffer
	pb := NewProgressBar("Processing", 10, WithWriter(&buf))

	// Test initial state
	pb.Start()
	output := buf.String()
	if !strings.Contains(output, "Processing") {
		t.Error("Progress bar should contain title")
	}
	if !strings.Contains(output, "0.0%") {
		t.Error("Progress bar should show 0.0%")
	}

	// Test update
	buf.Reset()
	pb.Update(5)
	output = buf.String()
	if !strings.Contains(output, "50.0%") {
		t.Error("Progress bar should show 50.0%")
	}
	if !strings.Contains(output, "(5/10)") {
		t.Error("Progress bar should show (5/10)")
	}

	// Test increment
	buf.Reset()
	pb.Increment()
	output = buf.String()
	if !strings.Contains(output, "60.0%") {
		t.Error("Progress bar should show 60.0% after increment")
	}
}

func TestProgressBarFinish(t *testing.T) {
	var buf bytes.Buffer
	pb := NewProgressBar("Done", 5, WithWriter(&buf))

	pb.Start()
	pb.Update(3)
	buf.Reset()

	pb.Finish()
	output := buf.String()

	if !strings.Contains(output, "100.0%") {
		t.Error("Finished progress bar should show 100.0%")
	}
	if !strings.Contains(output, "(5/5)") {
		t.Error("Finished progress bar should show (5/5)")
	}
	if !strings.Contains(output, "Done in") {
		t.Error("Finished progress bar should show completion time")
	}
}

func TestProgressBarOverflow(t *testing.T) {
	var buf bytes.Buffer
	pb := NewProgressBar("Test", 10, WithWriter(&buf))

	// Test update beyond total
	pb.Update(15)
	output := buf.String()
	if !strings.Contains(output, "100.0%") {
		t.Error("Progress bar should cap at 100% when exceeding total")
	}
}

func TestProgressBarZeroTotal(t *testing.T) {
	var buf bytes.Buffer
	pb := NewProgressBar("Test", 0, WithWriter(&buf))

	pb.Start()
	pb.Update(5)

	// Should not crash with zero total
	if pb.total != 0 {
		t.Error("Total should remain 0")
	}
}

func TestProgressBarWithoutOptions(t *testing.T) {
	var buf bytes.Buffer
	pb := NewProgressBar("Test", 10,
		WithWriter(&buf),
		WithoutPercent(),
		WithoutCount(),
	)

	pb.Start()
	output := buf.String()

	if strings.Contains(output, "%") {
		t.Error("Progress bar should not show percentage when disabled")
	}
	if strings.Contains(output, "(") {
		t.Error("Progress bar should not show count when disabled")
	}
}

func TestNewSpinner(t *testing.T) {
	spinner := NewSpinner("Loading...")

	if spinner.title != "Loading..." {
		t.Errorf("Expected title 'Loading...', got '%s'", spinner.title)
	}

	if len(spinner.frames) == 0 {
		t.Error("Spinner should have frames")
	}

	if spinner.interval <= 0 {
		t.Error("Spinner should have positive interval")
	}
}

func TestSpinnerOptions(t *testing.T) {
	var buf bytes.Buffer
	customFrames := []string{"a", "b", "c"}
	customInterval := 50 * time.Millisecond

	spinner := NewSpinner("Test",
		WithSpinnerWriter(&buf),
		WithSpinnerFrames(customFrames),
		WithSpinnerInterval(customInterval),
	)

	if spinner.writer != &buf {
		t.Error("Writer option not applied")
	}

	if len(spinner.frames) != 3 || spinner.frames[0] != "a" {
		t.Error("Custom frames not applied")
	}

	if spinner.interval != customInterval {
		t.Error("Custom interval not applied")
	}
}

func TestSpinnerStartStop(t *testing.T) {
	var buf bytes.Buffer
	spinner := NewSpinner("Processing", WithSpinnerWriter(&buf))

	spinner.Start()

	// Give it a moment to animate
	time.Sleep(150 * time.Millisecond)

	spinner.Stop()

	output := buf.String()
	if !strings.Contains(output, "Processing") {
		t.Error("Spinner output should contain title")
	}
	if !strings.Contains(output, "✓ Done in") {
		t.Error("Stopped spinner should show completion message")
	}
}

func TestSpinnerUpdateTitle(t *testing.T) {
	spinner := NewSpinner("Initial")

	spinner.UpdateTitle("Updated")

	if spinner.title != "Updated" {
		t.Errorf("Expected title 'Updated', got '%s'", spinner.title)
	}
}

func TestSpinnerMultipleStops(t *testing.T) {
	var buf bytes.Buffer
	spinner := NewSpinner("Test", WithSpinnerWriter(&buf))

	spinner.Start()
	time.Sleep(50 * time.Millisecond)

	spinner.Stop()
	initialOutput := buf.String()

	// Stop again - should not cause issues
	spinner.Stop()
	finalOutput := buf.String()

	// Output should not have changed after second stop
	if initialOutput != finalOutput {
		t.Error("Multiple stops should not change output")
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
	}
}

func TestProgressBarProgressCalculation(t *testing.T) {
	var buf bytes.Buffer
	pb := NewProgressBar("Test", 4, WithWriter(&buf))

	tests := []struct {
		current  int
		expected string
	}{
		{0, "0.0%"},
		{1, "25.0%"},
		{2, "50.0%"},
		{3, "75.0%"},
		{4, "100.0%"},
	}

	for _, test := range tests {
		buf.Reset()
		pb.Update(test.current)
		output := buf.String()
		if !strings.Contains(output, test.expected) {
			t.Errorf("Expected %s for current=%d, but output was: %s", test.expected, test.current, output)
		}
	}
}

func TestProgressBarBarRendering(t *testing.T) {
	var buf bytes.Buffer
	pb := NewProgressBar("Test", 10, WithWriter(&buf), WithWidth(10))

	pb.Update(5) // 50%
	output := buf.String()

	// Should have filled blocks
	if !strings.Contains(output, "█") {
		t.Error("Progress bar should contain filled blocks")
	}

	// Should have empty blocks
	if !strings.Contains(output, "░") {
		t.Error("Progress bar should contain empty blocks")
	}
}
