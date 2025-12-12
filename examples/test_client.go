package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// Simple test client to interact with MCP server
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_client.go <request.json>")
		fmt.Println("Example: go run test_client.go test_list_tools.json")
		os.Exit(1)
	}

	// Read request from file
	requestFile := os.Args[1]
	data, err := os.ReadFile(requestFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Start the server
	cmd := exec.Command("go", "run", "../cmd/server/main.go")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("Error creating stdin pipe: %v\n", err)
		os.Exit(1)
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Error creating stdout pipe: %v\n", err)
		os.Exit(1)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}

	// Send request
	fmt.Printf("Sending request: %s\n", string(data))
	if _, err := stdin.Write(data); err != nil {
		fmt.Printf("Error writing to stdin: %v\n", err)
		os.Exit(1)
	}
	stdin.Write([]byte("\n"))
	stdin.Close()

	// Read response
	scanner := bufio.NewScanner(stdout)
	if scanner.Scan() {
		var response map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &response); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			fmt.Printf("Raw response: %s\n", scanner.Text())
		} else {
			prettyJSON, _ := json.MarshalIndent(response, "", "  ")
			fmt.Printf("Response:\n%s\n", string(prettyJSON))
		}
	}

	cmd.Wait()
}
