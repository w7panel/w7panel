# K3s 本地镜像仓库与热修改系统 - 完整集成方案

> **目标**: 集成到 w7panel 项目  
> **基于**: Kimi 知识库需求文档

---

## 1. 现有项目分析

### 1.1 w7panel 架构特点

| 特性 | 实现方式 |
|------|----------|
| 框架 | w7-rangine-go (基于 Gin) |
| 路由注册 | `provider.go` 中的 `RegisterHttpRoutes` |
| 控制器 | `app/application/http/controller/` |
| 认证 | `middleware.Auth{}.Process` |
| 日志 | `log/slog` |
| 配置 | YAML 配置文件 + `facade.GetConfig()` |

### 1.2 现有 API 风格

```go
// 参考: app/application/provider.go
server.RegisterRouters(func(engine *gin.Engine) {
    apiGroup := engine.Group("/panel-api/v1")
    {
        apiGroup.GET("/namespaces", middleware.Auth{}.Process, controller.Namespaces{}.GetList)
        // ...
    }
})
```

---

## 2. 项目结构设计

### 2.1 目录结构

```
w7panel/
├── app/
│   └── k3s-registry/                    # 新增模块
│       ├── http/
│       │   └── controller/
│       │       ├── registry.go          # Registry API 控制器
│       │       ├── containers.go        # 容器列表 API
│       │       ├── exec.go             # 容器执行 API
│       │       └── commit.go           # 镜像提交 API
│       ├── logic/
│       │   ├── container.go            # 容器操作逻辑
│       │   ├── image.go               # 镜像操作逻辑
│       │   └── registry.go             # Registry 逻辑
│       ├── model/
│       │   └── types.go               # 数据类型定义
│       ├── internal/
│       │   ├── config.go              # 配置定义
│       │   ├── containerd/
│       │   │   ├── client.go          # containerd 客户端
│       │   │   └── container.go      # 容器操作
│       │   ├── metadata/
│       │   │   └── bolt.go           # BoltDB 操作
│       │   ├── content/
│       │   │   └── store.go          # Blob 存储
│       │   └── patcher/
│       │       ├── namespace.go       # setns 工具
│       │       ├── executor.go       # 命令执行
│       │       ├── commit.go          # 镜像提交
│       │       └── squash.go         # 层合并
│       └── provider.go               # 模块注册
├── common/
│   └── service/
│       └── k3s-registry/             # 服务层
│           └── registry.go
└── config.yaml                        # 新增配置
```

---

## 3. 模块设计

### 3.1 Provider 注册

```go
// app/k3s-registry/provider.go
package k3sregistry

import (
    "gitee.com/we7coreteam/k8s-offline/app/k3s-registry/http/controller"
    "gitee.com/we7coreteam/k8s-offline/common/middleware"
    "github.com/gin-gonic/gin"
    "github.com/we7coreteam/w7-rangine-go/v2/src/httpserver"
)

type Provider struct{}

func (p Provider) Register(httpServer *httpserver.Server) {
    p.RegisterHttpRoutes(httpServer)
}

func (p Provider) RegisterHttpRoutes(server *httpserver.Server) {
    server.RegisterRouters(func(engine *gin.Engine) {
        // Registry API - 镜像仓库
        registryGroup := engine.Group("/panel-api/v1/k3s-registry")
        registryGroup.Use(middleware.Auth{}.Process)
        {
            registryGroup.GET("/v2/", controller.Registry{}.Version)
            registryGroup.GET("/v2/_catalog", controller.Registry{}.Catalog)
            registryGroup.GET("/v2/:name/tags/list", controller.Registry{}.Tags)
            registryGroup.GET("/v2/:name/manifests/*reference", controller.Registry{}.Manifest)
            registryGroup.PUT("/v2/:name/manifests/*reference", controller.Registry{}.PushManifest)
            registryGroup.GET("/v2/:name/blobs/*digest", controller.Registry{}.Blob)
            registryGroup.HEAD("/v2/:name/blobs/*digest", controller.Registry{}.BlobExists)
            registryGroup.POST("/v2/:name/blobs/uploads/", controller.Registry{}.InitUpload)
            registryGroup.PUT("/v2/:name/blobs/uploads/:uuid", controller.Registry{}.CompleteUpload)
        }

        // Patch API - 容器操作
        patchGroup := engine.Group("/panel-api/v1/k3s-patch")
        patchGroup.Use(middleware.Auth{}.Process)
        {
            patchGroup.GET("/containers", controller.Containers{}.List)
            patchGroup.GET("/containers/:id", controller.Containers{}.Get)
            patchGroup.GET("/containers/:id/layers", controller.Containers{}.Layers)
            patchGroup.POST("/containers/:id/exec", controller.Exec{}.Run)
            patchGroup.POST("/containers/:id/commit", controller.Commit{}.Run)
        }
    })
}
```

