package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

// Fake command executor
type FakeCmd struct {
	Stdin  *bytes.Buffer
	Stdout *bytes.Buffer
}

func (f *FakeCmd) StdinPipe() (io.WriteCloser, error) {
	return nopWriteCloser{f.Stdin}, nil
}
func (f *FakeCmd) StdoutPipe() (io.ReadCloser, error) {
	return io.NopCloser(f.Stdout), nil
}
func (f *FakeCmd) Start() error { return nil }
func (f *FakeCmd) Wait() error  { return nil }

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }

func TestMainFlow(t *testing.T) {

	fakeStdout := bytes.NewBuffer(nil)

	fakeStdout.WriteString(`{"jsonrpc":"2.0","id":1,"result":"initialized"}` + "\n")
	fakeStdout.WriteString(`{"jsonrpc":"2.0","id":2,"result":"cluster registered"}` + "\n")
	fakeStdout.WriteString(`{"jsonrpc":"2.0","id":3,"result":{"pods":["p1","p2"]}}` + "\n")

	fakeCmd := &FakeCmd{
		Stdin:  bytes.NewBuffer(nil),
		Stdout: fakeStdout,
	}

	stdin, _ := fakeCmd.StdinPipe()
	stdout, _ := fakeCmd.StdoutPipe()

	fakeCmd.Start()

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]string{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}
	json.NewEncoder(stdin).Encode(initReq)

	if !scanner.Scan() {
		t.Fatalf("no response for initialize")
	}
	resp1 := scanner.Text()
	if !bytes.Contains([]byte(resp1), []byte("initialized")) {
		t.Errorf("unexpected init response: %s", resp1)
	}

	registerReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "k8s_cluster_register",
			"arguments": map[string]interface{}{
				"cluster_id":      "default",
				"kubeconfig_path": "path to kubeconfig",
			},
		},
	}
	json.NewEncoder(stdin).Encode(registerReq)

	if !scanner.Scan() {
		t.Fatalf("no response for register")
	}
	resp2 := scanner.Text()
	if !bytes.Contains([]byte(resp2), []byte("cluster registered")) {
		t.Errorf("unexpected register response: %s", resp2)
	}

	listPodsReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "k8s_pod_list",
			"arguments": map[string]interface{}{
				"cluster_id": "default",
				"namespace":  "test-mcp",
			},
		},
	}
	json.NewEncoder(stdin).Encode(listPodsReq)

	if !scanner.Scan() {
		t.Fatalf("no response for pod list")
	}
	resp3 := scanner.Text()
	if !bytes.Contains([]byte(resp3), []byte("p1")) {
		t.Errorf("unexpected pods response: %s", resp3)
	}

	stdin.Close()
	fakeCmd.Wait()
}
