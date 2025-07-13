# Workflow 

## Roles & Collaboration Model

This document outlines the collaboration model between the User (acting as Tech Lead) and the AI (Claude, acting as a senior developer/pair programmer). We will use a DIY Kanban system based on folders and Markdown files, managed with Git.

- Tech Lead (User) Responsibilities:
  - Define project architecture, goals and priorities.
  - Select which task to work on, create or refine
  - Create initial task Markdown files or request AI to draft them.
  - Review and approve or modify implementation plans proposed by the AI within task files.
  - Review agent's work
  - Answer agent's questions and resolve roadblocks.

- AI (Claude) Responsibilities:
  - Draft new task file in `kanban/backlog` (as per `kanban/backlog/T000_example.md`) on request
  - Refine selected tasks by proposing detailed "Implementation Plan & Progress" sections within the task's Markdown file. This includes sub-tasks, design considerations (e.g., function signatures, data models), and testing strategies.
  - Work only on tasks selected by the User.
  - Generate code, documentation, tests, or other artifacts as per the agreed sub-tasks.
  - Update the status of sub-tasks (`[ ]`, `[WIP]`, `[X]` (done), `[blocked]`) within the task's Markdown file.
  - Clearly indicate when a roadblock is encountered by marking the relevant sub-task `[blocked]` and adding a note in the "Roadblocks" section of the task file.
  - Create git commit messages according to the convention specified below.
  - Update task files with commit ids.
  - Always stop and wait for user input after modifying a task's Markdown file content (especially the plan or sub-task status) or when a stopping condition is met.
  - When submitting a task (or subtask) for user review, provide a detailed git commit message (see Commit Conventions). 

## Kanban System: Folders & Files

The project work is managed in `kanban/` and subdirectories: 

- kanban/
  - backlog/: Tasks not yet being worked on.
  - in-progress/: Tasks being actively worked on.
  - in-review/: Tasks the AI considers completed, pending user validation.
  - done/: Completed tasks, verified by the user.
  - archive/: Old completed tasks.

Each task is a single Markdown file (e.g., `T023_implement_option_parser.md`), where T023 is the task's unique ID (also stored in the file's frontmatter).

Tasks must adhere closely to the format of `../kanban/backlog/T000_example.md`.

## Task Dependencies & Relationships

Tasks often have relationships with each other that should be explicitly documented to ensure proper sequencing and coordination. These relationships include:

- **Dependency Types**:
  - Blocks: This task must be completed before another can start.
  - Depends: Another task must be completed before this task can start.
  - Related: Tasks share context / implementation details but don't block each other.
  - Part-of: Task is a sub-component of a larger program or epic.
  - Duplicates: Substantially overlaps with another task (one will typically be cancelled)

- Managing Dependencies:
  - Dependencies should be included in the task's Markdown frontmatter in the related_tasks field
  - The relationship type should be specified using the format: ["depends:T123", "blocks:task456"]
  - Dependencies should be mentioned in the task's Discussion Log when they impact progress
  - When a task or sub-task is blocked by a dependency, it should be marked as ["blocked"]

- Visualizing Dependencies:
  - Use a simple text diagram in the task file when complex dependencies exist
  - Example: T001 --> T002 --> T003

## AI Interaction with Task Markdown:

- When planning, AI will propose content for "3. Implementation Plan & Progress".
- When working, AI will update sub-task statuses (`[ ]`, `[WIP]`, `[x]`, `[blocked]`) in section 3.
- AI will add entries to "4. Roadblocks" and "5. Notes / Discussion Log".
- AI may edit task files in place, in which case it will always print the verbatim content of the changes and surrounding context to the user.

## Commit Checklist

When asked to "commit" or when work is complete on a task or subtask, follow the following checklist:
- Review the work completed since last commit (e.g. the current subtask), and
  - [ ] add any relevant notes to the Notes section of the task file, especially noting any unexpected issues encountered, human feedback, refactoring done, or key decisions made.
  - [ ] review files touched and important decisions / changes made, and add any Anchor Comments (per CLAUDE.md) as appropriate to aid future work.
  - [ ] format the code
  - [ ] lint the code and address any issues
  - [ ] stage and prepare a commit for all changes in the working directory
  - [ ] determine if the task requires or warrants user verification (manual testing or code review), and either
    - [ ] confirm with user before issuing commit command, or
    - [ ] commit, display committed message and commit hash, and check for acceptance.
  - [ ] add commit SHA to the task file for future reference.
  - [ ] note the next logical subtask to proceed with, if any

Then, if working on a task with multiple subtasks and another subtask is defined:
  - [ ] evaluate the detail captured in the next subtask
  - [ ] if it seems sufficiently clear, appropriate, and detailed, print it and suggest proceeding
  - [ ] if it looks like it could be improved, insert a new subtask to better plan the work.

### Git Workflow & Commit Conventions

