package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
)

func main() {
	cmd := exec.Command("./cmd/server/mcp-k8s-server.exe")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()

	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 2*1024*1024), 2*1024*1024)

	// Initialize
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	json.NewEncoder(stdin).Encode(initReq)
	scanner.Scan()

	// Register cluster
	registerReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params: map[string]any{
			"name": "k8s_cluster_register",
			"arguments": map[string]any{
				"cluster_id":      "default",
				"kubeconfig_path": "path to kubeconfig",
			},
		},
	}

	json.NewEncoder(stdin).Encode(registerReq)
	scanner.Scan()

	// Get deployment info
	getDeploymentReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]any{
			"name": "k8s_deployment_get_info",
			"arguments": map[string]any{
				"cluster_id":      "default",
				"namespace":       "test-mcp",
				"deployment_name": "test-mcp",
			},
		},
	}

	json.NewEncoder(stdin).Encode(getDeploymentReq)
	scanner.Scan()
	fmt.Println(scanner.Text())

	stdin.Close()
	cmd.Wait()
}