### 3.2 配置定义

```go
// app/k3s-registry/internal/config.go
package internal

import "github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"

type Config struct {
    // containerd socket 路径
    ContainerdSocket string
    
    // 数据目录
    DataDir string
    
    // Runtime 目录
    RuntimeDir string
    
    // Agent 目录
    AgentDir string
    
    // 是否启用
    Enabled bool
}

func LoadConfig() *Config {
    return &Config{
        ContainerdSocket: facade.GetConfig().GetString("k3s-registry.containerd_socket"),
        DataDir:          facade.GetConfig().GetString("k3s-registry.data_dir"),
        RuntimeDir:       facade.GetConfig().GetString("k3s-registry.runtime_dir"),
        AgentDir:         facade.GetConfig().GetString("k3s-registry.agent_dir"),
        Enabled:          facade.GetConfig().GetBool("k3s-registry.enabled"),
    }
}

// 默认值
func DefaultConfig() map[string]interface{} {
    return map[string]interface{}{
        "k3s-registry.enabled":           false,
        "k3s-registry.containerd_socket": "/run/k3s/containerd/containerd.sock",
        "k3s-registry.data_dir":         "/var/lib/rancher/k3s/agent/containerd",
        "k3s-registry.runtime_dir":      "/run/k3s/containerd/io.containerd.runtime.v2.task/k8s.io",
        "k3s-registry.agent_dir":        "/var/lib/rancher/k3s/agent",
    }
}
```

### 3.3 数据类型

```go
// app/k3s-registry/model/types.go
package model

// 容器信息
type ContainerInfo struct {
    ID          string            `json:"id"`
    Names       []string          `json:"names"`
    Image       string            `json:"image"`
    ImageID     string            `json:"image_id"`
    PID         uint32           `json:"pid"`
    Status      string            `json:"status"`
    Created     int64            `json:"created"`
    RootFS      string            `json:"rootfs"`
    SnapshotKey string            `json:"snapshot_key"`
    Env         []string         `json:"env"`
    Cmd         []string         `json:"cmd"`
    Overlay     *OverlayInfo     `json:"overlay,omitempty"`
}

type OverlayInfo struct {
    UpperDir string `json:"upper_dir"`
    WorkDir  string `json:"work_dir"`
}

// 镜像层信息
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

// 执行请求
type ExecRequest struct {
    Command []string `json:"command" binding:"required"`
    Env     []string `json:"env"`
    WorkDir string   `json:"workdir"`
    User    string   `json:"user"`
}

type ExecResponse struct {
    ExitCode  int    `json:"exit_code"`
    Stdout    string `json:"stdout"`
    Stderr    string `json:"stderr"`
    DurationMs int64 `json:"duration_ms"`
}

// 提交请求
type CommitRequest struct {
    NewTag          string   `json:"new_tag" binding:"required"`
    Command         []string `json:"command"`
    Squash          bool     `json:"squash"`
    SquashFrom      int      `json:"squash_from"`
    ReplaceOriginal bool    `json:"replace_original"`
    Restart         bool     `json:"restart"`
}

type CommitResponse struct {
    NewImage         string        `json:"new_image"`
    Digest           string        `json:"digest"`
    Size             int64         `json:"size"`
    CommandExecuted  bool          `json:"command_executed"`
    CommandResult    *ExecResponse `json:"command_result,omitempty"`
    SquashResult     *SquashResult `json:"squash_result,omitempty"`
    Replaced         bool          `json:"replaced"`
    Restarted        bool          `json:"restarted"`
}

type SquashResult struct {
    Performed       bool `json:"performed"`
    OriginalLayers  int  `json:"original_layers"`
    FinalLayers     int  `json:"final_layers"`
}
```

