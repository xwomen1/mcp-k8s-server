package main

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"testing"
)

// ----------------- Struct JSONRPCRequest -----------------
type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

func RunListNamespace(serverPath string, namespace string) (string, error) {
	cmd := exec.Command(serverPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 2*1024*1024), 2*1024*1024)

	//---------------------------------------------------------
	// 1) Initialize
	//---------------------------------------------------------
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

	//---------------------------------------------------------
	// 2) Register cluster
	//---------------------------------------------------------
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

	//---------------------------------------------------------
	// 3) Get namespace details
	//---------------------------------------------------------
	getNamespaceReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]any{
			"name": "k8s_namespace_get",
			"arguments": map[string]any{
				"cluster_id": "default",
				"namespace":  namespace,
			},
		},
	}
	json.NewEncoder(stdin).Encode(getNamespaceReq)
	scanner.Scan()

	resp := scanner.Text()

	stdin.Close()
	cmd.Wait()

	return resp, nil
}

// ----------------- Fake MCP server code -----------------
const fakeServerCode = `
package main
import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)
func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var req map[string]any
		json.Unmarshal(scanner.Bytes(), &req)
		id := req["id"]
		// Trả về tên namespace
		result := map[string]string{}
		if req["method"] == "tools/call" && req["params"] != nil {
			p := req["params"].(map[string]any)
			if p["name"] == "k8s_namespace_get" {
				args := p["arguments"].(map[string]any)
				ns := args["namespace"].(string)
				result["namespace"] = ns
			}
		}
		resJSON, _ := json.Marshal(map[string]any{"jsonrpc":"2.0","id":id,"result":result})
		fmt.Println(string(resJSON))
	}
}
`

// ----------------- Unit Test -----------------
func TestListNamespace(t *testing.T) {
	// Build fake server
	tmpDir := t.TempDir()
	src := tmpDir + "/fake_server.go"
	exe := tmpDir + "/fake_server.exe"

	if err := os.WriteFile(src, []byte(fakeServerCode), 0644); err != nil {
		t.Fatalf("write fake server failed: %v", err)
	}

	cmd := exec.Command("go", "build", "-o", exe, src)
	if err := cmd.Run(); err != nil {
		t.Fatalf("build fake server failed: %v", err)
	}

	// Run test
	namespace := "test-mcp"
	resp, err := RunListNamespace(exe, namespace)
	if err != nil {
		t.Fatalf("RunListNamespace failed: %v", err)
	}

	expected := `{"jsonrpc":"2.0","id":3,"result":{"namespace":"test-mcp"}}`
	if resp != expected {
		t.Fatalf("unexpected response:\n got=%s\nwant=%s", resp, expected)
	}
}
