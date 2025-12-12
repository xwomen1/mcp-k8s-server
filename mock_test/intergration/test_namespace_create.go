package main

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"testing"
)

func TestNamespaceCreate(t *testing.T) {

	// -------------------------------------------------
	// Start MCP server process
	// -------------------------------------------------
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

	// =================================================
	// 1. Initialize
	// =================================================
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
			t.Fatalf("no response for initialize")
		}

		var resp JSONRPCResponse
		json.Unmarshal(scanner.Bytes(), &resp)

		if resp.Error != nil {
			t.Fatalf("initialize error: %v", resp.Error.Message)
		}
	})

	// =================================================
	// 2. Register cluster
	// =================================================
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
			t.Fatalf("no response for cluster_register")
		}

		var resp JSONRPCResponse
		json.Unmarshal(scanner.Bytes(), &resp)

		if resp.Error != nil {
			t.Fatalf("cluster_register error: %v", resp.Error.Message)
		}
	})

	// =================================================
	// 3. Create namespace
	// =================================================
	t.Run("NamespaceCreate", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "k8s_namespace_create",
				"arguments": map[string]any{
					"cluster_id": "default",
					"namespace":  "test-mcp",
				},
			},
		}

		json.NewEncoder(stdin).Encode(req)
		if !scanner.Scan() {
			t.Fatalf("no response for namespace_create")
		}

		var resp JSONRPCResponse
		json.Unmarshal(scanner.Bytes(), &resp)

		if resp.Error != nil {
			t.Fatalf("namespace_create error: %v", resp.Error.Message)
		}
	})

	// =================================================
	// 4. Get namespace
	// =================================================
	t.Run("NamespaceGetAfterCreate", func(t *testing.T) {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      4,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "k8s_namespace_get",
				"arguments": map[string]any{
					"cluster_id": "default",
					"namespace":  "test-mcp",
				},
			},
		}

		json.NewEncoder(stdin).Encode(req)
		if !scanner.Scan() {
			t.Fatalf("no response for namespace_get")
		}

		var resp JSONRPCResponse
		json.Unmarshal(scanner.Bytes(), &resp)

		if resp.Error != nil {
			t.Fatalf("expected namespace to exist, but got error: %v", resp.Error.Message)
		}
	})

	stdin.Close()
	cmd.Wait()
}
