# Hangout — 产品设计与技术规格（Spec）

- **状态**: 已批准设计，待评审 spec
- **日期**: 2026-06-23
- **作者**: 产品 + 工程协作（brainstorming 产出）
- **暂定名**: Hangout
- **一句话定位**: 面向熟人圈子的社交日历与活动协调 Web 应用，灵感来自飞书日历，UI 风格对标 Luma。

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
| D1 | 微信登录 | 邮箱+密码为 P0；微信 OAuth 写入数据模型，但 P0 用 **Mock 微信登录**（假 openid/昵称/头像，接口与真实一致），真实 OAuth → P1 |
| D2 | 分享卡片 | 标准 OG / Twitter meta（P0），浏览器/多数平台显示卡片；微信内自定义卡片用 JS-SDK → P1 |
| D3 | RSVP 状态 | 三态：`going` / `not_going` / `maybe`；`maybe` 在日历显示为暂定忙 |
| D4 | MVP 首切片 | 核心闭环优先：活动 + 分享 + RSVP + 自动日程 |
| D5 | 匿名访问 | 未登录可**查看**活动详情（只读），**报名**需登录 |
| D6 | 群活动可见性 | 仅链接 / 群内可见（隐私优先） |
| D7 | 账号合并 | P0 邮箱与微信视为独立账号；手动「绑定账号」→ P2 |
| D8 | 通知 | P0 站内通知列表；推送 / 微信模板消息 → P2 |
| D9 | 容量已满 | P0 直接阻止报名；候补队列（waitlist）→ P1 |

---

## 3. 功能列表（按优先级）

### P0 — MVP（核心闭环可用）
1. **认证**
   - 邮箱+密码注册/登录/登出
   - Mock 微信登录（生成假 openid/昵称/头像，可一键切换多个测试身份）
   - JWT 存于 HttpOnly Cookie；鉴权中间件
2. **活动**
   - 创建活动（个人发起 / 群发起）
   - 字段：标题、类型（攀岩/吃饭/出行/游戏/其他）、开始时间、结束时间、地点、人数上限（可选）
   - 活动详情页（未登录可只读查看）
   - 唯一分享链接（slug）+ OG meta 标签
   - 取消活动（仅发起人；群活动发起人或群主）
3. **RSVP（报名）**
   - going / not_going / maybe
   - 报名 `going`（或 `maybe`）→ 自动在个人日程创建关联日程（默认 `busy_only`）
   - 取消报名 / 改为 not_going → 自动删除对应日程
   - 容量已满阻止新的 going
4. **个人日程**
   - 手动创建/编辑/删除个人日程（独立于活动）
   - 三种可见性：public / busy_only / private
   - 活动自动日程默认 busy_only，用户可手动改
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
8. **通知（站内）**
   - 好友申请、活动邀请、活动变更/取消的站内通知列表 + 未读计数

### P1 — 增强
- 真实微信 OAuth 登录
- 微信 JS-SDK 自定义分享卡片
- 日历筛选：全部 / 我的群 / 好友
- 查看好友日程（按可见性过滤）
- 容量候补队列（waitlist），有人退出自动递补
- 活动评论/留言
- 群活动详情显示「谁还没回应」（invited 但未 RSVP）
- 活动编辑（已存在的字段修改 + 变更通知参与者）
- 周视图 / 日视图
- 群主转让

### P2 — 锦上添花
- 账号绑定/合并（邮箱 ↔ 微信）
- 微信模板消息 / Web Push 推送
- 推荐时段（找朋友共同有空的时间）
- 活动复盘/相册
- 重复活动（周期性日程）
- 日历订阅（iCal 导出）
- 数据统计面板

---

## 4. 数据模型设计

> ORM: GORM。所有表含 `id (uint, PK)`、`created_at`、`updated_at`、`deleted_at (软删除, gorm.DeletedAt)`，下文不再重复列出。时间统一 UTC 存储，前端按 Asia/Shanghai 展示。

### 4.1 users
| 字段 | 类型 | 说明 |
|------|------|------|
| display_name | string | 昵称 |
| avatar_url | string | 头像 URL（可空） |
| email | string, unique, nullable | 邮箱（邮箱注册时有） |
| password_hash | string, nullable | bcrypt 哈希（邮箱注册时有） |
| status | string | `active` / `deactivated`（注销后显示「已注销用户」） |

### 4.2 auth_providers
统一账户体系的关键表，一个 user 可挂多个登录方式。
| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | uint, FK→users | |
| provider | string | `email` / `wechat` / `wechat_mock` |
| provider_uid | string | 微信为 openid；email 为邮箱 |
| 唯一约束 | (provider, provider_uid) | 防重复绑定 |

