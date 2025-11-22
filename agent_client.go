package maa

import (
	"time"

	"github.com/MaaXYZ/maa-framework-go/v2/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v2/internal/maa"
)

type AgentClient struct {
	handle uintptr
}

// NewAgentClient creates and initializes a new Agent client instance
func NewAgentClient(identifier string) *AgentClient {
	identifierStrBuf := buffer.NewStringBuffer()
	defer identifierStrBuf.Destroy()
	identifierStrBuf.Set(identifier)

	handle := maa.MaaAgentClientCreateV2(identifierStrBuf.Handle())
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

// Identifier returns the identifier of the current agent client
func (ac *AgentClient) Identifier() (string, bool) {
	buf := buffer.NewStringBuffer()
	defer buf.Destroy()
	ok := maa.MaaAgentClientIdentifier(ac.handle, buf.Handle())
	return buf.Get(), ok
}

// BindResource binds a resource object to the current client
func (ac *AgentClient) BindResource(res *Resource) bool {
	return maa.MaaAgentClientBindResource(ac.handle, res.handle)
}

func (ac *AgentClient) RegisterResourceSink(res *Resource) bool {
	return maa.MaaAgentClientRegisterResourceSink(ac.handle, res.handle)
}

func (ac *AgentClient) RegisterControllerSink(ctrl Controller) bool {
	return maa.MaaAgentClientRegisterControllerSink(ac.handle, ctrl.handle)
}

func (ac *AgentClient) RegisterTaskerSink(tasker Tasker) bool {
	return maa.MaaAgentClientRegisterTaskerSink(ac.handle, tasker.handle)
}

// Connect attempts to establish connection with agent service
func (ac *AgentClient) Connect() bool {
	return maa.MaaAgentClientConnect(ac.handle)
}

// Disconnect actively terminates current connection
func (ac *AgentClient) Disconnect() bool {
	return maa.MaaAgentClientDisconnect(ac.handle)
}

// Connected checks if the current agent client is in a connected state
func (ac *AgentClient) Connected() bool {
	return maa.MaaAgentClientConnected(ac.handle)
}

// Alive checks if the current agent client is in an alive state
func (ac *AgentClient) Alive() bool {
	return maa.MaaAgentClientAlive(ac.handle)
}

// SetTimeout sets the timeout duration for the current agent client
func (ac *AgentClient) SetTimeout(duration time.Duration) bool {
	if duration < 0 {
		return false
	}

	milliseconds := duration.Milliseconds()

	return maa.MaaAgentClientSetTimeout(ac.handle, milliseconds)
}
