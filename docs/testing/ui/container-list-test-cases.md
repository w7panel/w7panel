# 容器列表深度测试用例

## 测试用例清单

### 功能开关测试

| 用例ID | 功能 | 预期行为 | 测试方法 |
|--------|------|----------|----------|
| TC001 | 自动刷新开关 | 默认开启 | 检查开关默认状态 |
| TC002 | 开启自动刷新 | 启动 Watch 流 | 验证 fetch 请求带 watch=true |
| TC003 | 关闭自动刷新 | 停止 Watch 流 | 验证连接被中断 |
| TC004 | 切换开关 | 不触发额外请求 | 验证只改变状态，不立即请求 |

### 边界条件测试

| 用例ID | 功能 | 预期行为 | 测试方法 |
|--------|------|----------|----------|
| TC005 | 组件销毁 | 自动停止 Watch | 验证 beforeUnmount 清理连接 |
| TC006 | 切换命名空间 | 正确切换 Watch 目标 | 验证请求参数变更 |
| TC007 | 页面返回 | 停止 Watch 节省资源 | 验证连接释放 |

### UI 交互测试

| 用例ID | 功能 | 预期行为 | 测试方法 |
|--------|------|----------|----------|
| TC008 | 开关状态显示 | 与实际状态一致 | 验证 UI 与逻辑同步 |
| TC009 | Pod 状态更新 | 实时反映变化 | 验证 ADDED/MODIFIED/DELETED 事件 |

### 资源加载测试（通用）

| 用例ID | 功能 | 预期行为 | 测试方法 |
|--------|------|----------|----------|
| TC010 | 远程资源加载失败 | 显示默认占位符 | 检查控制台错误，验证 fallback |
| TC011 | 图片加载失败 | 显示默认图标/占位符 | 模拟网络错误验证 |
| TC012 | API 请求失败 | 显示错误提示 | 验证错误处理逻辑 |
| TC013 | 状态互斥 | 加载中/成功/失败不同时显示 | 验证条件互斥 |

---

## 问题1：为什么没测出来？

**应用图标加载失败问题没测出来的原因：**

1. **测试粒度太粗**：只检查元素是否存在，不检查资源是否加载成功
2. **缺少网络错误模拟**：没有模拟远程资源不可用的场景
3. **缺少控制台检查**：没有检查网络请求是否失败
4. **缺少 fallback 验证**：没有验证降级方案是否正常工作

**改进方向：**
- 添加资源加载失败的测试用例
- 增加控制台错误检查
- 验证 fallback 逻辑
- 验证状态互斥

## 问题2：为什么错误图标和默认图标同时显示？

**问题原因**：
- Vue 的 `v-if` 在初始渲染时判断条件，@error 触发后 img 元素仍存在于 DOM 中
- 使用 `v-if` + `v-else` 时，@error 回调执行时条件已判断完成

**通用解决方案**：

### 1. 资源加载失败 fallback 规范

```vue
<!-- 方案1：v-if 双重条件（推荐） -->
<img v-if="record.icon && !record.iconLoadError" :src="record.icon" @error="record.iconLoadError = true" />
<icon-common v-if="!record.icon || record.iconLoadError" class="icon" />

<!-- 方案2：v-show 用于 fallback（推荐） -->
<img v-if="record.icon" :src="record.icon" @error="record.iconLoadError = true" />
<icon-common v-show="!record.icon || record.iconLoadError" class="icon" />
```

### 2. 状态互斥规范

```vue
<!-- 错误做法：条件可能同时满足 -->
<div v-if="loading">加载中...</div>
<div v-if="success">成功</div>
<div v-if="error">失败</div>

<!-- 正确做法：使用互斥条件 -->
<div v-if="loading">加载中...</div>
<div v-else-if="success">成功</div>
<div v-else-if="error">失败</div>

<!-- 或使用计算属性确保互斥 -->
<div v-if="status === 'loading'">加载中...</div>
<div v-else-if="status === 'success'">成功</div>
<div v-else-if="status === 'error'">失败</div>
```

### 3. 测试验证方法

```bash
# 验证资源加载 fallback
agent-browser eval "
(function() {
    // 检查是否有多个图标同时显示
    const icons = document.querySelectorAll('.icon');
    const visibleIcons = Array.from(icons).filter(el => {
        const style = window.getComputedStyle(el);
        return style.display !== 'none' && style.visibility !== 'hidden';
    });
    return {
        totalIcons: icons.length,
        visibleIcons: visibleIcons.length,
        isCorrect: visibleIcons.length <= 1
    };
})()
"

# 验证状态互斥
agent-browser eval "
(function() {
    const loading = document.querySelector('.loading');
    const success = document.querySelector('.success');
    const error = document.querySelector('.error');
    return {
        loadingVisible: loading && loading.offsetParent !== null,
        successVisible: success && success.offsetParent !== null,
        errorVisible: error && error.offsetParent !== null,
        isMutuallyExclusive: (loading?.offsetParent !== null ? 1 : 0) + 
                            (success?.offsetParent !== null ? 1 : 0) + 
                            (error?.offsetParent !== null ? 1 : 0) <= 1
    };
})()
"
```

