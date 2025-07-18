
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

**C4 Styling:**
- **Person**: `shape: person`, fill: `#08427b`, font-color: white
- **Software System**: Default shape, fill: `#1168bd`, font-color: white  
- **External System**: Default shape, fill: `#999999`, font-color: white
- **Container**: Default shape, fill: `#1168bd`, font-color: white
- **Database**: `shape: cylinder`, fill: `#999999`, font-color: white
- **Component**: Default shape, fill: `#85bbf0`, font-color: white
- **Needs Implementation**: Add `style.stroke-dash: 5`, fill: `#ffaaaa`

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
