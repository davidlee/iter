<img src="https://r2cdn.perplexity.ai/pplx-full-logo-primary-dark%402x.png" class="logo" width="120"/>

## Adapting SM-2 and SRS Algorithms to Incremental Writing and Task Management

### Background: SM-2 and Incremental Practices

- **SM-2** is the original Spaced Repetition (SRS) algorithm used by SuperMemo and popularized in tools like Anki. It schedules items for review based on recall performance, dynamically extending intervals to optimize retention[^1][^2].
- **Incremental writing** involves the gradual development and refinement of ideas and documents, typically returning to drafts at spaced intervals, rather than linear completion[^3].
- **Task management with SRS** involves “submerging” tasks into a pool and resurfacing them for action or review at calculated intervals.


### Prior Art

#### 1. **Incremental Reading (SuperMemo, Obsidian, Anki)**

- SuperMemo pioneered incremental reading, where excerpts and notes are reviewed and expanded using SRS. This allows simultaneous progress on many texts, reprioritizing as knowledge/needs evolve[^4][^3][^5].
- Incremental writing has been integrated with SRS in various note-taking platforms (Obsidian plugins, Anki extensions). Notes or writing fragments are queued and scheduled for review or further elaboration[^6][^7][^8][^9].


#### 2. **Tasklists in SRS Context**

- SuperMemo’s tasklist manager combines prioritization and spaced review. Tasks can be reviewed, postponed, updated, or promoted based on changing priorities, with features for introducing tasks into the SM-2 schedule (“memorizing” a task)[^10][^11].
- Users have adapted SRS (Anki, RemNote, and others) to habit formation and recurring task reminders, treating each task as an “item” and using review intervals to refresh one's engagement with delayed or deferred tasks[^12][^13][^14].


#### 3. **Academic and Workflow Use Cases**

- Notes, research problems, and partially developed ideas are often added to an “inbox” or list. Review and work on these notes uses SRS to periodically surface items for attention, preventing them from stagnating or being forgotten[^15][^16][^17].


### Approaches for Adapting SM-2

| Approach | Description | Pros | Cons |
| :-- | :-- | :-- | :-- |
| **Item-level Incremental Writing** | Each draft, idea, or paragraph is treated as an “item” in an SM-2 schedule. Review sessions involve editing, expanding, or refactoring the item. | Easy to implement; smooth context-switching; supports vast “idea gardens.” | Item atomization is difficult; potential cognitive load if too granular. |
| **Task Submersion \& Surfacing** | Every task/subtask is inserted into an SRS queue. Review responses (e.g., progress, blocked, not started) dictate next interval. | Tasks won’t be forgotten; can batch similar tasks; adjustable to urgency. | SM-2’s focus on memory, not task urgency, may lead to out-of-sync priorities; not optimized for dependencies. |
| **Dynamic Prioritization** | Use SM-2’s “ease factor” and interval as a priority estimator: highly active or important tasks/items are reviewed frequently; others fade. | Hybridizes value-urgency; automates triage; fits both knowledge fragments and tasks. | May not suit deadlines; items could drop below awareness threshold (“eternal backlog”). |
| **Quality-based Feedback** | On reviewing an item/task, users rate “progress” or “clarity” (from 0-5 as in SM-2). Low scores reset interval or trigger deeper review/edit; high scores delay resurfacing. | Flexible; lets user surface only what most needs attention; handles “blocked” or “deferred” states well. | Requires regular, honest feedback; more subjective than rote memory tasks. |
| **Incremental Merging and Consolidation** | As in SuperMemo's writing process, multiple “atomic” ideas are eventually merged during the consolidation phase, with SRS ensuring previously outlined points return for context before integration[^3]. | Prevents lost threads; supports synthesis for complex documents. | Requires additional tooling for merging; cognitive complexity increases with note granularity. |

### Novel and Experimental Directions

