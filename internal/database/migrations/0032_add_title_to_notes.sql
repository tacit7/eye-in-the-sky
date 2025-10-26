-- Migration: Add title column to notes table
-- Notes will now have a title extracted from first two words of content

-- Add title column (allow NULL initially for migration)
ALTER TABLE notes ADD COLUMN title TEXT;

-- Generate titles for existing notes from first two words of content
-- This handles the migration of existing data
UPDATE notes
SET title = CASE
    WHEN LENGTH(content) > 0 THEN
        SUBSTR(
            content,
            1,
            CASE
                WHEN INSTR(SUBSTR(content, INSTR(content, ' ') + 1), ' ') > 0
                THEN INSTR(content, ' ') + INSTR(SUBSTR(content, INSTR(content, ' ') + 1), ' ')
                ELSE LENGTH(content)
            END
        )
    ELSE 'Untitled'
END
WHERE title IS NULL;

-- Create index on title for faster searching
CREATE INDEX IF NOT EXISTS idx_notes_title ON notes(title);
