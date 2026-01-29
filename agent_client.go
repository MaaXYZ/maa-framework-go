package maa

import (
	"time"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v3/internal/native"
)

type AgentClient struct {
	handle uintptr
}

// NewAgentClient creates an Agent client instance
// If identifier is empty, it will be automatically generated
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

// NewAgentClientTcp creates an Agent client instance using TCP connection
func NewAgentClientTcp(port uint16) *AgentClient {
	handle := native.MaaAgentClientCreateTcp(port)
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

// BindResource links the Agent client to the specified resource
func (ac *AgentClient) BindResource(res *Resource) bool {
	return native.MaaAgentClientBindResource(ac.handle, res.handle)
}

// RegisterResourceSink registers resource events to resource
func (ac *AgentClient) RegisterResourceSink(res *Resource) bool {
	return native.MaaAgentClientRegisterResourceSink(ac.handle, res.handle)
}

// RegisterControllerSink registers controller events to controller
func (ac *AgentClient) RegisterControllerSink(ctrl Controller) bool {
	return native.MaaAgentClientRegisterControllerSink(ac.handle, ctrl.handle)
}

// RegisterTaskerSink registers tasker events to tasker
func (ac *AgentClient) RegisterTaskerSink(tasker Tasker) bool {
	return native.MaaAgentClientRegisterTaskerSink(ac.handle, tasker.handle)
}

// Connect connects to the Agent server
func (ac *AgentClient) Connect() bool {
	return native.MaaAgentClientConnect(ac.handle)
}

// Disconnect disconnects from the Agent server
func (ac *AgentClient) Disconnect() bool {
	return native.MaaAgentClientDisconnect(ac.handle)
}

// Connected checks if the client is connected to the Agent server
func (ac *AgentClient) Connected() bool {
	return native.MaaAgentClientConnected(ac.handle)
}

// Alive checks if the Agent server is still responsive
func (ac *AgentClient) Alive() bool {
	return native.MaaAgentClientAlive(ac.handle)
}

// SetTimeout sets the timeout duration for the Agent server
func (ac *AgentClient) SetTimeout(duration time.Duration) bool {
	if duration < 0 {
		return false
	}

	milliseconds := duration.Milliseconds()

	return native.MaaAgentClientSetTimeout(ac.handle, milliseconds)
}

// GetCustomRecognitionList returns the custom recognition name list of the agent client
func (ac *AgentClient) GetCustomRecognitionList() ([]string, bool) {
	buf := buffer.NewStringListBuffer()
	defer buf.Destroy()

	ok := native.MaaAgentClientGetCustomRecognitionList(ac.handle, buf.Handle())
	return buf.GetAll(), ok
}

// GetCustomActionList returns the custom action name list of the agent client
func (ac *AgentClient) GetCustomActionList() ([]string, bool) {
	buf := buffer.NewStringListBuffer()
	defer buf.Destroy()

	ok := native.MaaAgentClientGetCustomActionList(ac.handle, buf.Handle())
	return buf.GetAll(), ok
}
