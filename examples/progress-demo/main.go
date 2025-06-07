// Package main demonstrates progress indicators and UI components
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/eugener/clix/cli"
	"github.com/eugener/clix/core"
)

// Configuration for commands with progress
type ProgressConfig struct {
	Count int    `posix:"c,count,Number of items to process,default=10"`
	Delay string `posix:"d,delay,Delay between items,default=100ms"`
}

type SpinnerConfig struct {
	Duration string `posix:"d,duration,How long to run,default=3s"`
}

func main() {
	app := cli.New("progress-demo").
		Version("1.0.0").
		Description("Demonstration of progress indicators and UI components").
		Recovery().
		WithCommands(
			// Progress bar example
			core.NewCommand("process", "Process items with progress bar", func(ctx context.Context, config ProgressConfig) error {
				delay, err := time.ParseDuration(config.Delay)
				if err != nil {
					return fmt.Errorf("invalid delay: %v", err)
				}

				// Create progress bar
				pb := cli.NewProgressBar("Processing items", config.Count)
				pb.Start()
				defer pb.Finish()

				// Simulate processing
				for i := 0; i < config.Count; i++ {
					time.Sleep(delay)
					pb.Update(i + 1)
				}

				fmt.Println("\n✅ All items processed successfully!")
				return nil
			}),

			// Spinner example
			core.NewCommand("load", "Simulate loading with spinner", func(ctx context.Context, config SpinnerConfig) error {
				duration, err := time.ParseDuration(config.Duration)
				if err != nil {
					return fmt.Errorf("invalid duration: %v", err)
				}

				// Create spinner
				spinner := cli.NewSpinner("Loading data from server...")
				spinner.Start()
				defer spinner.Stop()

				// Simulate phases of work
				phases := []struct {
					message  string
					duration time.Duration
				}{
					{"Connecting to server...", duration / 4},
					{"Authenticating...", duration / 4},
					{"Downloading data...", duration / 3},
					{"Processing results...", duration / 6},
				}

				for _, phase := range phases {
					spinner.UpdateTitle(phase.message)
					time.Sleep(phase.duration)
				}

				fmt.Println("✅ Data loaded successfully!")
				return nil
			}),

			// Different spinner styles
			core.NewCommand("spinners", "Show different spinner styles", func(ctx context.Context, config struct{}) error {
				styles := []struct {
					name   string
					frames []string
				}{
					{"Dots", cli.SpinnerDots},
					{"Line", cli.SpinnerLine},
					{"Arrows", cli.SpinnerArrows},
					{"Bounce", cli.SpinnerBounce},
					{"Circle", cli.SpinnerCircle},
					{"Square", cli.SpinnerSquare},
				}

				for _, style := range styles {
					fmt.Printf("🎨 %s style: ", style.name)
					spinner := cli.NewSpinner(fmt.Sprintf("Demo %s spinner", style.name),
						cli.WithSpinnerFrames(style.frames),
						cli.WithSpinnerInterval(200*time.Millisecond),
					)

					spinner.Start()
					time.Sleep(2 * time.Second)
					spinner.Stop()
				}

				return nil
			}),

			// Combined progress and output formatting
			core.NewCommand("export", "Export data with progress and structured output", func(ctx context.Context, config struct {
				Format string `posix:"f,format,Output format,default=table"`
				Items  int    `posix:"i,items,Number of items,default=5"`
			}) error {
				// Validate format
				if !cli.ValidFormat(config.Format) {
					return fmt.Errorf("invalid format: %s. Valid formats: json, yaml, table, text", config.Format)
				}

				// Create progress bar for data generation
				pb := cli.NewProgressBar("Generating export data", config.Items)
				pb.Start()

				// Generate sample data with progress
				var data []map[string]interface{}
				for i := 0; i < config.Items; i++ {
					time.Sleep(200 * time.Millisecond)
					
					item := map[string]interface{}{
						"id":       i + 1,
						"name":     fmt.Sprintf("Item %d", i+1),
						"status":   []string{"active", "pending", "completed"}[i%3],
						"created":  time.Now().Add(-time.Duration(i)*time.Hour).Format("2006-01-02"),
						"value":    float64(i+1) * 10.5,
					}
					data = append(data, item)
					pb.Update(i + 1)
				}

				pb.Finish()
				fmt.Println()

				// Create spinner for export processing
				spinner := cli.NewSpinner("Formatting export data...")
				spinner.Start()
				time.Sleep(500 * time.Millisecond) // Simulate formatting time
				spinner.Stop()

				// Output formatted data
				fmt.Printf("📊 Export complete! Here's your data in %s format:\n\n", config.Format)
				return cli.FormatAndOutput(data, cli.Format(config.Format))
			}),

			// Delayed spinner demo
			core.NewCommand("quick", "Quick operation with delayed spinner", func(ctx context.Context, config struct {
				Fast bool `posix:"f,fast,Make operation fast to test delay"`
			}) error {
				// Use delayed spinner that only shows if operation takes > 1 second
				spinner := cli.DelayedSpinner("Processing (might be quick)...", 1*time.Second)
				spinner.Start()
				defer spinner.Stop()

				if config.Fast {
					// Quick operation - spinner shouldn't show
					time.Sleep(100 * time.Millisecond)
					fmt.Println("✅ Quick operation completed!")
				} else {
					// Slow operation - spinner will show after delay
					time.Sleep(2 * time.Second)
					fmt.Println("✅ Slow operation completed!")
				}

				return nil
			}),

			// Help command
			cli.Cmd("help", "Show help information", func() error {
				fmt.Println("📖 Progress Demo Commands:")
				fmt.Println("  process   - Show progress bar with customizable count and delay")
				fmt.Println("  load      - Show spinner with customizable duration")
				fmt.Println("  spinners  - Demonstrate different spinner styles")
				fmt.Println("  export    - Combined progress + structured output")
				fmt.Println("  quick     - Delayed spinner demonstration")
				fmt.Println()
				fmt.Println("💡 Examples:")
				fmt.Println("  progress-demo process --count 20 --delay 200ms")
				fmt.Println("  progress-demo load --duration 5s")
				fmt.Println("  progress-demo export --format json --items 10")
				fmt.Println("  progress-demo quick --fast")
				return nil
			}),
		).
		Build()

	app.RunWithArgs(context.Background())
}