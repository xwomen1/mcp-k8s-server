# Kubernetes MCP Server

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)
![MCP](https://img.shields.io/badge/MCP-Protocol-blueviolet)

The Most Comprehensive Kubernetes MCP Server – Manage your entire Kubernetes ecosystem through AI assistants like Claude, ChatGPT, and more! 

---
## 📺 Live Demo
![K8s Assistant in action]
https://github.com/user-attachments/assets/56846b94-7ab7-4a81-9771-9cc67b200c3d 
*AI automatically validating YAML, deploying to 'test' namespace, and opening a tunnel.*

✨ **Comprehensive Coverage (65+ Tools!)**

Our project provides a **complete set of Kubernetes management tools**, including advanced workloads, configuration management, observability, and security.

🎯 **Production-Grade Features**
* **Complete Workload Management**: Deployments, StatefulSets, DaemonSets, Jobs, CronJobs
* **Advanced Scheduling**: Node taints management, resource quotas, limit ranges 
* **Security & RBAC**: Webhook configurations, ClusterRole listings, Secrets management
* **Observability**: Event tracking, pod logs, HPA metrics, node resource utilization
* **Multi-Cluster Support**: Register and manage multiple Kubernetes clusters

---

## 📋 Features Overview

### 🔧 Core Workload Operations
* **Pods**: List, get logs, execute commands
* **Deployments**: Scale, get info, restart
* **StatefulSets**: Scale, restart, delete
* **DaemonSets**: Restart, get pods, delete
* **Jobs & CronJobs**: Create, delete, trigger, suspend, get logs

### ⚙️ Configuration Management
* **ConfigMaps**: Create, read, update, delete
* **Secrets**: Secure management (keys only display)
* **ResourceQuotas**: List and get detailed quotas
* **LimitRanges**: Enforce resource limits per namespace

### 🌐 Networking
* **Services**: List, get details, delete
* **Ingress**: Full ingress controller management

### 📊 Monitoring & Debugging
* **Events**: List with filtering by involved object
* **Logs**: Real-time pod and job logs
* **Metrics**: Node resource utilization
* **HPA**: Autoscaling status and metrics

### 🛡️ Security & RBAC
* **Webhooks**: Mutating and Validating webhook configurations 
* **ClusterRoles**: List all cluster-wide RBAC policies
* **Node Taints**: Advanced scheduling control

---

## 🚀 Getting Started

### Quick Installation

```bash
# Clone the repository
git clone [https://github.com/yourusername/k8s-mcp-server.git](https://github.com/yourusername/k8s-mcp-server.git)
cd k8s-mcp-server

# Build the server
go build -o k8s-mcp-server
```

### Run with Default Configuration

```bash
./k8s-mcp-server
```

### Configuration Example

This configuration is typically used in the client (AI assistant) to point to your server instance.

```json
{
  "servers": {
    "k8s": {
      "command": "path/to/k8s-mcp-server",
      "env": {
        "KUBECONFIG_PATH": "/path/to/kubeconfig"
      }
    }
  }
}
```

---

## 📖 Usage Examples

Ask your AI Assistant (e.g., Claude, ChatGPT) to manage your cluster:
* "List all deployments in the production namespace"
* "Scale my-api deployment to 5 replicas"
* "Show me recent events in the default namespace"
* "Get logs from the failing payment-service pod"
* "Create a new Job to run a database migration"
* "Check HPA status for the frontend service"

## 🤖 AI Assistant Integration

This server is compatible with any client supporting the Model Context Protocol (MCP):
* Claude Desktop
* Cursor AI
* Windsurf
* Any MCP-compatible client

## 🏆 Developer Benefits

| Feature | Description |
|---------|-------------|
| ✅ Developer Experience | Intuitive Tool Names: Consistent k8s_<resource>_<action> naming, Detailed Descriptions, Smart Defaults for namespace and other parameters, Comprehensive error handling. |
| ✅ Production Ready | Multi-Cluster: Manage dev, staging, and production clusters; Security First: Never expose secret values, only metadata; Audit Trail: Event logging and monitoring tools; Resource Control: Quotas and limits. |
| ✅ Extensible Architecture | Easy to add new tools using the defined Domain-Use Case-Delivery structure. |

## 🔄 Roadmap

| Version | Features |
|---------|----------|
| v1.1 | CRD support, Helm management tools, Network Policies, Pod Disruption Budgets, Vertical Pod Autoscaling (VPA) |
| v1.2 | Multi-tenant namespace management, Cost optimization recommendations, Security scanning integration, Backup & Restore operations, GitOps synchronization |

## 🤝 Contributing

We welcome contributions! Here's how you can help:
* **Add New Tools**: See our guide for adding Kubernetes resource types
* **Improve Documentation**: Help make our README better
* **Report Issues**: Found a bug? Let us know
* **Feature Requests**: Suggest new tools

Check out CONTRIBUTING.md for detailed developer guidelines.

## 📚 Learn More

* [Model Context Protocol](https://modelcontextprotocol.io)
* [Kubernetes API Documentation](https://kubernetes.io/docs/reference)
* [Go MCP SDK](https://github.com/mark3labs/mcp-go)

## 🛡️ License

Apache 2.0 License - see [LICENSE](LICENSE) file for details.

## ⭐ Show Your Support

If this project helps you manage Kubernetes more effectively, please give it a star! It helps others discover the tool and motivates continued development.

Join the revolution in Kubernetes management through AI!
