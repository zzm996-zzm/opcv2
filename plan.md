# OPC v2 第一版实施计划

> **执行要求：** 按任务顺序实施，每完成一个阶段都必须通过对应测试和验收项后再进入下一阶段。

**目标：** 按 Notion 中的 OPC v2 UI 完成第一版可上线产品，跑通“用户登录 -> AI 免费分析 -> 天眼查企业获客 -> 加入 CRM -> 跟进客户 -> 会员/兑换控制权益”的核心闭环。

**架构：** 采用 Go 模块化单体，不拆微服务。HTTP API 和异步 Worker 使用同一套领域代码、分开进程运行；PostgreSQL 保存业务数据，Redis 负责验证码、限流、缓存和任务队列；前端按设计图逐屏实现。

**核心原则：**

- 第一版优先完成业务闭环，不建设复杂平台能力。
- 天眼查返回的企业名称、电话、邮箱、网址和来源直接展示。
- 不设计 `public / validated / verified / invalid` 联系方式状态。
- 不承诺联系方式一定可达，页面统一注明“信息来自公开数据源，请使用者自行核实”。
- 数据源按职责拆分：天眼查负责企业主体，Serper/SerpAPI 负责公开网页补充，不做多个企业库聚合。
- 不使用微服务、Kafka、Elasticsearch、Kubernetes。
- UI 结构、布局、文案层级、状态页和交互流程以 Notion 设计图为准。

---

## 1. 第一版范围

### 1.1 必须真实可用

1. 手机验证码登录/注册、退出登录。
2. 首页一框输入和四个快捷入口。
3. 免费分析双 Tab、多轮追问、生成结果、历史记录和详细报告。
4. VIP 获客条件输入、异步任务、天眼查企业搜索和结果展示。
5. 企业线索批量加入 CRM。
6. CRM 客户列表、详情、阶段、跟进记录和下次跟进时间。
7. 会员套餐、权益控制、积分账本和兑换码。
8. 管理后台的用户、套餐、兑换码、站点配置和任务查询。

### 1.2 UI 完整、数据简化

以下页面按设计图完成，但第一版使用后台配置或种子数据：

- 咨询通：文章分类、文章列表、文章详情。
- 工具箱：分类、搜索、工具卡片、收藏。
- 社群：免费社群、VIP 社群、二维码、活动和权益说明。
- 首页案例、品牌数字、社群数字。

### 1.3 第一版不做

- 联系方式真实性、可达性和人工核验。
- 微信扫码登录、微信支付和自动续费。
- 多组织协作、坐席分配和复杂 RBAC。
- 企业线索自动拨测、短信或邮件群发。
- 多个企业数据库聚合及跨企业库去重。
- 招投标、融资、招聘等复杂购买信号评分。
- 跨境获客和本地门店获客。
- 全网自动抓取资讯、自动发现 RSS 和工具自动收录。

---

## 2. 技术栈

### 2.1 服务端

| 能力 | 选择 | 原因 |
|---|---|---|
| 语言 | Go | 并发任务、部署简单、资源占用低 |
| HTTP | Gin | 成熟、资料多、适合中小型 API |
| ORM | GORM | 第一版开发速度优先，迁移和事务能力足够 |
| 数据库 | PostgreSQL | 事务、JSONB、模糊匹配和后续多租户扩展 |
| 缓存/限流 | Redis | 验证码、频控、临时状态、热点缓存 |
| 异步任务 | Asynq | 基于 Redis，适合 Go，避免额外引入消息系统 |
| 企业数据 | 天眼查开放平台 | 企业主体和公开联系方式 |
| Web 搜索 | Serper 主用、SerpAPI 备用 | 官网、新闻、招投标和其他公开证据链接 |
| 免费资讯 | GDELT + RSS | 咨询通和企业新闻补充，不作为联系方式来源 |
| 数据迁移 | golang-migrate | SQL 迁移明确、可审查、可回滚 |
| 参数校验 | go-playground/validator | 请求参数统一校验 |
| 日志 | slog + JSON Handler | 使用 Go 标准日志接口，减少依赖 |
| 配置 | 环境变量 + 配置结构体 | 开发、测试、生产配置分离 |
| API 文档 | OpenAPI/Swagger | 方便前后端联调 |
| 测试 | testing + testify + testcontainers | 单元测试和 PostgreSQL/Redis 集成测试 |

