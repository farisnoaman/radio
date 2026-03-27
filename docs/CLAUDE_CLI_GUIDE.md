# Claude Code CLI Quick Reference Guide

Complete guide to using the Claude Code CLI (`claude` command).

## Installation Status

✅ **Claude Code CLI**: Version 2.1.83
✅ **Installed at**: `~/.nvm/versions/node/v22.12.0/bin/claude`
✅ **Everything Plugin**: Version 1.9.0 (enabled)

## Basic Commands

### Version & Help

```bash
# Check version
claude --version

# Show help
claude --help

# Show help for a specific command
claude plugin --help
claude project --help
```

## Plugin Management

### List Plugins

```bash
# List all installed plugins
claude plugin list

# List all marketplaces
claude plugin marketplace list

# Search for plugins (if search is available)
claude plugin search <keyword>
```

### Install Plugins

```bash
# Install from official marketplace
claude plugin install <plugin-name>@<marketplace>

# Examples:
claude plugin install coderabbit@claude-plugins-official
claude plugin install everything-claude-code@everything-claude-code

# Install multiple plugins
claude plugin install plugin1 plugin2 plugin3
```

### Manage Marketplaces

```bash
# Add a new marketplace
claude plugin marketplace add <github-user/repo>

# Example (already done):
claude plugin marketplace add affaan-m/everything-claude-code

# List configured marketplaces
claude plugin marketplace list

# Remove a marketplace
claude plugin marketplace remove <marketplace-name>
```

### Update & Remove Plugins

```bash
# Update a specific plugin
claude plugin update <plugin-name>@<marketplace>

# Update all plugins
claude plugin update --all

# Remove/uninstall a plugin
claude plugin remove <plugin-name>@<marketplace>

# Force reinstall
claude plugin install --force <plugin-name>@<marketplace>
```

### Plugin Information

```bash
# Show plugin details (if command exists)
claude plugin show <plugin-name>@<marketplace>

# Check plugin status
claude plugin list | grep <plugin-name>
```

## Project Management

### Set Project Context

```bash
# Set current project (affects plugin scope)
claude project set /path/to/project

# Show current project
claude project get

# List recent projects
claude project list
```

**★ Insight ─────────────────────────────────────**
- **Project Scope**: Plugins can be scoped to `user` (global) or `project` (current directory only)
- **User plugins**: Available in all conversations and projects
- **Project plugins**: Only available when working in that specific project directory
`─────────────────────────────────────────────────`

## Configuration

### Settings Files

```bash
# Global settings
~/.claude/settings.json

# Project-specific settings
/path/to/project/.claude/settings.json

# Local overrides (gitignored)
~/.claude/settings.local.json
```

### View Configuration

```bash
# Show current configuration
claude config list

# Get specific config value
claude config get <key>

# Set config value
claude config set <key> <value>
```

### Common Config Options

```bash
# Set default model
claude config set model claude-sonnet-4-6

# Set output style
claude config set outputStyle explanatory

# Enable/disable features
claude config set featureFlag true
```

## Session Management

### Start New Session

```bash
# Start with specific project
claude --project /path/to/project

# Start with specific model
claude --model claude-opus-4-6

# Start with custom settings
claude --setting key=value
```

### Session History

```bash
# Show recent sessions
claude session list

# Resume session
claude session resume <session-id>

# Export session
claude session export <session-id> > session.json
```

## Aliases (Recommended)

Add these to your `~/.bashrc` for convenience:

```bash
# Claude Code CLI aliases
alias cc='claude'
alias ccp='claude plugin'
alias ccc='claude config'
alias ccm='claude plugin marketplace'
alias ccl='claude plugin list'
alias cci='claude plugin install'
alias ccu='claude plugin update'
alias ccr='claude plugin remove'
```

Then use them:

```bash
ccp list              # List plugins
ccp install <plugin>  # Install plugin
ccp update --all      # Update all plugins
ccc list              # List config
```

## Installed Plugins Summary

### Official Plugins (claude-plugins-official)

| Plugin | Purpose | Scope |
|--------|---------|-------|
| `agent-sdk-dev` | Agent SDK development | project |
| `asana` | Asana integration | project |
| `chrome-devtools-mcp` | Chrome DevTools via MCP | user |
| `claude-code-setup` | Initial setup wizard | user |
| `claude-md-management` | CLAUDE.md management | user |
| `code-review` | Code review tools | project |
| `coderabbit` | CodeRabbit integration | user |
| `context7` | Context documentation | project |
| `explanatory-output-style` | Educational explanations | project |
| `feature-dev` | Feature development | project |
| `firecrawl` | Web scraping | user |
| `frontend-design` | Frontend design | project |
| `github` | GitHub integration | project |
| `greptile` | Code analysis | project |
| `hookify` | Behavior prevention hooks | project |
| `learning-output-style` | Interactive learning | project |
| `notion` | Notion integration | project |
| `playwright` | Browser automation | user |
| `plugin-dev` | Plugin development | user |
| `pr-review-toolkit` | PR review tools | user |
| `security-guidance` | Security best practices | project |
| `sentry` | Error tracking | project |
| `serena` | Serena integration | project |
| `supabase` | Supabase integration | project |

