# AquaDX 协议等价性清单

> 本文件以 `/home/ubuntu/AquaDX` 当前源码为基准，记录 MaiGoDX 的 ALL.Net、AimeDB 与 maimai2 对照结果。每项必须有实现、测试和提交记录后才能标记为完成。

## 基线

AquaDX 的 maimai2 DSL 声明了 58 个 API。MaiGoDX 的加密端点路由表目前包含绝大部分既有端点，但仍缺少 16 个 AquaDX 命名端点；此外，已登记端点不代表响应模型与持久化语义完全一致，必须逐项核对。

## 已发现并已修复

| 范围 | 差异 | MaiGoDX 修复 |
|---|---|---|
| ALL.Net PowerOn | AquaDX 默认返回 `/g/{game}/{version}/`；仅在显式开启 Keychip 检查时返回 `/gs/{token}/...` | `allnet_check_keychip=false` 为默认值；开启后保留 `/gs` 会话保护，并在分发前改写为 `/g` |
| ALL.Net Secure 路由 | AquaDX 验证 `/gs/{token}` 后将请求改写为 `/g` 再进入 maimai 控制器 | MaiGoDX 已进行同样路径改写 |
| PowerOn 诊断 | AquaDX 日志会记录响应；游戏地址错误难以定位 | MaiGoDX PowerOn 日志已输出 `route=` 和完整 `uri=` |
| AimeDB Felica LookupV2 | AquaDX 将 IDm 转为 20 位十进制 Access Code，查卡后返回外部用户 ID 与卡号字段 | MaiGoDX 已实现同等 IDm 转换、数据库查卡和响应字段 |

## AquaDX 已声明但 MaiGoDX 路由表尚未登记的 maimai2 API

| 优先级 | API | AquaDX 语义概览 | 当前状态 |
|---:|---|---|---|
| P1 | `CreateToken` | 创建会话/令牌相关请求 | 待核对并实现 |
| P1 | `GetGameKaleidxScope` | Kaleidx 游戏侧范围数据 | 待实现 |
| P1 | `GetGameNationalData` | 全国数据 | 待实现 |
| P1 | `GetGameTournamentInfo` | 锦标赛信息 | 待实现 |
| P1 | `GetUserRegion` | 用户地区信息 | 待实现 |
| P1 | `GetUserRivalData` | 对手档案 | 待实现 |
| P1 | `GetUserRivalMusic` | 对手乐曲数据 | 待实现 |
| P1 | `GetUserShopStock` | 商店库存 | 待实现 |
| P2 | `CMGetSellingCard` | CardMaker 售卡信息 | 待实现 |
| P2 | `CMUpsertUserPrintlog` | CardMaker 打印日志 | 待实现 |
| P2 | `GetTransferFriend` | 好友转移 | 待实现 |
| P2 | `GetUserFriendBonus` | 好友奖励 | 待实现 |
| P2 | `GetUserFriendCheck` | 好友检查 | 待实现 |
| P2 | `GetUserNewItem` | 单项新道具 | 待实现 |
| P2 | `GetUserNewItemList` | 新道具列表 | 待实现 |
| P2 | `UserFriendRegist` | 好友登记 | 待实现 |

## 仍需核对的既有路由语义

| 范围 | AquaDX 基准 | MaiGoDX 风险点 | 状态 |
|---|---|---|---|
| PowerOn 地址 | 使用 `AllNet-Forwarded-From`、服务器主机配置、TLS 与隐藏端口选项构造 `uri` | MaiGoDX 当前使用请求 Host 并固定 HTTP；需改为数据库配置且覆盖 80/443 部署情形 | 待实现 |
| Keychip 策略 | `check-keychip` 和 permissive testing 均为显式开关 | MaiGoDX 机台登记逻辑目前始终要求 Keychip 存在 | 待配置化 |
| PowerOn 版本校验 | AquaDX 拒绝 `ver < 1.0` | MaiGoDX 尚未严格拒绝 | 待实现 |
| GameSetting 默认值 | AquaDX 使用固定的 2020 重启时间，并以空 URI 禁用资源/下载服务器 | MaiGoDX 旧数据库中的重启时间可能为空；需迁移策略 | 待实现 |
| AimeDB Keychip 策略 | AquaDX 在测试宽松模式下可接受未知 Keychip | MaiGoDX 目前始终拒绝未知 Keychip（除 type 0x13） | 待配置化 |
| AimeDB 卡片语义 | AquaDX 使用 Card 外部 ID、读取时不自动创建卡 | MaiGoDX 使用 UserCard/GameUserID；需逐请求核对注册、重复注册与查卡结果 | 待核对 |
| 加密端点 | AquaDX 对全部声明 API 与游戏变体计算 MD5 | MaiGoDX 的端点名单缺少 16 项，导致这些端点无法解密 | 待实现 |
| 响应模型 | AquaDX 通过 Kotlin 数据模型输出字段、空数组和错误码 | MaiGoDX 多处使用通用 no-op；需逐 API 替换 | 待实现 |

## 验收条件

1. 所有 AquaDX maimai2 API 都进入 MaiGoDX 的路由和 MD5 加密端点清单。
2. 每个 API 都有与 AquaDX 一致的请求必填字段、响应顶层结构、空值语义和持久化行为。
3. ALL.Net PowerOn 的 `g`/`gs`、版本、Keychip、主机、端口和 TLS 行为均可由数据库配置控制。
4. AimeDB 对 type `0x01/0x04/0x05/0x09/0x0b/0x0d/0x0f/0x11/0x13/0x64/0x66` 的响应与 AquaDX 对照通过。
5. 使用 KanadeDX 完成 PowerOn、标题/游戏服务器连接、Aime 读卡、建档和成绩上传的端到端验证。
