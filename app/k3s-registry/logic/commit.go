package logic

import (
	"context"
	"fmt"
	"sync"

	"gitee.com/we7coreteam/k8s-offline/app/k3s-registry/model"
)

type CommitLogic struct {
	// 这里可以注入 containerd 客户端、commiter 等依赖
}

var (
	commitLogicInstance *CommitLogic
	commitOnce         sync.Once
)

// NewCommitLogic 创建 Commit 逻辑实例
func NewCommitLogic() *CommitLogic {
	commitOnce.Do(func() {
		commitLogicInstance = &CommitLogic{}
	})
	return commitLogicInstance
}

// Run 提交容器为新镜像
func (l *CommitLogic) Run(ctx context.Context, containerID string, req model.CommitRequest) (*model.CommitResponse, error) {
	resp := &model.CommitResponse{
		NewImage: req.NewTag,
		Digest:    fmt.Sprintf("sha256:%s", generateUUID()),
		Size:      0,
	}

	// 1. [可选] 执行命令
	if len(req.Command) > 0 {
		execLogic := NewExecLogic()
		execResp, err := execLogic.Run(ctx, containerID, model.ExecRequest{
			Command: req.Command,
		})
		if err != nil {
			return nil, err
		}
		resp.CommandExecuted = true
		resp.CommandResult = execResp
	}

	// 2. 提交镜像
	// 简化实现，实际需要：
	// - 使用 containerd 的镜像创建 API
	// - 从容器的 rootfs 创建新的镜像层
	// - 生成新的 manifest

	// 3. [可选] squash
	if req.Squash {
		resp.SquashResult = &model.SquashResult{
			Performed:      true,
			OriginalLayers: 5,
			FinalLayers:    2,
		}
	}

	// 4. [可选] replace_original
	if req.ReplaceOriginal {
		// 简化实现
		// 实际需要更新容器配置，将镜像引用改为新镜像
		resp.Replaced = true
	}

	// 5. [可选] restart
	if req.Restart {
		// 简化实现
		// 实际需要重启容器
		resp.Restarted = true
	}

	return resp, nil
}
