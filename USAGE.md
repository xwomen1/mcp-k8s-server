# Hướng dẫn sử dụng MCP Kubernetes Server

## Sau khi chạy server (`go run .`), bạn có thể sử dụng theo các cách sau:

### Cách 1: Gửi JSON request trực tiếp qua stdin

1. **Giữ server đang chạy** trong terminal hiện tại

2. **Mở terminal mới** và gửi request:

```powershell
# Windows PowerShell - Test list tools
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | go run cmd/server/main.go
```

Hoặc **trong chính terminal đang chạy server**, bạn có thể gõ trực tiếp JSON và nhấn Enter:

```
{"jsonrpc":"2.0","id":1,"method":"tools/list"}
```

### Cách 2: Sử dụng file JSON

1. Tạo file `request.json`:
```json
{"jsonrpc":"2.0","id":1,"method":"tools/list"}
```

2. Pipe vào server:
```powershell
Get-Content request.json | go run cmd/server/main.go
```

### Cách 3: Test thủ công (Recommended cho lần đầu)

1. **Khởi động server** trong terminal 1:
```powershell
cd cmd/server
go run .
```

2. **Trong cùng terminal đó**, sau khi thấy log "Registered tool...", gõ JSON request:
```
{"jsonrpc":"2.0","id":1,"method":"tools/list"}
```
Nhấn Enter

3. Server sẽ trả về response dạng JSON trong cùng terminal.

### Các request mẫu:

#### 1. Liệt kê tất cả tools:
```json
{"jsonrpc":"2.0","id":1,"method":"tools/list"}
```

#### 2. Đăng ký Kubernetes cluster:
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "k8s_cluster_register",
    "arguments": {
      "cluster_id": "my-cluster",
      "kubeconfig_path": "C:\\Users\\linh.do\\.kube\\config"
    }
  }
}
```

#### 3. Lấy logs từ pod:
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "k8s_pod_get_logs",
    "arguments": {
      "cluster_id": "my-cluster",
      "namespace": "default",
      "pod_name": "my-pod",
      "tail_lines": 100
    }
  }
}
```

#### 4. Scale deployment:
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "k8s_deployment_scale",
    "arguments": {
      "cluster_id": "my-cluster",
      "namespace": "default",
      "deployment_name": "my-deployment",
      "replicas": 3
    }
  }
}
```

## Lưu ý:

- Server đọc từ **stdin** và ghi ra **stdout**
- Mỗi request phải là một dòng JSON hợp lệ
- Nhấn Enter sau mỗi request để server xử lý
- Server sẽ đợi request tiếp theo sau khi xử lý xong

## Tích hợp với Claude Desktop:

Để dùng với Claude Desktop, cần build binary và cấu hình trong Claude Desktop settings. Xem README.md để biết chi tiết.