- **Serendipitous Idea Collision:** By leveraging SRS’s pseudo-random resurfacing, you can prompt unexpected combinations of ideas or tasks during review, encouraging creativity or cross-pollination[^17][^3].
- **Value/Time Priority Heuristics:** As in SuperMemo's tasklists, integrate time-to-completion and estimated value directly into SRS scheduling, “promoting” items that become more urgent or valuable over time (functionally overlaying Eisenhower matrix style prioritization)[^10][^18].
- **Auto-archiving and Dormancy:** If an item repeatedly scores low (uninteresting, blocked), the system can auto-archive it, making room for active tasks or ideas, but allowing for periodic resurfacing to prevent permanent loss[^16][^10].
- **Context-aware Scheduling:** Algorithms can factor in upcoming deadlines or dependencies, adjusting “ease factor” or overriding SM-2 intervals to resurface items/tasks as required—a blend of SRS and traditional task managers[^19][^18].


### Tradeoffs

- **SM-2 for Non-memorization Tasks:** The fit is strongest where maintenance and gradual improvement are valued over immediate completion. Adapting it for deadlines or multitasking with dependencies may require augmenting with prioritization frameworks or manual overrides.
- **Cognitive Load:** Overloading the system with too many items (fragments/tasks) can lead to “review fatigue” if intervals or granularity aren’t well-regulated[^15][^20].
- **Subjectivity in Ratings:** Unlike factual recall, evaluating progress on ideas or tasks is subjective, potentially reducing algorithmic reliability if not paired with discipline and good review hygiene[^15][^16].
- **Discovery vs. Efficiency:** Random resurfacing can yield creative breakthroughs, but may feel inefficient compared to tightly managed, goal-driven workflows. Balancing serendipity and structure is key[^17][^3][^15].


### Summary

Adapting SM-2 (and similar SRS algorithms) to incremental writing and task management is already a part of advanced knowledge and productivity workflows, especially with tools inspired by SuperMemo. The core tradeoff is between maintaining a living backlog of ideas/tasks versus the precision needed for prioritization and completion. Using SM-2’s exponential scheduling and feedback ratings fits best where gradual progress and retention matter more than strict deadlines. Blending SRS with additional prioritization, dormancy controls, and creative surfacing strategies extends its utility to the incremental development of both documents and to-do lists[^15][^17][^16][^3][^10].

<div style="text-align: center">⁂</div>

[^1]: https://github.com/thyagoluciano/sm2

[^2]: https://en.wikipedia.org/wiki/SuperMemo

[^3]: https://supermemo.guru/wiki/Incremental_writing

[^4]: https://help.supermemo.org/wiki/Incremental_reading

[^5]: https://soki.ai/insights/what-is-incremental-reading

[^6]: https://forum.obsidian.md/t/implementing-spaced-repetition-for-incremental-writing-workflow/92361

[^7]: https://www.obsidianstats.com/plugins/obsidian-incremental-writing

[^8]: https://forums.ankiweb.net/t/incremental-reading-in-long-term-future/596

[^9]: https://github.com/bjsi/incremental-writing

[^10]: https://help.supermemo.org/wiki/Tasklist_manager

[^11]: https://super-memory.com/help/overviews.htm

[^12]: https://andrewtmckenzie.com/spaced_repetition

[^13]: https://www.reddit.com/r/todoist/comments/189osyn/spaced_repetition/

[^14]: https://notes.andymatuschak.org/Spaced_repetition_systems_can_be_used_to_program_attention

[^15]: https://notes.andymatuschak.org/Spaced_repetition_may_be_a_helpful_tool_to_incrementally_develop_inklings

[^16]: https://notes.andymatuschak.org/Spaced_repetition_may_be_a_helpful_tool_to_incrementally_develop_inklings?stackedNotes=zUP4GuzPF33dWkZPiu9N6V5

[^17]: https://cesarrodrig.github.io/garden/implementing-a-spaced-repetition-writing-system.html

