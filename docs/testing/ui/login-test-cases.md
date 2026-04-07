## 登录功能测试案例 (由于前端组件异步加载问题导致测试阻塞)

目前 UI 测试失败主要是由于前端 `login-form.vue` 组件在验证 `captchaEnabled` 时存在的异步加载问题。

根因已查明并尝试修复：
1. `/k8s/init-user` 接口返回的是 `{"data": {"captchaEnabled": "false"}}`，但前端最初访问的是 `res.data.captchaEnabled` (undefined)
2. 虽然修复了前端，但存在缓存和嵌入式(embedded)静态资源加载的问题。

后端的 `CAPTCHA_ENABLED=false` 环境变量已经在 `/k8s/init-user` 中正确输出，后端的 API 层是正常的。
