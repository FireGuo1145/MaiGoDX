# AquaDX 协议等价性清单

> 本文件以 `/home/ubuntu/AquaDX` 当前源码为基准，记录 MaiGoDX 的 ALL.Net、AimeDB 与 maimai2 对照结果。每项必须有实现、测试和提交记录后才能标记为完成。

## 基线

AquaDX 的 maimai2 DSL 声明了 58 个 API。MaiGoDX 的明文与 MD5 加密端点名称表现已覆盖全部 58 个 AquaDX API；此外，已登记端点不代表响应模型与持久化语义完全一致，必须逐项核对。

## 已发现并已修复

| 范围 | 差异 | MaiGoDX 修复 |
|---|---|---|
| ALL.Net PowerOn | AquaDX 默认返回 `/g/{game}/{version}/`；仅在显式开启 Keychip 检查时返回 `/gs/{token}/...` | `allnet_check_keychip=false` 为默认值；开启后保留 `/gs` 会话保护，并在分发前改写为 `/g` |
| ALL.Net Secure 路由 | AquaDX 验证 `/gs/{token}` 后将请求改写为 `/g` 再进入 maimai 控制器 | MaiGoDX 已进行同样路径改写 |
| PowerOn 诊断 | AquaDX 日志会记录响应；游戏地址错误难以定位 | MaiGoDX PowerOn 日志已输出 `route=` 和完整 `uri=` |
| AimeDB Felica LookupV2 | AquaDX 将 IDm 转为 20 位十进制 Access Code，查卡后返回外部用户 ID 与卡号字段 | MaiGoDX 已实现同等 IDm 转换、数据库查卡和响应字段 |
| maimai2 加密路由表 | AquaDX 为全部 API 计算 MD5 加密路径 | MaiGoDX 已补齐所有 58 个 AquaDX API 名称，支持明文和 MD5 路由解析 |
| 第一批缺失 API | AquaDX 提供售卡、地区、对手、商店、Kaleidx、全国/赛事、好友与新道具端点 | MaiGoDX 已补齐端点、售卡/地区模型、响应结构和端点回归测试 |
| 特殊端点 | AquaDX 控制器另注册头像、照片、CardMaker 卡片、收藏道具与音乐排行处理器 | MaiGoDX 已补齐头像读写、嵌套照片分块上传、CM 用户卡片、收藏道具与七日独立玩家乐曲排行 |
| CardMaker 预览与范围 | AquaDX 的 CM 预览使用精简字段，Kaleidx 默认解锁前六门 | MaiGoDX 已改为同形 CM 响应，并实现基于通关状态的 Kaleidx 闸门解锁 |

## maimai2 路由覆盖

MaiGoDX 的路由表现已包含全部 58 个 AquaDX DSL API，并补充控制器特殊注册的 `GetUserPortrait`、`UploadUserPortrait`、`CMGetUserCard` 与 `CMGetUserCardPrintError`；这些端点均可参与 SDGA/SDEZ/SDGB 的 MD5 加密端点计算。新增端点中，`CMGetSellingCard` 使用 `GameSellingCard` 数据表；`GetUserRegion` 与 `UserLogin` 使用 `UserRegion` 数据表；对手乐曲、售卡、商店库存和静态游戏数据均使用 AquaDX 同形响应结构。

## 仍需核对的既有路由语义

| 范围 | AquaDX 基准 | MaiGoDX 风险点 | 状态 |
|---|---|---|---|
| PowerOn 地址 | 使用 `AllNet-Forwarded-From`、服务器主机配置、TLS 与隐藏端口选项构造 `uri` | MaiGoDX 当前使用请求 Host 并固定 HTTP；需改为数据库配置且覆盖 80/443 部署情形 | 待实现 |
| Keychip 策略 | `check-keychip` 和 permissive testing 均为显式开关 | MaiGoDX 机台登记逻辑目前始终要求 Keychip 存在 | 待配置化 |
| PowerOn 版本校验 | AquaDX 拒绝 `ver < 1.0` | MaiGoDX 尚未严格拒绝 | 待实现 |
| GameSetting 默认值 | AquaDX 使用固定的 2020 重启时间，并以空 URI 禁用资源/下载服务器 | MaiGoDX 旧数据库中的重启时间可能为空；需迁移策略 | 待实现 |
| AimeDB Keychip 策略 | AquaDX 在测试宽松模式下可接受未知 Keychip | MaiGoDX 目前始终拒绝未知 Keychip（除 type 0x13） | 待配置化 |
| AimeDB 卡片语义 | AquaDX 使用 Card 外部 ID、读取时不自动创建卡 | MaiGoDX 使用 UserCard/GameUserID；需逐请求核对注册、重复注册与查卡结果 | 待核对 |
| 加密端点 | AquaDX 对全部声明 API 与游戏变体计算 MD5 | MaiGoDX 路由表已覆盖全部 AquaDX API；仍需为每个 API 增加哈希路由回归样本 | 进行中 |
| 响应模型 | AquaDX 通过 Kotlin 数据模型输出字段、空数组和错误码 | 主要特殊端点已对齐；仍需逐 API 审核普通用户数据、成绩上传与 `UpsertUserAll` 细节 | 进行中 |

## 验收条件

1. 所有 AquaDX maimai2 API 都进入 MaiGoDX 的路由和 MD5 加密端点清单。
2. 每个 API 都有与 AquaDX 一致的请求必填字段、响应顶层结构、空值语义和持久化行为。
3. ALL.Net PowerOn 的 `g`/`gs`、版本、Keychip、主机、端口和 TLS 行为均可由数据库配置控制。
4. AimeDB 对 type `0x01/0x04/0x05/0x09/0x0b/0x0d/0x0f/0x11/0x13/0x64/0x66` 的响应与 AquaDX 对照通过。
5. 使用 KanadeDX 完成 PowerOn、标题/游戏服务器连接、Aime 读卡、建档和成绩上传的端到端验证。
