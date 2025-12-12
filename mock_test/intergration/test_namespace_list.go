package main

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"testing"
)

func TestMCPNamespaceList(t *testing.T) {

	// Run server executable
	cmd := exec.Command("./cmd/server/mcp-k8s-server.exe")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin error: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout error: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("start error: %v", err)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 2*1024*1024), 2*1024*1024)

	// ---------------------------
	// 1. initialize
	// ---------------------------
	t.Run("Initialize", func(t *testing.T) {

		req := JSONRPCRequest{
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

		json.NewEncoder(stdin).Encode(req)

		if !scanner.Scan() {
			t.Fatalf("server sent no response")
		}

		var resp JSONRPCResponse
		json.Unmarshal(scanner.Bytes(), &resp)

		if resp.Error != nil {
			t.Fatalf("initialize error: %v", resp.Error.Message)
		}
	})

	// ---------------------------
	// 2. k8s_cluster_register
	// ---------------------------
	t.Run("ClusterRegister", func(t *testing.T) {

		req := JSONRPCRequest{
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

		json.NewEncoder(stdin).Encode(req)

		if !scanner.Scan() {
			t.Fatalf("server sent no response for register")
		}

		var resp JSONRPCResponse
		json.Unmarshal(scanner.Bytes(), &resp)

		if resp.Error != nil {
			t.Fatalf("cluster_register error: %v", resp.Error.Message)
		}
	})

	// ---------------------------
	// 3. k8s_namespace_list
	// ---------------------------
	t.Run("NamespaceList", func(t *testing.T) {

		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "k8s_namespace_list",
				"arguments": map[string]any{
					"cluster_id": "default",
				},
			},
		}

		json.NewEncoder(stdin).Encode(req)

		if !scanner.Scan() {
			t.Fatalf("server sent no response for namespace_list")
		}

		var resp JSONRPCResponse
		json.Unmarshal(scanner.Bytes(), &resp)

		if resp.Error != nil {
			t.Fatalf("namespace_list error: %v", resp.Error.Message)
		}

		if resp.Result == nil {
			t.Fatalf("namespace_list result is nil")
		}
	})

	// shutdown
	stdin.Close()
	cmd.Wait()
}