### Superpowers Marketplace

| Plugin | Purpose | Version | Scope |
|--------|---------|---------|-------|
| `superpowers` | Core superpowers skills | 5.0.5 | user |
| `superpowers-chrome` | Chrome browser tools | 1.6.1 | project |
| `superpowers-lab` | Experimental features | 0.1.0 | project |
| `claude-session-driver` | Session management | 1.0.1 | user |
| `double-shot-latte` | Productivity tools | 1.1.5 | user |
| `elements-of-style` | Writing improvements | 1.0.0 | project |
| `episodic-memory` | Conversation memory | 1.0.15 | project |

### Everything Plugin

| Plugin | Purpose | Version | Scope |
|--------|---------|---------|-------|
| `everything-claude-code` | Community enhancements | 1.9.0 | user |

## Troubleshooting

### Command Not Found

```bash
# If you get "claude: command not found"
export PATH=$PATH:/home/faris/.nvm/versions/node/v22.12.0/bin

# Or reload your shell
source ~/.bashrc
```

### Plugin Not Working

```bash
# Check if plugin is enabled
claude plugin list | grep <plugin-name>

# Reinstall the plugin
claude plugin remove <plugin-name>@<marketplace>
claude plugin install <plugin-name>@<marketplace>
```

### Update Issues

```bash
# Clear cache and update
claude plugin update --all --force

# Check for errors
claude plugin list 2>&1 | grep -i error
```

### PATH Issues

Add to `~/.bashrc`:

```bash
# NVM setup (already there, just verify)
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

# Claude CLI alias
alias claude="$HOME/.nvm/versions/node/v22.12.0/bin/claude"
```

## Advanced Usage

### Batch Operations

```bash
# Install multiple plugins at once
claude plugin install \
  everything-claude-code@everything-claude-code \
  coderabbit@claude-plugins-official \
  firecrawl@claude-plugins-official

# Update all marketplaces
for m in $(claude plugin marketplace list | awk '{print $1}'); do
  claude plugin marketplace update $m
done
```

### Scripting

```bash
#!/bin/bash
# Backup plugin configuration
claude plugin list > plugins-backup.txt
claude config list > config-backup.json

# Restore from backup
while read plugin; do
  claude plugin install $plugin
done < plugins-backup.txt
```

### Integration with Git

```bash
# Add plugin install to project setup
echo "post-install: claude plugin install everything-claude-code@everything-claude-code" >> package.json

# Or create a setup script
cat > setup.sh << 'EOF'
#!/bin/bash
claude plugin install everything-claude-code@everything-claude-code
claude plugin install coderabbit@claude-plugins-official
echo "Plugins installed successfully!"
EOF
chmod +x setup.sh
```

## Quick Reference Card

```
╔═══════════════════════════════════════════════════════════════╗
║                  Claude Code CLI - Quick Reference            ║
╠═══════════════════════════════════════════════════════════════╣
║  claude --version          Check CLI version                  ║
║  claude plugin list        List all plugins                   ║
║  claude plugin install X   Install plugin X                   ║
║  claude plugin update X    Update plugin X                    ║
║  claude plugin remove X    Remove plugin X                    ║
║  claude marketplace add X  Add marketplace X                  ║
║  claude config list        Show configuration                 ║
║  claude project set PATH   Set project directory              ║
╚═══════════════════════════════════════════════════════════════╝
```

## Resources

- **Official Docs**: https://docs.anthropic.com/claude-code
- **GitHub**: https://github.com/anthropics/claude-plugins-official
- **Superpowers**: https://github.com/obra/superpowers-marketplace
- **Everything Plugin**: https://github.com/affaan-m/everything-claude-code

## Tips & Best Practices

1. **Update Regularly**: Run `claude plugin update --all` weekly
2. **Use Aliases**: Set up bash aliases for frequently used commands
3. **Scope Wisely**: Use `user` scope for plugins you use everywhere, `project` for specific tools
4. **Test Plugins**: Test new plugins in a safe environment before using in production
5. **Backup Config**: Keep backups of your plugin configuration
6. **Check Compatibility**: Ensure plugins are compatible with your CLI version

## Getting Help

```bash
# General help
claude --help

# Plugin help
claude plugin --help

# Config help
claude config --help

# Check logs for issues
~/.claude/logs/
```

---

**Last Updated**: 2026-03-25
**CLI Version**: 2.1.83
**Everything Plugin**: 1.9.0
