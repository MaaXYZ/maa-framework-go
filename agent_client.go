package maa

import (
	"time"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v3/internal/native"
)

type AgentClient struct {
	handle uintptr
}

// NewAgentClient creates and initializes a new Agent client instance
func NewAgentClient(identifier string) *AgentClient {
	identifierStrBuf := buffer.NewStringBuffer()
	defer identifierStrBuf.Destroy()
	identifierStrBuf.Set(identifier)

	handle := native.MaaAgentClientCreateV2(identifierStrBuf.Handle())
	if handle == 0 {
		return nil
	}

	return &AgentClient{
		handle: handle,
	}
}

// Destroy releases underlying resources
func (ac *AgentClient) Destroy() {
	native.MaaAgentClientDestroy(ac.handle)
}

// Identifier returns the identifier of the current agent client
func (ac *AgentClient) Identifier() (string, bool) {
	buf := buffer.NewStringBuffer()
	defer buf.Destroy()
	ok := native.MaaAgentClientIdentifier(ac.handle, buf.Handle())
	return buf.Get(), ok
}

// BindResource binds a resource object to the current client
func (ac *AgentClient) BindResource(res *Resource) bool {
	return native.MaaAgentClientBindResource(ac.handle, res.handle)
}

func (ac *AgentClient) RegisterResourceSink(res *Resource) bool {
	return native.MaaAgentClientRegisterResourceSink(ac.handle, res.handle)
}

func (ac *AgentClient) RegisterControllerSink(ctrl Controller) bool {
	return native.MaaAgentClientRegisterControllerSink(ac.handle, ctrl.handle)
}

func (ac *AgentClient) RegisterTaskerSink(tasker Tasker) bool {
	return native.MaaAgentClientRegisterTaskerSink(ac.handle, tasker.handle)
}

// Connect attempts to establish connection with agent service
func (ac *AgentClient) Connect() bool {
	return native.MaaAgentClientConnect(ac.handle)
}

// Disconnect actively terminates current connection
func (ac *AgentClient) Disconnect() bool {
	return native.MaaAgentClientDisconnect(ac.handle)
}

// Connected checks if the current agent client is in a connected state
func (ac *AgentClient) Connected() bool {
	return native.MaaAgentClientConnected(ac.handle)
}

// Alive checks if the current agent client is in an alive state
func (ac *AgentClient) Alive() bool {
	return native.MaaAgentClientAlive(ac.handle)
}

// SetTimeout sets the timeout duration for the current agent client
func (ac *AgentClient) SetTimeout(duration time.Duration) bool {
	if duration < 0 {
		return false
	}

	milliseconds := duration.Milliseconds()

	return native.MaaAgentClientSetTimeout(ac.handle, milliseconds)
}
