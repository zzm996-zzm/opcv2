# 落地模块参考 UI 设计

## 目标

将任务中心、竞品全盘数据破解、竞品动态监测和增长测算的 27 张参考图复刻到现有 V4 工作台。保留现有 API 与业务动作；后端无数据或离线时，使用与参考图一致的示例内容维持完整布局。

## 页面与路由

- 任务中心：`/tasks`、`/tasks/board`、`/tasks/calendar`、`/tasks/new`、`/tasks/new/menu`、`/tasks/ai`、`/tasks/detail`、`/tasks/menu`
- 竞品破解：`/competitor-data`、`/competitor-data/progress`、`/competitor-data/history`、`/competitor-data/results/:section`
- 动态监测：`/competitor-monitoring`、`/competitor-monitoring/history`、`/competitor-monitoring/progress`、`/competitor-monitoring/analysis`
- 增长测算：`/growth-calculator`、`/growth-calculator/questions`、`/growth-calculator/history`、`/growth-calculator/report`

## 组件结构

新增共享落地工作台组件，负责：

- V4 主导航与侧栏复用
- 页面主区和右侧 Copilot 的稳定两栏布局
- 任务表格、看板、日历、表单和详情状态
- 查询首页、处理进度、历史列表和多维结果页
- 监测首页、过程、时间线与 AI 分析页
- 测算首页、补充提问、历史和报告页

现有 API 页面继续保留，新增参考页面只负责路由展示和状态组织。需要真实操作的主入口继续链接或调用已有页面动作。

## 验证

- 在 `1672x941` 桌面视口逐类截图核对
- 在 `390x844` 检查无横向溢出和内容遮挡
- 运行相关单元测试、ESLint、生产构建和全量测试
- 提交后使用仓库根目录 `update.sh` 部署并验证线上地址
