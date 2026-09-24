package main

import (
	"fmt"
	"os"
	"time"

	"github.com/MaaXYZ/maa-framework-go/v4"
)

func main() {
	maa.Init()

	socketID := os.Args[1]

	if err := maa.AgentServerRegisterCustomAction("TestAgentServer", NewAgentServerAction()); err != nil {
		fmt.Println(err)
		return
	}

	// 注册关闭回调：收到 ShutDownRequest 时先执行（可阻塞做收尾），
	// 返回后服务端才回 ShutDownResponse——客户端 Disconnect() 一返回
	// 就代表这里的收尾已经完成。必须在 AgentServerStartUp 之前调用。
	//
	// Register the shutdown callback: runs when a ShutDownRequest arrives
	// (can block for cleanup); the ShutDownResponse is sent only after it
	// returns — the client's Disconnect() returning means cleanup is done.
	// Must be called before AgentServerStartUp.
	if err := maa.AgentServerSetShutdownCallback(func() {
		fmt.Println("shutting down: doing cleanup...")
		// 演示阻塞收尾：客户端的 Disconnect() 会等到这里返回才收到确认。
		// 实际场景可以替换为等待在途通知发完、写状态文件等操作。
		//
		// Demonstrate blocking cleanup: the client's Disconnect() waits
		// until this returns. Replace with flushing in-flight notifications,
		// writing state files, or other real cleanup work.
		time.Sleep(2 * time.Second)
		fmt.Println("cleanup done")
	}); err != nil {
		fmt.Println(err)
		return
	}

	if err := maa.AgentServerStartUp(socketID); err != nil {
		fmt.Println(err)
		return
	}

	maa.AgentServerJoin()

	maa.AgentServerShutDown()
}

type AgentServerAction struct{}

// Run implements maa.CustomAction.
func (a *AgentServerAction) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {

	fmt.Println("Agent server custom action is running")

	return true
}

func NewAgentServerAction() maa.CustomActionRunner {
	return &AgentServerAction{}
}
