-- Add terminal_application column to agents table
-- This stores which terminal application the Claude Code session is running in
-- Examples: "iTerm2", "Terminal", "Warp", "Kitty", "Alacritty"

ALTER TABLE agents ADD COLUMN terminal_application TEXT;
