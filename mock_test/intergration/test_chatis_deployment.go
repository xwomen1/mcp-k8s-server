package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"testing"
)

func TestMainFlowDeploymentInfo(t *testing.T) {

	// Fake stdout: server giả trả 3 response JSONRPC
	fakeStdout := bytes.NewBuffer(nil)

	fakeStdout.WriteString(`{"jsonrpc":"2.0","id":1,"result":"initialized"}` + "\n")
	fakeStdout.WriteString(`{"jsonrpc":"2.0","id":2,"result":"cluster registered"}` + "\n")
	fakeStdout.WriteString(`{"jsonrpc":"2.0","id":3,"result":{"name":"chatis","replicas":3}}` + "\n")

	fakeCmd := &FakeCmd{
		Stdin:  bytes.NewBuffer(nil),
		Stdout: fakeStdout,
	}

	stdin, _ := fakeCmd.StdinPipe()
	stdout, _ := fakeCmd.StdoutPipe()

	fakeCmd.Start()

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 2*1024*1024), 2*1024*1024)

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

	if !scanner.Scan() {
		t.Fatalf("missing initialize response")
	}
	resp1 := scanner.Text()
	if !bytes.Contains([]byte(resp1), []byte("initialized")) {
		t.Errorf("unexpected initialize response: %s", resp1)
	}

	//--------------------------------------------------------

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

	if !scanner.Scan() {
		t.Fatalf("missing register response")
	}
	resp2 := scanner.Text()
	if !bytes.Contains([]byte(resp2), []byte("cluster registered")) {
		t.Errorf("unexpected cluster register response: %s", resp2)
	}

	//--------------------------------------------------------

	getDeploymentReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]any{
			"name": "k8s_deployment_get_info",
			"arguments": map[string]any{
				"cluster_id":      "default",
				"namespace":       "chatis",
				"deployment_name": "chatis",
			},
		},
	}

	json.NewEncoder(stdin).Encode(getDeploymentReq)

	if !scanner.Scan() {
		t.Fatalf("missing deployment info response")
	}
	resp3 := scanner.Text()

	// EXPECTED: response contains deployment "name" and "replicas"
	if !bytes.Contains([]byte(resp3), []byte(`"name":"chatis"`)) {
		t.Errorf("deployment info missing name: %s", resp3)
	}
	if !bytes.Contains([]byte(resp3), []byte(`"replicas":3`)) {
		t.Errorf("deployment info missing replicas: %s", resp3)
	}

	stdin.Close()
	fakeCmd.Wait()
}
