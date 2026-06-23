# Hangout — 产品设计与技术规格（Spec）

- **状态**: 已批准设计，待评审 spec
- **日期**: 2026-06-23
- **作者**: 产品 + 工程协作（brainstorming 产出）
- **暂定名**: Hangout
- **一句话定位**: 面向熟人圈子的社交日历与活动协调 Web 应用，灵感来自飞书日历，UI 风格对标 Luma。

> **变更记录（2026-06-23 架构修订）**：认证改为 Supabase Auth 托管；保留 Gin 后端承载核心业务逻辑，Gin 仅验签 JWT。因开发者为个人主体（无法注册微信开放平台网站应用、无法完成微信认证），**微信相关功能全部移除**（Mock 微信登录、真实微信 OAuth、JS-SDK 自定义分享卡片）。部署策略更新为：**前后端同机部署在香港阿里云 ECS**，Cloudflare 仅负责 DNS 解析，默认不开代理；正式域名上线前允许使用 ECS 公网 IP 做小范围内测。数据库/Auth 仍采用阿里云 Supabase，最终地域以上线前实测延迟为准。受影响章节：D1、D2、第 3 节、第 4 节、5.1、第 6/7/9/10 节、第 8 节。

---

## 1. 背景与目标

### 1.1 核心痛点
- 约活动的信息散落在微信群聊天记录里，容易被淹没。
- 不知道朋友什么时候有空。
- 凑不齐人、响应慢。

### 1.2 产品形态
Web 应用（移动端优先）。用户通过微信群分享活动链接，群成员点击链接进入参与（报名）。

### 1.3 成功标准（North Star）
核心闭环可用：**建活动 → 分享链接 → 群成员点进来报名 → 自动生成个人日程**。
衡量：活动从创建到凑齐人的平均时长、活动报名转化率（点击→报名）。

### 1.4 设计风格
- Luma 风格：大圆角卡片、克制的留白、活动封面/色块、清晰的层级、轻动效。
- 移动端优先（mobile-first），桌面端自适应增强。

---

## 2. 关键设计决策（已确认）

| # | 议题 | 决策 |
|---|------|------|
| D1 | 认证方式 | **Supabase Auth 托管**：邮箱+密码注册/登录/重置密码为 P0；前端直连 Supabase 完成认证，Gin 后端仅**验签** Supabase 签发的 JWT。微信登录因个人主体限制移除（见变更记录） |
| D2 | 分享卡片 | P0 提供标准 OG / Twitter meta，浏览器与多数平台可正常抓取。**微信内展示效果不做保证**：可能显示默认链接预览、普通链接或风险提示页；微信内自定义卡片（JS-SDK）需企业主体 + 微信认证，已移除 |
| D3 | RSVP 状态 | 三态：`going` / `not_going` / `maybe`；`maybe` 在日历显示为暂定忙 |
| D4 | MVP 首切片 | 核心闭环优先：活动 + 分享 + RSVP + 自动日程 |
| D5 | 匿名访问 | 未登录可**查看**活动详情（只读），**报名**需登录 |
| D6 | 群活动可见性 | 仅链接 / 群内可见（隐私优先） |
| D7 | 账号合并 | P0 仅邮箱单一登录方式，无需账号合并；多登录方式绑定/合并待引入第三方登录后再议 → P2 |
| D8 | 通知 | P0 站内通知列表；Web Push 推送 → P2 |
| D9 | 容量已满 | P0 直接阻止报名；候补队列（waitlist）→ P1 |
| D10 | 收件箱（统一待办） | P0 统一收件箱，集中处理活动邀请、好友申请、群组邀请等「待我处理」事项。被邀请进入 `invited` 待回应状态；批阅时一次交互完成报名 + 是否加入日历 |
| D11 | 报名联动日历 | 报名 going/maybe 时默认勾选「加入我的日历（busy_only）」，用户可取消勾选并当场切换可见性 |
| D12 | 日程冲突 | 软提示策略：允许时间重叠的多条日程并存，系统只检测并提示冲突（报名/建日程时 + 日历标记），不阻止 |
| D13 | 数据库环境 | **统一使用 PostgreSQL**：开发/测试使用 Supabase local，生产使用阿里云 Supabase Postgres；不使用 SQLite 作为开发替代，避免 uuid、jsonb、trigger、partial index 等能力差异导致迁移风险 |

---

## 3. 功能列表（按优先级）