### 4.3 friendships
| 字段 | 类型 | 说明 |
|------|------|------|
| requester_id | uint, FK→users | 申请方 |
| addressee_id | uint, FK→users | 被申请方 |
| status | string | `pending` / `accepted` / `rejected` |
| 唯一约束 | (requester_id, addressee_id) | |

> 查询「是否好友」：存在一条 `accepted` 且 {requester,addressee} 命中该用户对（无序）。

### 4.4 groups
| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 群名 |
| description | string | 群简介（可空） |
| owner_id | uint, FK→users | 群主 |
| invite_code | string, unique | 群邀请码（用于邀请链接） |
| status | string | `active` / `dissolved` |

### 4.5 group_members
| 字段 | 类型 | 说明 |
|------|------|------|
| group_id | uint, FK→groups | |
| user_id | uint, FK→users | |
| role | string | `owner` / `member` |
| 唯一约束 | (group_id, user_id) | |

### 4.6 events
| 字段 | 类型 | 说明 |
|------|------|------|
| owner_id | uint, FK→users | 发起人 |
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

### 4.7 event_participants
| 字段 | 类型 | 说明 |
|------|------|------|
| event_id | uint, FK→events | |
| user_id | uint, FK→users | |
| rsvp | string | `going` / `not_going` / `maybe` |
| source | string | `self`（自己点进来）/ `invited`（被邀请） |
| 唯一约束 | (event_id, user_id) | |

### 4.8 schedules（个人日程）
| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | uint, FK→users | 所属用户 |
| title | string | 标题（手动日程用户填；活动日程取活动标题） |
| start_at | datetime | |
| end_at | datetime, nullable | |
| location | string | 可空 |
| visibility | string | `public` / `busy_only` / `private` |
| source | string | `manual` / `event` |
| event_id | uint, FK→events, nullable | source=event 时关联活动 |
| 唯一约束 | (user_id, event_id) where source=event | 防止重复生成 |

> **联动规则**：RSVP=going/maybe → upsert 一条 source=event 的 schedule（默认 busy_only）；RSVP 改为 not_going 或取消 → 删除该 schedule。用户手动改过的可见性需保留（upsert 时不覆盖用户已改的 visibility）。

### 4.9 notifications
| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | uint, FK→users | 接收者 |
| type | string | `friend_request` / `friend_accepted` / `event_invite` / `event_changed` / `event_cancelled` |
| payload | jsonb/text | 关联实体 id 与展示数据 |
| read_at | datetime, nullable | 已读时间 |

---

## 5. API 接口设计

> 统一前缀 `/api`。鉴权用 JWT（HttpOnly Cookie）。响应体统一 `{ "data": ..., "error": null }`。错误码用 HTTP status + `{ "error": { "code", "message" } }`。

### 5.1 认证 Auth
| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| POST | /api/auth/register | 邮箱+密码注册 | 否 |
| POST | /api/auth/login | 邮箱+密码登录，设 Cookie | 否 |
| POST | /api/auth/logout | 登出，清 Cookie | 是 |
| POST | /api/auth/wechat/mock | Mock 微信登录（传 openid/昵称/头像，无则随机） | 否 |
| GET  | /api/auth/wechat/callback | 真实微信 OAuth 回调（P1） | 否 |
| GET  | /api/auth/me | 当前用户信息 | 是 |

### 5.2 用户 Users
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/users/me | 个人资料 |
| PATCH | /api/users/me | 更新昵称/头像 |
| DELETE | /api/users/me | 注销（软删除，置 deactivated） |
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
| DELETE | /api/groups/:id/members/:userId | 退群 / 移除成员 |
| GET | /api/groups/:id/events | 群活动列表 |

### 5.5 活动 Events
| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| POST | /api/events | 创建活动 | 是 |
| GET | /api/events/:slug | 活动详情（含参与者汇总） | 否（只读可匿名） |
| PATCH | /api/events/:id | 编辑活动（发起人，P1） | 是 |
| POST | /api/events/:id/cancel | 取消活动 | 是 |
| POST | /api/events/:id/rsvp | 报名/改状态 `{rsvp}` | 是 |
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

### 5.7 日历聚合 Calendar
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/calendar?from=&to=&filter= | 聚合月视图数据：我的日程 + 群活动 +（P1）好友。`filter`=all/groups/friends |