## 测试脚本

### test-auto-refresh.sh

```bash
#!/bin/bash
# 容器列表自动刷新功能深度测试

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/test-lib.sh"

echo "=== 容器列表自动刷新功能深度测试 ==="

# 1. 登录
ui_login || exit 1
sleep 3

# 2. 进入应用管理
agent-browser open "$BASE_URL/app/apps" 2>&1 | grep -v "^$" || true
sleep 5

# 3. 获取交互元素
agent-browser snapshot -i

# 4. 选择一个有 Pod 的应用（需要手动选择或使用已知的测试应用）
# 注意：这里需要用户手动选择应用，或通过 URL 直接进入

# 5. 测试自动刷新开关
# 5.1 检查开关默认状态
agent-browser eval "
(function() {
    const switchEl = document.querySelector('[class*=switch]') || 
                     document.querySelector('.arco-switch') ||
                     document.querySelector('[role=switch]');
    return {
        exists: !!switchEl,
        checked: switchEl?.ariaChecked === 'true' || 
                 switchEl?.classList.contains('arco-switch-checked') ||
                 switchEl?.getAttribute('aria-checked') === 'true',
        text: switchEl?.parentElement?.textContent?.trim() || ''
    };
})()
"

# 6. 测试切换开关
# 6.1 点击开关
agent-browser eval "
(function() {
    const switchEl = document.querySelector('.arco-switch') ||
                     document.querySelector('[role=switch]');
    if(switchEl) {
        switchEl.click();
        return 'clicked';
    }
    return 'not found';
})()
"

# 7. 检查浏览器控制台错误
agent-browser console | grep -iE "error|warning|fetch" || true

# 8. 验证 Pod 列表
agent-browser eval "
(function() {
    const pods = document.querySelectorAll('[class*=table] tbody tr');
    return {
        count: pods.length,
        hasData: pods.length > 0
    };
})()
"

echo "=== 测试完成 ==="
close_browser
```

---

## 修复验证

### 修复后预期

| 测试项 | 修复前 | 修复后 |
|--------|--------|--------|
| 自动刷新 | 需要开关控制，默认 OFF | 始终自动开启 |
| UI 元素 | 有开关 | 无开关 |
| 资源清理 | 无 | 组件销毁时自动清理 |

### 验证方法

1. **UI 验证**：
   - 页面不再显示自动刷新开关
   - Pod 列表自动实时更新

2. **控制台验证**：
   - 页面加载后应有 `watch=true` 请求
   - 离开页面应有 AbortController 调用

## 修复记录

### 2026-02-20 修复 (v2)

**优化决策**：完全删除自动刷新开关，始终使用自动刷新

| 问题 | 修复内容 | 文件位置 |
|------|----------|----------|
| 保留开关多余 | 删除 `<a-switch v-model="autoResetList">` | pod.vue:8-10 |
| 代码冗余 | 删除 `autoResetList` data 属性 | pod.vue:368 |
| 简化逻辑 | 直接使用 watch 模式 | pod.vue:510-517 |

### 修复后代码

```javascript
// 1. 删除 UI 开关 (原 L8-10)
<!-- 删除了整个自动刷新开关 -->

// 2. 删除 data 属性 (原 L368)
// autoResetList: true,  // 已删除

// 3. 简化 getList 方法 (L510-517)
getList(){
    if(!Object.keys(this.data)?.length){return}
    this.nativeList = [];
    let selector = this.data?.spec?.selector?.matchLabels || {};
    let label = Object.keys(selector).map(key=>`${key}=${selector[key]}`).join(',');
    
    // 始终使用 watch 模式
    this.stopWatch();
    
    const controller = new AbortController();
    this.watchController = controller;
    // ... watch 逻辑
}

// 4. 保留资源清理
beforeUnmount(){
    this.stopRequestStatus();
    this.stopWatch();  // 组件销毁时清理
}
```

### 优化效果

| 指标 | 修复前 | 修复后 |
|------|--------|--------|
| 用户操作步骤 | 需要手动开启 | 0 步（自动） |
| 代码行数 | 更多（开关逻辑） | 更少（简洁） |
| 维护成本 | 高（两种模式） | 低（单一模式） |
| 用户体验 | 一般 | 好 |

---

### 2026-02-20 修复 (v1)

| 问题 | 修复内容 | 文件位置 |
|------|----------|----------|
| 默认 OFF | `autoResetList: false` → `true` | pod.vue:372 |
| 切换触发请求 | 删除 `autoResetList` watch | pod.vue:388-390 |
| 无连接清理 | 添加 `watchController` + `stopWatch()` | pod.vue:370,420-426 |