### P0 — MVP（核心闭环可用）
1. **认证**
   - 邮箱+密码注册/登录/登出（Supabase Auth 托管）
   - 邮箱重置密码（Supabase 邮件服务）
   - 前端直连 Supabase 完成认证拿到 JWT；Gin 鉴权中间件**验签** Supabase JWT（HS256，用 project JWT secret 或 JWKS），不自行签发
2. **活动**
   - 创建活动（个人发起 / 群发起）
   - 字段：标题、类型（攀岩/吃饭/出行/游戏/其他）、开始时间、结束时间、地点、人数上限（可选）
   - 活动详情页（未登录可只读查看）
   - 唯一分享链接（slug）+ 标准 OG meta 标签（微信内展示不做保证）
   - 取消活动（仅发起人；群活动发起人或群主）
3. **RSVP（报名）**
   - 状态：`invited`（被邀请未回应）/ `going` / `not_going` / `maybe`
   - 自己点链接进来报名（source=self）直接进入 going/maybe/not_going
   - 被邀请（source=invited）先进入 `invited` 待回应，落入收件箱
   - 报名 `going`/`maybe` 时默认勾选「加入我的日历（busy_only）」，可取消勾选、当场切可见性
   - 勾选后 → 自动 upsert 关联日程；取消报名 / 改 not_going / 取消勾选 → 删除对应日程
   - 容量已满阻止新的 going
4. **个人日程**
   - 手动创建/编辑/删除个人日程（独立于活动）
   - 三种可见性：public / busy_only / private
   - 活动自动日程默认 busy_only，用户可手动改
   - **日程冲突软提示**：允许同一时间存在多条重叠日程（如活动 A 7:00–9:00 与 B 8:00–10:00 可同时参加）；系统仅在报名/创建/编辑日程时检测时间重叠并提示「与 X 个已有日程冲突」，不阻止操作；日历上对冲突时段做视觉标记
5. **日历视图**
   - 月视图，日期标记点区分事件类型
   - 颜色编码：蓝=群活动，绿=个人日程，灰=忙（busy_only）
   - 点击日期展开当天事件
6. **好友**
   - 发送好友申请 / 同意 / 拒绝（双向确认）
   - 好友列表
   - 点对点邀请好友参加活动
7. **群组**
   - 创建群 / 邀请加入 / 退群
   - 群成员列表
   - 任一群成员可在群内发起活动
8. **统一收件箱（待办中心）**
   - 集中展示「待我处理」事项：活动邀请（invited）、好友申请、群组邀请
   - 活动邀请行内批阅：报名 / 待定 / 不参加 + 「加入日历」勾选（默认勾选）+ 可见性切换
   - 时间冲突提示：邀请活动与已有日程撞车时标红
   - 好友申请 / 群组邀请：行内同意 / 拒绝
   - 已处理项移出待办区（灰显或归档）；未读计数
9. **通知（提醒入口）**
   - 站内通知列表（好友申请、活动邀请、活动变更/取消）+ 未读计数
   - 职责区分：通知=提醒触点（红点/未来推送），收件箱=待办处理

### P1 — 增强
- 日历筛选：全部 / 我的群 / 好友
- 查看好友日程（按可见性过滤）
- 容量候补队列（waitlist），有人退出自动递补
- 活动评论/留言
- 活动编辑（已存在的字段修改 + 变更通知参与者）
- 周视图 / 日视图
- 群主转让

### P2 — 锦上添花
- Web Push 推送
- 推荐时段（找朋友共同有空的时间）
- 活动复盘/相册
- 重复活动（周期性日程）
- 日历订阅（iCal 导出）
- 数据统计面板

> 注：微信生态相关能力（OAuth 登录、JS-SDK 自定义分享卡片、模板消息、账号绑定/合并）需企业主体 + 微信认证，当前个人主体阶段不纳入规划；将来注册公司主体后再评估引入。

---

## 4. 数据模型设计

> ORM: GORM。除 `users` 外，所有表含 `id (uint, PK)`、`created_at`、`updated_at`、`deleted_at (软删除, gorm.DeletedAt)`，下文不再重复列出。**`users.id` 为 `uuid`，与 Supabase `auth.users.id` 对齐**；因此所有指向用户的外键（`*→users`，即下文各 `user_id`/`owner_id`/`requester_id`/`addressee_id`/`inviter_id`/`invitee_id`）类型均为 `uuid`。时间统一 UTC 存储，前端按 Asia/Shanghai 展示。
>
> **认证与用户表关系**：用户的认证凭据（邮箱、密码、第三方身份）由 Supabase Auth 在 `auth` schema（`auth.users` / `auth.identities`）托管，应用不自建认证表。本表 `public.users` 是业务侧用户档案（profile），`id` 外键引用 `auth.users.id`；用户在 Supabase 注册成功后，通过数据库 trigger（`on auth.users insert`）自动 upsert 一条 `public.users` 记录。

