# 表单控件组件

## `CustomCheckbox`

源文件：`w7panel-ui/src/components/custom-checkbox.vue`

功能：支持自定义选中值和未选中值的 Checkbox，适合字段值不是简单 boolean 的表单。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `modelValue` | boolean/string/number | `false` | 当前值 |
| `checkedValue` | boolean/string/number | `true` | 选中时写入的值 |
| `uncheckedValue` | boolean/string/number | `false` | 取消选中时写入的值 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `update:modelValue` | checkedValue/uncheckedValue | 值变化时触发 |

使用示例：

```vue
<CustomCheckbox
  v-model="form.storageRwMode"
  checked-value="ReadWriteMany"
  unchecked-value="ReadWriteOnce"
>
  多写
</CustomCheckbox>
```

## `CronJob`

源文件：`w7panel-ui/src/components/cron-job.vue`

功能：用表单方式编辑 5 位 cron 表达式，支持每月、每周、每天、每小时、每隔 N 日、每隔 N 小时、每隔 N 分钟。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `value` | string | - | cron 表达式，例如 `0 2 * * *` |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `change` | string | 用户修改后输出新的 cron 表达式 |

使用示例：

```vue
<CronJob :value="form.cron" @change="value => form.cron = value" />
```

## `HealthProbe`

源文件：`w7panel-ui/src/components/health-probe.vue`

功能：健康检查配置组件，支持 liveness/readiness/startup 等 probe 数据编辑。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `data` | object | - | probe 数据 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `returnData` | object | 输出健康检查配置 |

## 其它表单抽屉

| 组件 | 源文件 | Props | Events | 说明 |
|------|--------|-------|--------|------|
| `DcformDrawer` | `src/components/dcform-drawer.vue` | `show` | `close(refreshList)` | Docker Compose 动态表单抽屉 |

## 维护要求

- 修改组件 Props、Events、Ref 方法或对外行为时，同步更新本文。
- 涉及 Wujie 事件的组件，只在本文记录组件职责，事件协议维护到 [wujie-events.md](./wujie-events.md)。
- 返回 [组件说明入口](./components.md)。