---

## 4. 核心实现

### 4.1 containerd 客户端

```go
// app/k3s-registry/internal/containerd/client.go
package containerd

import (
    "context"
    "fmt"
    
    "github.com/containerd/containerd"
    "github.com/containerd/containerd/containers"
    "github.com/containerd/containerd/leases"
    "github.com/containerd/containerd/snapshots"
)

type Client struct {
    client *containerd.Client
    config *internal.Config
}

func NewClient(cfg *internal.Config) (*Client, error) {
    client, err := containerd(cfg.ContainerdSocket)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to containerd: %w", err)
    }
    return &Client{client: client, config: cfg}, nil
}

// ListContainers 列出运行中的容器
func (c *Client) ListContainers(ctx context.Context) ([]*containers.Container, error) {
    return c.client.Containers(ctx)
}

// GetContainer 获取指定容器
func (c *Client) GetContainer(ctx context.Context, id string) (*containers.Container, error) {
    return c.client.LoadContainer(ctx, id)
}

// GetImage 获取镜像信息
func (c *Client) GetImage(ctx context.Context, ref string) (containerd.Image, error) {
    return c.client.GetImage(ctx, ref)
}

// ListImages 列出所有镜像
func (c *Client) ListImages(ctx context.Context) ([]containerd.Image, error) {
    return c.client.ListImages(ctx)
}
```

### 4.2 BoltDB 元数据操作

```go
// app/k3s-registry/internal/metadata/bolt.go
package metadata

import (
    "fmt"
    "sync"
    
    "go.etcd.io/bbolt"
)

type BoltStore struct {
    db   *bbolt.DB
    path string
}

var (
    defaultBucket = []byte("images")
    once         sync.Once
    store        *BoltStore
)

func NewBoltStore(path string) (*BoltStore, error) {
    var err error
    var db *bbolt.DB
    
    once.Do(func() {
        db, err = bbolt.Open(path, 0644, nil)
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to open bolt db: %w", err)
    }
    
    // 创建默认 bucket
    err = db.Update(func(tx *bbolt.Tx) error {
        _, err := tx.CreateBucketIfNotExists(defaultBucket)
        return err
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to create bucket: %w", err)
    }
    
    store = &BoltStore{db: db, path: path}
    return store, nil
}

func (b *BoltStore) GetCatalog() ([]string, error) {
    var images []string
    err := b.db.View(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket(defaultBucket)
        cursor := bucket.Cursor()
        for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
            images = append(images, string(k))
        }
        return nil
    })
    return images, err
}

func (b *BoltStore) GetTags(image string) ([]string, error) {
    var tags []string
    err := b.db.View(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket(defaultBucket)
        data := bucket.Get([]byte(image))
        if data == nil {
            return nil
        }
        // 解析 tags
        return nil
    })
    return tags, err
}

func (b *BoltStore) PutManifest(image, manifest string) error {
    return b.db.Update(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket(defaultBucket)
        return bucket.Put([]byte(image), []byte(manifest))
    })
}

func (b *BoltStore) GetManifest(image string) (string, error) {
    var manifest string
    err := b.db.View(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket(defaultBucket)
        data := bucket.Get([]byte(image))
        if data != nil {
            manifest = string(data)
        }
        return nil
    })
    return manifest, err
}
```

### 4.3 容器运行时信息

