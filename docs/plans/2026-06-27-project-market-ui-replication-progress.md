# 项目确定及拆解 UI 复刻进度

日期：2026-06-27

分支：`feature/bootstrap`

参考目录：

`/Users/zzm/Library/Containers/com.tencent.xinWeChat/Data/Documents/xwechat_files/wxid_5iu3dbhushlw22_7043/msg/file/2026-06/项目确定及拆解/`

## 已完成范围

本轮重点完成了 `项目确定及拆解/项目超市` 相关页面的一比一视觉复刻推进，主要落在 `apps/web/src/styles.css`，并配合少量页面结构调整与测试更新。

已覆盖页面/状态：

- 项目超市首页
- AI 匹配页
- 匹配结果页
- 项目详情页
- 项目对比页
- 导出匹配报告弹窗
- 付费解锁弹窗
- 机会探索页
- 真实案例库页
- AI 补充提问页
- 匹配历史与收藏页

同时已完成前序公共 UI 组：

- 首页与认证相关页面
- 登录页
- 注册资料页
- 个人资料与账户设置
- 会员页与升级弹窗
- 消息页
- 帮助与反馈相关状态

## 关键实现记录

- 受保护路由截图时曾临时加入 `__preview_auth=1` 本地预览开关。
- 每轮视觉验证后均已删除临时预览开关；当前 `apps/web/src/App.tsx` 无残留 diff。
- Playwright 截图文件与 `.playwright-mcp/` 已清理，不应进入提交。
- 右侧 Copilot、结果卡、详情指标、路径时间线、案例卡、历史记录表格等均做了桌面宽度下的布局修正，避免遮挡、竖排和溢出。

## 最新验证

已通过：

```bash
npm --prefix apps/web test -- ProjectsPage.test.tsx
npm --prefix apps/web test -- ProfilePage.test.tsx HelpPage.test.tsx MembershipPage.test.tsx MessagesPage.test.tsx ProjectsPage.test.tsx
npm --prefix apps/web run build
```

结果：

- `ProjectsPage.test.tsx`：11/11 通过
- 扩展回归：5 个测试文件，25/25 通过
- 构建通过
- 构建仍有既有 Vite 大 chunk 警告，未在本轮处理

## 注意事项

- 后端未启动时，历史页会出现 `/api/v1/projects/matches` 代理报错；页面 fallback 能渲染，前端测试和构建不受影响。
- `internal/membership/postgres_repository.go` 当前仅有一个空行格式改动，和 UI 复刻无业务关联。
- 后续继续复刻时，优先沿用已追加的 `project market` scoped CSS 覆盖方式，避免重排大 CSS 文件。

## 下一步建议

继续按参考目录推进其他模块页面：

1. 商业沙盘
2. 落地模块
3. 增长模块
4. 其他「项目确定及拆解」子目录里的弹窗和状态页

每个子模块建议保持当前节奏：

1. 先检查参考图和现有路由/组件
2. 尽量做 scoped CSS 覆盖
3. 用临时预览 auth 截图
4. 删除临时 auth
5. 跑相关测试与构建
6. 更新本进度记录或新增对应模块记录
