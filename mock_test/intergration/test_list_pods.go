package main

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"testing"
)

func TestMCPServerFullFlow(t *testing.T) {
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
		t.Fatalf("cannot start server exe: %v", err)
	}

	scanner := bufio.NewScanner(stdout)

	// 1) initialize

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
		t.Fatalf("no initialize response received")
	}
	t.Logf("Init response: %s", scanner.Text())

	// 2) register cluster

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
		t.Fatalf("no register response received")
	}
	t.Logf("Register response: %s", scanner.Text())

	// 3) list pods

	listPodsReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]any{
			"name": "k8s_pod_list",
			"arguments": map[string]any{
				"cluster_id": "default",
				"namespace":  "chatis",
			},
		},
	}

	if err := json.NewEncoder(stdin).Encode(listPodsReq); err != nil {
		t.Fatalf("failed to send list pods request: %v", err)
	}

	if !scanner.Scan() {
		t.Fatalf("no list pods response received")
	}

	listResp := scanner.Text()
	t.Logf("List pods response: %s", listResp)

	if len(listResp) == 0 {
		t.Errorf("list pods response is empty")
	}

	stdin.Close()
	_ = cmd.Wait()
}
