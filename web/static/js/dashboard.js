// Eye in the Sky Dashboard JavaScript

// Dashboard state management
let refreshInterval = null;
const REFRESH_INTERVAL = 30000; // 30 seconds

// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    initializeDashboard();
    startAutoRefresh();
    updateRelativeTimes();

    // Update relative times every 30 seconds
    setInterval(updateRelativeTimes, 30000);
});

// Initialize dashboard features
function initializeDashboard() {
    console.log('Eye in the Sky Dashboard initialized');

    // Add click handlers for agent rows
    document.querySelectorAll('.agent-row').forEach(row => {
        row.addEventListener('click', function(e) {
            // Don't navigate if clicking on buttons
            if (!e.target.closest('.btn-group')) {
                const agentId = this.querySelector('.agent-id').textContent;
                window.location.href = `/agent/${agentId}`;
            }
        });
    });

    // Initialize tooltips
    const tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
    tooltipTriggerList.map(function (tooltipTriggerEl) {
        return new bootstrap.Tooltip(tooltipTriggerEl);
    });
}

// Auto-refresh functionality
function startAutoRefresh() {
    refreshInterval = setInterval(() => {
        refreshDashboard();
    }, REFRESH_INTERVAL);
}

function stopAutoRefresh() {
    if (refreshInterval) {
        clearInterval(refreshInterval);
        refreshInterval = null;
    }
}

// Refresh dashboard data
async function refreshDashboard() {
    try {
        const response = await fetch(window.location.href, {
            headers: {
                'X-Requested-With': 'XMLHttpRequest'
            }
        });

        if (response.ok) {
            // For now, just reload the page
            // In a more sophisticated implementation, we'd update specific parts
            window.location.reload();
        }
    } catch (error) {
        console.error('Failed to refresh dashboard:', error);
        showToast('Failed to refresh dashboard', 'error');
    }
}

// Session management functions
async function resumeSession(agentId) {
    if (!confirm(`Resume suspended session for agent ${agentId}?`)) {
        return;
    }

    try {
        showToast(`Resuming session for agent ${agentId}...`, 'info');

        // Copy resume command to clipboard for user convenience
        const resumeCommand = `restart session with agent_id ${agentId}`;
        await navigator.clipboard.writeText(resumeCommand);

        showToast(`Resume command copied to clipboard: "${resumeCommand}"`, 'success');
        showToast('Paste this command in Claude Code to resume the session', 'info');

        // Optional: Try to bring Claude Code window to front if available
        setTimeout(() => {
            bringToFront('code');
        }, 1000);

    } catch (error) {
        console.error('Error resuming session:', error);
        showToast(`To resume: Say "restart session with agent_id ${agentId}" in Claude Code`, 'info');
    }
}

// Agent window management
async function bringAgentFront(agentId) {
    try {
        showToast(`Bringing agent ${agentId} window to front...`, 'info');

        const response = await fetch(`/api/agents/${agentId}/bring-front`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        const result = await response.json();
        if (result.success) {
            showToast(result.message, 'success');
        } else {
            showToast(`Failed to bring window to front: ${result.message}`, 'error');
        }
    } catch (error) {
        console.error('Error bringing agent window to front:', error);
        showToast('Failed to bring agent window to front', 'error');
    }
}

// Agent management functions
async function endSession(agentId) {
    if (!confirm(`Are you sure you want to end the session for agent ${agentId}? This will mark it as completed.`)) {
        return;
    }

    try {
        const response = await fetch(`/api/agents/${agentId}/end`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (response.ok) {
            showToast(`Session ended for agent ${agentId}`, 'success');
            setTimeout(() => refreshDashboard(), 1000);
        } else {
            throw new Error('Failed to end session');
        }
    } catch (error) {
        console.error('Error ending session:', error);
        showToast('Failed to end session', 'error');
    }
}

async function archiveAgent(agentId) {
    if (!confirm(`Are you sure you want to archive agent ${agentId}? This will hide it from the dashboard.`)) {
        return;
    }

    try {
        const response = await fetch(`/api/agents/${agentId}/archive`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (response.ok) {
            showToast(`Agent ${agentId} archived`, 'success');
            setTimeout(() => refreshDashboard(), 1000);
        } else {
            throw new Error('Failed to archive agent');
        }
    } catch (error) {
        console.error('Error archiving agent:', error);
        showToast('Failed to archive agent', 'error');
    }
}

async function updateStatus(agentId) {
    const newStatus = prompt('Enter new status (active/idle/working/completed/failed):');
    if (!newStatus || !['active', 'idle', 'working', 'completed', 'failed'].includes(newStatus)) {
        return;
    }

    try {
        const response = await fetch(`/api/agents/${agentId}/status`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ status: newStatus })
        });

        if (response.ok) {
            showToast(`Status updated for agent ${agentId}`, 'success');
            setTimeout(() => refreshDashboard(), 1000);
        } else {
            throw new Error('Failed to update status');
        }
    } catch (error) {
        console.error('Error updating status:', error);
        showToast('Failed to update status', 'error');
    }
}