### 4.1 users（业务档案 / profile）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uuid, PK | 等于 Supabase `auth.users.id`（外键引用） |
| display_name | string | 昵称 |
| avatar_url | string | 头像 URL（可空） |
| email | string, nullable | 邮箱副本（展示/搜索用，权威值在 `auth.users`） |
| status | string | `active` / `deactivated`（注销后显示「已注销用户」） |

> 密码、邮箱验证、第三方身份由 Supabase Auth 托管，本表不存 `password_hash`。

### 4.2 friendships
| 字段 | 类型 | 说明 |
|------|------|------|
| requester_id | uuid, FK→users | 申请方 |
| addressee_id | uuid, FK→users | 被申请方 |
| status | string | `pending` / `accepted` / `rejected` |
| 唯一约束 | (requester_id, addressee_id) | |

> 查询「是否好友」：存在一条 `accepted` 且 {requester,addressee} 命中该用户对（无序）。

### 4.3 groups
| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 群名 |
| description | string | 群简介（可空） |
| owner_id | uuid, FK→users | 群主 |
| invite_code | string, unique | 群邀请码（用于邀请链接） |
| status | string | `active` / `dissolved` |

### 4.4 group_members
| 字段 | 类型 | 说明 |
|------|------|------|
| group_id | uint, FK→groups | |
| user_id | uuid, FK→users | |
| role | string | `owner` / `member` |
| 唯一约束 | (group_id, user_id) | |

### 4.5 events
| 字段 | 类型 | 说明 |
|------|------|------|
| owner_id | uuid, FK→users | 发起人 |
| scope | string | `personal`（点对点好友邀请）/ `group`（群活动） |
| group_id | uint, FK→groups, nullable | scope=group 时有值 |
| title | string | 标题 |
| type | string | `climbing` / `dining` / `travel` / `gaming` / `other` |
| start_at | datetime | 开始时间 |
| end_at | datetime, nullable | 结束时间 |
| location | string | 地点（可空） |
| capacity | int, nullable | 人数上限（可空=不限） |
| status | string | `active` / `cancelled` / `expired`（过期由 start/end 派生或定时任务标记） |
| share_slug | string, unique | 分享链接短码 |

> **过期处理**：读取时若 `end_at`（无则 `start_at`）已过且状态非 cancelled，视为 `expired`（只读）。可选每日定时任务把过期活动状态落库为 `expired`。

### 4.6 event_participants
| 字段 | 类型 | 说明 |
|------|------|------|
| event_id | uint, FK→events | |
| user_id | uuid, FK→users | |
| rsvp | string | `invited`（被邀请未回应）/ `going` / `not_going` / `maybe` |
| source | string | `self`（自己点进来）/ `invited`（被邀请） |
| add_to_calendar | bool | 报名时是否联动生成日程（默认 true） |
| 唯一约束 | (event_id, user_id) | |

> **收件箱来源**：`rsvp=invited` 的参与记录即「待回应活动邀请」。发起人据此查看「谁还没回应」（P0 即可用）。

### 4.7 schedules（个人日程）
| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | uuid, FK→users | 所属用户 |
| title | string | 标题（手动日程用户填；活动日程取活动标题） |
| start_at | datetime | |
| end_at | datetime, nullable | |
| location | string | 可空 |
| visibility | string | `public` / `busy_only` / `private` |
| source | string | `manual` / `event` |
| event_id | uint, FK→events, nullable | source=event 时关联活动 |
| 唯一约束 | (user_id, event_id) where source=event | 防止重复生成 |

> **联动规则**：仅当参与记录 `rsvp ∈ {going, maybe}` 且 `add_to_calendar=true` 时，upsert 一条 source=event 的 schedule（默认 busy_only）；当 rsvp 改为 not_going/取消报名，或 add_to_calendar=false 时，删除该 schedule。用户手动改过的可见性需保留（upsert 时不覆盖用户已改的 visibility）。

