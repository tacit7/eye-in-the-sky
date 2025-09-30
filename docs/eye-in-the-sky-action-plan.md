# Eye in the Sky — Action Plan

## 🔧 Backend (Go + SQLite)

### 1. Project Structure
```
/cmd/server/main.go
/internal/db/          # SQLite helpers
/internal/handlers/    # HTTP route handlers
/templates/            # HTML templates
/static/               # Tailwind CSS, JS, assets
/data/agents.db        # SQLite database
```

### 2. Database Schema
- **agents**: id, name, status, task, source, updated_at  
- **sessions**: id, agent_id, started_at, ended_at  
- **logs**: id, session_id, timestamp, type, message  
- **notes**: id, session_id, created_at, content  
- **context**: id, session_id, key, value  

### 3. Handlers
- `GET /` → dashboard template (agents + recent logs feed)  
- `GET /agent/{id}` → agent detail template (logs, notes, context)  
- `GET /static/*` → serve CSS/JS  
- (Optional) `GET /ws/events` → WebSocket stream for real-time logs  

### 4. Templates
- Use Go `html/template`.  
- Data passed as structs:  
  - `DashboardData{Agents, RecentLogs}`  
  - `AgentData{Agent, Session, Logs, Notes, Context}`  

### 5. Real-Time (Phase 2)
- Implement WebSocket hub.  
- Push log updates to connected clients.  

---

## 🎨 Frontend (Go Templates + Tailwind)

### 1. Tailwind Setup
- Install Tailwind via Node: `npm install tailwindcss`  
- Configure `tailwind.config.js` with **dark theme tokens**:  
  - Background: `#1B1F28`  
  - Surface: `#242B38`  
  - Text primary: `#E5E9F0`  
  - Text secondary: `#A0A9B8`  
  - Status colors: Active `#3DDC97`, Idle `#E0B95B`, Failed `#F96363`, Completed `#9CA3AF`  
  - Accent: `#83AAFF`  

- Compile Tailwind into `/static/main.css`  

### 2. Components (as templates)
- **AgentCard** → name, status badge, task, source, last updated  
- **StatusBadge** → pill with color-coded state  
- **ActivityFeedItem** → timestamp, icon, message  
- **Tabs** → Logs | Notes | Context  

### 3. Pages
- **Dashboard (`/`)**:  
  - Grid of AgentCards  
  - ActivityFeed (recent logs)  
- **Agent Detail (`/agent/{id}`)**:  
  - Header: agent name, status, task  
  - Tabs: Logs (list), Notes (read-only), Context (table)  

### 4. Interactivity
- Phase 1: static server-rendered templates (manual refresh)  
- Phase 2: add WebSocket JS client for live feed updates  

---

## 🚀 Development Workflow
1. **Phase 1** — Setup SQLite schema + Go handlers with seed data  
2. **Phase 2** — Build Tailwind dark theme, implement templates  
3. **Phase 3** — Connect templates to DB queries for dynamic rendering  
4. **Phase 4** — Add WebSocket for live log updates (optional)  
5. **Phase 5** — Polish: responsive layout, small animations, icons  

---

## ✅ Deliverables
- Single Go binary serving dashboard + agent detail  
- SQLite database as storage  
- Tailwind-styled dark theme frontend  
- Activity feed + agent drill-down views  
- (Optional) WebSocket real-time updates  
