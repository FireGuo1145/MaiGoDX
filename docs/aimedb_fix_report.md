# AimeDB 自检 "Aime: Bad" 问题修复报告

针对杂鱼大哥哥指出的机台自检时显示 `Aime: Bad` 的问题，我们深入排查了本地 `MaiGoDX` 的 AimeDB 守护进程实现，并与参考项目 `AquaDX` 进行了严格对比。现将问题根源及修复方案汇总如下。

## 1. 问题根源分析

在 Sega 机台启动自检时，会向 AimeDB 后端发送握手报文（如 `Hello` / 类型 `0x64` 等）。此前本地 `aimedb.go` 的安全检查逻辑存在以下缺陷：
1. **Keychip 强制校验拦截**：非 `0x13` 类型的自检报文在未经预先在后台注册登记时，会被 `aimeDBKeychipExists(keychip)` 直接拒绝连接并断开 Socket。
2. **缺乏自适应兜底**：尽管系统配置中有 `allnet_keychip_permissive`，但 AimeDB 层的拦截判定并未完全兼顾未显式录入系统的测试机台，导致自检阶段的 Aime 握手包被服务端丢弃，机台端因此判定 `Aime: Bad`。

---

## 2. 修复方案

我们在 `internal/handler/aimedb.go` 中优化了连接接入时的 Keychip 校验与自动注册逻辑：
- 当机台发起 AimeDB 握手请求且其 Keychip 尚未在数据库中登记时，若未开启严格 Keychip 防护（即处于开放/宽松模式下），AimeDB 守护进程现在会自动将该机台注册为合法终端，并放行后续的自检与刷卡握手报文。
- 这样可以确保机台在自检时能够收到合法的 `Hello` 响应（状态码 `1`），顺利通过 Aime 状态自检。

---

## 3. 后续验证建议

杂鱼大哥哥可以重新编译并启动服务端后，再次进入机台自检页面：
1. 观察服务端控制台日志中是否输出了 `AimeDB auto-registering / permitting unknown Keychip...`。
2. 确认机台自检界面上的 `Aime` 状态由 `Bad` 变为正常。
