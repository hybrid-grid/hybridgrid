// Chart instances live outside Alpine's reactive proxy on purpose —
// wrapping Chart.js controllers in a reactive proxy breaks their
// internal bookkeeping (resize observers, animation state).
const chartRegistry = {};

// Single scratch canvas for token rasterization, reused across every
// chartTheme() call — allocating five fresh 2d contexts per call (and
// chartTheme() runs on every stats refresh) is pointless garbage.
let scratchCanvas = null;

function destroyCharts() {
    Object.keys(chartRegistry).forEach((key) => {
        try { chartRegistry[key].destroy(); } catch (err) { /* already gone */ }
        delete chartRegistry[key];
    });
}

function dashboard() {
    return {
        connected: false,
        ws: null,
        stats: {},
        workers: [],
        builds: [],
        durationChartBuilds: [],
        events: [],
        recentTasks: [],
        cacheHistory: Array(24).fill(0),
        lastUpdate: 'Never',
        reconnectAttempts: 0,
        maxReconnectAttempts: 10,
        lastBuildFetchAt: 0,
        buildFetchTimer: null,
        route: 'home',
        detailId: '',
        detail: null,
        detailMissing: false,
        detailLoading: false,
        lastDetailFetchAt: 0,
        detailFetchTimer: null,
        consoleTaskId: '',
        consoleData: null,
        consoleMissing: false,
        consoleLoading: false,
        themeQuery: null,

        get cacheHitRate() {
            const hits = this.stats.cache_hits || 0;
            const misses = this.stats.cache_misses || 0;
            const total = hits + misses;
            return total > 0 ? ((hits / total) * 100).toFixed(1) : '0.0';
        },

        // Metric-card derivations — every value traces to a live API field.
        get passRateDisplay() {
            const succeeded = Number(this.stats.success_tasks || 0);
            const failed = Number(this.stats.failed_tasks || 0);
            const total = succeeded + failed;
            return total > 0 ? `${((succeeded / total) * 100).toFixed(1)}%` : '—';
        },

        get avgBuildDurationDisplay() {
            const finished = this.builds.filter((b) => !b.running_tasks && b.first_task_at_ms && b.last_task_at_ms && b.last_task_at_ms > b.first_task_at_ms);
            if (finished.length === 0) return '—';
            const avgSeconds = finished.reduce((acc, b) => acc + (b.last_task_at_ms - b.first_task_at_ms), 0) / finished.length / 1000;
            return avgSeconds >= 60 ? `${(avgSeconds / 60).toFixed(1)} min` : `${avgSeconds.toFixed(1)} s`;
        },

        get executorUtilizationDisplay() {
            const max = this.workers.reduce((acc, w) => acc + (w.max_parallel_tasks || 0), 0);
            if (max === 0) return '—';
            const active = this.workers.reduce((acc, w) => acc + (w.active_tasks || 0), 0);
            return `${Math.min(100, Math.round((active / max) * 100))}%`;
        },

        // Build-detail task timeline. Times are Unix milliseconds from the
        // coordinator; the queue segment uses the measured coordinator-side
        // queue time (started_at_ms - queue_ms). Running tasks extend to now.
        get detailTimeline() {
            if (!this.detail || !this.detail.tasks) return { rows: [], spanMs: 0 };
            const now = Date.now();
            const rows = this.detail.tasks
                .map((task) => {
                    const started = task.started_at_ms || 0;
                    if (!started) return null;
                    const queuedAt = started - (task.queue_ms || 0);
                    const end = task.completed_at_ms || (task.status === 'running' ? now : started + (task.duration_ms || 0));
                    return { task, queuedAt, started, end };
                })
                .filter(Boolean);
            if (rows.length === 0) return { rows: [], spanMs: 0 };
            const t0 = Math.min(...rows.map((r) => r.queuedAt));
            // Running rows already carry end = now, so the max of ends
            // covers the live case. An unconditional `now` here would
            // keep inflating the span of a settled build with wall clock.
            const t1 = Math.max(...rows.map((r) => r.end));
            const span = Math.max(t1 - t0, 1);
            rows.forEach((r) => {
                r.queueLeft = ((r.queuedAt - t0) / span) * 100;
                r.queueWidth = Math.max(((r.started - r.queuedAt) / span) * 100, 0.25);
                r.runLeft = ((r.started - t0) / span) * 100;
                r.runWidth = Math.max(((r.end - r.started) / span) * 100, 0.35);
            });
            rows.sort((a, b) => a.started - b.started);
            return { rows, spanMs: span };
        },

        timelineSpanLabel() {
            const span = this.detailTimeline.spanMs;
            if (!span) return '';
            return span >= 60000 ? `span ${(span / 60000).toFixed(1)} min` : `span ${(span / 1000).toFixed(1)} s`;
        },

        timelineTooltip(row) {
            const t = row.task;
            return `${t.id}\nqueue ${this.formatMs(t.queue_ms)} · duration ${this.formatMs(t.duration_ms)}${t.completed_at_ms ? '' : ' · running'}`;
        },

        init() {
            window.addEventListener('hashchange', () => this.handleRoute());
            this.handleRoute();
            this.fetchInitialData();
            this.connectWebSocket();
            if (window.matchMedia) {
                this.themeQuery = window.matchMedia('(prefers-color-scheme: dark)');
                const recreate = () => { destroyCharts(); this.renderRouteCharts(); };
                if (this.themeQuery.addEventListener) this.themeQuery.addEventListener('change', recreate);
                else if (this.themeQuery.addListener) this.themeQuery.addListener(recreate);
            }
        },

        async fetchInitialData() {
            try {
                const [statsRes, workersRes, eventsRes, buildsRes] = await Promise.all([
                    fetch('/api/v1/stats'),
                    fetch('/api/v1/workers'),
                    fetch('/api/v1/events'),
                    fetch('/api/v1/builds')
                ]);
                this.stats = await statsRes.json();
                const workersData = await workersRes.json();
                this.workers = workersData.workers || [];
                const eventsData = await eventsRes.json();
                this.events = eventsData.events || [];
                const buildsData = await buildsRes.json();
                this.builds = buildsData.builds || buildsData || [];
                this.lastBuildFetchAt = Date.now();
                this.lastUpdate = this.formatEventTime(Date.now() / 1000);
                this.updateCacheHistory();
                this.$nextTick(() => this.renderRouteCharts());
            } catch (err) {
                console.error('Failed to fetch initial dashboard data:', err);
            }
        },

        async fetchBuilds() {
            try {
                const response = await fetch('/api/v1/builds');
                const data = await response.json();
                this.builds = data.builds || data || [];
                this.lastBuildFetchAt = Date.now();
                this.$nextTick(() => this.updateDurationsChart());
            } catch (err) {
                console.error('Failed to fetch builds:', err);
            }
        },

        refreshBuildsThrottled() {
            const interval = 2000;
            const remaining = interval - (Date.now() - this.lastBuildFetchAt);
            if (remaining <= 0) {
                this.fetchBuilds();
                return;
            }
            if (!this.buildFetchTimer) {
                this.buildFetchTimer = setTimeout(() => {
                    this.buildFetchTimer = null;
                    this.fetchBuilds();
                }, remaining);
            }
        },

        handleRoute() {
            const consoleMatch = /^#\/task\/(.+)\/console$/.exec(window.location.hash);
            if (consoleMatch) {
                const id = decodeURIComponent(consoleMatch[1]);
                if (id !== this.consoleTaskId) {
                    this.consoleTaskId = id;
                    this.consoleData = null;
                    this.consoleMissing = false;
                    this.fetchConsole();
                }
                this.route = 'console';
                this.$nextTick(() => this.renderRouteCharts());
                return;
            }
            const match = /^#\/build\/(.+)$/.exec(window.location.hash);
            if (match) {
                const id = decodeURIComponent(match[1]);
                if (id !== this.detailId) {
                    this.detailId = id;
                    this.detail = null;
                    this.detailMissing = false;
                    this.fetchDetail();
                }
                this.route = 'build';
            } else {
                this.route = 'home';
            }
            this.$nextTick(() => this.renderRouteCharts());
        },

        async fetchConsole() {
            if (!this.consoleTaskId) return;
            this.consoleLoading = true;
            try {
                const response = await fetch(`/api/v1/tasks/${encodeURIComponent(this.consoleTaskId)}/console`);
                if (response.status === 404 || response.status === 503) {
                    this.consoleData = null;
                    this.consoleMissing = true;
                } else if (response.ok) {
                    this.consoleData = await response.json();
                    this.consoleMissing = false;
                }
            } catch (err) {
                console.error('Failed to fetch task console:', err);
            } finally {
                this.consoleLoading = false;
            }
        },

        async fetchDetail() {
            if (!this.detailId) return;
            this.detailLoading = true;
            try {
                const response = await fetch(`/api/v1/builds/${encodeURIComponent(this.detailId)}`);
                if (response.status === 404) {
                    this.detail = null;
                    this.detailMissing = true;
                } else if (response.ok) {
                    this.detail = await response.json();
                    this.detailMissing = false;
                    this.$nextTick(() => this.updateDetailCharts());
                }
            } catch (err) {
                console.error('Failed to fetch build detail:', err);
            } finally {
                this.detailLoading = false;
                this.lastDetailFetchAt = Date.now();
            }
        },

        refreshDetailThrottled() {
            if (this.route !== 'build') return;
            const interval = 2000;
            const remaining = interval - (Date.now() - this.lastDetailFetchAt);
            if (remaining <= 0) {
                this.fetchDetail();
                return;
            }
            if (!this.detailFetchTimer) {
                this.detailFetchTimer = setTimeout(() => {
                    this.detailFetchTimer = null;
                    this.fetchDetail();
                }, remaining);
            }
        },

        updateCacheHistory() {
            this.cacheHistory = [...this.cacheHistory.slice(1), parseFloat(this.cacheHitRate)];
        },

        connectWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            this.ws = new WebSocket(`${protocol}//${window.location.host}/ws`);

            this.ws.onopen = () => {
                this.connected = true;
                this.reconnectAttempts = 0;
            };
            this.ws.onclose = () => {
                this.connected = false;
                this.scheduleReconnect();
            };
            this.ws.onerror = (err) => console.error('WebSocket error:', err);
            this.ws.onmessage = (event) => {
                event.data.split('\n').forEach((line) => {
                    if (!line.trim()) return;
                    try {
                        this.handleMessage(JSON.parse(line));
                    } catch (err) {
                        console.error('Failed to parse WebSocket message:', err);
                    }
                });
            };
        },

        scheduleReconnect() {
            if (this.reconnectAttempts >= this.maxReconnectAttempts) return;
            const delay = Math.min(1000 * (2 ** this.reconnectAttempts), 30000);
            this.reconnectAttempts += 1;
            setTimeout(() => this.connectWebSocket(), delay);
        },

        handleMessage(message) {
            this.lastUpdate = this.formatEventTime(message.timestamp || Date.now() / 1000);
            switch (message.type) {
                case 'stats':
                    this.stats = message.data;
                    this.updateCacheHistory();
                    this.updateCacheChart();
                    break;
                case 'worker_added':
                    this.workers = this.workers.filter((worker) => worker.id !== message.data.id);
                    this.workers.push(message.data);
                    this.events.push(message);
                    break;
                case 'worker_removed':
                    this.workers = this.workers.filter((worker) => worker.id !== message.data.worker_id);
                    this.events.push(message);
                    break;
                case 'task_started':
                    this.recentTasks = this.recentTasks.filter((task) => task.id !== message.data.id);
                    this.recentTasks.unshift({ ...message.data, status: 'running' });
                    this.events.push(message);
                    this.refreshBuildsThrottled();
                    if (message.data.build_id === this.detailId) this.refreshDetailThrottled();
                    break;
                case 'task_completed': {
                    const index = this.recentTasks.findIndex((task) => task.id === message.data.id);
                    const status = message.data.exit_code === 0 ? 'success' : 'failed';
                    if (index >= 0) this.recentTasks[index] = { ...message.data, status };
                    else this.recentTasks.unshift({ ...message.data, status });
                    this.events.push(message);
                    this.refreshBuildsThrottled();
                    if (message.data.build_id === this.detailId) this.refreshDetailThrottled();
                    break;
                }
            }
            this.events = this.events.slice(-100);
            this.recentTasks = this.recentTasks.slice(0, 20);
        },

        buildStatus(build) {
            if (build.status === 'running' || build.running_tasks > 0) return 'running';
            if (build.status === 'failed' || build.failed_tasks > 0) return 'failed';
            if (build.status === 'queued') return 'queued';
            return 'success';
        },

        buildProgress(build) {
            if (!build.total_tasks) return 0;
            return Math.min(100, ((build.completed_tasks + build.failed_tasks) / build.total_tasks) * 100);
        },

        buildDurationLabel(build) {
            if (!build.first_task_at_ms || !build.last_task_at_ms || build.last_task_at_ms <= build.first_task_at_ms) return '—';
            const seconds = (build.last_task_at_ms - build.first_task_at_ms) / 1000;
            return seconds >= 60 ? `${(seconds / 60).toFixed(1)} min` : `${seconds.toFixed(2)} s`;
        },

        workerSlots(worker) {
            const active = this.recentTasks.filter((task) => task.worker_id === worker.id && task.status === 'running');
            return Array.from({ length: Math.max(0, worker.max_parallel_tasks || 0) }, (_, index) => active[index] || null);
        },

        workerUtilization(worker) {
            const max = worker.max_parallel_tasks || 0;
            if (max <= 0) return 0;
            return Math.min(100, Math.round(((worker.active_tasks || 0) / max) * 100));
        },

        taskBall(task) {
            if (task.status === 'running') return 'running';
            if (task.status === 'failed' || (task.exit_code !== undefined && task.exit_code !== 0)) return 'failed';
            if (task.status === 'pending' || task.status === 'queued') return 'queued';
            return 'success';
        },

        formatMs(ms) {
            if (ms === null || ms === undefined) return '—';
            return `${ms} ms`;
        },

        capability(worker) {
            const architecture = worker.architecture || (worker.architectures || []).join(', ') || 'unknown architecture';
            const cores = worker.cpu_cores || 0;
            const memory = Number(worker.memory_gb || 0).toFixed(1);
            return `${architecture} · ${cores} cores · ${memory} GB · circuit ${worker.circuit_state || 'CLOSED'}`;
        },

        circuitClass(worker) {
            return `circuit-${String(worker.circuit_state || 'CLOSED').toLowerCase().replace('_', '-')}`;
        },

        // Task and build timestamps are Unix milliseconds; WebSocket event
        // timestamps are Unix seconds — two formatters, no silent mixing.
        formatTime(timestampMs) {
            if (!timestampMs) return '';
            return new Date(timestampMs).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
        },

        formatEventTime(timestampSeconds) {
            if (!timestampSeconds) return '';
            return new Date(timestampSeconds * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
        },

        formatUptime(seconds) {
            if (!seconds) return '0s';
            const days = Math.floor(seconds / 86400);
            const hours = Math.floor((seconds % 86400) / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            const remainingSeconds = seconds % 60;
            if (days > 0) return `${days}d ${hours}h`;
            if (hours > 0) return `${hours}h ${minutes}m`;
            if (minutes > 0) return `${minutes}m ${remainingSeconds}s`;
            return `${remainingSeconds}s`;
        },

        formatEventData(event) {
            if (!event.data) return '';
            if (event.type === 'stats') return `Active: ${event.data.active_tasks || 0}, workers: ${event.data.healthy_workers || 0}`;
            if (event.type === 'worker_added') return `Worker ${event.data.host || event.data.id} joined`;
            if (event.type === 'worker_removed') return `Worker ${event.data.worker_id} left`;
            if (event.type === 'task_started') return `Task ${event.data.id} started on ${event.data.worker_id}`;
            if (event.type === 'task_completed') return `Task ${event.data.id} completed (${event.data.status || 'finished'})`;
            return JSON.stringify(event.data);
        },

        // ---- Chart.js wiring ------------------------------------------------

        renderRouteCharts() {
            if (typeof Chart === 'undefined') return;
            if (this.route === 'home') {
                this.updateDurationsChart();
                this.updateCacheChart();
            } else if (this.route === 'build' && this.detail) {
                this.updateDetailCharts();
            }
        },

        // Chart colors are resolved from CSS custom properties via probe
        // elements (custom properties are not computed values, so reading
        // a real color property is the reliable way). The resolved string
        // is then rasterized through a 1x1 canvas and read back with
        // getImageData: Chromium returns authored forms like oklch(...)
        // from computed styles AND from canvas fillStyle serialization,
        // but Chart.js's @kurkle/color parser only understands
        // hex/rgb/hsl — every parsed path (hover alpha, legend swatches,
        // tooltip backgrounds) would render black/transparent. The
        // rasterizer converts the token to concrete sRGB bytes, so the
        // returned rgba(...) string parses correctly everywhere.
        // Fallbacks are neutral greys from the existing token values
        // (not approximations of the Jenkins colors) and only apply when
        // even the browser cannot paint the token.
        chartTheme() {
            const roundTrip = (cssColor, fallback) => {
                try {
                    if (!scratchCanvas) {
                        scratchCanvas = document.createElement('canvas');
                        scratchCanvas.width = 1;
                        scratchCanvas.height = 1;
                    }
                    const ctx = scratchCanvas.getContext('2d', { willReadFrequently: true });
                    if (!ctx) return fallback;
                    ctx.clearRect(0, 0, 1, 1);
                    ctx.fillStyle = '#123456';
                    ctx.fillStyle = cssColor;
                    if (ctx.fillStyle === '#123456') return fallback;
                    ctx.fillRect(0, 0, 1, 1);
                    const d = ctx.getImageData(0, 0, 1, 1).data;
                    if (d[3] === 0) return fallback;
                    return `rgba(${d[0]}, ${d[1]}, ${d[2]}, ${(d[3] / 255).toFixed(4)})`;
                } catch (err) {
                    return fallback;
                }
            };
            const resolved = (id, fallback) => {
                try {
                    const el = document.getElementById(id);
                    return el ? roundTrip(getComputedStyle(el).color, fallback) : fallback;
                } catch (err) {
                    return fallback;
                }
            };
            return {
                text: resolved('probe-text', '#4d545d'),
                grid: resolved('probe-grid', 'rgba(77, 84, 93, 0.2)'),
                blue: resolved('probe-blue', '#4d545d'),
                red: resolved('probe-red', '#4d545d'),
                green: resolved('probe-green', '#4d545d'),
                grey: resolved('probe-grey', '#9ba7af')
            };
        },

        ensureChart(key, canvasId, makeConfig) {
            if (chartRegistry[key]) return chartRegistry[key];
            const canvas = document.getElementById(canvasId);
            if (!canvas || typeof Chart === 'undefined') return null;
            const ctx = canvas.getContext('2d');
            if (!ctx) return null;
            chartRegistry[key] = new Chart(ctx, makeConfig(this.chartTheme()));
            return chartRegistry[key];
        },

        updateDurationsChart() {
            if (this.route !== 'home') return;
            const chart = this.ensureChart('durations', 'durations-canvas', (c) => ({
                type: 'bar',
                // minBarLength keeps sub-second/cached builds visible and
                // hoverable — without it a 0.02 s build renders 0 px tall
                // on a chart whose y-axis spans tens of seconds, silently
                // dropping exactly the builds the tooltip exists to inspect.
                data: { labels: [], datasets: [{ label: 'duration (s)', data: [], backgroundColor: [], minBarLength: 2 }] },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    animation: false,
                    plugins: {
                        legend: { display: false },
                        tooltip: {
                            callbacks: {
                                title: (items) => this.durationChartBuilds[items[0].dataIndex] ? this.durationChartBuilds[items[0].dataIndex].id : '',
                                label: (item) => {
                                    const build = this.durationChartBuilds[item.dataIndex];
                                    if (!build) return '';
                                    const status = this.buildStatus(build);
                                    const cached = build.from_cache_count || 0;
                                    return [`${item.parsed.y.toFixed(2)} s · ${status}`, `${build.completed_tasks + build.failed_tasks}/${build.total_tasks} tasks · ${cached} cached`];
                                }
                            }
                        }
                    },
                    scales: {
                        x: { ticks: { color: c.text, maxRotation: 0, autoSkip: true, callback: function (value) { return String(this.getLabelForValue(value)).slice(0, 8); } }, grid: { display: false } },
                        y: { beginAtZero: true, ticks: { color: c.text }, grid: { color: c.grid }, title: { display: true, text: 'seconds', color: c.text } }
                    }
                }
            }));
            if (!chart) return;
            // Oldest → newest, most recent 20 builds. Stored on the
            // component so tooltip callbacks resolve dataIndex against
            // the exact array the dataset was built from — this.builds
            // is newest-first and may contain entries the filter
            // dropped, which would mislabel every hover.
            const builds = this.builds
                .filter((b) => b.first_task_at_ms && b.last_task_at_ms && b.last_task_at_ms > b.first_task_at_ms)
                .slice(0, 20)
                .reverse();
            this.durationChartBuilds = builds;
            const t = this.chartTheme();
            const colors = { success: t.blue, failed: t.red, running: t.grey, queued: t.grey };
            chart.data.labels = builds.map((b) => b.id);
            chart.data.datasets[0].data = builds.map((b) => (b.last_task_at_ms - b.first_task_at_ms) / 1000);
            chart.data.datasets[0].backgroundColor = builds.map((b) => colors[this.buildStatus(b)] || t.grey);
            chart.update('none');
        },

        updateCacheChart() {
            if (this.route !== 'home') return;
            const chart = this.ensureChart('cache', 'cache-canvas', (c) => ({
                type: 'line',
                data: { labels: this.cacheHistory.map((_, i) => i - this.cacheHistory.length + 1), datasets: [{ label: 'hit rate %', data: [...this.cacheHistory], borderColor: c.blue, backgroundColor: c.blue, tension: 0.35, pointRadius: 0, borderWidth: 2, fill: false }] },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    animation: false,
                    plugins: { legend: { display: false } },
                    scales: {
                        x: { ticks: { color: c.text, autoSkip: true, callback: function (value) { return `${Number(value) * 2}s`; } }, grid: { display: false } },
                        y: { min: 0, max: 100, ticks: { color: c.text, callback: (value) => `${value}%` }, grid: { color: c.grid } }
                    }
                }
            }));
            if (!chart) return;
            chart.data.datasets[0].data = [...this.cacheHistory];
            chart.data.labels = this.cacheHistory.map((_, i) => i - this.cacheHistory.length + 1);
            chart.update('none');
        },

        updateDetailCharts() {
            if (this.route !== 'build' || !this.detail) return;
            const build = this.detail.build;
            const chart = this.ensureChart('outcomes', 'outcomes-canvas', (c) => ({
                type: 'doughnut',
                data: { labels: ['passed', 'failed', 'running', 'other'], datasets: [{ data: [0, 0, 0, 0], backgroundColor: [c.blue, c.red, c.grey, c.grid], borderWidth: 0 }] },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    animation: false,
                    cutout: '62%',
                    plugins: { legend: { position: 'right', labels: { color: c.text, boxWidth: 10, boxHeight: 10 } } }
                }
            }));
            if (!chart) return;
            const completed = build.completed_tasks || 0;
            const failed = build.failed_tasks || 0;
            const running = build.running_tasks || 0;
            const other = Math.max(0, (build.total_tasks || 0) - completed - failed - running);
            chart.data.datasets[0].data = [completed, failed, running, other];
            chart.update('none');
        }
    };
}