```go
// app/k3s-registry/internal/containerd/container.go
package containerd

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    
    "gitee.com/we7coreteam/k8s-offline/app/k3s-registry/model"
)

type ContainerManager struct {
    runtimeDir string
    agentDir  string
}

type TaskInfo struct {
    ID      string `json:"id"`
    PID     uint32 `json:"pid"`
    Status  string `json:"status"`
}

type ContainerStatus struct {
    Image   string   `json:"image"`
    RootFS  string   `json:"rootfs"`
    SnapshotKey string `json:"snapshotKey"`
    Env     []string `json:"env"`
    Cmd     []string `json:"cmd"`
}

func NewContainerManager(runtimeDir, agentDir string) *ContainerManager {
    return &ContainerManager{
        runtimeDir: runtimeDir,
        agentDir:  agentDir,
    }
}

// ListContainers 列出所有容器
func (m *ContainerManager) ListContainers() ([]model.ContainerInfo, error) {
    var containers []model.ContainerInfo
    
    entries, err := os.ReadDir(m.runtimeDir)
    if err != nil {
        return nil, err
    }
    
    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }
        
        stateFile := filepath.Join(m.runtimeDir, entry.Name(), "state.json")
        data, err := os.ReadFile(stateFile)
        if err != nil {
            continue
        }
        
        var state TaskInfo
        if err := json.Unmarshal(data, &state); err != nil {
            continue
        }
        
        // 获取容器配置
        configFile := filepath.Join(m.agentDir, "io.containerd.metadata.v1.bolt", "meta.db")
        
        containers = append(containers, model.ContainerInfo{
            ID:      state.ID,
            PID:     state.PID,
            Status:  state.Status,
        })
    }
    
    return containers, nil
}

// GetContainerInfo 获取容器详细信息
func (m *ContainerManager) GetContainerInfo(id string) (*model.ContainerInfo, error) {
    stateFile := filepath.Join(m.runtimeDir, id, "state.json")
    data, err := os.ReadFile(stateFile)
    if err != nil {
        return nil, fmt.Errorf("container not found: %w", err)
    }
    
    var state TaskInfo
    if err := json.Unmarshal(data, &state); err != nil {
        return nil, err
    }
    
    // 读取容器配置
    configFile := filepath.Join(m.agentDir, "io.containerd.metadata.v1.bolt", "meta.db")
    
    return &model.ContainerInfo{
        ID:     state.ID,
        PID:    state.PID,
        Status: state.Status,
    }, nil
}
```

### 4.4 命名空间操作 (setns)

```go
// app/k3s-registry/internal/patcher/namespace.go
package patcher

import (
    "fmt"
    "os"
    
    "golang.org/x/sys/unix"
)

type NamespaceExecutor struct {
    pid uint32
}

// NewNamespaceExecutor 创建命名空间执行器
func NewNamespaceExecutor(pid uint32) *NamespaceExecutor {
    return &NamespaceExecutor{pid: pid}
}

// Enter 进入容器命名空间
func (e *NamespaceExecutor) Enter() (restore func(), err error) {
    nsPath := fmt.Sprintf("/proc/%d/ns/mnt", e.pid)
    nsFile, err := os.Open(nsPath)
    if err != nil {
        return nil, fmt.Errorf("failed to open namespace: %w", err)
    }
    
    // 保存当前 namespace
    currentNs, err := os.Open("/proc/self/ns/mnt")
    if err != nil {
        nsFile.Close()
        return nil, err
    }
    
    restore = func() {
        unix.Setns(int(currentNs.Fd()), 0)
        currentNs.Close()
        nsFile.Close()
    }
    
    // 进入容器 namespace
    if err := unix.Setns(int(nsFile.Fd()), 0); err != nil {
        restore()
        return nil, fmt.Errorf("failed to setns: %w", err)
    }
    
    return restore, nil
}

// Exec 在容器内执行命令
func (e *NamespaceExecutor) Exec(command []string, workdir string) (int, string, string, error) {
    restore, err := e.Enter()
    if err != nil {
        return -1, "", "", err
    }
    defer restore()
    
    // 设置工作目录
    if workdir != "" {
        if err := os.Chdir(workdir); err != nil {
            return -1, "", "", err
        }
    }
    
    // 执行命令
    // 这里使用 syscall.Exec 或 exec.Command
    // ...
    
    return 0, "output", "", nil
}
```

### 4.5 镜像提交