### 2.2 前端

| 能力 | 选择 | 原因 |
|---|---|---|
| 框架 | React + TypeScript + Vite | 开发速度快，适合设计稿还原 |
| 路由 | React Router | 页面路由和登录保护 |
| 请求状态 | TanStack Query | 缓存、轮询、错误和加载状态统一 |
| 本地状态 | Zustand | 只保存登录态、UI 状态和未提交输入 |
| 样式 | Tailwind CSS | 便于精确还原设计图 |
| 基础组件 | Radix UI | 只提供无样式交互基础，视觉按设计图定制 |
| 表单 | React Hook Form + Zod | 表单状态和校验统一 |
| 图标 | Lucide React | 统一图标风格 |
| 图表 | Recharts | CRM 概览等简单图表 |
| 测试 | Vitest + Testing Library + Playwright | 组件、交互和核心流程测试 |

### 2.3 部署

- Docker Compose：`api`、`worker`、`web`、`postgres`、`redis`。
- Nginx/Caddy：HTTPS、静态资源和 API 反向代理。
- 对象存储第一版可不接；PDF 导出落临时文件并设置过期清理。
- 生产环境每天备份 PostgreSQL，保留 7 至 14 天。

---

## 3. 项目目录

```text
opcv2/
├── apps/
│   ├── api/                    # Go HTTP API 入口
│   ├── worker/                 # Go Asynq Worker 入口
│   └── web/                    # React 前端
├── internal/
│   ├── platform/               # 配置、数据库、Redis、日志、HTTP 公共能力
│   ├── auth/                   # 登录、验证码、Token
│   ├── user/                   # 用户资料
│   ├── entitlement/            # 套餐、权益、积分、兑换码
│   ├── analysis/               # AI 分析、追问、报告
│   ├── lead/                   # 获客任务、天眼查适配、企业线索
│   ├── search/                 # Serper/SerpAPI 公共搜索适配器
│   ├── crm/                    # 客户、阶段、跟进记录
│   ├── content/                # 咨询通、工具箱、社群配置
│   ├── admin/                  # 管理后台接口
│   └── ai/                     # 大模型 Provider 和结构化输出
├── migrations/                 # PostgreSQL SQL 迁移
├── api/openapi/                # OpenAPI 文档
├── deployments/               # Docker、反向代理、部署配置
├── tests/                      # 跨模块集成与端到端测试
├── go.mod
└── plan.md
```

模块内部统一采用：

```text
module/
├── model.go
├── repository.go
├── service.go
├── handler.go
├── dto.go
└── service_test.go
```

第一版不再拆 `domain/application/infrastructure` 多层目录，避免结构过重。

---

## 4. 核心数据模型

### 4.1 用户与登录

- `users`
  - `id`
  - `nickname`
  - `phone`
  - `wechat`
  - `status`
  - `created_at`
- `sms_codes`
  - 正式记录放 Redis，数据库只保存发送审计。
- `login_records`
  - `user_id`
  - `ip`
  - `user_agent`
  - `created_at`

手机号设置唯一索引。微信号如果允许登录，同样设置唯一索引。

### 4.2 套餐、权益和积分

- `plans`
  - 免费版、¥69、¥199、¥499。
  - 保存分析次数、获客次数、单次结果上限、CRM 上限等配置。
- `subscriptions`
  - `user_id`
  - `plan_id`
  - `starts_at`
  - `expires_at`
- `credit_accounts`
  - `user_id`
  - `balance`
- `credit_ledger`
  - `user_id`
  - `amount`
  - `type`
  - `biz_type`
  - `biz_id`
  - `idempotency_key`
- `redeem_codes`
  - 兑换码、权益内容、使用次数、有效期、状态。
- `redeem_records`
  - 用户、兑换码、兑换时间、发放结果。

积分扣减使用数据库事务和幂等键，但第一版不做冻结、结算、退款三段式账务。创建获客任务成功时直接扣减；任务因系统错误失败时写一笔补偿返还记录。

### 4.3 AI 分析

- `analysis_sessions`
  - 用户输入、分析类型、当前阶段、状态。
