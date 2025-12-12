package main

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"testing"
)

func TestGetPodLogs(t *testing.T) {

	cmd := exec.Command("./cmd/server/mcp-k8s-server.exe")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("cannot open stdin: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("cannot open stdout: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("cannot start MCP server: %v", err)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 2*1024*1024), 2*1024*1024)

	//  initialize

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
		t.Fatalf("failed to send initialize: %v", err)
	}

	if !scanner.Scan() {
		t.Fatalf("no initialize response")
	}
	t.Logf("Init response: %s", scanner.Text())

	//  register cluster

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
		t.Fatalf("failed to send register request: %v", err)
	}

	if !scanner.Scan() {
		t.Fatalf("no cluster register response")
	}
	t.Logf("Register response: %s", scanner.Text())

	//  get pod logs

	getLogsReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]any{
			"name": "k8s_pod_get_logs",
			"arguments": map[string]any{
				"cluster_id": "default",
				"namespace":  "test-mcp",
				"pod_name":   "test-mcp-8548bcfcf8-6gj49",
				"tail_lines": 50,
			},
		},
	}

	if err := json.NewEncoder(stdin).Encode(getLogsReq); err != nil {
		t.Fatalf("failed to send log request: %v", err)
	}

	if !scanner.Scan() {
		t.Fatalf("no pod logs response")
	}

	resp := scanner.Text()
	t.Logf("Pod logs response: %s", resp)

	//----------------------------------------------------------------------
	// BASIC ASSERT
	//----------------------------------------------------------------------
	if len(resp) == 0 {
		t.Errorf("expected logs, got empty response")
	}

	stdin.Close()
	_ = cmd.Wait()
}
