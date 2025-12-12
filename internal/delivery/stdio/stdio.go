package stdio

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/your-org/mcp-k8s-server/internal/infrastructure"
	"github.com/your-org/mcp-k8s-server/internal/usecase"
)

type Handler struct {
	toolManager    *usecase.ToolManager
	clusterUseCase *usecase.ClusterUseCase
	podUseCase     *usecase.PodUseCase
	logger         infrastructure.Logger
}

func NewHandler(
	toolManager *usecase.ToolManager,
	clusterUseCase *usecase.ClusterUseCase,
	podUseCase *usecase.PodUseCase,
	logger infrastructure.Logger,
) *Handler {
	return &Handler{
		toolManager:    toolManager,
		clusterUseCase: clusterUseCase,
		podUseCase:     podUseCase,
		logger:         logger,
	}
}

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (h *Handler) Start(ctx context.Context) error {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var req Request
			if err := decoder.Decode(&req); err != nil {
				if err == io.EOF {
					return nil
				}
				h.logger.Error("Failed to decode request", "error", err)
				continue
			}

			resp := h.handleRequest(ctx, &req)
			if err := encoder.Encode(resp); err != nil {
				h.logger.Error("Failed to encode response", "error", err)
				continue
			}
		}
	}
}

func (h *Handler) handleRequest(ctx context.Context, req *Request) *Response {
	switch req.Method {
	case "tools/list":
		return h.handleToolsList(req)
	case "tools/call":
		return h.handleToolsCall(ctx, req)
	default:
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &Error{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

func (h *Handler) handleToolsList(req *Request) *Response {
	tools := h.toolManager.ListTools()
	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]interface{}{"tools": tools},
	}
}

func (h *Handler) handleToolsCall(ctx context.Context, req *Request) *Response {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &Error{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	result, err := h.toolManager.ExecuteTool(ctx, params.Name, params.Arguments)
	if err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &Error{
				Code:    -32603,
				Message: err.Error(),
			},
		}
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}
