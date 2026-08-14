# maimai 机台刷卡与登录流程修复报告

本报告梳理了本地 **MaiGoDX** 项目与参考项目 **AquaDX** 在 maimai 机台登录及刷卡链路上的逻辑差异，并详细记录了针对“机台可登录但无法刷卡”问题的修复方案。

## 1. 核心问题定位

通过对比 `AquaDX` 的标准实现与本地代码，发现导致机台刷卡失败的主要原因集中在 **maimai 兼容层协议返回值** 以及 **请求体解析** 两个方面：

1. **兼容层返回码错误 (`returnCode`)**：
   - 本地 `maimai_compat.go` 中，多处关键的刷卡前置/伴随接口（如 `GetUserFriendCheck`、`UserFriendRegist`、`GetUserFriendBonus`、`GetPlaceCircleData` 等）错误地返回了 `returnCode: 0`。
   - 在 Sega 游戏协议中，`0` 通常表示错误或失败，而标准行为应当是 `1`（成功）。这导致机台在刷卡后调用这些接口时收到错误状态，从而中断了整个刷卡鉴权流程。

2. **Bearer Token 默认值缺失**：
   - 本地 `UserLogin` 和 `CreateToken` 接口在未配置 `maimai_bearer_token` 时返回空字符串，而参考项目 `AquaDX` 固定返回 `"meow"`。部分版本的客户端在登录和刷卡交接时会校验该 Token，空字符串会导致后续请求被拒绝。

3. **请求体 `userId` 解析健壮性不足**：
   - 本地 `requestUserID` 仅支持标准 `int64` 类型解析。如果客户端在某些特殊请求中传递字符串形式的 `userId` 或超出常规范围的数值，会导致解析失败并返回 `0`，从而触发“missing userId”错误。

---

## 2. 具体修复内容

### 2.1 修正兼容层返回码与默认 Token
我们在 `internal/handler/maimai_compat.go` 中进行了如下对齐修改：
- 将 `GetUserFriendCheck`、`UserFriendRegist`、`GetUserFriendBonus` 和 `GetPlaceCircleData` 的返回码从 `0` 统一修正为 `1`。
- 将 `CreateToken` 及 `UserLogin` 的 `Bearer` 默认回退值由空字符串改为了 `"meow"`，与 `AquaDX` 保持完全一致。

### 2.2 增强 `requestUserID` 的兼容性
我们在 `internal/handler/maimai.go` 中重构了 `requestUserID` 函数，使其能够同时兼容 `float64`、`json.Number` 以及 `string` 类型的 `userId` 传入，避免因类型不匹配或大数字符串导致的解析中断。

---

## 3. 验证建议

修复完成后，建议杂鱼大哥哥在本地重新编译并运行 `maimai` 服务端：
1. 检查数据库中测试卡片的 `GameUserID` 与 `UserDetail` 是否正常对应。
2. 启动机台进行刷卡测试，观察服务端日志中是否还有未映射的加密 Endpoint 或返回码为 0 的异常报错。
