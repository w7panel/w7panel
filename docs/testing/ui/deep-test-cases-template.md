# 深度测试用例模板

本文档提供深度测试用例的通用模板，适用于各种功能模块的测试。

---

## ⚠️ 测试前必须先读取菜单地图

**每次测试前必须先读取** `docs/testing/ui/ui-menu-map.md`

- 了解页面结构、菜单层级、路由规则
- 了解元素的定位方式和测试步骤
- **禁止直接开始测试，必须先规划测试路径**

### 正确测试流程

```
1. 读取菜单地图 → 了解页面结构
2. 规划测试路径 → 确定需要测试的功能入口
3. 编写测试用例 → 按照模板编写测试步骤
4. 执行测试 → 使用 agent-browser 执行
5. 分析问题 → 根因分析和修复
6. 验证修复 → 重新测试确认
7. 总结报告 → 更新测试用例文档
```

---

## ⚠️ 文档更新规范（重要！）

测试过程中如果发现了新功能、对项目有了新的理解，必须及时更新相关文档：

```
□ 1. 新功能发现
   - 发现系统新增了某个功能
   - 发现某个功能的新的使用方式
   - 发现代码结构或架构的变化

□ 2. 理解更新
   - 对系统工作原理有了新理解
   - 发现之前的理解有误
   - 发现新的技术细节

□ 3. 需要更新的文档
   - 菜单地图 (ui-menu-map.md)
   - 测试用例 (xxx-test-cases.md)
   - 测试报告 (xxx-test-report.md)
   - AGENTS.md 相关章节
   - 项目 README.md
```

### 文档更新检查清单

```
□ 测试完成后检查：
□ 发现新功能 → 更新菜单地图
□ 发现新路由 → 更新菜单地图
□ 发现新组件 → 更新测试用例
□ 发现新理解 → 更新相关说明文档
□ 确保文档与实际代码一致
```

---

## ⚠️ 问题分析检查清单（重要！）

发现问题时，必须执行以下检查：

```
□ 1. 理解问题本质
   - 这是什么类型的问题？（URL格式/路由/组件/数据...）
   - 问题的技术根因是什么？

□ 2. 检查同类所有
   - 搜索项目中所有类似的问题点
   - 使用 grep 批量检查
   - 示例：发现一个 URL 错误 → 检查所有 URL

□ 3. 举一反三
   - 如果这是路由问题，同类路由是否都有问题？
   - 如果这是组件问题，其他组件是否也有问题？
   - 如果这是格式问题，所有格式是否都正确？

□ 4. 更新相关文档
   - 找到问题的根因后，更新所有相关文档
   - 确保文档与代码一致
```

### 问题分析案例

**案例：发现 URL 使用了错误的 hash 格式**

| 检查项 | 操作 |
|--------|------|
| 问题类型 | URL 格式错误 |
| 技术根因 | 系统使用 history 模式，不是 hash 模式 |
| 同类检查 | 搜索所有 `#/` 开头的 URL |
| 结果 | 发现多个文档使用错误格式，已全部修复 |

---

## 通用测试用例清单

### 1. 功能开关测试

| 用例ID | 功能 | 预期行为 | 测试方法 |
|--------|------|----------|----------|
| TC-F001 | 开关默认状态 | 检查默认值是否符合产品设计 | 检查初始值 |
| TC-F002 | 开关开启行为 | 开启时功能正常 | 验证开启后行为 |
| TC-F003 | 开关关闭行为 | 关闭时功能停止 | 验证关闭后行为 |
| TC-F004 | 开关切换行为 | 切换不触发不必要请求 | 验证切换逻辑 |

### 2. 边界条件测试

| 用例ID | 功能 | 预期行为 | 测试方法 |
|--------|------|----------|----------|
| TC-B001 | 组件销毁 | 自动释放资源 | 验证 beforeUnmount 清理 |
| TC-B002 | 页面切换 | 正确切换上下文 | 验证页面切换逻辑 |
| TC-B003 | 数据为空 | 显示空状态 | 验证空数据处理 |
| TC-B004 | 数据超限 | 正确处理边界 | 验证边界值处理 |

### 3. UI 交互测试

| 用例ID | 功能 | 预期行为 | 测试方法 |
|--------|------|----------|----------|
| TC-U001 | UI 状态显示 | 与实际状态一致 | 验证 UI 与逻辑同步 |
| TC-U002 | 实时更新 | 反映数据变化 | 验证数据同步 |
| TC-U003 | 用户反馈 | 操作有响应 | 验证反馈机制 |

### 4. 资源加载测试（通用）

| 用例ID | 功能 | 预期行为 | 测试方法 |
|--------|------|----------|----------|
| TC-R001 | 远程资源加载失败 | 显示默认占位符 | 验证 fallback |
| TC-R002 | 图片加载失败 | 显示默认图标 | 验证错误处理 |
| TC-R003 | API 请求失败 | 显示错误提示 | 验证错误处理 |
| TC-R004 | 状态互斥 | 不同时显示多种状态 | 验证条件互斥 |

---

## 问题分析方法论

### 1. 为什么问题没测出来？

| 原因分类 | 具体表现 | 改进方向 |
|---------|---------|---------|
| **测试粒度太粗** | 只检查元素存在，不检查功能 | 增加功能验证 |
| **缺少边界测试** | 不测试空值、极限值 | 添加边界测试用例 |
| **缺少状态测试** | 不验证状态切换 | 增加状态验证 |
| **缺少错误测试** | 不模拟错误场景 | 增加错误场景测试 |
| **缺少 fallback 验证** | 不验证降级方案 | 增加 fallback 验证 |

