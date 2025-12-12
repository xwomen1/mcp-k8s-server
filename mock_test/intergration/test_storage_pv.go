package main

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"testing"
)

type JSONRPCResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      int       `json:"id"`
	Result  any       `json:"result"`
	Error   *RPCError `json:"error"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func TestK8sMCPServer(t *testing.T) {

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
		t.Fatalf("cannot start server: %v", err)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 2*1024*1024), 2*1024*1024)

	// ------------------------------
	// 1. Test Initialize
	// ------------------------------
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
			t.Fatalf("no response from server")
		}

		var resp JSONRPCResponse
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		if resp.Error != nil {
			t.Fatalf("initialize failed: %v", resp.Error.Message)
		}
	})

	// ------------------------------
	// 2. Test cluster register
	// ------------------------------
	t.Run("RegisterCluster", func(t *testing.T) {
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
			t.Fatalf("missing server response")
		}

		var resp JSONRPCResponse
		json.Unmarshal(scanner.Bytes(), &resp)

		if resp.Error != nil {
			t.Fatalf("cluster register failed: %s", resp.Error.Message)
		}
	})

	// ------------------------------
	// 3. Test list PersistentVolumes
	// ------------------------------
	t.Run("ListPV", func(t *testing.T) {

		req := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "k8s_persistentvolume_list",
				"arguments": map[string]any{
					"cluster_id": "default",
				},
			},
		}

		json.NewEncoder(stdin).Encode(req)
		if !scanner.Scan() {
			t.Fatalf("missing PV response")
		}

		var resp JSONRPCResponse
		json.Unmarshal(scanner.Bytes(), &resp)

		if resp.Error != nil {
			t.Fatalf("list PV failed: %s", resp.Error.Message)
		}

		// Optional: kiểm tra result có đúng dạng mong đợi
		if resp.Result == nil {
			t.Fatalf("PV list result is nil")
		}
	})

	stdin.Close()
	cmd.Wait()
}
