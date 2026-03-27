# Claude Code CLI - Cheat Sheet

**Quick reference for common commands. Print this for your desk!**

## 🔧 Essential Commands

| Command | Description |
|---------|-------------|
| `claude --version` | Check CLI version |
| `claude plugin list` | List all plugins |
| `claude plugin install X@Y` | Install plugin |
| `claude plugin update --all` | Update all plugins |
| `claude plugin remove X@Y` | Remove plugin |

## 📝 Quick Aliases (after running setup script)

| Alias | Full Command |
|-------|-------------|
| `cc` | `claude` |
| `ccp` | `claude plugin` |
| `ccc` | `claude config` |
| `ccp list` | `claude plugin list` |
| `cci X@Y` | `claude plugin install X@Y` |
| `ccu X@Y` | `claude plugin update X@Y` |
| `cc-update` | `claude plugin update --all` |
| `cc-all` | Show all plugins (formatted) |

## 🔌 Common Plugins

### Everything Plugin
```bash
claude plugin install everything-claude-code@everything-claude-code
```

### CodeRabbit
```bash
claude plugin install coderabbit@claude-plugins-official
```

### Firecrawl
```bash
claude plugin install firecrawl@claude-plugins-official
```

## 🛠️ Marketplace Commands

```bash
# List marketplaces
claude plugin marketplace list

# Add marketplace
claude plugin marketplace add user/repo

# Example:
claude plugin marketplace add affaan-m/everything-claude-code
```

## 📂 Project Commands

```bash
# Set project
claude project set /path/to/project

# Get current project
claude project get
```

## 🔍 Troubleshooting

### Command not found?
```bash
export PATH=$PATH:/home/faris/.nvm/versions/node/v22.12.0/bin
# or
source ~/.bashrc
```

### Plugin not working?
```bash
claude plugin remove X@Y
claude plugin install X@Y
```

### Update everything?
```bash
claude plugin update --all
```

## 🚀 Setup

Run the setup script:
```bash
cd /home/faris/Documents/lamees/radio
./scripts/setup_claude_cli.sh
```

Then reload:
```bash
source ~/.bashrc
```

## 📁 File Locations

| File | Location |
|------|----------|
| CLI binary | `~/.nvm/versions/node/v22.12.0/bin/claude` |
| Global config | `~/.claude/settings.json` |
| Project config | `project/.claude/settings.json` |
| Logs | `~/.claude/logs/` |
| Helper scripts | `~/.claude/bin/` |

## 🎯 Quick Workflows

### Install New Plugin
```bash
claude plugin install plugin-name@marketplace
claude plugin list | grep plugin-name
```

### Update All Plugins
```bash
claude plugin update --all
claude plugin list
```

### Backup Configuration
```bash
~/.claude/bin/backup-plugins.sh
```

### Restore Plugins
```bash
~/.claude/bin/restore-plugins.sh ~/.claude/backups/plugins-20260325.txt
```

## 💡 Pro Tips

1. **Use aliases** - Setup script adds them automatically
2. **Update weekly** - `cc-update` keeps plugins fresh
3. **Check scope** - `user` plugins are global, `project` are local
4. **Backup first** - Use `backup-plugins.sh` before major changes
5. **Test plugins** - Try new plugins in safe environment first

---

**Need more?** See [CLAUDE_CLI_GUIDE.md](./CLAUDE_CLI_GUIDE.md) for complete documentation.
