# Claude Code Scripts

Helper scripts for Claude Code CLI.

## Available Scripts

### setup_claude_cli.sh
Sets up aliases, helper functions, and bash completion for Claude Code CLI.

**Location:** `scripts/setup_claude_cli.sh`

**Usage:**
```bash
./scripts/setup_claude_cli.sh
```

**What it does:**
- Adds convenient aliases (cc, ccp, cci, etc.)
- Creates helper functions (cc-install, cc-remove, etc.)
- Sets up plugin backup/restore scripts
- Adds bash completion
- Updates PATH

**After running:**
```bash
source ~/.bashrc  # Reload shell
```

## Helper Functions (after setup)

### cc-install
Install a plugin quickly.
```bash
cc-install plugin-name@marketplace
```

### cc-remove
Remove a plugin.
```bash
cc-remove plugin-name@marketplace
```

### cc-info
Show plugin information.
```bash
cc-info plugin-name
```

### cc-all
List all installed plugins with nice formatting.
```bash
cc-all
```

### cc-upgrade
Update all plugins with summary.
```bash
cc-upgrade
```

## Backup & Restore Scripts

Located in `~/.claude/bin/` after setup.

### backup-plugins.sh
Backup current plugins and config.
```bash
backup-plugins.sh
```

Creates backups in `~/.claude/backups/` with timestamps.

### restore-plugins.sh
Restore plugins from backup.
```bash
restore-plugins.sh ~/.claude/backups/plugins-20260325.txt
```

## Quick Reference

| Task | Command |
|------|---------|
| List plugins | `ccp list` |
| Install plugin | `cci plugin@market` |
| Update all | `cc-update` |
| Remove plugin | `ccr plugin@market` |
| Show all | `cc-all` |
| Backup | `backup-plugins.sh` |

See [docs/CLAUDE_CLI_GUIDE.md](../../docs/CLAUDE_CLI_GUIDE.md) for complete documentation.
