package maa

import (
	"errors"
	"time"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v3/internal/native"
)

// AgentClient is used to connect to AgentServer, delegating custom recognition and
// action execution to a separate process. This allows separating MaaFW core from
// custom logic into independent processes.
type AgentClient struct {
	handle uintptr
}

const (
	agentClientCreationModeIdentifier = "identifier"
	agentClientCreationModeTcp        = "tcp"
)

type agentClientConfig struct {
	identifier string
	tcpPort    uint16
	lastSet    string
}

// AgentClientOption configures how an Agent client is created.
type AgentClientOption func(*agentClientConfig)

// WithIdentifier sets the client identifier for creating an agent client.
// The identifier is used to identify this specific client instance.
// The identifier creation mode uses IPC, and will fall back to TCP on older
// Windows versions that do not support AF_UNIX (builds prior to 17063).
// If empty, an identifier will be automatically generated.
//
// Priority: This option takes precedence for creation mode if specified
// after WithTcpPort. If specified before WithTcpPort, WithTcpPort will
// override it.
func WithIdentifier(identifier string) AgentClientOption {
	return func(cfg *agentClientConfig) {
		cfg.identifier = identifier
		cfg.lastSet = agentClientCreationModeIdentifier
	}
}

// WithTcpPort sets the TCP port for creating a TCP-based agent client.
// The client will connect to the agent server at the specified port.
//
// Priority: This option takes precedence for creation mode if specified
// after WithIdentifier. If specified before WithIdentifier, WithIdentifier
// will override it.
func WithTcpPort(port uint16) AgentClientOption {
	return func(cfg *agentClientConfig) {
		cfg.tcpPort = port
		cfg.lastSet = agentClientCreationModeTcp
	}
}

// NewAgentClient creates an Agent client instance with specified options.
// At least one creation option (WithIdentifier or WithTcpPort) should be provided.
// If none is provided, it defaults to identifier mode with an empty identifier.
//
// See WithIdentifier and WithTcpPort for priority rules when both are specified.
func NewAgentClient(opts ...AgentClientOption) (*AgentClient, error) {
	cfg := &agentClientConfig{lastSet: agentClientCreationModeIdentifier}
	for _, opt := range opts {
		opt(cfg)
	}

	var handle uintptr
	if cfg.lastSet == agentClientCreationModeTcp {
		handle = native.MaaAgentClientCreateTcp(cfg.tcpPort)
	} else {
		identifierStrBuf := buffer.NewStringBuffer()
		defer identifierStrBuf.Destroy()
		identifierStrBuf.Set(cfg.identifier)
		handle = native.MaaAgentClientCreateV2(identifierStrBuf.Handle())
	}

	if handle == 0 {
		return nil, errors.New("failed to create agent client")
	}

	return &AgentClient{
		handle: handle,
	}, nil
}

// Destroy releases underlying resources.
func (ac *AgentClient) Destroy() {
	native.MaaAgentClientDestroy(ac.handle)
}

// Identifier returns the identifier of the current agent client.
func (ac *AgentClient) Identifier() (string, bool) {
	buf := buffer.NewStringBuffer()
	defer buf.Destroy()
	ok := native.MaaAgentClientIdentifier(ac.handle, buf.Handle())
	return buf.Get(), ok
}

// BindResource links the Agent client to the specified resource.
func (ac *AgentClient) BindResource(res *Resource) bool {
	return native.MaaAgentClientBindResource(ac.handle, res.handle)
}

// RegisterResourceSink registers resource events for the resource.
func (ac *AgentClient) RegisterResourceSink(res *Resource) bool {
	return native.MaaAgentClientRegisterResourceSink(ac.handle, res.handle)
}

// RegisterControllerSink registers controller events for the controller.
func (ac *AgentClient) RegisterControllerSink(ctrl Controller) bool {
	return native.MaaAgentClientRegisterControllerSink(ac.handle, ctrl.handle)
}

// RegisterTaskerSink registers tasker events for the tasker.
func (ac *AgentClient) RegisterTaskerSink(tasker Tasker) bool {
	return native.MaaAgentClientRegisterTaskerSink(ac.handle, tasker.handle)
}

// Connect connects to the Agent server.
func (ac *AgentClient) Connect() bool {
	return native.MaaAgentClientConnect(ac.handle)
}

// Disconnect disconnects from the Agent server.
func (ac *AgentClient) Disconnect() bool {
	return native.MaaAgentClientDisconnect(ac.handle)
}

// Connected checks if the client is connected to the Agent server.
func (ac *AgentClient) Connected() bool {
	return native.MaaAgentClientConnected(ac.handle)
}

// Alive checks if the Agent server is still responsive.
func (ac *AgentClient) Alive() bool {
	return native.MaaAgentClientAlive(ac.handle)
}

// SetTimeout sets the timeout duration for the Agent server.
func (ac *AgentClient) SetTimeout(duration time.Duration) bool {
	if duration < 0 {
		return false
	}

	milliseconds := duration.Milliseconds()

	return native.MaaAgentClientSetTimeout(ac.handle, milliseconds)
}

// GetCustomRecognitionList returns the custom recognition name list of the agent client.
func (ac *AgentClient) GetCustomRecognitionList() ([]string, bool) {
	buf := buffer.NewStringListBuffer()
	defer buf.Destroy()

	ok := native.MaaAgentClientGetCustomRecognitionList(ac.handle, buf.Handle())
	return buf.GetAll(), ok
}

// GetCustomActionList returns the custom action name list of the agent client.
func (ac *AgentClient) GetCustomActionList() ([]string, bool) {
	buf := buffer.NewStringListBuffer()
	defer buf.Destroy()

	ok := native.MaaAgentClientGetCustomActionList(ac.handle, buf.Handle())
	return buf.GetAll(), ok
}
