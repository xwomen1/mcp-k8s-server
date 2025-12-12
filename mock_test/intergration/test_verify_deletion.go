package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"testing"
)

func jsonContains(jsonStr, substring string) bool {
	return bytes.Contains([]byte(jsonStr), []byte(substring))
}

func TestNamespaceDeletionFlow(t *testing.T) {

	// Fake stdout simulating MCP server response:
	// 1. initialize
	// 2. register
	// 3. namespace list (server returns NO "test-mcp")
	fakeStdout := bytes.NewBuffer(nil)
	fakeStdout.WriteString(`{"jsonrpc":"2.0","id":1,"result":"initialized"}` + "\n")
	fakeStdout.WriteString(`{"jsonrpc":"2.0","id":2,"result":"cluster registered"}` + "\n")
	fakeStdout.WriteString(`{"jsonrpc":"2.0","id":3,"result":{"namespaces":["default","kube-system","prod"]}}` + "\n")

	fakeCmd := &FakeCmd{
		Stdin:  bytes.NewBuffer(nil),
		Stdout: fakeStdout,
	}

	stdin, _ := fakeCmd.StdinPipe()
	stdout, _ := fakeCmd.StdoutPipe()
	fakeCmd.Start()

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 2*1024*1024), 2*1024*1024)

	//  Initialize
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
	scanner.Scan() // get server response

	//  Cluster register

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

	//  List namespaces
	listReq := JSONRPCRequest{
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
	json.NewEncoder(stdin).Encode(listReq)
	scanner.Scan()

	result := scanner.Text()

	if jsonContains(result, "test-mcp") {
		t.Errorf("expected namespace 'test-mcp' to be deleted but still exists: %s", result)
	}
}