- `analysis_messages`
  - 用户回答和 AI 追问。
- `analysis_reports`
  - 标题、摘要、结构化报告 `JSONB`、模型、Prompt 版本。
- `opportunities`
  - 方向名称、匹配度、适合理由、市场机会、启动难度和排序。
- `action_items`
  - 7 天行动计划及完成状态。

AI 返回必须使用结构化 JSON Schema；解析失败允许自动重试一次，仍失败则记录错误并提示用户重试。

### 4.3.1 免费分析输出契约

产品补充的三张图片属于免费分析功能规格，不是三个新增导航页面。对应关系如下：

| 产品图片 | 对应 Notion UI | 作用 |
|---|---|---|
| A1 Tab A 输出结构 | 免费分析结果概览 | 定义 3 张方向卡的必填字段 |
| 第二层详细报告 + 第三层 7 天计划 | 免费分析详细报告 | 定义点击方向卡后的正文和右栏行动计划 |
| A2 Tab B + A3 追问机制 | 免费分析输入、分析中追问、竞品报告 | 定义竞品拆解内容和何时追问 |

#### Tab A：我有什么，适合做什么

第一层一次生成 3 张方向卡，每张必须包含：

- `name`：一句话说明方向。
- `match_score`：0 至 100 分。
- `match_reasons`：固定 3 条，并引用用户输入。
- `market_opportunity`：必须附可点击来源，不允许只写无法追溯的市场数字。
- `startup_difficulty`：低/中/高，加关键卡点。
- `benchmarks`：2 至 3 个真实可查询案例。
- `first_three_actions`：三步行动，每一步能在当天开始执行。
- `upgrade_hook`：从免费分析进入 VIP 获客的自然提示。

第二层详细报告包含 7 个固定章节：

1. 为什么特别适合你：约 300 字，逐项对应用户资源和技能。
2. 市场机会与竞争格局：约 500 字，包含市场规模、主要玩家、竞争密度和未来趋势。
3. 0 到 1 起步路径：30 天计划，按周拆分并给出具体平台、工具和内容。
4. 推荐对标公司或个人：3 至 5 个，包含业务、起步方式、收入估算依据和可模仿动作。
5. 启动资金分配建议：按用户预算给出金额或比例。
6. 风险与应对：3 至 5 个真实风险和对应动作。
7. 第一批客户从哪里来：2 至 3 个具体渠道及每个渠道的第一步。

产品图中章节前的方框是规格清单标记，正式 UI 按报告目录和正文呈现，不做成可勾选控件。

第三层是详细报告右栏的 7 天行动计划：

- 每天 1 至 3 个可勾选任务。
- 每项必须写明在哪个平台执行什么操作。
- 勾选状态保存到 `action_items`，刷新和再次登录后不丢失。

#### Tab B：拆解一个对标公司

用户可输入公司名、公众号、抖音或小红书链接。报告固定包含：

1. 商业模式一句话总结。
2. 获客渠道拆解：平台、内容类型、转化路径。
3. 爆款内容分析：选择 3 至 5 个公开可访问案例，分析标题、结构和钩子。
4. 定价与变现方式：产品梯度、前端引流品和后端高价品。
5. 可模仿点：普通人或小团队可执行的 3 条动作，并标难度。
6. 差异化切入建议：避免正面竞争的切入策略。

如果公开搜索没有足够资料，报告必须明确显示“公开信息不足”，不得由模型虚构案例、收入或市场数字。

#### 追问规则

- 最多追问 2 轮，每轮最多 3 个问题。
- 每个问题提供快捷选项和自由输入。
- 用户输入少于 30 个字，或预算、投入时间、线上/线下偏好、目标地区四项中缺少任意两项时触发。
- 可问：启动资金、每周投入时间、线上/线下偏好、目标客户地区。
- 不再询问：已填写的行业经验、登录时已有的姓名和联系方式。
- 用户可以选择“跳过并生成”，但页面提示结果可能不够精准。

### 4.4 获客任务与企业线索

- `lead_tasks`
  - `user_id`
  - 原始输入
  - 行业、地区、关键词
  - 状态：`pending/running/succeeded/failed/cancelled`
  - 结果数量、错误摘要、积分消耗。
