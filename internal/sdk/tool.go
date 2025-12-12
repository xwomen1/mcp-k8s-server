package sdk

type ToolCall struct {
	Name      string
	Arguments map[string]interface{}
}

type ToolResponse struct {
	Content []Content
}

type Content struct {
	Type string
	Text string
}

type ToolHandlerFunc func(call ToolCall) (*ToolResponse, error)
