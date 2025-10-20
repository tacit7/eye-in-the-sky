#!/usr/bin/env python3
"""Update Claude Desktop config to remove -db argument from eye-in-the-sky MCP server"""

import json
import sys
from pathlib import Path

def update_config(config_path: Path):
    """Update Claude Desktop config file"""

    # Read current config
    with open(config_path, 'r') as f:
        config = json.load(f)

    # Track changes
    changes_made = 0

    # Update all project configs that have eye-in-the-sky MCP server
    if 'projects' not in config:
        print("No 'projects' key found in config")
        return

    for project_path, project_config in config['projects'].items():
        if isinstance(project_config, dict) and 'mcpServers' in project_config:
            if 'eye-in-the-sky' in project_config['mcpServers']:
                old_config = project_config['mcpServers']['eye-in-the-sky']

                # Remove args if present
                if 'args' in old_config and old_config['args']:
                    print(f"Updating {project_path}")
                    print(f"  Old args: {old_config['args']}")
                    old_config['args'] = []
                    print(f"  New args: []")
                    changes_made += 1

    if changes_made == 0:
        print("No changes needed - config already up to date")
        return

    # Backup original
    backup_path = config_path.with_suffix('.json.backup')
    with open(backup_path, 'w') as f:
        json.dump(config, f, indent=2)
    print(f"\n✅ Backup created: {backup_path}")

    # Write updated config
    with open(config_path, 'w') as f:
        json.dump(config, f, indent=2)

    print(f"✅ Updated {changes_made} project config(s)")
    print(f"✅ Configuration saved to {config_path}")
    print("\n⚠️  Please restart Claude Desktop for changes to take effect")

if __name__ == '__main__':
    config_path = Path.home() / '.claude.json'

    if not config_path.exists():
        print(f"Error: Config file not found at {config_path}")
        sys.exit(1)

    update_config(config_path)