- `provider_requests`
  - 天眼查、Serper、SerpAPI 的请求类型、响应状态、费用计数和耗时。
  - 敏感 Token 不落库。
- `lead_evidence`
  - `lead_id`
  - `provider`
  - `evidence_type`：`official_site/news/tender/contact_page/other`
  - `title`
  - `url`
  - `snippet`
  - `published_at`
- `leads`
  - `lead_task_id`
  - `provider`
  - `provider_company_id`
  - 企业名称
  - 法定代表人
  - 行业
  - 地区
  - 地址
  - 电话
  - 邮箱
  - 网站
  - 经营范围
  - 成立日期
  - 注册资本
  - 原始数据 `JSONB`
  - 来源页面或接口标识

去重仅做确定性规则：

1. 优先使用天眼查企业 ID。
2. 没有企业 ID 时使用统一社会信用代码。
3. 两者都没有时使用标准化企业名称。

不做相似企业智能合并，不做联系方式校验。

### 4.5 CRM

- `customers`
  - 来源线索、客户名称、联系人信息、阶段、负责人、备注。
- `customer_activities`
  - 跟进方式、内容、跟进时间、创建人。
- `followup_reminders`
  - 下次跟进时间、状态。
- `script_records`
  - AI 生成的开场或跟进话术。

CRM 阶段第一版固定为：

```text
待跟进 -> 已沟通 -> 有意向 -> 已成交 -> 已放弃
```

### 4.6 内容配置

- `insight_articles`
- `tool_categories`
- `tool_items`
- `tool_favorites`
- `community_configs`
- `site_configs`

内容默认由管理后台录入；可增加白名单 RSS/GDELT 定时导入，但不建设通用网页采集平台。

---

## 5. 数据源接入方案

### 5.1 天眼查：企业主体源

1. 企业高级搜索：根据地区、行业、关键词获取候选企业。
2. 企业基本信息/联系方式：按企业 ID 或名称补充电话、邮箱和网址。

具体购买哪个接口套餐，以天眼查商务最终授权和返回字段为准。开发前必须先获得：

- 正式 App Key/Token。
- 商业使用授权确认。
- 接口限流和日/月配额。
- 是否允许将返回结果保存到本产品 CRM。
- 是否允许向最终用户展示返回字段及所需来源标识。

### 5.2 Serper/SerpAPI：Web 证据源

旧项目 `/Users/zzm/data/www/opc` 中有可参考的实现：

- `server/internal/service/source_search.go`
- `server/internal/service/source_search_test.go`
- `LeadSourceSearchProvider` 接口。
- `serper` 和 `serpapi` 两种实现。
- `OPC_SEARCH_PROVIDER/API_KEY/BASE_URL/TIMEOUT_SECONDS/MAX_RESULTS` 配置。

新项目不迁移旧文件、不照搬旧结构，只参考其 Provider 接口、请求解析和 Mock 测试的有效做法，重新按 v2 模块边界实现：

- `SearchProvider`：通用搜索接口。
- Serper 为主 Provider。
- SerpAPI 为备用 Provider，仅在主 Provider 未配置、配额耗尽或服务失败时切换。
- Key 只能从环境变量读取，禁止写入 YAML、源码、日志或数据库。
- 旧项目明文配置过的 Key 在 v2 启用搜索前必须轮换。

每家企业使用固定查询模板，第一版最多执行以下查询：

```text
"{企业全名}" 官网
"{企业全名}" 联系我们
"{企业全名}" 新闻
site:gov.cn "{企业全名}" 招标 OR 采购
site:ccgp.gov.cn "{企业全名}"
```

搜索结果仅保存标题、URL、摘要、发布时间和 Provider，不抓取整页，不把摘要中的电话或邮箱覆盖到天眼查主体数据中。

### 5.3 免费数据源

第一版可以接入但不阻塞主流程：

1. **GDELT**
   - 免费的全球新闻事件和文档检索。
   - 用于补充企业相关新闻和咨询通内容。
   - 不用于企业主体、电话或邮箱。
2. **RSS/Atom**
   - 接入明确允许订阅的政府、行业协会、媒体或企业官方 Feed。
   - 用于咨询通文章候选池。
   - 第一版使用白名单 Feed，不做全网自动发现。
