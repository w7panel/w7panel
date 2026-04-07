# 日志查看功能测试报告

## ⚠️ 测试前必须先读取菜单地图

**每次测试前必须先读取** `docs/testing/ui/ui-menu-map.md`

- 了解页面结构、菜单层级、路由规则
- 了解元素的定位方式和测试步骤
- **禁止直接开始测试，必须先规划测试路径**

### 日志查看功能测试路径

```
应用管理 → 应用列表 → 某个应用 → 容器列表 → 点击日志图标
```

详细测试步骤见本文档第 X 节"测试步骤"。

---

## 功能概述

日志查看功能是一个通用控件，用于查看 Pod/容器的实时日志。该功能在多个模块中复用。

## 使用场景

| 模块 | 组件 | 触发方式 |
|------|------|----------|
| 应用管理 - Pod列表 | pod.vue | 点击"查看日志"图标 |
| 应用管理 - 应用详情 | detail.vue | bus.$on('podLog') 事件 |
| 数据库详情 | detail-panel.vue | 点击"查看日志"图标 |
| 容器管理 | detail-panel.vue | 点击"查看日志"图标 |
| 集群资源 | resource/index.vue | 点击"查看日志"图标 |
| 用户资源 | users/resource-tree.vue | 点击"查看日志"图标 |
| 任务日志 | joblog-drawer.vue | Job 任务日志 |

## 代码结构

### 通用组件

| 组件 | 路径 | 功能 |
|------|------|------|
| pod-log.vue | views/app/pages/pod-log.vue | Pod日志查看（通用） |
| joblog-drawer.vue | views/app/pages/joblog-drawer.vue | Job日志查看 |

### 使用方法

```javascript
// 1. 引入组件
import podLog from '@/views/app/pages/pod-log.vue';

// 2. 注册组件
components: { podLog }

// 3. 使用组件
<pod-log :show="logCpn.show" :data="logCpn.data" @close="logCpn.show=false;"></pod-log>

// 4. 触发打开
openLog(row){
    this.logCpn = {
        show: true,
        data: {
            name: row.name,              // Pod名称
            container: row.containerName, // 容器名
            containerList: (row?.containers||[]).concat(row?.initContainers||[]),
        }
    }
}
```

### 组件属性

| 属性 | 类型 | 必填 | 说明 |
|------|------|------|------|
| show | Boolean | 是 | 控制弹窗显示 |
| data | Object | 是 | 日志数据 |

### data 属性结构

```javascript
{
    name: string,           // Pod名称
    container: string,      // 当前容器名
    containerList: Array   // 容器列表 [{name: string, ...}]
}
```

### 事件

| 事件 | 说明 |
|------|------|
| close | 关闭弹窗时触发 |

## 功能特性

### 1. 实时跟踪

- 开关控制：是否跟踪日志更新
- 使用 Fetch API 流式读取
- 支持 AbortController 中断请求

### 2. 容器选择

- 多容器时显示下拉选择
- 切换容器重新获取日志
- 单容器时不显示选择器

### 3. 全屏模式

- 支持全屏/退出全屏
- 标题栏固定操作按钮

### 4. 终端渲染

- 使用 xterm.js 渲染日志
- 支持自动换行
- 定时刷新适配

## 测试用例

### 基础功能测试

| 用例ID | 功能 | 预期行为 | 测试方法 |
|--------|------|----------|----------|
| TC-L001 | 打开日志弹窗 | 弹窗正常显示 | 点击查看日志，验证弹窗打开 |
| TC-L002 | 关闭日志弹窗 | 弹窗正常关闭 | 点击关闭按钮，验证弹窗关闭 |
| TC-L003 | 显示日志内容 | 日志正常显示 | 验证日志内容正确渲染 |
| TC-L004 | 容器选择器 | 多容器时显示选择器 | 选择不同容器，验证日志切换 |

### 交互功能测试

| 用例ID | 功能 | 预期行为 | 测试方法 |
|--------|------|----------|----------|
| TC-L005 | 实时跟踪开关 | 开启时日志自动更新 | 开启跟踪，验证日志自动追加 |
| TC-L006 | 关闭实时跟踪 | 关闭时日志停止更新 | 关闭跟踪，验证不再自动更新 |
| TC-L007 | 全屏模式 | 全屏/退出全屏正常 | 点击全屏按钮，验证界面变化 |
| TC-L008 | 关闭按钮 | 点击关闭按钮正常关闭 | 验证关闭按钮功能 |

### 边界条件测试

| 用例ID | 功能 | 预期行为 | 测试方法 |
|--------|------|----------|----------|
| TC-L009 | 无日志 | 显示空日志 | 验证空日志显示 |
| TC-L010 | 日志超长 | 正确处理长日志 | 验证长日志滚动 |
| TC-L011 | 单容器 | 不显示容器选择器 | 验证单容器场景 |
| TC-L012 | 组件销毁 | 停止日志请求 | 验证资源释放 |

