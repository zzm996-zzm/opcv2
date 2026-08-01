# Shared Route Shell UI Fixes Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Remove the first-batch blank-route and loading-shell defects without changing the existing navigation or page layout.

**Architecture:** Keep `V4PageShell` as the shared application chrome. Protected routes render a shell-level restore placeholder until authentication is ready, aliases reuse existing page components, and the final wildcard route renders an authenticated in-app 404. Global Copilot visibility is gated by authentication readiness. Reference routes render one visual page and no longer mount hidden legacy pages alongside it.

**Tech Stack:** React 18, React Router 7, TypeScript, Vitest, Testing Library, Vite.

---

### Task 1: Add shared restore and not-found views

**Files:**
- Create: `apps/web/src/pages/NotFoundPage.tsx`
- Modify: `apps/web/src/App.tsx`
- Modify: `apps/web/src/styles.css`

Render a shell-preserving restore state and an authenticated 404 with return-to-workbench and browser-back actions. Keep the restore state free of the floating Copilot until auth restoration completes.

### Task 2: Add route aliases and registration entry mode

**Files:**
- Modify: `apps/web/src/pages/LoginPage.tsx`
- Modify: `apps/web/src/App.tsx`

Allow `/register` to open the existing registration tab, and redirect `/account/settings` to `/profile/settings` without introducing a second layout.

### Task 3: Remove hidden duplicate page mounts

**Files:**
- Modify: `apps/web/src/App.tsx`

Keep the current reference page as the single rendered visual surface for task, competitor-data, competitor-monitoring, growth-calculator, and enterprise home routes. Remove imports that only served hidden legacy mounts.

### Task 4: Add regression tests

**Files:**
- Modify: `apps/web/src/App.test.tsx`
- Modify: `apps/web/src/pages/LoginPage.test.tsx`

Cover restore-shell rendering, orb suppression during restore, `/register`, `/account/settings`, wildcard 404, and the absence of hidden duplicate route pages.

### Task 5: Verify and commit

Run `npm test` and `npm run build` from `apps/web`. Review the diff, ensure no unrelated files changed, then commit the completed batch.