3. **公开网页定向搜索**
   - 通过 Serper/SerpAPI 查询政府采购网、`gov.cn`、企业官网和 1688 等公开页面。
   - 只展示原始链接和摘要，不批量抓取受限制页面。
4. **Brave Search API**
   - 官方当前按月提供少量免费额度，可作为后续第三个 `SearchProvider`。
   - MVP 先保留适配器接口，不阻塞 Serper/SerpAPI 上线。

搜索额度参考以开发时官方页面为准：

- Serper：新账户当前提供 2,500 次免费查询。
- SerpAPI：当前免费档为每月 250 次搜索。
- Brave Search API：当前每月提供 5 美元免费额度，按搜索价格折算约 1,000 次请求。

暂不作为 MVP 数据源：

- OpenStreetMap/Nominatim：国内企业和联系方式覆盖不足，公共实例也不适合批量商业查询。
- 国家企业信用信息公示系统：中国大陆权威但没有适合本 MVP 的稳定公开商业 API，不做爬虫。
- 高德/百度地图：等待商业数据保存和展示授权后再评估。
- 企查查：作为天眼查备用企业库，MVP 不同时采购。
- Apollo：跨境版本再接。

### 5.4 请求控制

- `TianyanchaProvider` 是独立适配器，业务层不直接拼接天眼查参数。
- `SearchProvider` 与企业 Provider 分离，不能用 Web 搜索替代企业主体查询。
- 每个获客任务限制最大企业数，套餐决定 20/50/100 等上限。
- Worker 串行或低并发补充企业详情，遵守 Provider QPS。
- 对 `429` 和临时 `5xx` 指数退避重试，最多 3 次。
- 对无结果、参数错误、配额耗尽分别返回明确错误。
- Redis 记录当日调用量，数据库记录每次 Provider 请求用于成本统计。
- 同一企业的证据搜索结果按规范化 URL 去重。
- MVP 每个企业最多保留 5 条 Web 证据，避免搜索成本失控。

### 5.5 页面展示

结果卡片直接展示：

- 企业名称、行业、地区、地址。
- 公开电话、邮箱和网站。
- 注册资本、成立时间、经营范围。
- 数据来源：天眼查。
- 官网、新闻、招投标等公开证据链接，标记对应搜索 Provider。
- 提示：“信息来自公开数据源，请使用者自行核实并依法合规使用。”

不展示真实性徽章、验证状态或可达性评分。

---

## 6. API 规划

### 6.1 Auth

```text
POST /api/v1/auth/sms/send
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/me
```

### 6.2 Analysis

```text
POST /api/v1/analyses
POST /api/v1/analyses/:id/answers
GET  /api/v1/analyses
GET  /api/v1/analyses/:id
POST /api/v1/analyses/:id/regenerate
PATCH /api/v1/action-items/:id
```

第一版使用普通 HTTP 创建任务，加轮询获取进度；AI 文本生成页面可使用 SSE。不要为了统一而全站使用 WebSocket。

### 6.3 Leads

```text
POST /api/v1/lead-tasks
GET  /api/v1/lead-tasks
GET  /api/v1/lead-tasks/:id
POST /api/v1/lead-tasks/:id/cancel
POST /api/v1/lead-tasks/:id/retry
GET  /api/v1/lead-tasks/:id/leads
POST /api/v1/leads/import-to-crm
```

### 6.4 CRM

```text
GET    /api/v1/customers
POST   /api/v1/customers
GET    /api/v1/customers/:id
PATCH  /api/v1/customers/:id
POST   /api/v1/customers/:id/activities
GET    /api/v1/followups
POST   /api/v1/customers/:id/scripts
```

### 6.5 Membership

```text
GET  /api/v1/plans
GET  /api/v1/entitlements
GET  /api/v1/credits
GET  /api/v1/credits/ledger
POST /api/v1/redeem
```

### 6.6 Content/Admin

```text
GET /api/v1/insights
GET /api/v1/tools
GET /api/v1/community

/api/v1/admin/users
/api/v1/admin/plans
/api/v1/admin/redeem-codes
/api/v1/admin/content
/api/v1/admin/site-configs
/api/v1/admin/tasks
```

---

## 7. UI 页面实施清单

所有页面先建立设计 Token：

