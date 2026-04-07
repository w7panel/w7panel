# w7panel 部署指南

## 启动方式（推荐使用 w7panel-ctl.sh）

启动脚本会自动检测并设置正确的环境变量：

```bash
# ========== 开发模式 (需要 kubeconfig.yaml) ==========
export KUBECONFIG=/path/to/kubeconfig.yaml
./w7panel-ctl.sh start

# ========== 生产模式 (使用 ServiceAccount) ==========
export LOCAL_MOCK=false
./w7panel-ctl.sh start
```

## 环境变量说明

| 变量名 | 开发模式 | 生产模式 | 说明 |
|--------|---------|---------|------|
| `CAPTCHA_ENABLED` | false | false | 验证码开关 |
| `LOCAL_MOCK` | true | false | K8s 访问方式 |
| `KO_DATA_PATH` | ./kodata | ./kodata | 静态资源目录 |
| `KUBECONFIG` | 必填 | 不需要 | kubeconfig 文件路径 |

## 常见错误

### 权限错误

**错误信息**：
```
serviceaccounts "admin" is forbidden: User "system:serviceaccount:default:default" 
cannot get resource "serviceaccounts" in API group "" in the namespace "default"
```

**原因**：未正确设置模式，系统使用了 ServiceAccount 而非 kubeconfig。

**解决**：
- 开发模式：设置 `KUBECONFIG` 环境变量
- 生产模式：确认 ServiceAccount 权限配置正确
