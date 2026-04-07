# 开发指南

## 项目结构

```
$BASE_DIR/
├── w7panel/                      # 后端 (Go)
│   ├── app/application/http/controller/  # 控制器
│   ├── common/service/             # 业务服务
│   ├── common/middleware/          # 中间件
│   ├── install/charts/             # Helm Charts
│   ├── scripts/                    # 构建脚本
│   └── kodata/                    # 静态资源 (开发用)
│
├── w7panel-ui/                     # 前端 (Vue 3 + Arco Design)
│   ├── src/api/                   # API接口
│   ├── src/views/                  # 页面组件
│   ├── src/components/             # 公共组件
│   ├── src/hooks/                  # Hooks
│   └── src/router/                 # 路由配置
│
├── codeblitz/                      # Web IDE (React + Codeblitz)
│   ├── src/                        # 源码
│   └── node_modules/@codeblitzjs/ide-core/bundle/ # WASM文件
│
├── dist/                           # 编译输出
│   ├── w7panel                    # 可执行文件
│   ├── kodata/                    # 静态资源
│   └── runtime/                    # 运行时数据
│
├── kubeconfig.yaml                # K8S集群配置
├── docs/                          # 项目文档
└── tests/                         # 测试脚本
```

## 代码规范

### 后端 (Go)

- 日志使用 `log/slog` 键值对格式
```go
slog.Info("操作成功", "user", userID, "action", "create")
```

- 目录结构：
  - 控制器: `app/{module}/http/controller/`
  - 服务: `common/service/`
  - 中间件: `common/middleware/`

### 前端 (Vue 3)

- 目录结构：
  - API: `src/api/`
  - 页面: `src/views/`
  - 组件: `src/components/`

- API示例：
```typescript
// src/api/cluster.ts
export function compressFiles(compressUrl: string, sources: string[], output: string) {
    return axios.post(`${compressUrl}/compress`, { sources, output });
}
```

## 开发流程

### 1. 前后端同步开发

修改后端后，必须检查前端：
```bash
# 检查前端是否使用该字段
grep -r "新字段名" w7panel-ui/src/
```

### 2. 提交前检查清单

| 检查项 | 命令 |
|--------|------|
| 后端编译 | `cd w7panel && go build` |
| 前端编译 | `cd w7panel-ui && npm run build` |
| 服务启动测试 | `./w7panel-ctl.sh start` |

### 3. 功能测试

完成开发后必须进行功能测试：
- 后端编译通过
- 后端接口返回正确
- 前端编译通过
- 前端功能正常使用

## 常用开发命令

```bash
# 启动开发服务（需要 kubeconfig.yaml）
cd $BASE_DIR/dist
export KUBECONFIG=$BASE_DIR/kubeconfig.yaml
./w7panel-ctl.sh start

# 后端热编译
cd $BASE_DIR/w7panel
go build -o $BASE_DIR/dist/w7panel .

# 前端开发模式
cd $BASE_DIR/w7panel-ui
npm run dev

# 运行测试
cd $BASE_DIR/tests
bash compress-ui-test.sh all
```

## 添加新功能

### 1. 添加后端API

```go
// app/application/http/controller/example.go
package controller

type Example struct {
    controller.Abstract
}

func (self Example) List(http *gin.Context) {
    // 业务逻辑
    self.JsonResponseWithoutError(http, result)
}
```

### 2. 注册路由

```go
// app/application/provider.go
localApiGroup.GET("/example", middleware.Auth{}.Process, controller2.Example{}.List)
```

### 3. 添加前端API

```typescript
// src/api/example.ts
export function getExampleList() {
    return axios.get('/api/example');
}
```

### 4. 创建页面

```vue
<!-- src/views/example/index.vue -->
<template>
  <div>Example Page</div>
</template>

<script setup lang="ts">
import { getExampleList } from '@/api/example';
</script>
```

## 调试技巧

### 查看日志
```bash
# 服务日志
tail -f /tmp/w7panel.log

# 浏览器控制台（UI测试时）
agent-browser console
```

### API测试
```bash
# 获取Token
TOKEN=$(curl -s -X POST "http://localhost:8080/k8s/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}' | jq -r '.token')

# 测试API
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/example
```
