# API 接口文档

## API 路由规范

### 路由分层

| 类型 | 前缀 | 说明 |
|------|------|------|
| 面板业务 | `/panel-api/v1/` | Helm、配置、密钥、事件、代理等 |
| K8s 代理 | `/k8s-proxy/` | 仅转发 K8s API (api/v1/*, apis/*) |
| 未授权公开 | `/panel-api/v1/noauth/site/*` | 公开接口（只返回业务字段） |

### 开发环境 vs 生产环境

| 环境 | WebDAV路径 | 说明 |
|------|-----------|------|
| 开发 (LOCAL_MOCK=true) | `/panel-api/v1/files/webdav-agent/{pid}/agent` | 直接访问本地文件系统 |
| 生产 | `/panel-api/v1/files/webdav-agent/{pid}/agent` | 通过代理访问Agent Pod |

## 快速开始

### 认证流程

```
1. 调用登录接口获取 token
2. 在请求头中添加 Authorization: Bearer {token}
3. 调用其他 API
```

## 认证

### 登录
```
POST /panel-api/v1/login
Content-Type: application/x-www-form-urlencoded

Request:
username=admin&password=123456

Response:
{
  "token": "eyJhbGciOi...",
  "refreshToken": "xxx",
  "expire": 1771222291,
  "isK3kUser": false,
  "isClusterUser": false
}
```

### 获取初始化配置 (公开接口)
```
GET /panel-api/v1/noauth/site/init-user
无需认证

Response:
{
  "canInitUser": "false",
  "allowConsoleRegister": "true",
  "captchaEnabled": "false"
}
```

### 获取备案信息 (公开接口)
```
GET /panel-api/v1/noauth/site/beian
无需认证

Response:
{
  "icpnumber": "xxx",
  "number": "xxx",
  "location": "xxx"
}
```

### 获取K3K配置 (公开接口)
```
GET /panel-api/v1/noauth/site/k3k-config
无需认证

Response:
{
  "indexpage": "login"
}
```

## 文件管理

### WebDAV 文件操作
```
# 基础路径 (使用 /panel-api/v1 前缀)
# 生产环境: 通过代理访问
# 开发环境: 直接访问

# 读取目录
PROPFIND /panel-api/v1/files/webdav-agent/{pid}/agent{path}
Headers:
  Authorization: Bearer {token}
  Depth: 1
  Content-Type: text/xml; charset=utf-8

# 读取文件
GET /panel-api/v1/files/webdav-agent/{pid}/agent{path}
Headers:
  Authorization: Bearer {token}

# 创建/写入文件
PUT /panel-api/v1/files/webdav-agent/{pid}/agent{path}
Headers:
  Authorization: Bearer {token}
  Content-Type: {mime-type}

# 创建目录
MKCOL /panel-api/v1/files/webdav-agent/{pid}/agent{path}
Headers:
  Authorization: Bearer {token}

# 删除
DELETE /panel-api/v1/files/webdav-agent/{pid}/agent{path}
Headers:
  Authorization: Bearer {token}

# 重命名/移动
MOVE /panel-api/v1/files/webdav-agent/{pid}/agent{old_path}
Headers:
  Authorization: Bearer {token}
  Destination: /panel-api/v1/files/webdav-agent/{pid}/agent{new_path}

# 复制
COPY /panel-api/v1/files/webdav-agent/{pid}/agent{path}
Headers:
  Authorization: Bearer {token}
  Destination: /panel-api/v1/files/webdav-agent/{pid}/agent{copy_path}
```

### 压缩/解压
```
# 压缩
POST /panel-api/v1/files/compress-agent/{pid}/compress
Headers:
  Authorization: Bearer {token}
  Content-Type: application/json

Request:
{
  "sources": ["/path/file1", "/path/file2"],
  "output": "/path/archive.tar.gz"
}

# 解压
POST /panel-api/v1/files/compress-agent/{pid}/extract
Headers:
  Authorization: Bearer {token}
  Content-Type: application/json

Request:
{
  "source": "/path/archive.tar.gz",
  "target": "/path/extract_dir"
}

# 支持的格式
压缩: zip, tar, tar.gz, tar.xz
解压: zip, tar, tar.gz, tar.bz2, tar.xz, 7z
```

### 权限修改
```
# chmod
POST /panel-api/v1/files/permission-agent/{pid}/chmod
Headers:
  Authorization: Bearer {token}
  Content-Type: application/json

Request:
{
  "path": "/path/file",
  "mode": "755"
}

# chown
POST /panel-api/v1/files/permission-agent/{pid}/chown
Headers:
  Authorization: Bearer {token}
  Content-Type: application/json

Request:
{
  "path": "/path/file",
  "owner": "root"
}
```

## 获取Pod信息
```
GET /panel-api/v1/pid?namespace=default&HostIp=10.0.0.1&containerName=app&podName=app-xxx
Headers:
  Authorization: Bearer {token}

Response:
{
  "pid": 12345,
  "webdavUrl": "/panel-api/v1/files/webdav-agent/12345/agent",
  "webdavToken": "xxx",
  "compressUrl": "/panel-api/v1/files/compress-agent/12345",
  "permissionUrl": "/panel-api/v1/files/permission-agent/12345"
}
```

## K8s 资源 API

### 通过代理访问 K8s 资源
```
# 使用 /k8s-proxy/ 前缀访问 K8s 资源

# 获取命名空间列表
GET /k8s-proxy/api/v1/namespaces
Headers:
  Authorization: Bearer {token}

# 获取 Pod 列表
GET /k8s-proxy/api/v1/namespaces/{namespace}/pods
Headers:
  Authorization: Bearer {token}

# 获取 Service 列表
GET /k8s-proxy/api/v1/namespaces/{namespace}/services
Headers:
  Authorization: Bearer {token}

# 获取 ConfigMap
GET /k8s-proxy/api/v1/namespaces/{namespace}/configmaps/{name}
Headers:
  Authorization: Bearer {token}

# 获取 Secret
GET /k8s-proxy/api/v1/namespaces/{namespace}/secrets/{name}
Headers:
  Authorization: Bearer {token}
```

## 应用管理

### Helm 操作
```
# 获取应用列表
GET /panel-api/v1/helm/releases
Headers:
  Authorization: Bearer {token}

# 获取应用详情
GET /panel-api/v1/helm/releases/{name}
Headers:
  Authorization: Bearer {token}

# 安装应用
POST /panel-api/v1/helm/releases/{name}
Headers:
  Authorization: Bearer {token}
  Content-Type: application/json

# 卸载应用
DELETE /panel-api/v1/helm/releases/{name}
Headers:
  Authorization: Bearer {token}
```

## 监控数据

```
# 获取 CPU/内存使用情况
GET /panel-api/v1/metrics/usage/normal
Headers:
  Authorization: Bearer {token}

# 获取磁盘使用情况
GET /panel-api/v1/metrics/usage/disk
Headers:
  Authorization: Bearer {token}

# 获取监控状态
GET /panel-api/v1/metrics/state
Headers:
  Authorization: Bearer {token}

# 检查是否安装监控
GET /panel-api/v1/metrics/installed
Headers:
  Authorization: Bearer {token}
```

## 集群管理

### K3K 集群信息
```
GET /panel-api/v1/k3k/info
Headers:
  Authorization: Bearer {token}
```

### 超卖资源配置
```
# 获取超卖配置
GET /panel-api/v1/k3k/overselling/config
Headers:
  Authorization: Bearer {token}

# 获取当前资源
GET /panel-api/v1/k3k/overselling/current-resource
Headers:
  Authorization: Bearer {token}
```
