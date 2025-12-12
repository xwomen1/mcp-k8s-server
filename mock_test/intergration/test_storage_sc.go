package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"testing"
)

func TestStorageClassListFlow(t *testing.T) {

	fakeStdout := bytes.NewBuffer(nil)

	fakeStdout.WriteString(`{"jsonrpc":"2.0","id":1,"result":"initialized"}` + "\n")
	fakeStdout.WriteString(`{"jsonrpc":"2.0","id":2,"result":"cluster registered"}` + "\n")
	fakeStdout.WriteString(`{"jsonrpc":"2.0","id":3,"result":{"storageClasses":["standard","gp2","ssd-fast"]}}` + "\n")

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
	scanner.Scan() // response id=1

	//  Register cluster

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
	scanner.Scan() // response id=2

	//  List StorageClasses

	listSCReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]any{
			"name": "k8s_storageclass_list",
			"arguments": map[string]any{
				"cluster_id": "default",
			},
		},
	}

	json.NewEncoder(stdin).Encode(listSCReq)
	scanner.Scan()

	// Response như server giả trả về
	result := scanner.Text()

	//------------------------------------------------
	// 4️⃣ Validate kết quả
	//------------------------------------------------
	if !bytes.Contains([]byte(result), []byte("gp2")) {
		t.Errorf("expected storageclass 'gp2' in result, got: %s", result)
	}

	if !bytes.Contains([]byte(result), []byte("ssd-fast")) {
		t.Errorf("expected storageclass 'ssd-fast' in result, got: %s", result)
	}

	stdin.Close()
	fakeCmd.Wait()
}