### 5.8 通知 Notifications
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/notifications | 通知列表 + 未读数 |
| POST | /api/notifications/:id/read | 标记已读 |
| POST | /api/notifications/read-all | 全部已读 |

### 5.9 分享 OG 卡片
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /e/:slug | 服务端渲染含 OG/Twitter meta 的 HTML（标题、时间、报名人数），随后 hydrate 到前端详情页 |

---

## 6. 前端页面列表与核心交互

> 移动端优先，底部 Tab 导航；Luma 风格大卡片。

| 页面 | 路由 | 核心交互 |
|------|------|---------|
| 启动/登录 | `/login` | 邮箱登录、邮箱注册切换、Mock 微信一键登录（开发态多身份切换） |
| 日历（首页） | `/` | 月视图 + 标记点 + 颜色图例；点日期展开当天事件抽屉；筛选 全部/我的群/好友（P1）；右下角「+」创建活动/日程 |
| 活动详情 | `/e/:slug` | 封面/类型色块、时间地点、报名人数与头像墙、RSVP 三态按钮、邀请好友、分享按钮（复制链接/微信）；未登录显示只读 + 登录后报名引导；满员禁用 going |
| 创建活动 | `/events/new` | 选发起范围（个人/群）、标题、类型选择器、时间选择、地点、人数上限；提交后跳详情并弹分享 |
| 我的活动 | `/events` | 我发起/参与，分「即将开始 / 已过期」 |
| 创建/编辑日程 | `/schedules/new`、`/schedules/:id` | 标题、时间、地点、可见性选择（public/busy_only/private 带说明） |
| 好友 | `/friends` | 好友列表、申请红点、搜索加好友、同意/拒绝；进入好友日程（P1） |
| 群组列表 | `/groups` | 我的群、创建群、通过邀请码加入 |
| 群详情 | `/groups/:id` | 成员列表、群活动列表、群内发起活动、退群/解散/复制邀请码 |
| 通知 | `/notifications` | 通知流、未读高亮、点击跳转对应实体 |
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
| 自动日程可见性 | 用户手动改过后，后续报名联动 upsert 不覆盖用户设定 |
| 匿名访问详情 | 可只读；点 RSVP 时引导登录，登录后回到原活动 |
| 微信内打开 | 微信屏蔽自定义 OG → 走默认抓取；P1 用 JS-SDK 优化卡片 |
| 时区 | 后端 UTC，前端 Asia/Shanghai 展示与输入转换 |

---

## 8. 技术栈

- **后端**: Go + Gin + GORM；SQLite（开发）/ PostgreSQL（生产）
- **前端**: React 18 + TypeScript + Vite + Tailwind CSS + shadcn/ui
- **认证**: JWT（HttpOnly Cookie，SameSite=Lax）
- **设计**: 移动端优先，Luma 风格
- **测试**: 后端 Go `testing` + httptest（服务层 + handler）；前端组件测试（Vitest + Testing Library）

---

## 9. MVP 开发顺序建议

> 每步可独立交付与测试，遵循核心闭环优先。

1. **项目骨架**：Go+Gin 工程结构、GORM 连接、migration；React+Vite+Tailwind+shadcn 脚手架；前后端联调跑通 `/api/health`。
2. **认证（P0）**：users + auth_providers，邮箱注册/登录、Mock 微信登录、JWT Cookie 中间件、`/me`。
3. **活动 + 分享（P0）**：events 表、创建/详情、share_slug、`/e/:slug` OG meta、匿名只读详情页。
4. **RSVP + 自动日程（P0，核心闭环）**：event_participants、三态 RSVP、schedules 表与联动 upsert/删除、满员阻止。
5. **个人日程 + 日历月视图（P0）**：手动日程 CRUD、可见性、`/api/calendar` 聚合、月视图 UI + 颜色编码。
6. **好友（P0）**：friendships、申请/同意/拒绝、好友列表、点对点邀请。
7. **群组（P0）**：groups/group_members、创建/加入/退群、群内发起活动、群活动列表。
8. **通知（P0）**：notifications，关键事件触发（好友申请、邀请、变更/取消）+ 未读计数。
9. **打磨**：Luma 风格视觉细化、空状态、加载态、移动端适配回归。
10. **P1 起步**：真实微信 OAuth、好友日程查看、日历筛选、JS-SDK 分享卡片。

---

## 10. 待评审 / 后续可深化
- 推荐时段算法（P2）的具体规则。
- 候补队列递补的通知与时限策略（P1）。
- 真实微信 OAuth 与账号绑定的合并冲突处理（P1/P2）。
