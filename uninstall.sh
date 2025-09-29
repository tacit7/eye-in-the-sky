#!/bin/bash

# Eye in the Sky - Uninstall Script
# Claude Code Multi-Agent Management System

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="$HOME/.eye-in-the-sky"
BIN_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/Library/Application Support/Claude"
CONFIG_FILE="$CONFIG_DIR/claude_desktop_config.json"

echo -e "${BLUE}🗑️  Eye in the Sky - Uninstall Script${NC}"
echo -e "${BLUE}=====================================${NC}"
echo ""

# Confirm uninstall
confirm_uninstall() {
    echo "This will remove Eye in the Sky from your system:"
    echo "  - Binaries: $BIN_DIR/eye-in-the-sky*"
    echo "  - Installation: $INSTALL_DIR"
    echo "  - Database: $INSTALL_DIR/data/agents.db"
    echo "  - Claude Desktop MCP configuration"
    echo ""
    read -p "Are you sure you want to uninstall? (y/N): " -n 1 -r
    echo

    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Uninstall cancelled."
        exit 0
    fi
}

# Stop running processes
stop_processes() {
    echo -e "${BLUE}🛑 Stopping Eye in the Sky processes...${NC}"

    # Kill any running Eye in the Sky processes
    if pgrep -f "eye-in-the-sky" > /dev/null; then
        echo "Stopping running Eye in the Sky processes..."
        pkill -f "eye-in-the-sky" || true
        sleep 2
        echo -e "${GREEN}✅ Processes stopped${NC}"
    else
        echo -e "${GREEN}✅ No running processes found${NC}"
    fi
}

# Remove binaries
remove_binaries() {
    echo -e "${BLUE}🗂️  Removing binaries...${NC}"

    removed_count=0

    for binary in "eye-in-the-sky" "eye-in-the-sky-integrated" "eye-in-the-sky-start"; do
        if [ -f "$BIN_DIR/$binary" ]; then
            rm "$BIN_DIR/$binary"
            echo "Removed: $BIN_DIR/$binary"
            ((removed_count++))
        fi
    done

    if [ $removed_count -gt 0 ]; then
        echo -e "${GREEN}✅ Removed $removed_count binaries${NC}"
    else
        echo -e "${YELLOW}⚠️  No binaries found to remove${NC}"
    fi
}

# Remove installation directory
remove_installation() {
    echo -e "${BLUE}📁 Removing installation directory...${NC}"

    if [ -d "$INSTALL_DIR" ]; then
        # Show what will be removed
        if [ -f "$INSTALL_DIR/data/agents.db" ]; then
            agent_count=$(sqlite3 "$INSTALL_DIR/data/agents.db" "SELECT COUNT(*) FROM agents;" 2>/dev/null || echo "0")
            echo "Database contains $agent_count agents"
        fi

        rm -rf "$INSTALL_DIR"
        echo -e "${GREEN}✅ Installation directory removed${NC}"
    else
        echo -e "${YELLOW}⚠️  Installation directory not found${NC}"
    fi
}

# Remove Claude Desktop configuration
remove_claude_config() {
    echo -e "${BLUE}🔧 Removing Claude Desktop configuration...${NC}"

    if [ -f "$CONFIG_FILE" ]; then
        # Create backup
        cp "$CONFIG_FILE" "$CONFIG_FILE.backup.$(date +%s)"
        echo "Created backup: $CONFIG_FILE.backup.$(date +%s)"

        # Remove Eye in the Sky MCP server configuration
        python3 -c "
import json
import sys

config_file = '$CONFIG_FILE'
try:
    with open(config_file, 'r') as f:
        config = json.load(f)
except:
    print('Error reading Claude config file')
    sys.exit(1)

if 'mcpServers' in config and 'eye-in-the-sky' in config['mcpServers']:
    del config['mcpServers']['eye-in-the-sky']

    with open(config_file, 'w') as f:
        json.dump(config, f, indent=2)

    print('✅ Removed Eye in the Sky from Claude Desktop configuration')
else:
    print('⚠️  Eye in the Sky not found in Claude Desktop configuration')
" 2>/dev/null || echo -e "${YELLOW}⚠️  Could not update Claude Desktop configuration (manual removal needed)${NC}"
    else
        echo -e "${YELLOW}⚠️  Claude Desktop configuration file not found${NC}"
    fi
}