- **Commit Conventions:**
  - ALL commits have a title (in conventional format) 
  - Non-trivial commit messages follow the title with a synopsis, then a bullet list summary of changes, separated by newlines.
  - **Commit message title** (first line)
    - Format: `type(scope): [TASK-ID/SUBTASK-ID] description`
      - Type: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`
      - Scope: Component or module affected (optional)
      - Task ID: Reference to the task this commit relates to
      - Subtask ID: Reference to the subtask this commit relates to (if any)
      - Description: Concise explanation in present tense
    - Examples:
      - `feat(auth)[T123/1.1]: implement login form`
      - `fix(api)[T234]: correct response format in user service`
      - `docs(readme)[T567/2.3]: update installation instructions`
  - Synopsis: a short description of the main purpose of the commit (1-2 sentences).
  - Bulleted Summary (include a subsection for each each relevant concern):
    - functional changes: feature additions
    - dependencies: libraries, plugins or packages added / removed / upgraded
    - refactoring: improvements made to existing code to improve code quality / adaptability
    - tests: automated tests written or modified
    - QA: other QA activities undertaken; manual test plans created
    - docs: documentation added or modified
    - feedback: user feedback addressed during work; summary of user inputs
    - security: security considerations: potential or addressed
    - performance: performance considerations, benchmarks 
    - maintainability: considerations for future extension

- **Commit Workflow:**
  - AI should suggest commit messages but not execute commits without user validation.
  - When a task or subtask is completed and validated by the user, AI can proceed with creating the commit.
  - After commit, the task file should be updated with the git commit ID for traceability.
  - Multiple related changes can be grouped in a single commit if they form a logical unit.

### Core Workflow

1. **Session Start / Context Restoration:**
  - AI will read all the files in the categories in-progress and all other files that will help building the context for this.
  - User: "Let's work on `task name`."
  - AI will move this task in the in progress folder and will create or improve an implementation plan if needed

2. **Task Creation (Optional - If AI assists):**
  - User: "Suggest a task for implementing X."
  - AI: Proposes a new task by drafting the full Markdown content for a new file (e.g., `kanban/backlog/new_task.md`), including basic sections (Goal, ACs) and an empty Implementation Plan.
  - IMPORTANT: before creating a new task, AI will ALWAYS check existing task IDs using `fd 'T[0-9]{3,3}\w*.md' kanban` to ensure unique, monotonic task IDs.
  - User: Reviews, modifies, saves the file to `backlog/`, and commits or can ask AI to proceed to some changes.

3. **Planning Phase (for a selected task):**
  - User: "Let's plan `in-progress/T123.md`." (User typically moves file to `in-progress/` before/during planning).
  - AI: Analyzes Goal & ACs. Proposes the "3. Implementation Plan & Progress" section with detailed sub-tasks, design notes, and testing strategies.
  - AI: **STOPS.** "I have updated `T123.md` with a proposed implementation plan. Please review. Here is the updated content: ... [provides full MD content] ..."
  - User: Reviews the plan *within the task.md file*. Makes edits directly or asks AI for revisions. Commits changes to the task.md file.

4. **Implementation Phase (Sub-task by Sub-task):**
  - User: "The plan for `T123.md` is approved. Let's start with sub-task 1.1: [Sub-task description]."
  - AI:
    1. Updates sub-task 1.1 status to `[ongoing]` in the Markdown.
    2. Focuses on sub-task 1.1: asks clarifying questions if needed, then generates code, documentation, test cases, etc., as per the design.
    3. Places generated content in section 6 or as appropriate.
    4. Updates sub-task 1.1 status to `[x]` (done) in the Markdown.
    5. Updates "Overall Status" in section 3 if a major phase is complete.
    6. AI: **STOPS.** "Sub-task 1.1 is complete. `T123.md` has been updated. Here is the new content: ... [provides full MD content] ... Ready for the next sub-task or your review. I suggest committing these changes with a message like: `feat(T123): Complete sub-task 1.1 - [brief description]`."
    *  User: Reviews AI's work, integrates code into the project, runs tests. Commits changes. "Okay, proceed with sub-task 1.2."

5. **Handling Roadblocks:**
  - AI (during a sub-task): "I've encountered a roadblock on sub-task X.Y: [description]."
  - AI: Updates sub-task X.Y status to `[blocked]` in the Markdown. Adds details to section "4. Roadblocks".
  - AI: **STOPS.** "Roadblock encountered and noted in `T123.md`. Here is the updated content: ... Please advise."
  - User: Resolves roadblock, provides guidance. Updates task.md if necessary (e.g., unblocks sub-task, modifies plan). Commits. "Okay, you can now proceed with sub-task X.Y."

6. **Task Completion:**
  - AI (after the last sub-task): "All sub-tasks for `T123.md` are complete. The overall status is now `[done]`."
  - AI: Updates the task.md file accordingly.
  - AI: **STOPS.** "`T123.md` is now complete. Here is the final content: ... I recommend moving it to the `done/` folder and committing. Suggested commit: `feat(T123): Complete [Task Title]`."
  - User: Moves file to `done/`. Makes final commits.

7. **"No Code Before Approved Plan":**
  - The AI must not generate implementation code or detailed artifacts for a task or sub-task if the "3. Implementation Plan & Progress" section for that task/sub-task has not been filled out and implicitly or explicitly approved by the User (i.e., User says "proceed with this plan" or "start sub-task X").

8. **Stopping Conditions (AI must stop and wait for User):**
  - After proposing or editing any part of the "Implementation Plan & Progress" section in a task.md.
  - After marking any sub-task as `[WIP]`, `[x]` (done), or `[blocked]` and providing the associated output/update.
  - When a roadblock is identified and documented.
  - When explicitly asked to stop by the User.
  - After completing all sub-tasks for a main task.