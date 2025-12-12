package mcpserver

type Server struct{}

type Config struct {
	Name    string
	Version string
}

func NewServer(cfg Config) (*Server, error) {
	return &Server{}, nil
}

func (s *Server) SetToolsHandler(handler interface{}) {}
