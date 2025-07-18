
### Architecture Diagrams

**C4 Model + D2 Tooling**

Use D2 (d2lang.com) for all C4 architecture diagrams; follow C4 model conventions. 

See https://c4model.com/diagrams for reference.

- **Context Diagrams**: Show system boundaries and external dependencies
- **Container Diagrams**: Internal system components and their relationships  
- **Component Diagrams**: Detailed component architecture within containers
- **Flow Diagrams**: Process flows and decision points

**D2 C4 Conventions:**
- Place diagram source files in `doc/diagrams/` with `.d2` extension
- Generate SVG output (default): `d2 diagram_name.d2`
- Use markdown labels with C4 structure: `|md **Name** [Type] Description |`
- Include title with level indicator: "System Name - Context Level"

Invoking D2:
- dark-mode responsive diagrams: `d2 --theme=0 --dark-theme=200 input.d2`
- design docs and kanban cards should use `--sketch`; specifications should not.

**C4 Styling:**
- **Person**: `shape: person`
- **Software System**: Default shape
- **External System**: Default shape
- **Container**: Default shape
- **Database**: `shape: cylinder`
- **Component**: Default shape, fill: `#85bbf0`
- **Needs Implementation**: Add `style.stroke-dash: 5`

**Relationship Labels:**
- Use markdown format: `|md Description [Technology] |`
- Include technology type in brackets: [Function calls], [File I/O], [Library calls]

**Legends:**
- Include status legend for implementation status  
- Use consistent styling matching component colors
- Position: `near: bottom-right`

**Naming Convention:**
- Use descriptive names: `entry_system_context.d2`, `field_input_hierarchy.d2`
- Match markdown section titles for easy reference
- Reference SVG files in markdown: `![Description](/doc/diagrams/filename.svg)`