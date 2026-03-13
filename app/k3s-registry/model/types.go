package model

// ContainerInfo 容器信息
type ContainerInfo struct {
	ID          string        `json:"id"`
	Names       []string      `json:"names"`
	Image       string        `json:"image"`
	ImageID     string        `json:"image_id"`
	PID         uint32        `json:"pid"`
	Status      string        `json:"status"`
	Created     int64         `json:"created"`
	RootFS      string        `json:"rootfs"`
	SnapshotKey string        `json:"snapshot_key"`
	Env         []string      `json:"env"`
	Cmd         []string      `json:"cmd"`
	Overlay     *OverlayInfo  `json:"overlay,omitempty"`
}

type OverlayInfo struct {
	UpperDir string `json:"upper_dir"`
	WorkDir  string `json:"work_dir"`
}

// LayerInfo 镜像层信息
type LayerInfo struct {
	Digest string `json:"digest"`
	Size   int64  `json:"size"`
}

type ContainerLayers struct {
	ContainerID string      `json:"container_id"`
	Image       string      `json:"image"`
	Layers      []LayerInfo `json:"layers"`
	TotalSize   int64       `json:"total_size"`
	LayerCount  int         `json:"layer_count"`
}

// ExecRequest 执行请求
type ExecRequest struct {
	Command []string `json:"command" binding:"required"`
	Env     []string `json:"env"`
	WorkDir string   `json:"workdir"`
	User    string   `json:"user"`
}

type ExecResponse struct {
	ExitCode   int    `json:"exit_code"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	DurationMs int64  `json:"duration_ms"`
}

// CommitRequest 提交请求
type CommitRequest struct {
	NewTag          string   `json:"new_tag" binding:"required"`
	Command         []string `json:"command"`
	Squash          bool     `json:"squash"`
	SquashFrom      int      `json:"squash_from"`
	ReplaceOriginal bool     `json:"replace_original"`
	Restart         bool     `json:"restart"`
}

type CommitResponse struct {
	NewImage        string        `json:"new_image"`
	Digest          string        `json:"digest"`
	Size            int64         `json:"size"`
	CommandExecuted bool          `json:"command_executed"`
	CommandResult   *ExecResponse `json:"command_result,omitempty"`
	SquashResult    *SquashResult `json:"squash_result,omitempty"`
	Replaced        bool          `json:"replaced"`
	Restarted       bool          `json:"restarted"`
}

type SquashResult struct {
	Performed      bool `json:"performed"`
	OriginalLayers int  `json:"original_layers"`
	FinalLayers    int  `json:"final_layers"`
}