### 2. 常见问题根因分析

| 问题类型 | 典型表现 | 根因分析 |
|---------|---------|---------|
| **初始状态错误** | 首次打开显示错误状态 | 初始化逻辑有问题 |
| **状态不同步** | UI 与数据不一致 | 状态管理有问题 |
| **条件渲染失效** | v-if 动态状态不更新 | 条件判断不包含动态状态 |
| **资源泄露** | 组件销毁后资源仍存在 | 缺少清理逻辑 |
| **性能问题** | 操作响应慢 | 可能是请求/渲染问题 |

---

## Vue 组件规范

### 1. 资源加载 fallback 规范

```vue
<!-- 方案1：v-if 双重条件 -->
<img v-if="record.src && !record.loadError" :src="record.src" @error="record.loadError = true" />
<placeholder v-if="!record.src || record.loadError" />

<!-- 方案2：v-show 用于 fallback -->
<img v-if="record.src" :src="record.src" @error="record.loadError = true" />
<placeholder v-show="!record.src || record.loadError" />

<!-- 方案3：计算属性确保互斥 -->
<loading v-if="status === 'loading'" />
<success v-else-if="status === 'success'" />
<error v-else-if="status === 'error'" />
```

### 2. 资源清理规范

```javascript
// 确保组件销毁时释放资源
beforeUnmount() {
    // 停止定时器
    this.stopTimers();
    
    // 取消网络请求
    this.cancelRequests();
    
    // 关闭 WebSocket 连接
    this.closeWebSocket();
    
    // 停止 Watch 流
    this.stopWatch();
}
```

### 3. 状态管理规范

```javascript
// 使用计算属性确保状态互斥
const displayStatus = computed(() => {
    if (loading.value) return 'loading';
    if (error.value) return 'error';
    return 'success';
});

// 或使用互斥的标志位
const status = ref('idle'); // idle | loading | success | error
```

---

## 测试验证方法

### 1. 资源加载验证

```bash
# 验证资源加载 fallback
agent-browser eval "
(function() {
    const elements = document.querySelectorAll('.resource-item');
    const results = [];
    elements.forEach(el => {
        const img = el.querySelector('img');
        const placeholder = el.querySelector('.placeholder');
        results.push({
            hasImg: !!img,
            imgVisible: img && img.offsetParent !== null,
            placeholderVisible: placeholder && placeholder.offsetParent !== null,
            isCorrect: !(img && img.offsetParent !== null && placeholder && placeholder.offsetParent !== null)
        });
    });
    return results;
})()
"
```

### 2. 状态互斥验证

```bash
# 验证状态互斥
agent-browser eval "
(function() {
    const states = ['loading', 'success', 'error'];
    let visibleCount = 0;
    let visibleState = '';
    
    for (const state of states) {
        const el = document.querySelector('.' + state);
        if (el && el.offsetParent !== null) {
            visibleCount++;
            visibleState = state;
        }
    }
    
    return {
        visibleCount,
        visibleState,
        isMutuallyExclusive: visibleCount <= 1
    };
})()
"
```

### 3. 资源泄露验证

```bash
# 验证组件销毁后资源释放
# 方法1：检查控制台是否有未清理的定时器/请求
agent-browser console | grep -iE "timer|request|websocket"

# 方法2：验证离开页面后网络请求停止
# 在页面时监控网络
# 离开页面后再次监控
# 对比请求数量
```

### 4. 功能开关验证

```bash
# 验证开关默认状态
agent-browser eval "
(function() {
    const switchEl = document.querySelector('[role=switch]');
    return {
        exists: !!switchEl,
        checked: switchEl?.getAttribute('aria-checked') === 'true',
        defaultBehavior: '应符合产品设计'
    };
})()
"

# 验证切换行为（切换不应触发请求）
agent-browser eval "
(function() {
    // 记录切换前的请求数
    const beforeCount = window.performance?.getEntriesByType('resource')?.length || 0;
    
    // 切换开关
    const switchEl = document.querySelector('[role=switch]');
    switchEl?.click();
    
    // 等待短时间
    const afterCount = window.performance?.getEntriesByType('resource')?.length || 0;
    
    return {
        beforeCount,
        afterCount,
        hasUnnecessaryRequest: afterCount > beforeCount
    };
})()
"
```

---

## 测试用例总结模板

每次深度测试完成后，使用以下模板总结：

```markdown
## 测试用例总结

### 测试范围
- 测试模块：
- 测试环境：
- 测试数据：

### 发现的问题

| 问题 | 严重程度 | 根因 | 修复方案 |
|------|---------|------|----------|
|      |          |      |          |

### 遗漏的测试点

| 测试点 | 原因 | 改进建议 |
|--------|------|----------|
|        |      |          |

### 代码规范改进

| 规范类型 | 具体建议 |
|---------|---------|
| 状态管理 |          |
| 资源清理 |          |
| 条件渲染 |          |

### 相关文件

- 测试脚本：
- 被测组件：
- 测试报告：
```

---

## 相关文档

- [UI审美检查清单](../AGENTS.md#ui审美检查清单)
- [深度测试规范](../AGENTS.md#深度测试规范)
- [测试用例位置](./README.md)