### 4.8 group_invites
群组邀请，进入收件箱供被邀请人批阅（区别于用邀请码自助加入）。
| 字段 | 类型 | 说明 |
|------|------|------|
| group_id | uint, FK→groups | |
| inviter_id | uuid, FK→users | 邀请人 |
| invitee_id | uuid, FK→users | 被邀请人 |
| status | string | `pending` / `accepted` / `rejected` |
| 唯一约束 | (group_id, invitee_id) | |

### 4.9 notifications
| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | uuid, FK→users | 接收者 |
| type | string | `friend_request` / `friend_accepted` / `event_invite` / `event_changed` / `event_cancelled` / `group_invite` |
| payload | jsonb/text | 关联实体 id 与展示数据 |
| read_at | datetime, nullable | 已读时间 |

> **收件箱 vs 通知**：收件箱是「待办视图」，由可处理的源数据动态聚合 —— `event_participants(rsvp=invited)` + `friendships(status=pending, addressee=本人)` + `group_invites(status=pending, invitee=本人)`。notifications 是「提醒流」，承载红点/未读与未来推送。两者各自独立，不互相依赖。

---

## 5. API 接口设计

> 统一前缀 `/api`。鉴权用 **Supabase 签发的 JWT**（前端在 `Authorization: Bearer <token>` 携带，Gin 中间件验签）。响应体统一 `{ "data": ..., "error": null }`。错误码用 HTTP status + `{ "error": { "code", "message" } }`。

### 5.1 认证 Auth
> 注册/登录/登出/重置密码**由前端直连 Supabase Auth（supabase-js）完成**，后端不实现这些端点。Gin 侧只做 JWT 验签中间件 + 读取当前用户档案。

> **认证边界说明**：前端引入 `supabase-js` 的目的仅限于调用 Supabase Auth 能力（如 `signUp`、`signInWithPassword`、`resetPasswordForEmail`、`getSession`、`signOut`）。**前端不得使用 `supabase.from(...)` 或 `supabase.rpc(...)` 直接读写业务表/业务逻辑**；活动、RSVP、日程、好友、群组、收件箱、通知等业务数据统一通过 Gin 暴露的 `/api/*` 访问。

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| GET | /api/auth/me | 当前用户档案（验签 Supabase JWT 后查 `public.users`） | 是 |

> **验签方式**：用 Supabase project 的 JWT secret 验 HS256，或拉取项目 JWKS 验签。token 内 `sub` 即 `auth.users.id`，与 `public.users.id` 一致。前端登出由 supabase-js 清理本地 session，无需后端端点。

### 5.2 用户 Users
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/users/me | 个人资料 |
| PATCH | /api/users/me | 更新昵称/头像 |
| DELETE | /api/users/me | 注销（软删除 `public.users` 置 deactivated，并调用 Supabase Admin API 删除/禁用 `auth.users`） |
| GET | /api/users/search?q= | 按昵称/邮箱搜索（加好友用） |

### 5.3 好友 Friends
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/friends | 好友列表 |
| GET | /api/friends/requests | 收到/发出的申请 |
| POST | /api/friends/requests | 发送好友申请 `{addressee_id}` |
| POST | /api/friends/requests/:id/accept | 同意 |
| POST | /api/friends/requests/:id/reject | 拒绝 |
| DELETE | /api/friends/:userId | 删除好友 |
| GET | /api/friends/:userId/schedules?from=&to= | 查看好友日程（按可见性过滤，P1） |

### 5.4 群组 Groups
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/groups | 我的群列表 |
| POST | /api/groups | 创建群 |
| GET | /api/groups/:id | 群详情 + 成员 |
| PATCH | /api/groups/:id | 改群名/简介（群主） |
| DELETE | /api/groups/:id | 解散群（群主） |
| POST | /api/groups/:id/members | 通过邀请码加入 |
| POST | /api/groups/:id/invites | 邀请用户入群 `{user_ids}`（进收件箱） |
| POST | /api/groups/invites/:id/accept | 同意入群邀请 |
| POST | /api/groups/invites/:id/reject | 拒绝入群邀请 |
| DELETE | /api/groups/:id/members/:userId | 退群 / 移除成员 |
| GET | /api/groups/:id/events | 群活动列表 |

