# Manual Testing Guide for T009/1.4

## Testing Strategy

**Automated Testing**: All business logic is comprehensively tested via headless unit and integration tests (42 tests covering all combinations).

**Interactive UI Testing**: Manual verification of the actual CLI interface to ensure proper user experience.

## Interactive CLI Testing

The CLI UI framework requires interactive input. Test these scenarios manually:

### 1. Boolean + Manual (Quick Path)
```bash
iter goal add --dry-run
```
**User inputs**: simple → boolean → manual → "Did you exercise today?"

### 2. Boolean + Automatic  
```bash
iter goal add --dry-run
```
**User inputs**: simple → boolean → automatic → "Did you exercise today?"

### 3. Numeric + Manual with Constraints
```bash
iter goal add --dry-run
```
**User inputs**: simple → numeric → unsigned_int → "reps" → yes → "10" → "100" → manual → "How many push-ups?"

### 4. Numeric + Automatic with Range
```bash
iter goal add --dry-run
```
**User inputs**: simple → numeric → unsigned_decimal → "hours" → no → automatic → range → "7.0" → "9.0" → yes → "How many hours did you sleep?"

### 5. Time + Automatic
```bash
iter goal add --dry-run
```
**User inputs**: simple → time → automatic → before → "07:00" → "What time did you wake up?"

### 6. Duration + Automatic
```bash
iter goal add --dry-run
```
**User inputs**: simple → duration → automatic → greater_than_or_equal → "20m" → "How long did you meditate?"

### 7. Text + Manual (Multiline)
```bash
iter goal add --dry-run
```
**User inputs**: simple → text → yes → manual → "What did you write about today?"

## Expected Results

All interactive sessions should:
1. Guide user through appropriate form steps
2. Skip unnecessary steps (e.g., field config for time/duration)
3. Complete without errors  
4. Display "✅ Goal created successfully" message
5. Show valid YAML output
6. Pass goal validation

## Notes

- **Piped input will NOT work** due to TTY requirement
- All business logic is already validated by automated tests
- This is purely for UX verification of the interactive interface
- Use `--dry-run` to avoid modifying actual goal files