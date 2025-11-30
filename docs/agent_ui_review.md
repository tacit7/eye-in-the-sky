# Agent Detail UI Critique & Improvement Plan

This document reviews the current Agent Detail UI (DaisyUI + Tailwind) and proposes concrete improvements. The goal is to move from a generic “CRUD admin” feel to a focused “agent mission control” interface.

## 1. Current Strengths

- Typography is clean and legible.
- Session metadata is present and grouped.
- Tabs for different resources are visible and aligned.

## 2. Main UX / UI Problems

### 2.1 Weak visual hierarchy

- Everything sits on the same background with similar elevation.
- The commit row resembles an input field.
- Elements feel disconnected.

### 2.2 Header layout feels disconnected

- Layout lacks structure.
- Action buttons float without anchor.

### 2.3 Tabs are visually weak

- Poor active-state contrast.
- Count badges blend into text.

### 2.4 Commit list row is underdesigned

- No metadata or actions, just a sentence.
- Lacks semantic cues.

### 2.5 Spacing & empty state issues

- Excessive whitespace.
- No proper empty state design.

## 3. Design Goals

1. Introduce clear structure.
2. Strengthen the header.
3. Improve tabs for scannability.
4. Present commits as structured rows.
5. Add depth with tone and shadow.

## 4. Suggested Layout & Styling Changes

### 4.1 Page frame and main card

Use a subtle background and a raised content card.

### 4.2 Header structure

Organize the header with clear left/right regions and metadata.

### 4.3 Tabs with hierarchy

Use DaisyUI tabs with stronger active state and count badges.

### 4.4 Commit list as structured rows

Replace the pill with a row that includes:
- status badge
- commit message
- relative timestamp
- actions

### 4.5 Spacing & empty states

Add empty state messaging when needed.

## 5. Styling Guidelines

- Use rounded-lg for most elements.
- Use semantic DaisyUI colors.
- Keep typography consistent.