```go
// app/k3s-registry/internal/patcher/commit.go
package patcher

import (
    "context"
    "fmt"
    
    "gitee.com/we7coreteam/k8s-offline/app/k3s-registry/internal/containerd"
    "gitee.com/we7coreteam/k8s-offline/app/k3s-registry/model"
)

type Committer struct {
    client      *containerd.Client
    containerMgr *containerd.ContainerManager
}

func NewCommitter(cfg *internal.Config) (*Committer, error) {
    client, err := containerd.NewClient(cfg)
    if err != nil {
        return nil, err
    }
    
    return &Committer{
        client:       client,
        containerMgr: containerd.NewContainerManager(cfg.RuntimeDir, cfg.AgentDir),
    }, nil
}

// Commit 提交容器为新镜像
func (c *Committer) Commit(ctx context.Context, req model.CommitRequest) (*model.CommitResponse, error) {
    resp := &model.CommitResponse{}
    
    // 1. 获取容器信息
    container, err := c.client.GetContainer(ctx, req.ID)
    if err != nil {
        return nil, fmt.Errorf("container not found: %w", err)
    }
    
    // 2. [可选] 执行命令
    if len(req.Command) > 0 {
        execResp, err := c.execInContainer(ctx, req.ID, req.Command)
        if err != nil {
            return nil, err
        }
        resp.CommandExecuted = true
        resp.CommandResult = execResp
    }
    
    // 3. 提交镜像
    newImage, err := c.createImage(ctx, container, req.NewTag)
    if err != nil {
        return nil, fmt.Errorf("failed to create image: %w", err)
    }
    
    resp.NewImage = newImage
    resp.Digest = "sha256:..."
    
    // 4. [可选] squash
    if req.Squash {
        squashResp, err := c.squashImage(ctx, newImage, req.SquashFrom)
        if err != nil {
            return nil, err
        }
        resp.SquashResult = squashResp
    }
    
    // 5. [可选] replace_original
    if req.ReplaceOriginal {
        if err := c.replaceContainerImage(ctx, req.ID, newImage); err != nil {
            return nil, err
        }
        resp.Replaced = true
    }
    
    // 6. [可选] restart
    if req.Restart {
        if err := c.restartContainer(ctx, req.ID); err != nil {
            return nil, err
        }
        resp.Restarted = true
    }
    
    return resp, nil
}

func (c *Committer) createImage(ctx context.Context, container containerd.Container, tag string) (string, error) {
    // 实现镜像创建逻辑
    // ...
    return tag, nil
}

func (c *Committer) squashImage(ctx context.Context, image string, fromLayer int) (*model.SquashResult, error) {
    // 实现 squash 逻辑
    // ...
    return &model.SquashResult{
        Performed:      true,
        OriginalLayers: 5,
        FinalLayers:    2,
    }, nil
}

func (c *Committer) replaceContainerImage(ctx context.Context, containerID, newImage string) error {
    // 实现镜像替换逻辑
    // ...
    return nil
}

func (c *Committer) restartContainer(ctx context.Context, containerID string) error {
    // 实现容器重启逻辑
    // ...
    return nil
}
```

---

## 5. 控制器实现

### 5.1 Registry API 控制器

