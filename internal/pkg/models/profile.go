package models

// McpServer defines an MCP server configuration.
type McpServer struct {
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

// Profile holds configuration for a Claude Code profile.
type Profile struct {
	Name       string            `json:"name"`
	EnvVars    map[string]string `json:"env_vars,omitempty"`
	McpServers []McpServer       `json:"mcp_servers,omitempty"`
	Model      string            `json:"model,omitempty"`
}
