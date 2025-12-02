#!/bin/bash

# Eye in the Sky - Installation Script
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

echo -e "${BLUE}🔍 Eye in the Sky - Installation Script${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go is not installed. Please install Go 1.21 or later.${NC}"
        echo "Visit: https://golang.org/dl/"
        exit 1
    fi

    GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
    echo -e "${GREEN}✅ Go ${GO_VERSION} found${NC}"
}

# Create directories
create_directories() {
    echo -e "${BLUE}📁 Creating directories...${NC}"
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$BIN_DIR"
    mkdir -p "$INSTALL_DIR/data"
    echo -e "${GREEN}✅ Directories created${NC}"
}

# Download and build
build_application() {
    echo -e "${BLUE}⚡ Building Eye in the Sky...${NC}"

    # Clone or update repository
    if [ -d "$INSTALL_DIR/.git" ]; then
        echo "Updating existing installation..."
        cd "$INSTALL_DIR"
        git pull origin main
    else
        echo "Downloading Eye in the Sky..."
        git clone https://github.com/yourusername/eye-in-the-sky.git "$INSTALL_DIR"
        cd "$INSTALL_DIR"
    fi

    # Build the applications
    echo "Building MCP server..."
    go build -o "$BIN_DIR/eye-in-the-sky" ./cmd/server/main.go

    echo "Building integrated server..."
    go build -o "$BIN_DIR/eye-in-the-sky-integrated" ./main.go

    echo -e "${GREEN}✅ Build completed${NC}"
}

# Configure Claude Desktop
configure_claude_desktop() {
    echo -e "${BLUE}🔧 Configuring Claude Desktop...${NC}"

    # Create Claude config directory if it doesn't exist
    mkdir -p "$CONFIG_DIR"

    # Check if config file exists
    if [ ! -f "$CONFIG_FILE" ]; then
        echo "Creating new Claude Desktop configuration..."
        cat > "$CONFIG_FILE" << EOF
{
  "mcpServers": {}
}
EOF
    fi

    # Add Eye in the Sky MCP server configuration
    python3 -c "
import json
import sys

config_file = '$CONFIG_FILE'
try:
    with open(config_file, 'r') as f:
        config = json.load(f)
except:
    config = {'mcpServers': {}}

if 'mcpServers' not in config:
    config['mcpServers'] = {}

config['mcpServers']['eye-in-the-sky'] = {
    'command': '$BIN_DIR/eye-in-the-sky',
    'args': [],
    'env': {
        'PATH': '/usr/local/bin:/usr/bin:/bin',
        'HOME': '$HOME'
    }
}

with open(config_file, 'w') as f:
    json.dump(config, f, indent=2)

print('✅ Claude Desktop configuration updated')
"
}

# Create start script
create_start_script() {
    echo -e "${BLUE}📝 Creating start script...${NC}"

    cat > "$BIN_DIR/eye-in-the-sky-start" << 'EOF'
#!/bin/bash

# Eye in the Sky - Start Script

INSTALL_DIR="$HOME/.eye-in-the-sky"
PORT=${1:-8080}

echo "🔍 Starting Eye in the Sky Dashboard..."
echo "📊 Dashboard will be available at: http://localhost:$PORT"
echo "📂 Database: ~/.config/eye-in-the-sky/agents.db"
echo ""
echo "Press Ctrl+C to stop"

cd "$INSTALL_DIR"
exec "$HOME/.local/bin/eye-in-the-sky-integrated" -port "$PORT"
EOF

    chmod +x "$BIN_DIR/eye-in-the-sky-start"
    echo -e "${GREEN}✅ Start script created at $BIN_DIR/eye-in-the-sky-start${NC}"
}

# Add to PATH
update_path() {
    echo -e "${BLUE}🛤️  Updating PATH...${NC}"

    # Detect shell
    SHELL_RC=""
    if [ -n "$ZSH_VERSION" ]; then
        SHELL_RC="$HOME/.zshrc"
    elif [ -n "$BASH_VERSION" ]; then
        SHELL_RC="$HOME/.bashrc"
    else
        SHELL_RC="$HOME/.profile"
    fi

    # Add to PATH if not already there
    if ! echo "$PATH" | grep -q "$BIN_DIR"; then
        echo "" >> "$SHELL_RC"
        echo "# Eye in the Sky" >> "$SHELL_RC"
        echo "export PATH=\"$BIN_DIR:\$PATH\"" >> "$SHELL_RC"
        echo -e "${GREEN}✅ Added $BIN_DIR to PATH in $SHELL_RC${NC}"
        echo -e "${YELLOW}⚠️  Please restart your terminal or run: source $SHELL_RC${NC}"
    else
        echo -e "${GREEN}✅ $BIN_DIR already in PATH${NC}"
    fi
}

# Create desktop shortcut (macOS)
create_desktop_shortcut() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo -e "${BLUE}🖥️  Creating desktop shortcut...${NC}"

        cat > "$HOME/Desktop/Eye in the Sky.command" << EOF
#!/bin/bash
exec "$BIN_DIR/eye-in-the-sky-start"
EOF

        chmod +x "$HOME/Desktop/Eye in the Sky.command"
        echo -e "${GREEN}✅ Desktop shortcut created${NC}"
    fi
}

# Test installation
test_installation() {
    echo -e "${BLUE}🧪 Testing installation...${NC}"

    # Test MCP server
    if [ -x "$BIN_DIR/eye-in-the-sky" ]; then
        echo -e "${GREEN}✅ MCP server binary found${NC}"
    else
        echo -e "${RED}❌ MCP server binary not found${NC}"
        exit 1
    fi

    # Test integrated server
    if [ -x "$BIN_DIR/eye-in-the-sky-integrated" ]; then
        echo -e "${GREEN}✅ Integrated server binary found${NC}"
    else
        echo -e "${RED}❌ Integrated server binary not found${NC}"
        exit 1
    fi

    # Test database directory
    if [ -d "$INSTALL_DIR/data" ]; then
        echo -e "${GREEN}✅ Data directory created${NC}"
    else
        echo -e "${RED}❌ Data directory not found${NC}"
        exit 1
    fi

    echo -e "${GREEN}✅ Installation test passed${NC}"
}

# Main installation
main() {
    echo "This script will install Eye in the Sky to: $INSTALL_DIR"
    echo "Binaries will be installed to: $BIN_DIR"
    echo ""
    read -p "Continue with installation? (y/N): " -n 1 -r
    echo

    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Installation cancelled."
        exit 0
    fi

    echo ""
    check_go
    create_directories
    build_application
    configure_claude_desktop
    create_start_script
    update_path
    create_desktop_shortcut
    test_installation

    echo ""
    echo -e "${GREEN}🎉 Installation completed successfully!${NC}"
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo "1. Restart your terminal (or run: source ~/.zshrc)"
    echo "2. Restart Claude Desktop to load MCP configuration"
    echo "3. Start the dashboard: eye-in-the-sky-start"
    echo "4. Visit: http://localhost:8080"
    echo ""
    echo -e "${BLUE}For Claude Code CLI:${NC}"
    echo "- The 'eye-in-the-sky' MCP server is now available"
    echo "- Use 'claude mcp list' to verify"
    echo ""
    echo -e "${BLUE}For Claude Desktop:${NC}"
    echo "- Create a new project and use Eye in the Sky tools"
    echo "- Example: 'Register as Claude Code agent working on my project'"
    echo ""
    echo -e "${YELLOW}Need help? Check the documentation at:${NC}"
    echo "https://github.com/yourusername/eye-in-the-sky"
}

# Run main function
main "$@"