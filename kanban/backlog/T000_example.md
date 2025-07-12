---
title: "Descriptive Task Title"
type: ["feature"] # feature | fix | documentation | testing | refactor | chore
tags: ["parser"] # optional
related_tasks: ["depends-on:task123", "blocks:task456", "related-to:task789"] # Optional with relationship type
context_windows: ["./*.go", Claude.md, Architecture.md] # List of glob patterns useful to build the context window required for this task
---

# {{ title }}

**Context (Background)**:
*AI to complete*

**Context (Significant Code Files)**:
*AI to complete*

## 1. Goal / User Story

Brief description of what needs to be achieved, why, and for whom. Why is this task important?

## 2. Acceptance Criteria

A checklist of conditions that must be met for the task to be considered complete. User-focused.

- [ ] Criterion 1
- [ ] Criterion 2

---
## 3. Implementation Plan & Progress

**Overall Status:** `Not Started` | `In Progress` | `Completed` | `Blocked`
*AI updates this based on sub-task progress or user instruction*

**Sub-tasks:**
*(Sub-task status: `[ ]` = todo, `[WIP]` = currently being worked on by AI , `[x]` = done, `[blocked]` = blocked)*

- [ ] **High-Level Phase/Component 1**: (Optional description)
  - [ ] **Detailed Sub-task 1.1:** (Description of sub-task)
    - *Design:* (Brief notes on approach, function signatures, data structures, API endpoints, UI elements, etc. AI proposes this.)
    - *Code/Artifacts to be created or modified:* (e.g., `user_service.go`, API documentation update)
    - *Testing Strategy:* (Unit tests, integration tests, manual checks to be performed by user. AI proposes this.)
    - *AI Notes:* (Internal thoughts, questions for this sub-task, recommended future changes not in scope)
    - [ ] **Detailed Sub-task 1.2:** ...
- [ ] **High-Level Phase/Component 2**
  - [ ] **Detailed Sub-task 2.1:** ...
    - *Design:* ...
    - *Testing Strategy:* ...

## 4. Roadblocks

*(Timestamped list of any impediments. AI adds here when a sub-task is marked `[blocked]`)*
- `YYYY-MM-DD HH:MM - [Sub-task ID/Name]:` Description of roadblock.

## 5. Notes / Discussion Log

*(Timestamped notes, decisions, clarifications from User or AI during the task's lifecycle)*
- `YYYY-MM-DD HH:MM - User:` ...
- `YYYY-MM-DD HH:MM - AI:` ...

## 6. Code Snippets & Artifacts 

*(AI will place larger generated code blocks or references to files here if planned / directed. User will then move these to actual project files.)*