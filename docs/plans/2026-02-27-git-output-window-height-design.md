# Git Output Window Height Fix Design

Date: 2026-02-27

## Summary
On `/Git`, the output window leaves ~100px blank space at the bottom. We will make it stick to the bottom and automatically fill remaining height using layout-driven CSS, avoiding fragile JS height math.

## Context and Root Cause
- The `shell.calculateShellDivHeight` function computes height based on an element with id `mainCard`.
- The Git page (`web/src/components/Git.vue`) does not render an element with id `mainCard`, so the calculation returns early and leaves `shellController.divHeight` at its default (250).
- The output area is then constrained to this fixed height, leaving unused space below.

## Goals
- Output window on `/Git` should be flush with the bottom of the viewport.
- Output window should expand/shrink with viewport size changes.
- No behavioral changes to Git actions or shell streaming.

## Non-Goals
- Refactor global shell height logic across all pages.
- Change any backend API or SSE behavior.

## Proposed Approach (Recommended)
Use CSS layout to define height:
- Ensure the Git page container is a full-height flex column.
- Keep the header fixed height (flex-shrink: 0).
- Make the output card `flex: 1` and `min-height: 0` so it fills the remaining space.
- Let the output content use `height: 100%` of the output card without relying on JS.

This removes the dependence on the missing `mainCard` element and makes layout responsive by default.

## Alternatives Considered
1. **Adjust JS height calculation** by targeting Git-specific elements.
   - Pros: minimal CSS change.
   - Cons: brittle and still tied to DOM structure.
2. **Hybrid CSS + JS**: CSS for layout, JS for inner scrollbar.
   - Pros: robust layout with existing scrolling logic.
   - Cons: extra complexity, more moving parts.

## Risks
- If parent containers do not have a constrained height, the flex layout may not take effect.
  - Mitigation: verify parent container height rules and add minimal fixes if needed.

## Testing Plan
- Manual: open `/Git` at common viewport sizes; verify output window touches bottom.
- Manual: resize browser window; output area should grow/shrink smoothly.

