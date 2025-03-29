package maa

import (
	"github.com/MaaXYZ/maa-framework-go/v2/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v2/internal/maa"
)

type AgentClient struct {
	handle uintptr
}

func NewAgentClient() *AgentClient {
	handle := maa.MaaAgentClientCreate()
	if handle == 0 {
		return nil
	}

	return &AgentClient{
		handle: handle,
	}
}

func (ac *AgentClient) Destroy() {
	maa.MaaAgentClientDestroy(ac.handle)
}

func (ac *AgentClient) BindResource(res *Resource) bool {
	return maa.MaaAgentClientBindResource(ac.handle, res.handle)
}

func (ac *AgentClient) CreateSocket(identifier string) bool {
	identifierStrBuf := buffer.NewStringBuffer()
	defer identifierStrBuf.Destroy()
	identifierStrBuf.Set(identifier)
	return maa.MaaAgentClientCreateSocket(ac.handle, identifierStrBuf.Handle())
}

func (ac *AgentClient) Connect() bool {
	return maa.MaaAgentClientConnect(ac.handle)
}

func (ac *AgentClient) Disconnect() bool {
	return maa.MaaAgentClientDisconnect(ac.handle)
}
