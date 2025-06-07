package main

import (
	"context"
	"fmt"
	"time"

	"github.com/eugener/clix/cli"
	"github.com/eugener/clix/core"
)

// Configuration structures for different commands
type ContainerListConfig struct {
	All    bool   `posix:"a,all,Show all containers (default shows just running)"`
	Format string `posix:"f,format,Output format (table|json|yaml),default=table"`
}

type ContainerRunConfig struct {
	Image      string `posix:",image,Container image,required"`
	Name       string `posix:"n,name,Container name"`
	Detach     bool   `posix:"d,detach,Run container in background"`
	Port       string `posix:"p,port,Port mapping (host:container)"`
	Format     string `posix:"f,format,Output format,default=table"`
}

type ImageListConfig struct {
	All    bool   `posix:"a,all,Show all images"`
	Format string `posix:"f,format,Output format,default=table"`
}

type ImagePullConfig struct {
	Image  string `posix:",image,Image to pull,required"`
	Format string `posix:"f,format,Output format,default=table"`
}

type GetPodsConfig struct {
	Namespace string `posix:"n,namespace,Kubernetes namespace,default=default"`
	Format    string `posix:"f,format,Output format,default=table"`
}

type GetServicesConfig struct {
	Namespace string `posix:"n,namespace,Kubernetes namespace,default=default"`
	Format    string `posix:"f,format,Output format,default=table"`
}

func main() {
	// Create Docker-style nested commands using the unified command system
	
	// Container management subgroup - just a command with subcommands, no runner
	containerCommand := core.NewCommand[struct{}]("container", "Manage containers", nil)
	containerCommand.AddSubcommand(core.NewCommand("ls", "List containers", 
		func(ctx context.Context, config ContainerListConfig) error {
			fmt.Printf("🐳 Listing containers (all=%v, format=%s)\n", config.All, config.Format)
			
			containers := []map[string]interface{}{
				{"id": "abc123", "name": "web-server", "status": "running", "ports": "80:8080"},
				{"id": "def456", "name": "database", "status": "stopped", "ports": ""},
			}

			if !config.All {
				// Filter to show only running containers
				running := make([]map[string]interface{}, 0)
				for _, container := range containers {
					if container["status"] == "running" {
						running = append(running, container)
					}
				}
				containers = running
			}

			return cli.FormatAndOutput(containers, cli.Format(config.Format))
		}))
	
	containerCommand.AddSubcommand(core.NewCommand("run", "Run a new container",
		func(ctx context.Context, config ContainerRunConfig) error {
			fmt.Printf("🚀 Running container from image: %s\n", config.Image)
			if config.Name != "" {
				fmt.Printf("   Container name: %s\n", config.Name)
			}
			if config.Detach {
				fmt.Printf("   Running in background (detached)\n")
			}
			if config.Port != "" {
				fmt.Printf("   Port mapping: %s\n", config.Port)
			}

			// Simulate container startup
			time.Sleep(500 * time.Millisecond)

			result := map[string]interface{}{
				"container_id": "xyz789",
				"image":        config.Image,
				"name":         config.Name,
				"status":       "running",
				"created":      time.Now().Format(time.RFC3339),
			}

			return cli.FormatAndOutput(result, cli.Format(config.Format))
		}))

	// Image management subgroup
	imageCommand := core.NewCommand[struct{}]("image", "Manage container images", nil)
	imageCommand.AddSubcommand(core.NewCommand("ls", "List images",
		func(ctx context.Context, config ImageListConfig) error {
			fmt.Printf("🖼️  Listing images (all=%v, format=%s)\n", config.All, config.Format)
			
			images := []map[string]interface{}{
				{"repository": "nginx", "tag": "latest", "size": "133MB"},
				{"repository": "postgres", "tag": "13", "size": "314MB"},
			}

			return cli.FormatAndOutput(images, cli.Format(config.Format))
		}))

	imageCommand.AddSubcommand(core.NewCommand("pull", "Pull an image",
		func(ctx context.Context, config ImagePullConfig) error {
			fmt.Printf("📥 Pulling image: %s\n", config.Image)
			
			// Simulate image pull
			time.Sleep(1 * time.Second)
			
			result := map[string]interface{}{
				"image":   config.Image,
				"status":  "Downloaded",
				"pulled":  time.Now().Format(time.RFC3339),
			}

			return cli.FormatAndOutput(result, cli.Format(config.Format))
		}))

	// Main docker group - just another command that has subcommands
	dockerCommand := core.NewCommand[struct{}]("docker", "Docker container management", nil)
	dockerCommand.AddSubcommand(containerCommand)
	dockerCommand.AddSubcommand(imageCommand)

	// Create Kubernetes-style nested commands
	
	// Get subgroup
	getCommand := core.NewCommand[struct{}]("get", "Display one or many resources", nil)
	getCommand.AddSubcommand(core.NewCommand("pods", "List pods",
		func(ctx context.Context, config GetPodsConfig) error {
			fmt.Printf("☸️  Getting pods in namespace: %s (format=%s)\n", config.Namespace, config.Format)
			
			pods := []map[string]interface{}{
				{"name": "web-pod-1", "status": "Running", "namespace": config.Namespace},
				{"name": "api-pod-2", "status": "Running", "namespace": config.Namespace},
			}

			return cli.FormatAndOutput(pods, cli.Format(config.Format))
		}))

	getCommand.AddSubcommand(core.NewCommand("services", "List services", 
		func(ctx context.Context, config GetServicesConfig) error {
			fmt.Printf("🔗 Getting services in namespace: %s (format=%s)\n", config.Namespace, config.Format)
			
			services := []map[string]interface{}{
				{"name": "web-service", "type": "ClusterIP", "namespace": config.Namespace},
				{"name": "api-service", "type": "LoadBalancer", "namespace": config.Namespace},
			}

			return cli.FormatAndOutput(services, cli.Format(config.Format))
		}))

	// Main kubectl group
	kubectlCommand := core.NewCommand[struct{}]("kubectl", "Kubernetes command-line tool", nil)
	kubectlCommand.AddSubcommand(getCommand)

	// Build and run the CLI - parent commands are just commands!
	cli.New("nested-demo").
		Version("1.0.0").
		Description("Demonstration of nested commands with Docker and Kubernetes-style CLIs").
		Interactive().
		Recovery().
		WithCommands(dockerCommand, kubectlCommand).  // Using unified command approach
		RunWithArgs(context.Background())
}