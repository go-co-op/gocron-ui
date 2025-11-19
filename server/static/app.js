// State
let jobs = [];
let ws = null;
let isConnected = false;
let expandedSchedules = new Set(); // Track which job schedules are expanded

// API Base URL
const API_BASE = window.location.origin + '/api';
const WS_PROTOCOL = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const WS_URL = `${WS_PROTOCOL}//${window.location.host}/ws`;

// initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    loadConfig();
    connectWebSocket();
});

// load server configuration
async function loadConfig() {
    try {
        const response = await fetch(`${API_BASE}/config`);
        if (response.ok) {
            const config = await response.json();
            if (config.title) {
                // update page title
                document.title = config.title;

                // update header title
                const mainTitle = document.getElementById('main-title');
                const poweredBy = document.getElementById('powered-by');

                if (mainTitle) {
                    mainTitle.textContent = config.title;
                }

                if (poweredBy && config.title !== 'GoCron UI') {
                    poweredBy.style.display = 'block';
                }
            }
        }
    } catch (err) {
        console.error('Failed to load config:', err);
        // fallback to default title
    }
}

// webSocket connection
function connectWebSocket() {
    ws = new WebSocket(WS_URL);

    ws.onopen = () => {
        console.log('WebSocket connected');
        isConnected = true;
        updateConnectionStatus(true);
        hideError();
    };

    ws.onmessage = (event) => {
        try {
            const message = JSON.parse(event.data);
            if (message.type === 'jobs') {
                jobs = message.data || [];
                renderJobs();
            }
        } catch (err) {
            console.error('Failed to parse WebSocket message:', err);
        }
    };

    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        showError('Connection error. Retrying...');
    };

    ws.onclose = () => {
        console.log('WebSocket disconnected');
        isConnected = false;
        updateConnectionStatus(false);
        // reconnect after 3 seconds
        setTimeout(connectWebSocket, 3000);
    };
}

function updateConnectionStatus(connected) {
    const statusEl = document.getElementById('connection-status');
    if (connected) {
        statusEl.className = 'status-indicator connected';
        statusEl.textContent = '● Connected';
    } else {
        statusEl.className = 'status-indicator disconnected';
        statusEl.textContent = '○ Disconnected';
    }
}

// error handling
function showError(message) {
    const banner = document.getElementById('error-banner');
    const messageEl = document.getElementById('error-message');
    messageEl.textContent = message;
    banner.style.display = 'flex';
}

function hideError() {
    document.getElementById('error-banner').style.display = 'none';
}

// API Functions
async function deleteJob(id) {
    const response = await fetch(`${API_BASE}/jobs/${id}`, {
        method: 'DELETE',
    });

    if (!response.ok) {
        throw new Error('Failed to delete job');
    }
}

async function runJob(id) {
    const response = await fetch(`${API_BASE}/jobs/${id}/run`, {
        method: 'POST',
    });

    if (!response.ok) {
        throw new Error('Failed to run job');
    }
}

// job actions
async function handleRunJob(id) {
    try {
        await runJob(id);
        hideError();
    } catch (err) {
        showError(err.message);
    }
}

async function handleDeleteJob(id, name) {
    if (!confirm(`Are you sure you want to delete job "${name}"?`)) {
        return;
    }

    try {
        await deleteJob(id);
    } catch (err) {
        showError(err.message);
    }
}

