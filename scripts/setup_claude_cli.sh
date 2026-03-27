#!/bin/bash
################################################################################
# Claude Code CLI Setup Script
#
# Adds useful aliases and helper functions for Claude Code CLI
################################################################################

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_header() {
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}\n"
}

# Main setup
print_header "Claude Code CLI Setup"

BASHRC="$HOME/.bashrc"
BACKUP="$HOME/.bashrc.backup.$(date +%Y%m%d_%H%M%S)"

# Backup existing bashrc
if [ ! -f "$BASHRC.backup" ]; then
    cp "$BASHRC" "$BACKUP"
    print_success "Backed up .bashrc to $BACKUP"
fi

# Add aliases
print_info "Adding Claude Code CLI aliases to .bashrc..."

cat >> "$BASHRC" << 'EOF'

# ═══════════════════════════════════════════════════════════════
# Claude Code CLI Aliases
# ═══════════════════════════════════════════════════════════════

# Main CLI aliases
alias cc='claude'
alias ccp='claude plugin'
alias ccc='claude config'
alias ccm='claude plugin marketplace'
alias ccl='claude plugin list'
alias cci='claude plugin install'
alias ccu='claude plugin update'
alias ccr='claude plugin remove'

# Quick commands
alias cc-update='claude plugin update --all'
alias cc-list='claude plugin list'
alias cc-markets='claude plugin marketplace list'

# Helper functions
cc-install() {
    if [ -z "$1" ]; then
        echo "Usage: cc-install <plugin-name>@<marketplace>"
        return 1
    fi
    claude plugin install "$1"
}

cc-remove() {
    if [ -z "$1" ]; then
        echo "Usage: cc-remove <plugin-name>@<marketplace>"
        return 1
    fi
    claude plugin remove "$1"
}

cc-info() {
    if [ -z "$1" ]; then
        echo "Usage: cc-info <plugin-name>"
        return 1
    fi
    claude plugin list | grep -i "$1"
}

# List all installed plugins with status
cc-all() {
    echo "╔═══════════════════════════════════════════════════════════════╗"
    echo "║           Installed Claude Code Plugins                       ║"
    echo "╚═══════════════════════════════════════════════════════════════╝"
    claude plugin list
}

# Update all plugins and show summary
cc-upgrade() {
    echo "Updating all plugins..."
    claude plugin update --all
    echo ""
    echo "Update complete! Run 'cc-all' to see current plugins."
}

EOF

print_success "Aliases added to .bashrc"

# Add completion scripts (if they exist)
if [ -f "$HOME/.nvm/versions/node/v22.12.0/lib/node_modules/@anthropic-ai/claude-code/completions/bash.sh" ]; then
    print_info "Adding bash completion..."
    echo "source $HOME/.nvm/versions/node/v22.12.0/lib/node_modules/@anthropic-ai/claude-code/completions/bash.sh" >> "$BASHRC"
    print_success "Bash completion added"
fi

# Create helper scripts directory
mkdir -p "$HOME/.claude/bin"

# Create plugin backup script
cat > "$HOME/.claude/bin/backup-plugins.sh" << 'EOF'
#!/bin/bash
BACKUP_DIR="$HOME/.claude/backups"
mkdir -p "$BACKUP_DIR"
DATE=$(date +%Y%m%d_%H%M%S)
claude plugin list > "$BACKUP_DIR/plugins-$DATE.txt"
claude config list > "$BACKUP_DIR/config-$DATE.json"
echo "Plugins backed up to $BACKUP_DIR"
EOF
chmod +x "$HOME/.claude/bin/backup-plugins.sh"

# Create plugin restore script
cat > "$HOME/.claude/bin/restore-plugins.sh" << 'EOF'
#!/bin/bash
if [ -z "$1" ]; then
    echo "Usage: restore-plugins.sh <backup-file>"
    exit 1
fi
while read plugin; do
    echo "Installing $plugin..."
    claude plugin install $plugin
done < "$1"
echo "Plugin restore complete!"
EOF
chmod +x "$HOME/.claude/bin/restore-plugins.sh"

print_success "Helper scripts created in ~/.claude/bin/"

# Add bin directory to PATH if not already there
if ! grep -q "$HOME/.claude/bin" "$BASHRC"; then
    echo "" >> "$BASHRC"
    echo "# Claude Code helper scripts" >> "$BASHRC"
    echo "export PATH=\"\$PATH:\$HOME/.claude/bin\"" >> "$BASHRC"
    print_success "Added ~/.claude/bin to PATH"
fi

# Summary
print_header "Setup Complete!"

echo ""
print_success "Claude Code CLI aliases configured"
echo ""
echo "Quick commands:"
echo "  cc              - Run claude CLI"
echo "  ccp list        - List plugins"
echo "  cci <plugin>    - Install plugin"
echo "  ccu <plugin>    - Update plugin"
echo "  cc-update       - Update all plugins"
echo "  cc-all          - Show all plugins"
echo "  cc-upgrade      - Update all plugins (with summary)"
echo ""
echo "Helper functions:"
echo "  cc-install X    - Install plugin X"
echo "  cc-remove X     - Remove plugin X"
echo "  cc-info X       - Show info about plugin X"
echo ""
echo "Backup/restore:"
echo "  backup-plugins.sh      - Backup current plugins"
echo "  restore-plugins.sh     - Restore from backup"
echo ""

print_info "To use the new aliases, run:"
echo "  source ~/.bashrc"
echo ""
print_info "Or open a new terminal window."
echo ""

# Quick test
if command -v claude &> /dev/null; then
    print_success "Claude Code CLI is ready!"
    echo ""
    echo "Current version:"
    claude --version
    echo ""
    echo "Installed plugins:"
    claude plugin list | head -5
    echo "  ..."
else
    echo "⚠️  Claude CLI not found in PATH"
    echo "   Make sure to reload your shell: source ~/.bashrc"
fi

echo ""
print_success "Setup complete! 🎉"