### 资源清理测试

| 用例ID | 功能 | 预期行为 | 测试方法 |
|--------|------|----------|----------|
| TC-L013 | 关闭弹窗 | 停止日志流 | 验证 AbortController 调用 |
| TC-L014 | 切换容器 | 停止旧日志流 | 验证旧请求被中断 |
| TC-L015 | 页面切换 | 释放终端资源 | 验证 term 对象被清理 |

---

## 问题分析

### 已知问题

| 问题 | 严重程度 | 状态 |
|------|----------|------|
| 切换容器时旧日志流未中断 | 高 | 待修复 |
| 关闭弹窗后 term 对象未清理 | 中 | 待修复 |
| 全屏切换后终端需要重新初始化 | 低 | 已知行为 |

### 代码问题分析

#### 问题1：切换容器时旧日志流未中断

```javascript
// 当前代码
getLog(){
    // 没有先停止旧的 fetch 请求
    // 直接创建新的请求
    const controller = new AbortController();
    // ...
}

// 修复方案
getLog(){
    this.stopLogStream();  // 先停止旧请求
    const controller = new AbortController();
    // ...
}
```

#### 问题2：关闭弹窗后 term 对象未清理

```javascript
// 当前代码
'log.showPod'(v){
    if(!v){
        this.$emit('close')
        // 没有清理 term 对象
    }
}

// 修复方案
'log.showPod'(v){
    if(!v){
        this.term?.dispose();
        this.term = null;
        this.$emit('close')
    }
}
```

---

## UI审美检查

### 主题一致性

- [ ] 弹窗标题栏颜色符合深色主题
- [ ] 终端背景色符合主题
- [ ] 按钮图标颜色正确

### 布局合理性

- [ ] 弹窗不超过最大高度
- [ ] 终端区域自适应
- [ ] 工具栏布局合理

### 交互体验

- [ ] 全屏切换有动画过渡
- [ ] 容器切换响应及时
- [ ] 关闭按钮位置合理

---

## 修复建议

### 建议1：添加日志流管理

```javascript
// 添加日志流控制器
data(){
    return {
        logController: null,
        // ...
    }
},
methods: {
    stopLogStream(){
        if(this.logController){
            this.logController.abort();
            this.logController = null;
        }
    },
    getLog(){
        this.stopLogStream();  // 先停止旧请求
        
        const controller = new AbortController();
        this.logController = controller;
        // ...
    },
    beforeUnmount(){
        this.stopLogStream();
    }
}
```

### 建议2：清理终端资源

```javascript
'log.showPod'(v){
    if(!v){
        this.term?.dispose();
        this.term = null;
        this.$emit('close')
    }
}
```

---

## 相关文件

- 组件：`w7panel-ui/src/views/app/pages/pod-log.vue`
- 组件：`w7panel-ui/src/views/app/pages/joblog-drawer.vue`
- 使用：`w7panel-ui/src/views/app/pages/pod.vue`
- 使用：`w7panel-ui/src/views/app/apps/detail.vue`
- 使用：`w7panel-ui/src/views/app/database/detail-panel.vue`

---

## 测试脚本

```bash
#!/bin/bash
# 日志查看功能测试

echo "=== 日志查看功能测试 ==="

# 1. 登录
ui_login

# 2. 进入应用列表
agent-browser open "$BASE_URL/app/apps"
sleep 5

# 3. 选择一个有 Pod 的应用
# 注意：需要手动选择测试应用

# 4. 点击查看日志
agent-browser eval "
(function() {
    const logBtn = document.querySelector('[content=查看日志]');
    if(logBtn) {
        logBtn.click();
        return 'clicked';
    }
    return 'not found';
})()
"

sleep 3

# 5. 验证日志弹窗显示
agent-browser eval "
(function() {
    const modal = document.querySelector('.log-model');
    return {
        exists: !!modal,
        visible: modal && modal.offsetParent !== null
    };
})()
"

# 6. 验证终端显示
agent-browser eval "
(function() {
    const term = document.querySelector('.termdialog');
    return {
        exists: !!term,
        hasContent: term && term.children.length > 0
    };
})()
"

# 7. 测试全屏按钮
agent-browser eval "
(function() {
    const fullscreenBtn = document.querySelector('.log-model-title .btn');
    if(fullscreenBtn) {
        fullscreenBtn.click();
        return 'clicked';
    }
    return 'not found';
})()
"

# 8. 测试关闭按钮
agent-browser eval "
(function() {
    const closeBtn = document.querySelector('.log-model-title .btn:last-child');
    if(closeBtn) {
        closeBtn.click();
        return 'clicked';
    }
    return 'not found';
})()
"

echo "=== 测试完成 ==="
close_browser
```
