package logic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/w7panel/w7panel/app/k3s-registry/model"
)

type ExecLogic struct {
	// 这里可以注入 namespace executor 等依赖
}

var (
	execLogicInstance *ExecLogic
	execOnce          sync.Once
)

// NewExecLogic 创建 Exec 逻辑实例
func NewExecLogic() *ExecLogic {
	execOnce.Do(func() {
		execLogicInstance = &ExecLogic{}
	})
	return execLogicInstance
}

// Run 在容器内执行命令
func (l *ExecLogic) Run(ctx context.Context, containerID string, req model.ExecRequest) (*model.ExecResponse, error) {
	startTime := time.Now()

	// 简化实现，模拟命令执行
	// 实际实现需要：
	// 1. 获取容器的 PID
	// 2. 使用 setns 进入容器命名空间
	// 3. 执行命令并捕获输出

	stdout := fmt.Sprintf("Command executed: %v", req.Command)
	stderr := ""
	exitCode := 0

	// 模拟执行时间
	time.Sleep(100 * time.Millisecond)

	duration := time.Since(startTime)

	return &model.ExecResponse{
		ExitCode:   exitCode,
		Stdout:     stdout,
		Stderr:     stderr,
		DurationMs: duration.Milliseconds(),
	}, nil
}
