# Tools UI Parity Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Bring the tool library, tool detail, recommendations, and recommendation plan routes into close visual and behavioral parity with the supplied tool-room reference images.

**Architecture:** Keep the existing `ToolsPage` route variants and content API integration. Add deterministic local reference fallbacks for known catalog tools, expose the existing full-library route from the Copilot control, render page-specific Copilot content, and use styles scoped under `.tools-shell` so other V4 pages are unaffected.

**Tech Stack:** React 18, React Router, TypeScript, lucide-react, Vitest, Testing Library, Vite, CSS.

---

### Task 1: Lock the missing states with tests

**Files:**
- Modify: `apps/web/src/pages/ToolsPage.test.tsx`

**Step 1: Add a failing detail fallback test**

Render `/tools/detail?tool=midjourney-ai` with a failed content request and assert that the page still renders `Midjourney`, `专业 AI 图像生成工具`, and no synchronization warning.

**Step 2: Add a failing complete-library navigation test**

Render the library variant and assert that the Copilot collapse control links to `/tools/all`; render the full variant and assert all eight reference tools, four filters, and the floating Copilot entry are present.

**Step 3: Add page-specific Copilot assertions**

Verify that recommendation and plan variants expose the reference-specific prompts instead of only the generic quick actions.

**Step 4: Run focused tests and confirm the new assertions fail**

Run: `npm test -- ToolsPage.test.tsx`

Expected: failures for the missing alias fallback, full-state navigation, floating entry, and variant-specific assistant content.

### Task 2: Correct the page data and interaction states

**Files:**
- Modify: `apps/web/src/pages/ToolsPage.tsx`

**Step 1: Add known-tool fallback resolution**

Resolve API failures by normalized slug/name aliases and build complete local `ContentTool` detail data for Midjourney. Only show a synchronization warning when neither API nor local reference data can represent the requested tool.

**Step 2: Expose the full-library state**

Make the Copilot collapse control on the library route navigate to `/tools/all`, preserving the existing route as the reference's full-width catalog state.

**Step 3: Match recommendation content and actions**

Render the single-line reference summary, keep the scenario field accessible for API submission, remove the extra bottom refresh button, and move refresh into the recommendation-specific Copilot prompt list.

**Step 4: Match plan and detail semantics**

Use the reference badge, labels, structured detail rows, and correct tool-specific links. Replace text glyphs with lucide icons where the design uses familiar interface symbols.

**Step 5: Implement page-specific Copilot bodies**

Render library, detail, recommendation, and plan assistant messages and suggested actions from a variant configuration, including the plan CTA card.

### Task 3: Align the scoped visual system

**Files:**
- Modify: `apps/web/src/styles.css`

**Step 1: Remove the narrow centered tools canvas**

Let the tools workspace use the available V4 content width and size the Copilot rail responsively to match the reference proportions.

**Step 2: Align library cards and filters**

Match the flat white surfaces, compact radii, card dimensions, four-column full state, three-column Copilot state, pagination, and filter alignment.

**Step 3: Align detail, recommendation, and plan layouts**

Flatten nested cards where the reference uses divided rows, adjust spacing and type sizes, restore the actual preview image, and position flow arrows and side panels consistently.

**Step 4: Add the floating Copilot launcher**

Render the circular branded launcher on all tools variants without overlapping the assistant input or page actions.

**Step 5: Preserve responsive behavior**

At tablet and mobile breakpoints, collapse multi-column layouts cleanly and keep controls, text, and cards within the viewport.

### Task 4: Verify and commit

**Files:**
- Verify: `apps/web/src/pages/ToolsPage.test.tsx`
- Verify: `apps/web/src/pages/ToolsPage.tsx`
- Verify: `apps/web/src/styles.css`

**Step 1: Run focused tests**

Run: `npm test -- ToolsPage.test.tsx`

Expected: all tool-page tests pass.

**Step 2: Run the full frontend checks**

Run: `npm test && npm run build && npm run lint`

Expected: all tests pass, Vite production build succeeds, and ESLint reports no errors.

**Step 3: Perform visual verification**

Run the Vite development server and capture `/tools`, `/tools/all`, `/tools/detail?tool=midjourney`, `/tools/recommend`, and `/tools/recommendation-plan` at `1672x941`, plus one mobile check at `390x844`. Compare the desktop captures against the supplied reference images and correct overflow, overlap, or blank-canvas issues before completion.

**Step 4: Commit the implementation**

Stage only the tools page, its tests, and scoped style changes. Commit with a concise message describing the UI parity work.
