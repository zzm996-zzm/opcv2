# Shared Navigation Shell

## Decision

All authenticated application pages use `V4PageShell` as the single owner of the sidebar and top navigation. `ReferenceShell` remains only as a compatibility adapter for the home and community page APIs; it delegates rendering to `V4PageShell` and does not contain navigation markup or route data.

The canonical sidebar uses one menu definition for every route. Task and CRM pages must not substitute labels, destinations, or group names. Active state is derived from the current pathname while the menu structure remains unchanged.

## Styling Contract

`v4-shell-consistency.css` is loaded after page styles and owns application chrome dimensions, spacing, colors, responsive behavior, account presentation, notification icon geometry, and active navigation indicators. Desktop pages use a 260px fixed sidebar and 78px top bar. At 1180px and below, the sidebar is hidden and the top navigation becomes a horizontally scrollable 64px bar.

Page styles may control only their content inside `.v4-page-main`. They must not resize or reposition `.v4-sidebar`, `.v4-workspace`, `.v4-topbar`, `.v4-topnav`, or account controls. Route entry animation is disabled for the shell so navigation does not shift during page changes; content animations remain available inside the main area.

## Verification

Component tests assert that task, CRM, and Copilot routes render the same canonical menu and that `ReferenceShell` delegates to the V4 structure. Browser checks compare sidebar, top bar, and top navigation rectangles across home, Copilot, Insights, Community, CRM, Projects, and Sandbox routes at 1672x941, plus a 390x844 mobile overflow check.