- 颜色、字号、圆角、阴影、间距、容器宽度。
- 顶部导航、按钮、输入框、卡片、标签、弹窗、空状态。
- 桌面端优先，还原 Notion 设计图；移动端保证可用但不重新设计一套视觉。

### 7.1 页面路由

```text
/login
/
/analysis
/analysis/:id/progress
/analysis/:id
/analysis/:id/report
/leads
/leads/tasks
/leads/tasks/:id
/crm
/crm/followups
/insights
/tools
/community
/membership
/redeem
/admin/*
```

### 7.2 按设计图实现的状态

1. 登录页。
2. 未登录首页。
3. 登录态首页。
4. 免费分析输入页。
5. 分析中 Stepper 和 AI 追问。
6. 分析结果概览。
7. 分析详细报告。
8. VIP 获客输入页。
9. 获客任务列表。
10. 线索结果列表。
11. CRM 客户管理。
12. 全部跟进看板。
13. 咨询通。
14. 工具箱。
15. 社群。
16. 会员计划。
17. 会员兑换。

每个页面必须包含：

- 加载状态。
- 空状态。
- 错误状态。
- 权限或额度不足状态。
- 与设计图一致的成功状态。

### 7.3 免费分析动画与过渡

动画只用于表达系统正在处理，不制造虚假进度。统一使用 CSS/Framer Motion 实现，不引入大型动画引擎。

#### 分析中 Stepper

固定展示四个阶段：

```text
读取你的输入
-> 理解需求
-> 补充关键信息
-> 生成方向报告
```

- 当前步骤使用品牌色高亮、轻微呼吸光和环形进度动画。
- 已完成步骤显示勾选和 150 至 250ms 的完成过渡。
- 步骤切换使用 200 至 300ms 淡入与位移动画。
- 百分比只在后端返回真实阶段事件时前进，不使用定时器伪造到 99%。
- 后端长时间无事件时显示循环状态文案，不继续增加百分比。

#### AI 追问

- 问题卡片逐张淡入，上一题回答后再进入下一题。
- 快捷选项点击后显示选中反馈，允许修改。
- 每轮结束显示“正在结合你的补充信息”，然后回到 Stepper。
- 支持“跳过并生成”，避免用户被追问流程卡住。

#### 结果揭示

- 3 张方向卡按建议顺序依次进入，间隔 80 至 120ms。
- 分数环从 0 动画到实际分数，但无障碍减弱动画模式下直接显示结果。
- 点击方向卡进入详细报告时使用普通页面过渡，不做复杂 3D 翻转。
- 详细报告章节按流式输出逐段出现；没有内容的章节不提前展示空壳。
- 7 天计划勾选只做轻量完成反馈，不使用庆祝动画。

#### 动画验收

- 支持 `prefers-reduced-motion`。
- 动画不能阻止用户阅读、滚动、返回或重试。
- 普通手机上保持流畅，单个页面不加载视频背景。
- E2E 测试可关闭动画，避免测试因时间差不稳定。

---

## 8. 实施阶段

### 阶段 0：项目骨架与质量门槛

**创建：**

- `go.mod`
- `apps/api/main.go`
- `apps/worker/main.go`
- `apps/web/`
- `internal/platform/`
- `deployments/docker-compose.yml`
- `.env.example`
- CI 配置。

**完成项：**

- [x] API、Worker、Web、PostgreSQL、Redis 可本地一键启动。
- [x] `/health/live` 和 `/health/ready` 可用。
- [x] 数据库迁移可正向执行和回滚。
- [x] Go lint、单元测试、前端 lint 和构建进入 CI。

**当前实现说明（2026-06-11）：**

- 本地开发使用 `docker compose up --build`。
- 服务器部署使用 `.env.server`、`docker-compose.server.yml` 和 `scripts/deploy_server.sh`，Web 统一暴露端口并反代 API。
- 生产短信 Provider 尚未实现，因此服务器示例仍以 `OPCV2_ENV=development` 跑 MVP 演示；正式上线前必须切换真实短信服务商。

### 阶段 1：UI 基础与账号

**后端：**

- [x] 用户、登录记录和验证码限流。
- [x] Access Token + Refresh Token。
- [x] 手机号唯一约束。
- [ ] 开发环境提供固定测试验证码，生产环境必须接短信服务商。