### 5.5 活动 Events
| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| POST | /api/events | 创建活动 | 是 |
| GET | /api/events/:slug | 活动详情（含参与者汇总） | 否（只读可匿名） |
| PATCH | /api/events/:id | 编辑活动（发起人，P1） | 是 |
| POST | /api/events/:id/cancel | 取消活动 | 是 |
| POST | /api/events/:id/rsvp | 报名/改状态 `{rsvp, add_to_calendar?, visibility?}` | 是 |
| DELETE | /api/events/:id/rsvp | 取消报名 | 是 |
| POST | /api/events/:id/invite | 邀请好友 `{user_ids}` | 是 |
| GET | /api/events/mine | 我发起/参与的活动 | 是 |

### 5.6 个人日程 Schedules
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/schedules?from=&to= | 我的日程（月视图区间） |
| POST | /api/schedules | 手动创建日程 |
| PATCH | /api/schedules/:id | 编辑（含改可见性） |
| DELETE | /api/schedules/:id | 删除（仅 manual；event 日程通过取消报名删） |
| GET | /api/schedules/conflicts?start=&end=&exclude_id= | 检测给定时间段与本人现有日程的重叠，返回冲突日程列表 |

> **冲突检测**：判定为 `existing.start < new.end && new.start < existing.end`（半开区间，相邻不算冲突）。`exclude_id` 用于编辑时排除自身。报名 `POST /events/:id/rsvp` 与建/改日程的响应体附带 `conflicts: [...]`，前端据此弹软提示，不阻断。

### 5.7 日历聚合 Calendar
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/calendar?from=&to=&filter= | 聚合月视图数据：我的日程 + 群活动 +（P1）好友。`filter`=all/groups/friends |

### 5.8 收件箱 Inbox（统一待办）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/inbox | 聚合待办：活动邀请(invited) + 好友申请(pending) + 群组邀请(pending) + 各类计数 |

> 收件箱本身不新增写操作；批阅动作复用各域已有端点（`/events/:id/rsvp`、`/friends/requests/:id/accept|reject`、`/groups/invites/:id/accept|reject`）。

### 5.9 通知 Notifications
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/notifications | 通知列表 + 未读数 |
| POST | /api/notifications/:id/read | 标记已读 |
| POST | /api/notifications/read-all | 全部已读 |

### 5.10 分享 OG 卡片
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /e/:slug | 服务端渲染含 OG/Twitter meta 的 HTML（标题、时间、报名人数），随后 hydrate 到前端详情页 |

---

## 6. 前端页面列表与核心交互

> 移动端优先，底部 Tab 导航；Luma 风格大卡片。

| 页面 | 路由 | 核心交互 |
|------|------|---------|
| 启动/登录 | `/login` | 邮箱登录、邮箱注册切换、忘记密码/重置（均经 Supabase Auth） |
| 日历（首页） | `/` | 月视图 + 标记点 + 颜色图例；点日期展开当天事件抽屉；冲突时段做视觉标记（如角标/叠层）；筛选 全部/我的群/好友（P1）；右下角「+」创建活动/日程 |
| 活动详情 | `/e/:slug` | 封面/类型色块、时间地点、报名人数与头像墙、RSVP 三态按钮 +「加入日历」勾选与可见性、邀请好友、分享按钮（复制链接，粘贴到微信群等渠道）；报名时若与已有日程冲突弹软提示（可继续）；未登录显示只读 + 登录后报名引导；满员禁用 going；发起人可见「未回应」名单 |
| 创建活动 | `/events/new` | 选发起范围（个人/群）、标题、类型选择器、时间选择、地点、人数上限；提交后跳详情并弹分享 |
| 我的活动 | `/events` | 我发起/参与，分「即将开始 / 已过期」 |
| 创建/编辑日程 | `/schedules/new`、`/schedules/:id` | 标题、时间、地点、可见性选择（public/busy_only/private 带说明）；选好时间后实时提示是否与现有日程冲突（不阻止保存） |
| 收件箱（待办） | `/inbox` | 统一处理活动邀请、好友申请、群组邀请；活动邀请行内批阅（报名/待定/不参加 + 加入日历勾选 + 可见性 + 时间冲突标红）；好友/群邀请行内同意/拒绝；已处理项灰显/归档；底部 Tab 带未读红点 |
| 好友 | `/friends` | 好友列表、搜索加好友、发起申请；（待处理的好友申请统一在收件箱） |
| 群组列表 | `/groups` | 我的群、创建群、通过邀请码加入、邀请好友入群 |
| 群详情 | `/groups/:id` | 成员列表、群活动列表、群内发起活动、邀请入群、退群/解散/复制邀请码 |
| 通知 | `/notifications` | 通知流、未读高亮、点击跳转对应实体或收件箱 |
| 个人中心 | `/me` | 资料编辑、登出、注销账号 |