// rendering
function renderJobs() {
    const container = document.getElementById('jobs-container');

    // update job count in info banner
    const jobCount = document.getElementById('job-count');
    if (jobCount) {
        jobCount.textContent = jobs ? jobs.length : 0;
    }

    // clean up expanded state for deleted jobs
    if (jobs && jobs.length > 0) {
        const currentJobIds = new Set(jobs.map(j => j.id));
        expandedSchedules.forEach(id => {
            if (!currentJobIds.has(id)) {
                expandedSchedules.delete(id);
            }
        });
    } else {
        // no jobs, clear all expanded state
        expandedSchedules.clear();
    }

    if (!jobs || jobs.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <div class="empty-icon">⏰</div>
                <h2>No Jobs Running</h2>
                <p>Jobs are defined in your Go code. Once you start your application with scheduled jobs, they will appear here.</p>
                <div class="empty-hint">
                    💡 <strong>Tip:</strong> Check out the <a href="https://pkg.go.dev/github.com/go-co-op/gocron/v2" target="_blank">gocron documentation</a> to create jobs programmatically.
                </div>
            </div>
        `;
        return;
    }

    container.innerHTML = `
        <div class="job-list">
            <h2 class="job-list-title">Scheduled Jobs (${jobs.length})</h2>
            <div class="job-grid">
                ${jobs.map(job => renderJobCard(job)).join('')}
            </div>
        </div>
    `;
}

function renderJobCard(job) {
    const nextRun = job.nextRun ? formatDateTime(job.nextRun) : 'Never';
    const lastRun = job.lastRun ? formatDateTime(job.lastRun) : 'Never';
    const timeUntil = job.nextRun ? getTimeUntil(job.nextRun) : '';

    return `
        <div class="job-card">
            <div class="job-card-header">
                <h3 class="job-name">${escapeHtml(job.name)}</h3>
                <div class="job-actions">
                    <button
                        class="btn btn-success btn-sm"
                        onclick="handleRunJob('${job.id}')"
                        title="Run now"
                    >
                        ▶️
                    </button>
                    <button
                        class="btn btn-danger btn-sm"
                        onclick="handleDeleteJob('${job.id}', '${escapeHtml(job.name)}')"
                        title="Delete"
                    >
                        🗑️
                    </button>
                </div>
            </div>

            ${job.tags && job.tags.length > 0 ? `
                <div class="job-tags">
                    ${job.tags.map(tag => `<span class="tag">🏷️ ${escapeHtml(tag)}</span>`).join('')}
                </div>
            ` : ''}

            ${job.schedule ? `
                <div class="job-info-item" style="margin-bottom: 1rem;">
                    <span class="job-info-label">Schedule:</span>
                    <div class="job-info-value">
                        <span class="schedule-badge">${escapeHtml(job.schedule)}</span>
                        ${job.scheduleDetail ? `
                            <span class="schedule-detail">${escapeHtml(job.scheduleDetail)}</span>
                        ` : ''}
                    </div>
                </div>
            ` : ''}

            <div class="job-info">
                <div class="job-info-item">
                    <span class="job-info-label">Next Run:</span>
                    <span class="job-info-value">
                        ${nextRun}
                        ${timeUntil ? `<span class="time-until">⏱️ ${timeUntil}</span>` : ''}
                    </span>
                </div>
                <div class="job-info-item">
                    <span class="job-info-label">Last Run:</span>
                    <span class="job-info-value">${lastRun}</span>
                </div>
                <div class="job-info-item">
                    <span class="job-info-label">Job ID:</span>
                    <span class="job-info-value job-id">${job.id}</span>
                </div>
            </div>

            ${job.nextRuns && job.nextRuns.length > 0 ? `
                <div class="job-schedule">
                    <button class="schedule-toggle" onclick="toggleSchedule('${job.id}')">
                        <span id="toggle-icon-${job.id}">${expandedSchedules.has(job.id) ? '🔽' : '▶️'}</span> Upcoming Runs
                    </button>
                    <div id="schedule-${job.id}" class="schedule-details" style="display: ${expandedSchedules.has(job.id) ? 'block' : 'none'};">
                        ${job.nextRuns.map(run => `
                            <div class="schedule-item">📌 ${formatDateTime(run)}</div>
                        `).join('')}
                    </div>
                </div>
            ` : ''}
        </div>
    `;
}

function toggleSchedule(jobId) {
    const details = document.getElementById(`schedule-${jobId}`);
    const icon = document.getElementById(`toggle-icon-${jobId}`);

    if (details.style.display === 'none') {
        details.style.display = 'block';
        icon.textContent = '🔽';
        expandedSchedules.add(jobId); // remember this is expanded
    } else {
        details.style.display = 'none';
        icon.textContent = '▶️';
        expandedSchedules.delete(jobId); // remember this is collapsed
    }
}

// utility functions
function formatDateTime(dateStr) {
    if (!dateStr) return 'Never';
    const date = new Date(dateStr);
    return date.toLocaleString();
}

function getTimeUntil(dateStr) {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    const now = new Date();
    const diff = date.getTime() - now.getTime();

    if (diff < 0) return 'Overdue';

    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) return `in ${days}d ${hours % 24}h`;
    if (hours > 0) return `in ${hours}h ${minutes % 60}m`;
    if (minutes > 0) return `in ${minutes}m`;
    return `in ${seconds}s`;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
