function dashboard() {
    return {
        connected: false,
        ws: null,
        stats: {},
        workers: [],
        builds: [],
        events: [],
        recentTasks: [],
        cacheHistory: Array(12).fill(0),
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

        get cacheHitRate() {
            const hits = this.stats.cache_hits || 0;
            const misses = this.stats.cache_misses || 0;
            const total = hits + misses;
            return total > 0 ? ((hits / total) * 100).toFixed(1) : '0.0';
        },

        init() {
            window.addEventListener('hashchange', () => this.handleRoute());
            this.handleRoute();
            this.fetchInitialData();
            this.connectWebSocket();
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
                this.lastUpdate = this.formatTime(Date.now() / 1000);
                this.updateCacheHistory();
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
            this.lastUpdate = this.formatTime(message.timestamp || Date.now() / 1000);
            switch (message.type) {
                case 'stats':
                    this.stats = message.data;
                    this.updateCacheHistory();
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

        workerSlots(worker) {
            const active = this.recentTasks.filter((task) => task.worker_id === worker.id && task.status === 'running');
            return Array.from({ length: Math.max(0, worker.max_parallel_tasks || 0) }, (_, index) => active[index] || null);
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

        formatTime(timestamp) {
            if (!timestamp) return '';
            return new Date(timestamp * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
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
        }
    };
}