---

## 7. 边界情况处理

| 场景 | 处理 |
|------|------|
| 活动过期 | 读取时派生 `expired`，详情只读，禁止新 RSVP；日历显示为灰；可选定时任务落库 |
| 活动取消 | 仅发起人/群主；级联删除所有参与者的 event 日程；给参与者发 `event_cancelled` 通知 |
| 容量已满 | P0 阻止新 going（返回 409）；waitlist → P1 |
| 群解散 | 群主操作；未来群活动 → cancelled，过去的归档只读；成员关系标记失效 |
| 群主退群 | 若有其他成员，需先转让群主（P1）；否则自动解散 |
| 账号注销 | 软删除，display_name 显示「已注销用户」；个人发起的活动取消、群活动保留归属群；好友关系移除；个人日程清除 |
| 重复报名 | (event,user) 唯一约束 + upsert，幂等 |
| 邀请已是参与者 | 已 RSVP（going/maybe/not_going）的用户再被邀请：不回退为 invited，幂等忽略 |
| 邀请未注册用户 | P0 仅支持邀请站内好友/群成员（已注册）；外部仅靠分享链接 |
| 收件箱已处理项 | 源状态非待办（rsvp≠invited、申请非 pending）即移出待办区，避免重复处理 |
| 邀请活动已取消/过期 | 收件箱中的对应邀请置灰只读，不可再报名 |
| 日程时间冲突 | 软提示，不阻止：允许多条重叠日程并存；报名/建/改日程时返回 conflicts 列表供前端提示；日历对冲突时段视觉标记。判定用半开区间，相邻不算冲突 |
| 自动日程可见性 | 用户手动改过后，后续报名联动 upsert 不覆盖用户设定 |
| 匿名访问详情 | 可只读；点 RSVP 时引导登录，登录后回到原活动 |
| 微信内打开 | 标准 OG meta 会正常提供，但微信内实际展示效果不做保证：可能显示默认链接预览、普通链接或风险提示页。境外域名在微信内也存在访问不稳定的可能，P0 接受该风险，不承诺微信内最佳体验（自定义卡片需企业主体，已移除） |
| 时区 | 后端 UTC，前端 Asia/Shanghai 展示与输入转换 |

---

## 8. 技术栈

- **后端**: Go + Gin + GORM；Supabase local（开发/测试，本地 PostgreSQL + Auth）/ 阿里云 Supabase Postgres（生产，地域待上线前拍板，当前倾向杭州）
- **前端**: React 18 + TypeScript + Vite + Tailwind CSS + shadcn/ui；Auth 用 supabase-js 直连 Supabase
- **认证**: Supabase Auth 托管（邮箱+密码/重置密码）；Gin 验签 Supabase JWT（`Authorization: Bearer`），不自行签发
- **设计**: 移动端优先，Luma 风格
- **测试**: 后端 Go `testing` + httptest（服务层 + handler），依赖 Supabase local / PostgreSQL test DB 验证数据库行为；前端组件测试（Vitest + Testing Library）

> **数据库一致性原则**：开发、测试、生产均以 PostgreSQL 语义为准。Supabase local 用于本地复现 `auth` schema、`auth.users` trigger、uuid、jsonb、partial unique index 等关键能力；不引入 SQLite，避免因方言和类型系统差异产生仅在生产暴露的问题。

### 8.1 部署架构（香港 ECS 单机部署）

> 当前以 **香港阿里云 ECS 单机部署** 为基准方案：先把架构跑通与验证真实访问质量，再决定是否需要迁移或拆分。因服务器位于香港，当前默认**不备案**；Cloudflare 仅承担 DNS 管理职责，不作为架构核心依赖。

