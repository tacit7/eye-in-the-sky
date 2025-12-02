# Nerd Font Toggle

The Eye in the Sky TUI supports both **Nerd Fonts** and **plain ASCII** rendering for maximum terminal compatibility.

## Font Modes

### Nerd Font Mode (Default)

Uses Unicode glyphs and box-drawing characters for a polished look:

**Status Icons:**
- `●` Active
- `◉` Working
- `○` Idle
- `✓` Completed
- `✗` Failed

**Borders:**
- `─` Horizontal lines
- `│` Vertical pipes
- `═` Double lines (when applicable)

**Tree Symbols:**
- `│` Subagent indicator
- `✓` Bookmark marker

### Plain ASCII Mode

Uses standard ASCII characters for compatibility with terminals that don't support Nerd Fonts:

**Status Icons:**
- `*` Active
- `@` Working
- `o` Idle
- `+` Completed
- `x` Failed

**Borders:**
- `-` Horizontal lines
- `|` Vertical pipes

**Tree Symbols:**
- `|` Subagent indicator
- `*` Bookmark marker

## Configuration

### Environment Variable (Recommended)

Disable Nerd Fonts by setting the `NERD_FONTS` environment variable to `0`:

```bash
# Run with plain ASCII mode
NERD_FONTS=0 ./bin/eye-ui

# Or export for the session
export NERD_FONTS=0
./bin/eye-ui
```

By default (when `NERD_FONTS` is not set or set to any value other than `0`), Nerd Fonts mode is enabled.

### Future Configuration Options

The following configuration methods are planned but not yet implemented:

- Config file: `~/.config/eye-in-the-sky/config.yaml`
- CLI flag: `--nerdfonts` / `--plain`
- Runtime toggle: Press `F` key to switch modes

## Examples

### With Nerd Fonts (Default)

```
  Status      Session    Agent ID   Description
  ─────────────────────────────────────────────────────────
  ● active    84af194b   372f5a60   Implementing Phase 6
  │ ◉ working 572c2417   5f8088fe   Auto-refresh logs
  ○ idle      31a1155d   a1b2c3d4   Waiting for input
```

### Plain ASCII Mode

```
  Status      Session    Agent ID   Description
  ---------------------------------------------
  * active    84af194b   372f5a60   Implementing Phase 6
  | @ working 572c2417   5f8088fe   Auto-refresh logs
  o idle      31a1155d   a1b2c3d4   Waiting for input
```

## Terminal Compatibility

**Nerd Fonts work best with:**
- iTerm2
- Kitty
- Alacritty
- WezTerm
- Windows Terminal (with Nerd Font installed)

**Use Plain ASCII mode for:**
- Basic terminal emulators
- SSH sessions to systems without font support
- Screen/tmux in some configurations
- Terminals with missing glyph support

## Implementation Details

The font mode is detected at startup by checking the `NERD_FONTS` environment variable. The choice affects:

1. **Agent status icons** - All status indicators in the agent list
2. **Table borders** - Box-drawing characters in all tables
3. **Tree hierarchy symbols** - Parent/child agent indicators

The configuration is global and applies to all UI components throughout the application.
