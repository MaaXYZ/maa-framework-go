package maa

import (
	"github.com/MaaXYZ/maa-framework-go/v2/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v2/internal/maa"
)

type AgentClient struct {
	handle uintptr
}

// NewAgentClient creates and initializes a new Agent client instance
func NewAgentClient() *AgentClient {
	handle := maa.MaaAgentClientCreate()
	if handle == 0 {
		return nil
	}

	return &AgentClient{
		handle: handle,
	}
}

// Destroy releases underlying resources
func (ac *AgentClient) Destroy() {
	maa.MaaAgentClientDestroy(ac.handle)
}

// BindResource binds a resource object to the client
func (ac *AgentClient) BindResource(res *Resource) bool {
	return maa.MaaAgentClientBindResource(ac.handle, res.handle)
}

// CreateSocket creates a socket connection with specified identifier
func (ac *AgentClient) CreateSocket(identifier string) bool {
	identifierStrBuf := buffer.NewStringBuffer()
	defer identifierStrBuf.Destroy()
	identifierStrBuf.Set(identifier)
	return maa.MaaAgentClientCreateSocket(ac.handle, identifierStrBuf.Handle())
}

// Connect attempts to establish connection with Agent service
func (ac *AgentClient) Connect() bool {
	return maa.MaaAgentClientConnect(ac.handle)
}

// Disconnect actively terminates current connection
func (ac *AgentClient) Disconnect() bool {
	return maa.MaaAgentClientDisconnect(ac.handle)
}
