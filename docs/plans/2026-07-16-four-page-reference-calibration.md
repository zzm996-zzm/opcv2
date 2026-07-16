# Four-Page Reference Calibration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Recalibrate the login, signed-in workbench, project market, and AI community pages against the supplied 2048px-wide UI references while preserving existing routes, APIs, and interactions.

**Architecture:** Introduce a dedicated `ReferenceShell` and a single `reference-ui.css` stylesheet using new `ref-*` class names, then migrate only the four baseline pages to that structure. This prevents the existing duplicated `v4-*`, login, project-market, and community rules from participating in the new layouts; obsolete page-specific rules are removed after migration instead of being overridden again.

**Tech Stack:** React 18, TypeScript, Vite, CSS, Vitest, in-app browser screenshots.

---

### Task 1: Establish the desktop reference frame

**Files:**
- Create: `apps/web/src/components/ReferenceShell.tsx`
- Create: `apps/web/src/reference-ui.css`
- Modify: `apps/web/src/main.tsx`
- Test: browser layout measurements at `2048x1123`

**Steps:**
1. Add shared reference variables for a 360px sidebar, 98px top bar, compact page padding, and the reference typography scale.
2. Build the shared sidebar, top navigation, account area, and workspace with unique `ref-*` classes.
3. Keep existing mobile breakpoints intact and add a desktop fallback below the full reference width.
4. Verify sidebar, top bar, account area, and content origins in the browser.

### Task 2: Recompose the login page

**Files:**
- Modify: `apps/web/src/pages/LoginPage.tsx`
- Modify: `apps/web/src/reference-ui.css`
- Test: `apps/web/src/App.test.tsx`

**Steps:**
1. Preserve the existing login/register form behavior and validation.
2. Match the reference split, brand position, story block, feature list, sculpture placement, and form-card dimensions.
3. Keep registration fields scrollable within the same card without changing the desktop login state.
4. Verify the login screen at `2048x1114` and a narrower desktop width.

### Task 3: Calibrate the signed-in workbench

**Files:**
- Modify: `apps/web/src/pages/HomePage.tsx`
- Modify: `apps/web/src/reference-ui.css`
- Test: `apps/web/src/pages/HomePage.test.tsx`

**Steps:**
1. Reuse the current data mapping and signed-in empty/error states.
2. Match the reference greeting row, metric strip, three primary domain panels, recommendation grid, and fixed Copilot panel.
3. Ensure content remains clipped and scrollable inside the reference frame instead of widening the viewport.
4. Verify `/\?devAuth=1` at `2048x1123`.

### Task 4: Calibrate the project market

**Files:**
- Modify: `apps/web/src/pages/ProjectsPage.tsx`
- Modify: `apps/web/src/reference-ui.css`
- Test: `apps/web/src/pages/ProjectsPage.test.tsx`

**Steps:**
1. Preserve all existing project-market routes and API-backed states.
2. Match the hero height, image crop, search bar overlap, badge strip, core-entry cards, and right Copilot width.
3. Use the existing project-market image assets as the visible hero and card artwork.
4. Verify `/projects?devAuth=1` at `2048x1123`.

### Task 5: Calibrate AI community

**Files:**
- Modify: `apps/web/src/components/CommunityReferencePage.tsx`
- Modify: `apps/web/src/reference-ui.css`
- Test: `apps/web/src/pages/CommunityPage.test.tsx`

**Steps:**
1. Preserve join flows, API configuration, and modal behavior.
2. Match the two community entry cards, lower three-column information row, growth strip, and fixed right Copilot panel.
3. Keep the enterprise card emphasis and supplied community artwork.
4. Verify `/community?devAuth=1` at `2048x1123`.

### Task 6: Regression and delivery

**Files:**
- Verify: `apps/web/src/**/*.test.tsx`
- Verify: `apps/web/src/reference-ui.css`

**Steps:**
1. Run the focused page tests.
2. Run `npm run build` and `npm run lint`.
3. Capture one screenshot for each reference page at the supplied dimensions.
4. Inspect text overflow, content overlap, horizontal scrolling, missing assets, and console errors.
5. Remove migrated legacy style blocks that no longer have consumers.
6. Commit the completed calibration.