async function markIdle(agentId) {
    try {
        const response = await fetch(`/api/agents/${agentId}/status`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ status: 'idle' })
        });

        if (response.ok) {
            showToast(`Agent ${agentId} marked as idle`, 'success');
            setTimeout(() => refreshDashboard(), 1000);
        } else {
            throw new Error('Failed to mark as idle');
        }
    } catch (error) {
        console.error('Error marking as idle:', error);
        showToast('Failed to mark as idle', 'error');
    }
}

// Utility functions
function updateRelativeTimes() {
    document.querySelectorAll('.last-activity').forEach(element => {
        const timestamp = element.getAttribute('data-timestamp');
        if (timestamp) {
            const relativeTime = getRelativeTime(new Date(timestamp));
            element.textContent = relativeTime;
        }
    });
}

function getRelativeTime(date) {
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    return `${diffDays}d ago`;
}

function showToast(message, type = 'info') {
    const toastContainer = document.getElementById('toastContainer') || createToastContainer();

    const toast = document.createElement('div');
    toast.className = `toast align-items-center text-white bg-${type === 'error' ? 'danger' : 'success'} border-0`;
    toast.setAttribute('role', 'alert');
    toast.innerHTML = `
        <div class="d-flex">
            <div class="toast-body">
                <i class="bi bi-${type === 'error' ? 'exclamation-triangle' : 'check-circle'}"></i>
                ${message}
            </div>
            <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
        </div>
    `;

    toastContainer.appendChild(toast);
    const bsToast = new bootstrap.Toast(toast, { delay: 3000 });
    bsToast.show();

    // Remove toast element after it's hidden
    toast.addEventListener('hidden.bs.toast', () => {
        toast.remove();
    });
}

function createToastContainer() {
    const container = document.createElement('div');
    container.id = 'toastContainer';
    container.className = 'toast-container position-fixed bottom-0 end-0 p-3';
    document.body.appendChild(container);
    return container;
}

// Copy agent ID to clipboard
function copyAgentId(agentId) {
    navigator.clipboard.writeText(agentId).then(() => {
        showToast(`Agent ID ${agentId} copied to clipboard`, 'success');
    }).catch(() => {
        showToast('Failed to copy agent ID', 'error');
    });
}

// Keyboard shortcuts
document.addEventListener('keydown', function(e) {
    // Ctrl/Cmd + R to refresh
    if ((e.ctrlKey || e.metaKey) && e.key === 'r') {
        e.preventDefault();
        refreshDashboard();
    }

    // Escape to go back to main dashboard
    if (e.key === 'Escape' && window.location.pathname !== '/') {
        window.location.href = '/';
    }
});

// Visibility API to pause/resume auto-refresh when tab is not visible
document.addEventListener('visibilitychange', function() {
    if (document.hidden) {
        stopAutoRefresh();
    } else {
        startAutoRefresh();
        refreshDashboard(); // Refresh immediately when tab becomes visible
    }
});

// Window focus/blur events
window.addEventListener('focus', function() {
    if (!refreshInterval) {
        startAutoRefresh();
    }
});

window.addEventListener('blur', function() {
    // Keep refreshing even when window is not focused
    // Some users might have dashboard in a separate monitor
});

// Window management functions
async function getWindowId(application = 'current', windowTitle = '') {
    try {
        const response = await fetch('/api/window/get-id', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                application: application,
                window_title: windowTitle
            })
        });

        const result = await response.json();
        if (result.success) {
            showToast(`Window ID: ${result.window_id} (${result.window_info})`, 'success');
            return result;
        } else {
            showToast(`Failed to get window ID: ${result.message}`, 'error');
            return null;
        }
    } catch (error) {
        console.error('Error getting window ID:', error);
        showToast('Failed to get window ID', 'error');
        return null;
    }
}

