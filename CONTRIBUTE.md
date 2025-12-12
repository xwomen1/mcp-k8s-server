# 🤝 Contributing to MCP Kubernetes Server

Thank you for your interest in contributing to the Model Context Protocol (MCP) based Kubernetes Server! Your efforts help make managing Kubernetes via AI models safer, smarter, and more efficient.

This document outlines the project's structure, coding standards, and the required process for contributing new features or bug fixes.

## 1. 🏗️ Project Architecture and Directory Structure

This project adopts a **Layered Architecture** pattern, similar to Clean Architecture or Domain-Driven Design (DDD). This structure ensures separation of concerns, testability, and maintainability. 

The core logic resides within the `internal` directory:

| Directory | Layer | Core Responsibility | Key Packages Used |
| :--- | :--- | :--- | :--- |
| `internal/domain` | **Domain** | **Source of Truth (Data Models).** Defines all standardized Go structs (`Node`, `HPA`, `Deployment`, etc.) that represent Kubernetes resources *within* the application. **No business logic or external calls.** | Standard Go types |
| `internal/usecase` | **Use Case** | **Business Logic/Orchestration.** Contains functions that implement specific business workflows (e.g., `ListPods`, `DeleteHPA`). Interacts with the Kubernetes client-go SDK and transforms raw API objects into `domain` structs. | `k8s.io/client-go`, `internal/domain` |
| `internal/delivery` | **Delivery** | **Interface/Adapter.** Handles external communication protocols. This is where input is parsed and output is formatted. | `internal/usecase`, `github.com/modelcontextprotocol/go-sdk/mcp` |
| `internal/delivery/mcp` | **MCP Handler** | **Primary Tooling Interface.** Defines and registers MCP Tools, parses `map[string]any` input arguments, calls the Use Case layer, and formats the output into `mcp.CallToolResult`. | `internal/usecase` |
| `internal/infrastructure` | **Infrastructure** | **External Resources.** Holds components like the Cluster Manager, which handles client creation (`k8s.io/client-go`) and connection logic. | `k8s.io/client-go` |

## 2. 📝 Contribution Workflow (The 3-Step Flow)

To add a new feature (e.g., managing Deployments), you must follow this sequential flow across the architecture:

### Step A: Define the Domain Model (`internal/domain`)

1.  **Create/Update Struct:** Define the canonical Go struct for the resource (e.g., `Deployment`, `Service`).
2.  **Fields:** Include only the necessary fields for reporting and modification. Use accurate JSON tags.

### Step B: Implement the Use Case (`internal/usecase`)

1.  **New File:** Create a file (e.g., `deployment_uc.go`) within the `internal/usecase` package.
2.  **Logic Functions:** Implement the CRUD functions, ensuring all data passed is a `domain` struct.
    * Example: `ListDeployments(ctx context.Context, clusterID string) ([]domain.Deployment, error)`
3.  **Transformation:** Implement a helper function (e.g., `convertK8sDeploymentToDomain`) to convert the raw `appsv1.Deployment` API object into your cleaner `domain.Deployment` struct.

### Step C: Implement the MCP Handler and Register Tool (`internal/delivery/mcp`)

1.  **Handler Function:** Create the handler function:
    ```go
    func (m *MCPServer) handleListDeployments(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
        // 1. Extract raw input from args map[string]any
        // 2. Call m.k8sUC.ListDeployments(...)
        // 3. Format summary text
        // 4. Return *mcp.CallToolResult (text summary + full JSON data)
    }
    ```
2.  **Tool Registration:** Update the `setupTools()` function within the MCP Server to register the new tool.
    * Define the **Tool Name** (e.g., `k8s_deployment_list`).
    * Define a clear **Description**.
    * Provide the precise **InputSchema** (JSON Schema).

## 3. ⚠️ Coding Standards and Guidelines

### Error Handling

* **Error Wrapping:** Always wrap errors returning from external SDKs (like `client-go`) using `fmt.Errorf("failed to list pods: %w", err)`. This preserves the error chain for debugging.
* **Final Error Result:** Handlers must use the provided `errorResult(err)` helper to format errors for the MCP output.

### Data Types

* **Domain Structs:** Use `domain` structs exclusively when passing data between the **Use Case** and **Delivery** layers. Avoid passing raw `map[string]any` or complex API structs outside of the **Use Case** layer.
* **Input (`args`):** Inputs from the MCP handler (`args map[string]any`) must be explicitly converted and type-asserted within the Handler before calling the Use Case.

### Kubernetes Logic

* **Caching:** When implementing `List` operations, if read-heavy operations are expected, consider implementing a local cache or shared informer logic within the `infrastructure` layer.

### Testing

* **Unit Tests:** Place unit tests within `internal/usecase/...` to test the core transformation and business logic.
* **Integration Tests:** Place integration tests (which require a running K8s cluster or Mock API) within the dedicated `mock_test/integration` directory.

## 4. 🚀 Getting Started

1.  **Fork** the repository.
2.  **Clone** your forked repository.
3.  **Create a New Branch:** `git checkout -b feature/add-deployment-crud`
4.  **Make Changes:** Follow the 3-Step Workflow above.
5.  **Test:** Run all tests locally (`go test ./...`).
6.  **Commit:** Write clear, concise commit messages.
    * *Example:* `feat: Add CRUD for Deployment resource`
7.  **Push:** Push your changes to your fork.
8.  **Pull Request (PR):** Submit a Pull Request targeting the main branch of this repository. Be sure to link any relevant issues!

We look forward to your contributions!