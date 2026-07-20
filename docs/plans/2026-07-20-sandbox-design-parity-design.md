# 商业沙盘九状态还原与真实流转设计

## 目标

以 `/Users/zzm/Desktop/增长/项目确定及拆解/商业沙盘/` 中的 9 张设计图为验收基准，补齐商业沙盘从首页、智能提问、回答完成、角色选择、准备启动、推演过程、报告、历史到额度弹窗的完整业务流。所有会话信息、问题、回答、启动设置、历史记录和报告均通过 `/api/v1/sandbox/*` 接口读取或写入，生产页面不保存前端演示业务数据。

## 方案选择

### 方案 A：继续修补现有单文件页面

改动最少，但 `SandboxPage.tsx` 已同时承担路由、请求、状态和九类视图，现有沙盘 CSS 又位于超大号全局样式文件尾部。继续追加覆盖规则会放大状态串页和响应式回归风险，不适合作为上线版本。

### 方案 B：隔离式沙盘模块，复用现有业务能力

这是采用方案。保留现有认证、会员额度、沙盘会话、异步任务和 AI 报告能力；新增结构化会话上下文、问题回答、启动设置和演示目录接口。前端把九个状态拆成沙盘专用视图，并使用独立的 `sandbox-reference.css`，避免与旧 `sandbox-*` 规则竞争。

### 方案 C：按设计图写静态状态页

视觉实现最快，但无法形成真实流程，也会重复此前“有页面、无入口、按钮无行为”的问题，因此不采用。

## 页面与流转

主流程：

`/sandbox -> /sandbox/setup?session=:id -> /sandbox/questions?session=:id -> /sandbox/setup?session=:id&complete=1 -> /sandbox/roles?session=:id -> /sandbox/start?session=:id -> /sandbox/run?session=:id -> /sandbox/sessions/:id/report`

辅助流程：

- 首页和 Copilot 可进入 `/sandbox/history`。
- 历史记录的已完成会话进入具体报告；草稿进入上次未完成步骤；失败或取消会话进入运行页重试。
- 启动接口返回 `quota_exceeded` 或本地额度摘要耗尽时，保留当前页面并打开额度弹窗，不再跳到孤立的调试 URL。
- `/sandbox/quota` 保留为可分享的额度状态路由，但弹窗的关闭、升级和返回均可操作。

## 数据模型

在 `sandbox_sessions` 增加两个 JSONB 字段：

- `context`：初始描述、行业、目标用户、产品定位、价格区间、销售场景，以及结构化问题回答。
- `settings`：推演深度、输出风格、是否生成大纲和可选高级设置。

会话仍以现有 `goal`、`target_users`、`product`、`roles` 作为核心运行字段，保证现有 AI 任务和历史数据兼容。旧记录读取时为 `context` 与 `settings` 补默认值，不要求数据回填。

新增接口：

- `GET /api/v1/sandbox/questions`：返回服务端问题目录和识别字段定义。
- `PUT /api/v1/sandbox/sessions/:id/answers`：保存当前问题答案并返回更新后的会话。
- `GET /api/v1/sandbox/examples`：返回明确标记为演示数据的历史与报告，用于空账号的产品验收，不混入真实用户记录。

扩展 `PATCH /sessions/:id/draft` 保存上下文与设置；扩展 `GET /sessions` 支持状态、角色、关键词和时间范围筛选，同时返回稳定分页元数据。

## 前端结构

- `SandboxPage.tsx` 只负责识别路由状态、加载会话和协调 API。
- `components/sandbox/` 放置页面框架、Copilot、步骤条、角色卡、状态徽标和额度弹窗。
- `pages/sandbox/` 放置九个业务视图，所有控件必须有真实状态与行为。
- `sandbox-reference.css` 使用 `sb-*` 前缀和设计图的 1672 x 941 比例建立桌面布局；在 1180px 以下转为内容单列，在 760px 以下隐藏固定侧栏并保证无横向溢出。
- 首页使用已有 `/sandbox/home-hero.jpg`；角色卡使用已有 `/sandbox/role-*.jpg`，不再用 CSS 占位图形。

## 视觉验收

- 桌面基准为 1672 x 941，侧栏、顶栏、主栏和 Copilot 的宽度与设计图误差不超过 8px；主要卡片和 CTA 误差不超过 6px。
- 九个状态分别截图；动态状态允许内容文本不同，但信息层级、控件位置、密度、选中态和空/错/加载态一致。
- 390 x 844 检查文本溢出、横向滚动、卡片重排、弹窗和底部操作区。

## 错误处理与兼容

- 每个接口失败显示就地错误并保留用户已输入内容。
- 刷新时根据 `session` 查询参数恢复服务端会话，不依赖 `location.state`。
- queued/running 会话持续轮询；completed 自动展示完整结果入口；failed/canceled 提供重试。
- 空历史优先展示真实空状态，并提供“查看演示记录”切换；演示记录有显著标识，避免误认为用户数据。

## 测试

- Go：迁移、仓储扫描、草稿扩展、问答保存、筛选分页、演示数据和兼容旧记录。
- Web：九个路由、主流程跳转、问题逐题保存、设置更新、额度弹窗、历史筛选、推演轮询、报告和错误状态。
- 运行商业沙盘专项测试、Web 全量测试、Go 全量测试、lint 和生产构建。