async function bringToFront(application, windowTitle = '') {
    try {
        const response = await fetch('/api/window/bring-to-front', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                application: application,
                window_title: windowTitle
            })
        });

        const result = await response.json();
        if (result.success) {
            showToast(result.message, 'success');
            return true;
        } else {
            showToast(`Failed to bring window to front: ${result.message}`, 'error');
            return false;
        }
    } catch (error) {
        console.error('Error bringing window to front:', error);
        showToast('Failed to bring window to front', 'error');
        return false;
    }
}

async function listWindows(application = 'ghostty') {
    try {
        const response = await fetch(`/api/window/list?application=${encodeURIComponent(application)}`);
        const result = await response.json();

        if (result.success) {
            console.log('Windows for', application, ':', result.windows);
            return result.windows;
        } else {
            showToast(`Failed to list windows: ${result.message}`, 'error');
            return [];
        }
    } catch (error) {
        console.error('Error listing windows:', error);
        showToast('Failed to list windows', 'error');
        return [];
    }
}

// Focus dashboard window (bring browser to front)
async function focusDashboard() {
    return await bringToFront('dashboard');
}

// Get current dashboard window ID
async function getCurrentDashboardWindowId() {
    return await getWindowId('current');
}

// Enhanced keyboard shortcuts
document.addEventListener('keydown', function(e) {
    // Ctrl/Cmd + R to refresh
    if ((e.ctrlKey || e.metaKey) && e.key === 'r') {
        e.preventDefault();
        refreshDashboard();
    }

    // Escape to go back to main dashboard
    if (e.key === 'Escape' && window.location.pathname !== '/') {
        window.location.href = '/';
    }

    // Ctrl/Cmd + Shift + F to focus dashboard
    if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === 'F') {
        e.preventDefault();
        focusDashboard();
    }

    // Ctrl/Cmd + Shift + W to get current window ID
    if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === 'W') {
        e.preventDefault();
        getCurrentDashboardWindowId();
    }

    // Ctrl/Cmd + Shift + G to bring Ghostty to front
    if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === 'G') {
        e.preventDefault();
        bringToFront('ghostty');
    }
});

// Add window management controls to the dashboard
function addWindowControls() {
    const navbar = document.querySelector('.navbar .container');
    if (navbar) {
        const windowControls = document.createElement('div');
        windowControls.className = 'navbar-nav me-3';
        windowControls.innerHTML = `
            <div class="nav-item dropdown">
                <a class="nav-link dropdown-toggle text-light" href="#" role="button" data-bs-toggle="dropdown">
                    <i class="bi bi-window"></i> Windows
                </a>
                <ul class="dropdown-menu">
                    <li><a class="dropdown-item" href="#" onclick="getCurrentDashboardWindowId()">
                        <i class="bi bi-info-circle"></i> Get Window ID
                    </a></li>
                    <li><a class="dropdown-item" href="#" onclick="focusDashboard()">
                        <i class="bi bi-arrow-up"></i> Focus Dashboard
                    </a></li>
                    <li><hr class="dropdown-divider"></li>
                    <li><a class="dropdown-item" href="#" onclick="bringToFront('ghostty')">
                        <i class="bi bi-terminal"></i> Focus Ghostty
                    </a></li>
                    <li><a class="dropdown-item" href="#" onclick="bringToFront('terminal')">
                        <i class="bi bi-terminal-fill"></i> Focus Terminal
                    </a></li>
                    <li><a class="dropdown-item" href="#" onclick="bringToFront('safari')">
                        <i class="bi bi-globe"></i> Focus Safari
                    </a></li>
                    <li><hr class="dropdown-divider"></li>
                    <li><a class="dropdown-item" href="#" onclick="listWindows('ghostty')">
                        <i class="bi bi-list"></i> List Windows
                    </a></li>
                </ul>
            </div>
        `;

        // Insert before the existing nav content
        const existingNav = navbar.querySelector('.navbar-nav');
        if (existingNav) {
            navbar.insertBefore(windowControls, existingNav);
        }
    }
}

// Initialize window controls when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    addWindowControls();
});

// Export functions for global access
window.dashboardFunctions = {
    resumeSession,
    endSession,
    archiveAgent,
    updateStatus,
    markIdle,
    refreshDashboard,
    copyAgentId,
    bringAgentFront,
    getWindowId,
    bringToFront,
    listWindows,
    focusDashboard,
    getCurrentDashboardWindowId
};