- **前端**: React 构建产物（`dist`）直接部署在香港 ECS，由 Nginx 或 Caddy 托管静态文件
- **后端**: Gin 与前端同机部署在香港 ECS，通过反向代理将 `/api/*` 转发给 Gin
- **数据库 + Auth**: 阿里云 Supabase，承载 Postgres + Supabase Auth；项目需开通公网访问。地域待上线前以香港 ECS 到各候选地域（优先杭州）的实测延迟与稳定性决定
- **域名**: 域名托管在 Cloudflare，**默认灰云（DNS only）**，不默认开代理；正式域名解析到香港 ECS 公网 IP
- **HTTPS**: 正式上线后由 Nginx/Caddy + Let's Encrypt（或等价方案）签发证书；在未购买域名前，允许通过 ECS 公网 IP 进行小范围内测
- **访问路径**: 优先采用**同域部署**，例如 `example.com` 提供前端页面，`example.com/api/*` 转发 Gin，减少 CORS 和多域配置复杂度
- **边界**: 前端只通过 Supabase 做 Auth（拿 JWT）+ 可选 Realtime；**所有业务数据读写一律走 Gin**，避免 RLS 与 Gin 两套权限模型分裂。Gin 用 service_role 连接（绕过 RLS，业务鉴权在 Gin 内完成）
- **前端允许使用的 Supabase 能力**: 仅限认证与 session 管理（注册、登录、登出、邮箱验证、重置密码、获取当前 session）；**不允许**在前端直接查询 `events`、`schedules`、`friendships`、`groups`、`notifications` 等业务表，也不允许把复杂业务逻辑下沉到前端直接调用的 Supabase RPC
- **运维策略**: 默认先全灰云跑通；若后续出现源站暴露、被扫或抗攻击需求，再评估开启 Cloudflare 代理。若发生域名访问异常，优先准备备用域名而不是依赖裸 IP 作为正式入口
- **性能判断**: 目标用户主要在广东，香港链路通常可用，但稳定性以后续实测数据为准，不在 spec 中预设绝对结论

---

## 9. MVP 开发顺序建议

> 每步可独立交付与测试，遵循核心闭环优先。

1. **项目骨架**：Go+Gin 工程结构、GORM 连接（开发/测试 Supabase local PostgreSQL，生产阿里云 Supabase Postgres）、migration；React+Vite+Tailwind+shadcn 脚手架 + supabase-js；前后端联调跑通 `/api/health`。
2. **认证（P0）**：配置 Supabase Auth（邮箱+密码/重置密码）；前端 supabase-js 接入登录/注册；`public.users` 档案表 + `on auth.users insert` trigger 同步；Gin JWT 验签中间件、`/api/auth/me`。
3. **活动 + 分享（P0）**：events 表、创建/详情、share_slug、`/e/:slug` OG meta、匿名只读详情页。
4. **RSVP + 自动日程（P0，核心闭环）**：event_participants（含 invited 状态、add_to_calendar）、报名批阅、schedules 表与联动 upsert/删除、满员阻止。
5. **个人日程 + 日历月视图（P0）**：手动日程 CRUD、可见性、`/api/calendar` 聚合、月视图 UI + 颜色编码。
6. **好友（P0）**：friendships、申请/同意/拒绝、好友列表、点对点邀请（生成 invited）。
7. **群组（P0）**：groups/group_members/group_invites、创建/加入/退群/邀请入群、群内发起活动、群活动列表。
8. **统一收件箱 + 通知（P0）**：`/api/inbox` 聚合活动邀请/好友申请/群组邀请，行内批阅；notifications 提醒流与未读计数。
9. **打磨**：Luma 风格视觉细化、空状态、加载态、移动端适配回归。
10. **部署上线**：香港 ECS 安装 Nginx/Caddy，托管前端静态文件并反代 Gin；先用 ECS 公网 IP 做小范围内测，验证广东用户访问与登录链路；确认无明显问题后购买域名，由 Cloudflare 做 DNS 解析并配置 HTTPS，同时更新 Supabase 的 `Site URL` / `Redirect URLs`。
11. **P1 起步**：好友日程查看、日历筛选、活动编辑/评论、周/日视图、候补队列。

---

## 10. 待评审 / 后续可深化
- 推荐时段算法（P2）的具体规则。
- 候补队列递补的通知与时限策略（P1）。
- 注销账号时 `auth.users` 与业务数据（个人活动取消、群活动归属、好友关系、日程）的级联清理细节。
- 将来注册企业主体后，微信生态能力（OAuth 登录、JS-SDK 分享卡片、模板消息）的引入与是否需迁回大陆备案。
- Supabase 最终地域选择（优先杭州，但以上线前从香港 ECS 的实际延迟、抖动与登录成功率压测结果为准）。
- 从「公网 IP 内测」切换到「正式域名 + HTTPS」时，Supabase 回调地址、OG 分享链接与站点配置的切换清单。
