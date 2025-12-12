package main

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
	"testing"
	"time"
)

func executeAndScan(t *testing.T, stdin io.Writer, scanner *bufio.Scanner, req JSONRPCRequest, logMsg string) string {
	if err := json.NewEncoder(stdin).Encode(req); err != nil {
		t.Fatalf("failed to send %s request: %v", logMsg, err)
	}

	ch := make(chan string)
	go func() {
		if scanner.Scan() {
			ch <- scanner.Text()
		} else {
			ch <- ""
		}
	}()

	select {
	case resp := <-ch:
		if resp == "" {
			t.Fatalf("no %s response received", logMsg)
		}
		t.Logf("%s response: %s", logMsg, resp)

		var rpcResp JSONRPCResponse
		if err := json.Unmarshal([]byte(resp), &rpcResp); err != nil {
			t.Errorf("failed to unmarshal JSONRPC response for %s: %v", logMsg, err)
		} else if rpcResp.Error != nil {
			t.Errorf("%s returned error: %v", logMsg, rpcResp.Error)
		}

		return resp

	case <-time.After(5 * time.Second):
		t.Fatalf("timeout waiting for %s response", logMsg)
		return ""
	}
}

func TestMCPServerK8sFullFlow(t *testing.T) {

	cmd := exec.Command("./cmd/server/mcp-k8s-server.exe")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("cannot open stdin: %v", err)
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("cannot open stdout: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("cannot start server exe: %v", err)
	}
	defer cmd.Wait()

	scanner := bufio.NewScanner(stdout)

	const clusterID = "default"
	const kubeconfigPath = "/path/to/your/valid/kubeconfig"

	t.Run("InitializeServer", func(t *testing.T) {
		initReq := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "initialize",
			Params: map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]any{},
				"clientInfo":      map[string]any{"name": "test-client", "version": "1.0.0"},
			},
		}
		executeAndScan(t, stdin, scanner, initReq, "initialize")
	})

	t.Run("RegisterCluster", func(t *testing.T) {
		registerReq := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "k8s_cluster_register",
				"arguments": map[string]any{
					"cluster_id":      clusterID,
					"kubeconfig_path": kubeconfigPath,
				},
			},
		}
		executeAndScan(t, stdin, scanner, registerReq, "register cluster")
	})

	t.Run("ListPods", func(t *testing.T) {
		listPodsReq := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "k8s_pod_list",
				"arguments": map[string]any{
					"cluster_id": clusterID,
					"namespace":  "default",
				},
			},
		}
		executeAndScan(t, stdin, scanner, listPodsReq, "list pods")
	})

	// 4) k8s_node_list
	var firstNodeName string
	t.Run("ListNode", func(t *testing.T) {
		listNodesReq := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      4,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "k8s_node_list",
				"arguments": map[string]any{
					"cluster_id": clusterID,
				},
			},
		}

		resp := executeAndScan(t, stdin, scanner, listNodesReq, "list nodes")

		var rpcResp JSONRPCResponse
		if err := json.Unmarshal([]byte(resp), &rpcResp); err == nil && rpcResp.Result != nil {
			resultMap, ok := rpcResp.Result.(map[string]any)
			if !ok {
				t.Fatalf("Result is not a map[string]any")
			}
			// Giả định rằng 'data' chứa kết quả trả về từ Use Case, trong đó có mảng 'nodes'
			if data, ok := resultMap["data"].(map[string]any); ok {
				if nodes, ok := data["nodes"].([]any); ok && len(nodes) > 0 {
					if node, ok := nodes[0].(map[string]any); ok {
						if name, ok := node["name"].(string); ok {
							firstNodeName = name
							t.Logf("Found first node: %s", firstNodeName)
						}
					}
				}
			}
		}

		if firstNodeName == "" {
			t.Logf("Could not extract first node name, skipping get metrics test.")
		}
	})

	if firstNodeName != "" {
		t.Run("GetNodeMetrics", func(t *testing.T) {
			getNodeMetricsReq := JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      5,
				Method:  "tools/call",
				Params: map[string]any{
					"name": "k8s_node_get_metrics",
					"arguments": map[string]any{
						"cluster_id": clusterID,
						"node_name":  firstNodeName,
					},
				},
			}

			resp := executeAndScan(t, stdin, scanner, getNodeMetricsReq, "get node metrics")

			if len(resp) < 100 {
				t.Errorf("GetNodeMetrics response seems too short.")
			}
			if !json.Valid([]byte(resp)) {
				t.Errorf("GetNodeMetrics response is not valid JSON.")
			}
		})
	} else {
		t.Logf("Skipping GetNodeMetrics test as no node name was found.")
	}

	t.Run("ShutdownServer", func(t *testing.T) {
		shutdownReq := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      6,
			Method:  "shutdown",
			Params:  map[string]any{},
		}
		executeAndScan(t, stdin, scanner, shutdownReq, "shutdown")
	})

	t.Log("Test completed. Closing stdin and waiting for process to exit.")

	stdin.Close()
}
