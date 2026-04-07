# 测试文档

## 目录结构

```
docs/testing/                    # 测试相关文档和报告
│
├── README.md                  # 本文件 - 测试规范索引
│
├── ui/                       # UI测试报告
│   ├── ui-menu-map.md               # UI菜单地图
│   ├── deep-test-cases-template.md  # 深度测试用例模板（通用）
│   ├── log-viewer-test-report.md    # 日志查看功能测试报告
│   ├── container-list-ui-report.md  # 容器列表测试报告
│   ├── container-list-test-cases.md # 容器列表测试用例
│   ├── ui-performance-report.md     # UI性能分析报告
│   ├── editor-deep-test-report.md  # 编辑器测试报告
│   ├── file-manager-test-report.md # 文件管理测试报告
│   ├── text-editor-deep-test-report.md  # 文本编辑器测试
│   └── ...
│
├── performance/              # 性能测试报告
│   ├── backend.md                  # 后端性能分析
│   └── frontend.md                 # 前端性能分析
│
└── backend/                  # 后端测试报告 (备用)
```

## 定位说明

| 目录 | 内容 | 说明 |
|------|------|------|
| `/tests/` | 测试脚本 | `.sh` 可执行脚本 |
| `/docs/testing/` | 测试文档和报告 | 规范、报告等文档 |

## 测试类型

| 类型 | 报告位置 | 脚本位置 |
|------|----------|----------|
| UI自动化测试 | `docs/testing/ui/` | `tests/panel-ui-test.sh` |
| 深度测试 | `docs/testing/ui/deep-test-cases-template.md` | - |
| 性能测试 | `docs/testing/performance/` | - |
| API测试 | - | `tests/webdav.sh` |
| 功能测试 | - | `tests/*.sh` |

## 快速开始

### UI测试

```bash
# 运行面板功能测试
bash tests/panel-ui-test.sh all

# 运行压缩功能测试
bash tests/compress-ui-test.sh all
```

### 深度测试

参考 [深度测试用例模板](ui/deep-test-cases-template.md) 创建测试用例：

```bash
# 通用测试用例结构
## 功能开关测试
| TC-F001 | 开关默认状态 | 检查默认值是否符合产品设计 |

## 边界条件测试  
| TC-B001 | 组件销毁 | 自动释放资源 |

## UI交互测试
| TC-U001 | UI状态显示 | 与实际状态一致 |

## 资源加载测试
| TC-R001 | 远程资源加载失败 | 显示默认占位符 |
| TC-R002 | 状态互斥 | 不同时显示多种状态 |
```

### 查看测试报告

```bash
# UI测试报告
cat docs/testing/ui/ui-menu-map.md

# 深度测试用例模板
cat docs/testing/ui/deep-test-cases-template.md

# 性能测试报告
cat docs/testing/performance/backend.md
```

## 规范文档

- [UI菜单地图](ui/ui-menu-map.md) - 菜单定位参考
- [深度测试用例模板](ui/deep-test-cases-template.md) - 通用测试用例模板
- [UI测试规范](ui/) - 测试报告索引
- [性能测试报告](performance/) - 性能分析

详见 AGENTS.md 中的"UI自动化测试规范"章节。