```go
// app/k3s-registry/http/controller/registry.go
package controller

import (
    "net/http"
    
    "gitee.com/we7coreteam/k8s-offline/app/k3s-registry/internal"
    "gitee.com/we7coreteam/k8s-offline/app/k3s-registry/logic"
    "github.com/gin-gonic/gin"
    "github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Registry struct {
    controller.Abstract
}

var registryLogic = logic.NewRegistryLogic()

// Version 返回 Registry API 版本
func (c Registry) Version(ctx *gin.Context) {
    ctx.JSON(http.StatusOK, gin.H{
        "schemas": []string{"https://docs.docker.com/spec/api/v2/"},
    })
}

// Catalog 返回镜像列表
func (c Registry) Catalog(ctx *gin.Context) {
    catalog, err := registryLogic.GetCatalog(ctx)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"repositories": catalog})
}

// Tags 返回镜像标签
func (c Registry) Tags(ctx *gin.Context) {
    name := ctx.Param("name")
    tags, err := registryLogic.GetTags(ctx, name)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"name": name, "tags": tags})
}

// Manifest 获取镜像 manifest
func (c Registry) Manifest(ctx *gin.Context) {
    name := ctx.Param("name")
    reference := ctx.Param("reference")
    
    manifest, err := registryLogic.GetManifest(ctx, name, reference)
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    
    ctx.Data(http.StatusOK, "application/vnd.docker.distribution.manifest.v2+json", []byte(manifest))
}

// PushManifest 推送镜像 manifest
func (c Registry) PushManifest(ctx *gin.Context) {
    name := ctx.Param("name")
    reference := ctx.Param("reference")
    
    var manifest string
    if err := ctx.ShouldBindJSON(&manifest); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid manifest"})
        return
    }
    
    if err := registryLogic.PushManifest(ctx, name, reference, manifest); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    ctx.Header("Location", "/v2/"+name+"/manifests/"+reference)
    ctx.JSON(http.StatusCreated, gin.H{})
}

// Blob 获取 blob
func (c Registry) Blob(ctx *gin.Context) {
    name := ctx.Param("name")
    digest := ctx.Param("digest")
    
    data, err := registryLogic.GetBlob(ctx, name, digest)
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    
    ctx.Data(http.StatusOK, "application/octet-stream", data)
}

// InitUpload 初始化 blob 上传
func (c Registry) InitUpload(ctx *gin.Context) {
    name := ctx.Param("name")
    
    uuid, err := registryLogic.InitUpload(ctx, name)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    ctx.Header("Location", "/v2/"+name+"/blobs/uploads/"+uuid)
    ctx.Header("Range", "bytes=0-0")
    ctx.JSON(http.StatusAccepted, gin.H{})
}

// CompleteUpload 完成 blob 上传
func (c Registry) CompleteUpload(ctx *gin.Context) {
    name := ctx.Param("name")
    uuid := ctx.Param("uuid")
    digest := ctx.Query("digest")
    
    body, _ := ctx.GetRawData()
    
    if err := registryLogic.CompleteUpload(ctx, name, uuid, digest, body); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    ctx.Header("Location", "/v2/"+name+"/blobs/"+digest)
    ctx.JSON(http.StatusCreated, gin.H{})
}
```

### 5.2 Patch API 控制器

```go
// app/k3s-registry/http/controller/containers.go
package controller

import (
    "net/http"
    
    "gitee.com/we7coreteam/k8s-offline/app/k3s-registry/logic"
    "github.com/gin-gonic/gin"
    "github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Containers struct {
    controller.Abstract
}

var containersLogic = logic.NewContainersLogic()

// List 获取容器列表
func (c Containers) List(ctx *gin.Context) {
    containers, err := containersLogic.List(ctx)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, containers)
}

// Get 获取容器详情
func (c Containers) Get(ctx *gin.Context) {
    id := ctx.Param("id")
    
    container, err := containersLogic.Get(ctx, id)
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "container not found"})
        return
    }
    ctx.JSON(http.StatusOK, container)
}

// Layers 获取容器镜像层
func (c Containers) Layers(ctx *gin.Context) {
    id := ctx.Param("id")
    
    layers, err := containersLogic.GetLayers(ctx, id)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, layers)
}
```

---

## 6. 配置文件

### 6.1 config.yaml 新增配置

```yaml
# w7panel/config.yaml

# K3s 本地镜像仓库配置
k3s-registry:
  enabled: true                                    # 是否启用
  containerd_socket: "/run/k3s/containerd/containerd.sock"
  runtime_dir: "/run/k3s/containerd/io.containerd.runtime.v2.task/k8s.io"
  agent_dir: "/var/lib/rancher/k3s/agent"
```

### 6.2 默认配置

在 `config.yaml` 加载前，设置默认值：

```go
// app/k3s-registry/provider.go
func init() {
    // 注册默认配置
    facade.RegisterDefaultConfig(internal.DefaultConfig())
}
```

---

## 7. 部署配置

### 7.1 K8s 部署

由于需要访问 containerd socket 和本地文件系统，需要以 DaemonSet 方式部署：