**前端：**

- [x] 设计 Token 和公共组件。
- [x] 顶部六入口导航。
- [x] 登录页、首页、登录态首页。
- [x] 路由保护和头像菜单。

**当前实现说明（2026-06-11）：**

- 开发环境固定验证码为 `246810`，生产环境会拒绝使用开发短信配置；真实短信 Provider 尚未接入。
- Access Token 仅保存在前端内存，Refresh Token 使用 HttpOnly Cookie，并在刷新时轮换。
- 六个产品入口已建立受保护路由；业务页面将在对应后续阶段替换当前占位内容。
- 用户协议与隐私政策已有 MVP 说明页，正式上线前仍需按实际业务和法律意见定稿。

**验收：**

- 新用户通过验证码自动注册并登录。
- 未登录访问受限页面跳转登录，登录后返回原页面。
- UI 与登录、首页设计图逐项对照。

### 阶段 2：会员、积分和兑换码

- [x] 套餐和权益配置。
- [x] 用户订阅和权益查询。
- [x] 积分账户与账本。
- [x] 兑换码生成、兑换、防重复使用。
- [x] 会员计划页和兑换页。
- [ ] 管理后台套餐、兑换码和用户权益页面。

**当前实现说明（2026-06-11）：**

- 套餐目录先由服务端固定定义，并在迁移中写入 `membership_plans`，MVP 包含 Free 和 Pro。
- 用户默认 Free；兑换码可发放积分、开通 Pro，或两者同时发放。
- 兑换码按 `user_id + redemption_code_id` 幂等，同一用户重复兑换不会重复加积分。
- 积分余额与 `credit_transactions` 在同一事务内更新，当前验收已校验余额等于账本汇总。
- 已有会员中心与兑换码页面；完整运营管理后台还未实现，当前测试兑换码通过数据库插入。

**验收：**

- 权益完全由服务端判断。
- 同一兑换请求重复提交不会重复发放。
- 积分余额等于账本汇总结果。

### 阶段 3：AI 免费分析

- [ ] AI Provider 接口和一个正式模型实现。
- [ ] Tab A 方向卡、详细报告和 7 天计划的 Prompt、版本和 JSON Schema。
- [ ] Tab B 竞品拆解的 Prompt、版本和 JSON Schema。
- [ ] 信息完整度判断和最多两轮追问，每轮最多三个问题。
- [ ] 分析会话、历史记录、结果概览、详细报告。
- [ ] 7 天行动计划。
- [ ] SSE 输出真实阶段事件和流式报告。
- [ ] Stepper、追问卡片、结果揭示和减弱动画模式。
- [ ] 对应 4 张分析 UI。

**验收：**

- 输入信息不足时进入追问，而不是生成空洞报告。
- 追问不超过两轮，用户可跳过并生成。
- Tab A 每张方向卡的 7 个字段完整，详细报告包含 7 个章节。
- Tab B 搜索资料不足时明确提示，不虚构竞品数据。
- AI 非法 JSON 自动重试一次。
- 报告刷新后仍可从数据库恢复。
- 免费次数和会员权限生效。

### 阶段 4：天眼查获客与 Web 证据

- [ ] `LeadProvider` 接口。
- [ ] `TianyanchaProvider` 实现。
- [ ] 参考旧项目的有效做法，重新实现 `SearchProvider`、Serper、SerpAPI 和对应测试。
- [ ] 增加主 Provider 失败后的备用 Provider 切换。
- [ ] 获客任务创建、扣积分和 Asynq 入队。
- [ ] Worker 搜索企业并补充联系方式。
- [ ] Worker 按固定模板补充官网、新闻、联系页和招投标证据。
- [ ] 确定性去重和结果落库。
- [ ] 任务列表、进度、失败重试和取消。
- [ ] VIP 获客、任务列表、线索结果 3 张 UI。

**验收：**

- 输入“成都 教培”可创建异步任务并返回企业结果。
- 线索直接显示天眼查提供的电话、邮箱和网站。
- 线索可显示 Serper/SerpAPI 返回的官网、新闻和招投标原始链接。
- 页面没有联系方式验证状态或“100% 可达”承诺。
- Provider 配额耗尽时不无限重试，也不吞掉错误。
- 系统失败会返还本次任务消耗的积分。

