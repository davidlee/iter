# Manual Testing Guide for T009/1.4

## Dry-Run Testing Commands

Test the interactive CLI to ensure all field types and scoring combinations work:

### Boolean + Manual (Quick Path)
```bash
echo -e "simple\nboolean\nmanual\nDid you exercise today?\n" | iter goal add --dry-run
```

### Boolean + Automatic
```bash
echo -e "simple\nboolean\nautomatic\nDid you exercise today?\n" | iter goal add --dry-run
```

### Numeric + Automatic with Range
```bash
echo -e "simple\nnumeric\nunsigned_decimal\nhours\nyes\n7\n9\nautomatic\nrange\n7.0\n9.0\nyes\nHow many hours did you sleep?\n" | iter goal add --dry-run
```

### Time + Automatic
```bash
echo -e "simple\ntime\nautomatic\nbefore\n07:00\nWhat time did you wake up?\n" | iter goal add --dry-run
```

### Duration + Automatic
```bash
echo -e "simple\nduration\nautomatic\ngreater_than_or_equal\n20m\nHow long did you meditate?\n" | iter goal add --dry-run
```

## Expected Results

All commands should:
1. Complete without errors
2. Display "✅ Goal created successfully" message  
3. Show valid YAML output
4. Pass goal validation

## Notes

- TTY limitation prevents piped input testing in CI
- These commands are for manual verification only
- All business logic is already covered by unit + integration tests