# Remove from PATH
remove_from_path() {
    echo -e "${BLUE}🛤️  Removing from PATH...${NC}"

    # Detect shell
    SHELL_RC=""
    if [ -n "$ZSH_VERSION" ]; then
        SHELL_RC="$HOME/.zshrc"
    elif [ -n "$BASH_VERSION" ]; then
        SHELL_RC="$HOME/.bashrc"
    else
        SHELL_RC="$HOME/.profile"
    fi

    if [ -f "$SHELL_RC" ]; then
        # Remove Eye in the Sky PATH entries
        if grep -q "# Eye in the Sky" "$SHELL_RC"; then
            # Create backup
            cp "$SHELL_RC" "$SHELL_RC.backup.$(date +%s)"

            # Remove the lines
            sed -i.tmp '/# Eye in the Sky/,+1d' "$SHELL_RC" && rm "$SHELL_RC.tmp"
            echo -e "${GREEN}✅ Removed from PATH in $SHELL_RC${NC}"
            echo -e "${YELLOW}⚠️  Please restart your terminal or run: source $SHELL_RC${NC}"
        else
            echo -e "${YELLOW}⚠️  Eye in the Sky PATH not found in $SHELL_RC${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️  Shell configuration file not found${NC}"
    fi
}

# Remove desktop shortcuts
remove_desktop_shortcuts() {
    echo -e "${BLUE}🖥️  Removing desktop shortcuts...${NC}"

    if [[ "$OSTYPE" == "darwin"* ]]; then
        desktop_shortcut="$HOME/Desktop/Eye in the Sky.command"
        if [ -f "$desktop_shortcut" ]; then
            rm "$desktop_shortcut"
            echo -e "${GREEN}✅ Desktop shortcut removed${NC}"
        else
            echo -e "${YELLOW}⚠️  Desktop shortcut not found${NC}"
        fi
    fi
}

# Verify uninstall
verify_uninstall() {
    echo -e "${BLUE}🧪 Verifying uninstall...${NC}"

    failed=0

    # Check binaries
    for binary in "eye-in-the-sky" "eye-in-the-sky-integrated" "eye-in-the-sky-start"; do
        if command -v "$binary" &> /dev/null; then
            echo -e "${RED}❌ Binary still found: $binary${NC}"
            ((failed++))
        fi
    done

    # Check installation directory
    if [ -d "$INSTALL_DIR" ]; then
        echo -e "${RED}❌ Installation directory still exists: $INSTALL_DIR${NC}"
        ((failed++))
    fi

    # Check processes
    if pgrep -f "eye-in-the-sky" > /dev/null; then
        echo -e "${RED}❌ Eye in the Sky processes still running${NC}"
        ((failed++))
    fi

    if [ $failed -eq 0 ]; then
        echo -e "${GREEN}✅ Uninstall verification passed${NC}"
    else
        echo -e "${RED}❌ Uninstall verification failed ($failed issues)${NC}"
        exit 1
    fi
}

# Show post-uninstall instructions
show_post_uninstall() {
    echo ""
    echo -e "${GREEN}🎉 Eye in the Sky uninstalled successfully!${NC}"
    echo ""
    echo -e "${BLUE}Post-uninstall notes:${NC}"
    echo "1. Restart your terminal to update PATH"
    echo "2. Restart Claude Desktop to apply MCP configuration changes"
    echo "3. Configuration backups created with .backup extension"
    echo ""
    echo -e "${BLUE}Manual cleanup (if needed):${NC}"
    echo "- Check Claude Desktop config: $CONFIG_FILE"
    echo "- Check shell config: ~/.zshrc or ~/.bashrc"
    echo ""
    echo -e "${YELLOW}To reinstall Eye in the Sky:${NC}"
    echo "curl -fsSL https://raw.githubusercontent.com/yourusername/eye-in-the-sky/main/install.sh | bash"
}

# Main uninstall function
main() {
    confirm_uninstall
    echo ""
    stop_processes
    remove_binaries
    remove_installation
    remove_claude_config
    remove_from_path
    remove_desktop_shortcuts
    verify_uninstall
    show_post_uninstall
}

# Run main function
main "$@"