// cmd/server/main.go
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/your-org/mcp-k8s-server/internal/delivery/mcp"
	"github.com/your-org/mcp-k8s-server/internal/infrastructure"
	"github.com/your-org/mcp-k8s-server/internal/usecase"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to configuration file")
	flag.Parse()

	// Initialize logger
	logger := infrastructure.NewLogger()

	// Initialize cluster manager
	clusterManager := infrastructure.GetClusterManager(logger)

	// Initialize repository (in-memory)
	clusterRepo := infrastructure.NewInMemoryClusterRepository()

	// Initialize use cases
	clusterUseCase := usecase.NewClusterUseCase(clusterManager, clusterRepo, logger)
	k8sUseCase := usecase.NewK8sUseCase(clusterRepo, clusterManager, logger)

	// Create MCP server
	mcpServer, err := mcp.NewMCPServer(clusterUseCase, k8sUseCase, logger)
	if err != nil {
		logger.Error("Failed to create MCP server", "error", err)
		os.Exit(1)
	}

	// // Register tools
	// tools.RegisterClusterTools(mcpServer, clusterUseCase)
	// tools.RegisterPodTools(mcpServer, clusterUseCase)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down server...")

		// Cleanup resources
		clusterManager.CloseAll()
		cancel()
	}()

	// Run MCP server
	if err := mcpServer.Run(ctx); err != nil {
		logger.Error("MCP server error", "error", err)
		os.Exit(1)
	}
}
