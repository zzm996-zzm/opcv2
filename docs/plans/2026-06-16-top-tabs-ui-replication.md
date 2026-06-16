# Top Tabs UI Replication Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Continue one-to-one replication of the supplemental top-tab UI screenshots into working protected V4 routes.

**Architecture:** The app is a React/Vite frontend under `apps/web`. Protected routes are declared in `apps/web/src/App.tsx` and generally render inside `V4PageShell`. Each screenshot should map to a route, a focused page component under `apps/web/src/pages`, page-level tests, and scoped CSS in `apps/web/src/styles.css`.

**Tech Stack:** React, React Router, Vitest, Testing Library, Vite, shared CSS.

---

## Current Status

Branch: `feature/bootstrap`

Primary reference inventory:
- `docs/ui-inventory/top-tabs.md`

Supplemental UI source directory:
- `/Users/zzm/Library/Containers/com.tencent.xinWeChat/Data/Documents/xwechat_files/wxid_5iu3dbhushlw22_7043/msg/file/2026-06/顶部tab`

Completed top-tab screenshots:
- `AI教学/首页.png` -> `/learning`
- `AI教学/能力诊断.png` -> `/learning/diagnosis`

Implemented files for these two pages:
- `apps/web/src/pages/LearningPage.tsx`
- `apps/web/src/pages/LearningPage.test.tsx`
- `apps/web/src/pages/LearningDiagnosisPage.tsx`
- `apps/web/src/pages/LearningDiagnosisPage.test.tsx`
- `apps/web/src/App.tsx`
- `apps/web/src/styles.css`

Broader V4 shell and earlier UI work already exists in this same pending change set:
- `apps/web/src/components/V4PageShell.tsx`
- Home, auth, profile, membership, messages, help, projects, tasks, tools pages and tests under `apps/web/src/pages`

Latest verification before handoff:
- Command: `npm test && npm run build && npm run lint`
- Directory: `apps/web`
- Result before commit: 19 test files passed, 42 tests passed, production build passed, ESLint passed.

## Visual QA Notes

For protected-route screenshots, inject a development auth session in the browser before navigating:

```js
const { authSession } = await import('/src/lib/authSession.ts');
authSession.set({
  access_token: 'visual-token',
  access_token_expires_at: '2026-06-16T12:00:00Z',
  is_new_user: false,
  user: { id: 7, nickname: '张婧', phone: '13800138000', status: 'active' }
});
history.pushState(null, '', '/learning/diagnosis');
window.dispatchEvent(new PopStateEvent('popstate'));
```

Use at least:
- Desktop: 1920 x 1080
- Mobile: 390 x 844

Clean up after browser checks:

```bash
rm -rf .playwright-mcp learning-*.png
```

## Next Task: AI教学 / 能力评估

Target screenshot:
- `/Users/zzm/Library/Containers/com.tencent.xinWeChat/Data/Documents/xwechat_files/wxid_5iu3dbhushlw22_7043/msg/file/2026-06/顶部tab/AI教学/能力评估.png`

Target route:
- `/learning/assessment`

Recommended implementation:
1. Inspect the screenshot first with the image viewer.
2. Create `apps/web/src/pages/LearningAssessmentPage.tsx`.
3. Add route in `apps/web/src/App.tsx` before the generic `productRoutes` mapping.
4. Add `apps/web/src/pages/LearningAssessmentPage.test.tsx`.
5. Reuse existing diagnosis classes where visually identical, but keep page-specific classes for unique sections.
6. Run `npm test && npm run build && npm run lint` in `apps/web`.
7. Do desktop and mobile visual checks, then clean temporary screenshots and `.playwright-mcp`.
8. Update `docs/ui-inventory/top-tabs.md` to mark row 3 as completed and row 4 as next.

## Remaining Order After Ability Assessment

Continue exactly in `docs/ui-inventory/top-tabs.md` order:

1. `AI教学/差距分析.png` -> `/learning/gap-analysis`
2. `AI教学/推荐方案.png` -> `/learning/recommendation`
3. `AI教学/课程规划.png` -> `/learning/plan`
4. `AI教学/生成报告.png` -> `/learning/report`
5. `AI教学/全部课程.png` -> `/learning/courses`
6. `AI教学/全部推荐课程.png` -> `/learning/recommended-courses`
7. `AI教学/课程介绍页.png` -> `/learning/courses/intro`
8. `AI教学/课程内页.png` -> `/learning/courses/detail`
9. `AI教学/学习历史进度.png` -> `/learning/history`
10. Then continue with AI社群, 咨询通, 工具间, 智活 Copilot sections.

## Guardrails

- Preserve the user's requirement: UI should be as close to one-to-one as practical.
- Prefer existing V4 shell patterns and local page conventions.
- Keep each screenshot as a focused, testable route/state.
- Do not leave generated screenshots or `.playwright-mcp` in the repo.
- Before claiming completion, run fresh verification in `apps/web`:

```bash
npm test && npm run build && npm run lint
```