### 阶段 5：CRM 闭环

- [ ] 单条和批量线索加入 CRM。
- [ ] 客户列表、筛选、详情。
- [ ] 阶段修改和跟进记录。
- [ ] 下次跟进提醒。
- [ ] AI 跟进话术。
- [ ] CRM 详情和全部跟进 2 张 UI。

**验收：**

- 重复导入同一条线索不会创建重复客户。
- 客户阶段修改有活动记录。
- 到期和逾期跟进可以正确筛选。

### 阶段 6：内容页面和后台配置

- [ ] 咨询通文章管理和展示。
- [ ] 可选：白名单 RSS/GDELT 定时导入文章候选，默认关闭。
- [ ] 工具分类、工具卡片和收藏。
- [ ] 社群二维码、权益、活动配置。
- [ ] 站点品牌数字和案例配置。
- [ ] 咨询通、工具箱、社群 UI。

**验收：**

- 运营可通过后台修改内容，无需发布代码。
- 没有内容时显示设计统一的空状态。

### 阶段 7：上线准备

- [ ] 安全检查：密钥、默认账号、JWT、CORS、限流、日志脱敏。
- [ ] 手机号、微信号、邮箱在日志中脱敏。
- [ ] 数据库备份和恢复演练。
- [ ] Provider 调用量、失败率和费用告警。
- [ ] 核心流程 Playwright E2E。
- [ ] 性能测试和慢 SQL 检查。
- [ ] 用户协议、隐私政策和公开数据使用提示。

---

## 9. 测试计划

### 9.1 Go 单元测试

- 权益判断。
- 积分扣减、幂等和失败补偿。
- 兑换码并发兑换。
- AI JSON 解析与重试。
- 天眼查响应映射。
- Serper/SerpAPI 请求、解析、故障切换和 URL 去重。
- 线索确定性去重。
- CRM 阶段变更。

### 9.2 集成测试

- PostgreSQL 唯一约束和事务。
- Redis 验证码、限流和 Asynq。
- API 鉴权和权限。
- 天眼查使用 Mock Server，不在 CI 调用真实计费 API。

### 9.3 前端测试

- 登录表单和协议勾选。
- AI 追问流程。
- 任务状态轮询。
- 线索批量选择并加入 CRM。
- 兑换码重复提交。

### 9.4 E2E 核心流程

```text
登录
-> 首页提交问题
-> 完成 AI 追问
-> 查看分析报告
-> 创建获客任务
-> 查看企业线索
-> 加入 CRM
-> 添加跟进记录
```

---

## 10. 非功能要求

- 普通 API：目标 `p95 < 500ms`，不包含 AI 和第三方 Provider。
- AI 和获客任务必须异步或流式，不阻塞普通 HTTP 请求。
- 第一版目标支持 1,000 日活、100 个并发用户。
- PostgreSQL 数据恢复目标：RPO 24 小时，RTO 4 小时。
- 所有写接口支持请求 ID，关键扣费接口支持幂等键。
- 日志必须包含 `request_id`、`user_id`、`task_id`，不得记录验证码、Token 和完整手机号。

---

## 11. 上线验收标准

第一版满足以下条件才可上线：

1. 17 个设计页面或状态均已实现，核心页面完成视觉对照验收。
2. 用户可以完成完整核心闭环，期间不需要管理员手工改数据库。
3. 天眼查正式授权、配额和数据保存/展示规则已书面确认。
4. 线索页面直接展示公开数据，不出现虚假的验证状态或可达承诺。
5. 会员权益、积分扣减和兑换码无法通过前端绕过。
6. AI、Provider 或 Worker 失败时，用户能看到明确状态并可重试。
7. 核心 Go 测试、前端测试和 E2E 流程全部通过。
8. 生产密钥不在代码仓库，数据库能够备份和恢复。

---

## 12. 开发顺序结论

严格按以下顺序推进：

```text
项目骨架
-> 登录与 UI 基础
-> 会员/积分
-> AI 免费分析
-> 天眼查获客与 Web 证据
-> CRM
-> 内容页面
-> 上线检查
```

不要先把全部静态 UI 做完再接后端。每个阶段都应交付一条可操作、可测试的小闭环，避免最后集中联调。
