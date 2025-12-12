package main

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"testing"
)

func TestMCPServerFlow(t *testing.T) {
	// Start external MCP server process
	cmd := exec.Command("./cmd/server/mcp-k8s-server.exe")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to get stdin: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start MCP server: %v", err)
	}

	scanner := bufio.NewScanner(stdout)

	// ---- Test 1: Initialize ----
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

	if err := json.NewEncoder(stdin).Encode(initReq); err != nil {
		t.Fatalf("Failed to send init request: %v", err)
	}

	if !scanner.Scan() {
		t.Fatalf("No init response received")
	}

	initResp := scanner.Text()
	t.Logf("Init response: %s", initResp)

	// Optional: basic validate
	if len(initResp) == 0 {
		t.Errorf("Init response is empty")
	}

	// ---- Test 2: register cluster ----
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

	if err := json.NewEncoder(stdin).Encode(registerReq); err != nil {
		t.Fatalf("Failed to send register request: %v", err)
	}

	if !scanner.Scan() {
		t.Fatalf("No register response received")
	}

	registerResp := scanner.Text()
	t.Logf("Register response: %s", registerResp)

	// Optional assertion
	if len(registerResp) == 0 {
		t.Errorf("Register response is empty")
	}

	// close
	stdin.Close()
	cmd.Wait()
}