```yaml
# w7panel/install/charts/w7panel/templates/daemonset.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: w7panel
  labels:
    app: w7panel
spec:
  selector:
    matchLabels:
      app: w7panel
  template:
    metadata:
      labels:
        app: w7panel
    spec:
      hostNetwork: true
      hostPID: true
      containers:
      - name: w7panel
        image: w7panel:xxx
        securityContext:
          privileged: true
        env:
        - name: K3S_REGISTRY_ENABLED
          value: "true"
        volumeMounts:
        - name: containerd-sock
          mountPath: /run/k3s/containerd
        - name: k3s-agent
          mountPath: /var/lib/rancher/k3s/agent
          readOnly: true
        - name: proc
          mountPath: /host/proc
          mountPropagation: HostToContainer
        - name: run
          mountPath: /run
      volumes:
      - name: containerd-sock
        hostPath:
          path: /run/k3s/containerd
      - name: k3s-agent
        hostPath:
          path: /var/lib/rancher/k3s/agent
      - name: proc
        hostPath:
          path: /proc
      - name: run
        hostPath:
          path: /run
```

---

## 8. 前端集成

### 8.1 API 接口

根据 w7panel-ui 的现有模式，添加以下 API：

```typescript
// w7panel-ui/src/api/k3s-registry.ts
import axios from '@/api';

export function getRegistryCatalog() {
  return axios.get('/panel-api/v1/k3s-registry/v2/_catalog');
}

export function getImageTags(name: string) {
  return axios.get(`/panel-api/v1/k3s-registry/v2/${name}/tags/list`);
}

export function listContainers() {
  return axios.get('/panel-api/v1/k3s-patch/containers');
}

export function getContainer(id: string) {
  return axios.get(`/panel-api/v1/k3s-patch/containers/${id}`);
}

export function getContainerLayers(id: string) {
  return axios.get(`/panel-api/v1/k3s-patch/containers/${id}/layers`);
}

export function execInContainer(id: string, data: {
  command: string[];
  env?: string[];
  workdir?: string;
  user?: string;
}) {
  return axios.post(`/panel-api/v1/k3s-patch/containers/${id}/exec`, data);
}

export function commitContainer(id: string, data: {
  new_tag: string;
  command?: string[];
  squash?: boolean;
  replace_original?: boolean;
  restart?: boolean;
}) {
  return axios.post(`/panel-api/v1/k3s-patch/containers/${id}/commit`, data);
}
```

---

## 9. 实施计划

### Phase 1: 基础架构 (Week 1)

| 任务 | 描述 |
|------|------|
| T1.1 | 创建 `app/k3s-registry` 模块目录结构 |
| T1.2 | 实现 `internal/config.go` 配置管理 |
| T1.3 | 实现 `internal/containerd/client.go` containerd 客户端 |
| T1.4 | 实现 `internal/metadata/bolt.go` BoltDB 操作 |
| T1.5 | 创建 `provider.go` 模块注册 |

### Phase 2: Registry API (Week 2)

| 任务 | 描述 |
|------|------|
| T2.1 | 实现版本检查和目录 API |
| T2.2 | 实现镜像标签 API |
| T2.3 | 实现 manifest 读写 API |
| T2.4 | 实现 blob 上传/下载 API |

### Phase 3: Patch API (Week 3)

| 任务 | 描述 |
|------|------|
| T3.1 | 实现容器列表 API |
| T3.2 | 实现容器详情 API |
| T3.3 | 实现镜像层信息 API |
| T3.4 | 实现容器内命令执行 API |

### Phase 4: 镜像提交 (Week 4)

| 任务 | 描述 |
|------|------|
| T4.1 | 实现 setns 工具 |
| T4.2 | 实现镜像提交核心逻辑 |
| T4.3 | 实现 squash 层合并 |
| T4.4 | 实现镜像替换和容器重启 |

### Phase 5: 集成与测试 (Week 5)

| 任务 | 描述 |
|------|------|
| T5.1 | 注册到 w7panel 主程序 |
| T5.2 | 前端页面开发 |
| T5.3 | 集成测试 |
| T5.4 | 部署配置更新 |

---

## 10. 依赖项

需要在 `go.mod` 中添加：

```go
require (
    github.com/containerd/containerd v1.7.x
    go.etcd.io/bbolt v1.3.x
)
```

---

## 11. 验收标准

- [ ] Registry API 完全兼容 Docker Registry API V2
- [ ] Patch API 容器操作正常
- [ ] 镜像提交功能完整可用
- [ ] 与 w7panel 认证系统集成
- [ ] 在 k3s server/agent 节点正常工作
- [ ] 零外部依赖（不调用 crictl/ctr/docker）