[^18]: https://help.supermemo.org/wiki/Features

[^19]: https://en.wikipedia.org/wiki/Spaced_repetition

[^20]: https://supermemo.guru/wiki/Advantages_of_incremental_reading

[^21]: https://www.reddit.com/r/Anki/comments/1hmiker/how_do_i_incremental_read_with_anki/

[^22]: https://masterhowtolearn.wordpress.com/2018/10/30/is-sm-17-in-supermemo-better-than-sm-2-in-anki/

[^23]: https://trainingindustry.com/articles/strategy-alignment-and-planning/boost-learning-with-a-simple-cognitive-trick-spaced-repetition/

[^24]: https://news.ycombinator.com/item?id=35511357

[^25]: https://forums.ankiweb.net/t/sm-2-algorithm-pseudo-code/8350

[^26]: https://www.cultureamp.com/blog/spaced-repetition-learning-development

[^27]: https://pmc.ncbi.nlm.nih.gov/articles/PMC8594904/

[^28]: https://www.youtube.com/watch?v=v2asudkSFek

[^29]: https://e-student.org/spaced-repetition/

[^30]: https://news.ycombinator.com/item?id=34152100

[^31]: https://forum.artofmemory.com/t/anyone-interested-in-trying-supermemo-incremental-reading/54883

[^32]: https://snappify.com/blog/spaced-repetition

[^33]: http://manuals.ipaustralia.gov.au/patent/8.4.2-prior-art-information

[^34]: https://stackoverflow.com/questions/49047159/spaced-repetition-algorithm-from-supermemo-sm-2

[^35]: https://www.reddit.com/r/ObsidianMD/comments/mv6lek/has_anyone_managed_to_implement_a_spaced/

[^36]: https://www.geeksforgeeks.org/how-to-write-a-good-srs-for-your-project/

[^37]: https://groups.google.com/g/tiddlywiki/c/HoJ_4oVtxkk

[^38]: https://www.stephenmwangi.com/obsidian-spaced-repetition/resources/

[^39]: https://www.reddit.com/r/PKMS/comments/1ecy6la/combat_information_overload_and_forgetting_with/

[^40]: https://controlaltbackspace.org/spacing-algorithm/

[^41]: https://www.reddit.com/r/Anki/comments/vy9k7k/sm2_algorithm_and_cards_scheduled_for_the_past/

[^42]: https://www.reddit.com/r/Anki/comments/18w5b2c/incremental_reading/

[^43]: https://www.ollielovell.com/spaced-repetition-incremental-reading-anki-dendro/

[^44]: https://help.remnote.com/en/articles/6026144-the-anki-sm-2-spaced-repetition-algorithm

[^45]: https://noji.io/pl/blog/spaced-repetition/?from=ankipro

[^46]: https://www.bcu.ac.uk/exams-and-revision/best-ways-to-revise/spaced-repetition

[^47]: https://training.safetyculture.com/blog/how-spaced-repetition-works/

[^48]: https://parm.com/en/prioritizing-tasks-with-the-eisenhower-matrix/

[^49]: https://www.bmc.net/blog/management-and-leadership-articles/task-prioritization-in-management-and-leadership

[^50]: https://www.supermemo.com/en/blog/customize-your-learning-with-supermemo

[^51]: https://www.reddit.com/r/TooAfraidToAsk/comments/w7pras/job_interview_question_how_do_you_handle_a/

[^52]: https://uk.indeed.com/career-advice/interviewing/conflicting-priorities-interview-question

[^53]: https://getrapl.com/blog/revolutionize-training-top-10-spaced-repetition-platforms-and-their-benefits/

[^54]: https://www.youtube.com/watch?v=tkIwKk8toX4

[^55]: https://www.monitask.com/en/blog/master-your-task-management-how-the-1-3-5-rule-revolutionizes-to-do-lists

[^56]: https://supermemo.guru/wiki/Planning_a_perfect_productive_day_without_stress

