import { useSettingsStore } from '@/store/useSettingsStore'

const en: Record<string, Record<string, string>> = {
  common: {
    live: 'Live',
    connecting: 'Connecting',
    reconnecting: 'Reconnecting',
    loading: 'Loading…',
    empty: '(empty)',
    justNow: 'just now',
    secondsAgo: '{n}s ago',
    minutesAgo: '{n}m ago',
    hoursAgo: '{n}h ago',
    daysAgo: '{n}d ago',
    clear: 'Clear',
    cache: 'cache',
  },
  nav: {
    overview: 'Overview',
    builds: 'Builds',
    workers: 'Workers',
    brandSubtitle: 'Compile farm control',
  },
  status: {
    STATUS_QUEUED: 'Queued',
    STATUS_RUNNING: 'Running',
    STATUS_COMPLETED: 'Success',
    STATUS_FAILED: 'Failed',
    STATUS_TIMEOUT: 'Timeout',
  },
  circuit: {
    CLOSED: 'Closed',
    HALF_OPEN: 'Half-open',
    OPEN: 'Open',
  },
  overview: {
    eyebrow: 'Control room',
    title: 'Fleet overview',
    metricActive: 'Active',
    metricQueued: 'Queued',
    metricSucceeded: 'Succeeded',
    metricFailed: 'Failed',
    metricCacheHitRate: 'Cache hit rate',
    metricHealthyWorkers: 'Healthy workers',
    metricAvgBuildDuration: 'Avg build duration',
    metricUptime: 'Coordinator uptime',
    hintHitsMisses: '{hits} hits · {misses} misses',
    hintBuildsRetained: '{n} build(s) retained',
    hintQueueActive: 'queue {queued} · active {active}',
    workerFleetEyebrow: 'Worker fleet',
    workerFleetTitle: 'Live heartbeat',
    noWorkers: 'No workers registered yet.',
    activityEyebrow: 'Activity',
    eventLogTitle: 'Event log',
    waitingEvents: 'Waiting for events…',
    cacheStatsEyebrow: 'Cache',
    cacheStatsTitle: 'Cache statistics',
    cacheOverallLabel: 'Overall hit rate',
    cacheFlutterLabel: 'Flutter',
    cacheUnityLabel: 'Unity',
    clusterActivityEyebrow: 'Cluster',
    clusterActivityTitle: 'Cluster activity',
    clusterActivityNote: '{peak} peak concurrent · newest {n} task(s)',
    noClusterActivity: 'No dispatched tasks yet — bars appear as builds run.',
    idle: 'idle',
    taskCount: '{n} task(s)',
  },
  builds: {
    eyebrow: 'History',
    title: 'Builds',
    colBuild: 'Build',
    colType: 'Type',
    colStatus: 'Status',
    colTasks: 'Tasks',
    colCache: 'Cache',
    colLast: 'Last activity',
    loading: 'Loading…',
    empty: 'No builds recorded yet — every task so far ran through local fallback.',
    failedSuffix: '({n} failed)',
    durationsEyebrow: 'Wall-clock span per build',
    durationsTitle: 'Build durations',
    durationsEmpty: 'No builds yet.',
  },
  buildDetail: {
    backLabel: 'Builds',
    loading: 'Loading…',
    notFound: 'Build not found.',
    metaTitle: 'Status',
    metaStatus: 'Status',
    metaType: 'Type',
    metaTasks: 'Tasks',
    metaFailed: 'Failed',
    metaCache: 'From cache',
    metaFirstTask: 'First task',
    metaLastActivity: 'Last activity',
    metaSpan: 'Span',
    outcomesTitle: 'Outcomes',
    timelineTitle: 'Timeline',
    timelineEmpty: 'No task timestamps retained for this build.',
    legendQueue: 'queue (coordinator)',
    legendSuccess: 'compile',
    legendFailed: 'failed',
    legendRunning: 'running',
    tasksTitle: '{n} compile units',
    colTask: 'Task',
    colWorker: 'Worker',
    colStatus: 'Status',
    colDuration: 'Duration',
    consoleTitle: 'Console',
    consoleSelectTask: 'Pick a task to view its console output.',
    consoleOpenFull: 'Open full console',
    stdout: 'stdout',
    stderr: 'stderr',
  },
  workers: {
    eyebrow: 'Fleet',
    title: 'Workers',
    loading: 'Loading…',
    empty: 'No workers registered.',
    cores: '{n} cores',
    tasksLabel: 'Tasks',
    successLabel: 'Success',
    latencyLabel: 'Latency',
    seenLabel: 'Seen {t}',
    slotsLabel: 'Executor slots',
    utilizationLabel: '{active}/{max} busy',
    addWorker: 'Add worker',
    dialogTitle: 'Add a worker to the fleet',
    dialogDescription:
      'Workers join by running the hg-worker binary and pointing it at this coordinator — there is no remote-provisioning step here.',
    step1Title: '1. Get the binary',
    step1Body: 'Build from source or use the pre-built image on the machine that will run the worker.',
    step2Title: '2. Start it, pointed at this coordinator',
    step2Body: 'Run this on the worker machine (same LAN or reachable network):',
    step3Title: '3. Docker Compose alternative',
    step3Body: 'If you run the bundled compose stack, just scale the existing service:',
    step4Title: '4. Verify',
    step4Body:
      'The new worker appears in this list within a few seconds once it registers (heartbeat every ~30s). If it never shows up, check --advertise-address on machines behind NAT/Docker/K8s so the coordinator can dial back into it.',
  },
  console: {
    title: 'Console output',
    back: 'Back',
    truncatedNote: 'output truncated at 256 KiB per stream',
    loading: 'Loading console…',
    notAvailable: 'Console output not retained for this task.',
    provenance: 'Output is captured once at task completion — the RPC is unary, not a live tail.',
  },
}

const vi: typeof en = {
  common: {
    live: 'Trực tuyến',
    connecting: 'Đang kết nối',
    reconnecting: 'Đang kết nối lại',
    loading: 'Đang tải…',
    empty: '(trống)',
    justNow: 'vừa xong',
    secondsAgo: '{n} giây trước',
    minutesAgo: '{n} phút trước',
    hoursAgo: '{n} giờ trước',
    daysAgo: '{n} ngày trước',
    clear: 'Xoá',
    cache: 'cache',
  },
  nav: {
    overview: 'Tổng quan',
    builds: 'Builds',
    workers: 'Workers',
    brandSubtitle: 'Điều khiển trại compile',
  },
  status: {
    STATUS_QUEUED: 'Đang chờ',
    STATUS_RUNNING: 'Đang chạy',
    STATUS_COMPLETED: 'Thành công',
    STATUS_FAILED: 'Thất bại',
    STATUS_TIMEOUT: 'Hết giờ',
  },
  circuit: {
    CLOSED: 'Đóng',
    HALF_OPEN: 'Nửa mở',
    OPEN: 'Mở',
  },
  overview: {
    eyebrow: 'Phòng điều khiển',
    title: 'Tổng quan hạm đội',
    metricActive: 'Đang chạy',
    metricQueued: 'Đang chờ',
    metricSucceeded: 'Thành công',
    metricFailed: 'Thất bại',
    metricCacheHitRate: 'Tỷ lệ cache hit',
    metricHealthyWorkers: 'Worker khoẻ mạnh',
    metricAvgBuildDuration: 'Thời gian build TB',
    metricUptime: 'Uptime coordinator',
    hintHitsMisses: '{hits} hit · {misses} miss',
    hintBuildsRetained: '{n} build đang lưu',
    hintQueueActive: 'chờ {queued} · đang chạy {active}',
    workerFleetEyebrow: 'Hạm đội worker',
    workerFleetTitle: 'Nhịp tim trực tiếp',
    noWorkers: 'Chưa có worker nào đăng ký.',
    activityEyebrow: 'Hoạt động',
    eventLogTitle: 'Nhật ký sự kiện',
    waitingEvents: 'Đang chờ sự kiện…',
    cacheStatsEyebrow: 'Cache',
    cacheStatsTitle: 'Thống kê cache',
    cacheOverallLabel: 'Tỷ lệ hit tổng',
    cacheFlutterLabel: 'Flutter',
    cacheUnityLabel: 'Unity',
    clusterActivityEyebrow: 'Cluster',
    clusterActivityTitle: 'Hoạt động cluster',
    clusterActivityNote: 'đỉnh {peak} đồng thời · {n} task mới',
    noClusterActivity: 'Chưa có task nào được phân phối — thanh sẽ xuất hiện khi build chạy.',
    idle: 'rảnh',
    taskCount: '{n} task',
  },
  builds: {
    eyebrow: 'Lịch sử',
    title: 'Builds',
    colBuild: 'Build',
    colType: 'Loại',
    colStatus: 'Trạng thái',
    colTasks: 'Tasks',
    colCache: 'Cache',
    colLast: 'Hoạt động cuối',
    loading: 'Đang tải…',
    empty: 'Chưa có build nào — mọi task đến giờ đều chạy qua local fallback.',
    failedSuffix: '({n} lỗi)',
    durationsEyebrow: 'Thời lượng thực tế mỗi build',
    durationsTitle: 'Thời gian build',
    durationsEmpty: 'Chưa có build nào.',
  },
  buildDetail: {
    backLabel: 'Builds',
    loading: 'Đang tải…',
    notFound: 'Không tìm thấy build.',
    metaTitle: 'Trạng thái',
    metaStatus: 'Trạng thái',
    metaType: 'Loại',
    metaTasks: 'Tasks',
    metaFailed: 'Lỗi',
    metaCache: 'Từ cache',
    metaFirstTask: 'Task đầu tiên',
    metaLastActivity: 'Hoạt động cuối',
    metaSpan: 'Khoảng thời gian',
    outcomesTitle: 'Kết quả',
    timelineTitle: 'Dòng thời gian',
    timelineEmpty: 'Không còn dấu thời gian nào được lưu cho build này.',
    legendQueue: 'chờ (coordinator)',
    legendSuccess: 'biên dịch',
    legendFailed: 'lỗi',
    legendRunning: 'đang chạy',
    tasksTitle: '{n} đơn vị biên dịch',
    colTask: 'Task',
    colWorker: 'Worker',
    colStatus: 'Trạng thái',
    colDuration: 'Thời gian',
    consoleTitle: 'Console',
    consoleSelectTask: 'Chọn 1 task để xem console output.',
    consoleOpenFull: 'Mở console đầy đủ',
    stdout: 'stdout',
    stderr: 'stderr',
  },
  workers: {
    eyebrow: 'Hạm đội',
    title: 'Workers',
    loading: 'Đang tải…',
    empty: 'Chưa có worker nào đăng ký.',
    cores: '{n} nhân',
    tasksLabel: 'Tasks',
    successLabel: 'Tỷ lệ thành công',
    latencyLabel: 'Độ trễ',
    seenLabel: 'Thấy lúc {t}',
    slotsLabel: 'Khe thực thi',
    utilizationLabel: '{active}/{max} đang bận',
    addWorker: 'Thêm worker',
    dialogTitle: 'Thêm worker vào hạm đội',
    dialogDescription:
      'Worker gia nhập bằng cách chạy binary hg-worker và trỏ nó tới coordinator này — không có bước cấp phát từ xa nào ở đây.',
    step1Title: '1. Lấy binary',
    step1Body: 'Build từ source hoặc dùng image dựng sẵn trên máy sẽ chạy worker.',
    step2Title: '2. Khởi động, trỏ vào coordinator này',
    step2Body: 'Chạy lệnh này trên máy worker (cùng LAN hoặc mạng có thể truy cập được):',
    step3Title: '3. Dùng Docker Compose',
    step3Body: 'Nếu bạn chạy stack compose có sẵn, chỉ cần scale service hiện có:',
    step4Title: '4. Xác nhận',
    step4Body:
      'Worker mới sẽ xuất hiện trong danh sách này sau vài giây khi nó đăng ký (heartbeat mỗi ~30s). Nếu không thấy, kiểm tra --advertise-address trên các máy sau NAT/Docker/K8s để coordinator gọi ngược lại được.',
  },
  console: {
    title: 'Console output',
    back: 'Quay lại',
    truncatedNote: 'output đã bị cắt ở 256 KiB mỗi luồng',
    loading: 'Đang tải console…',
    notAvailable: 'Console output không còn được lưu cho task này.',
    provenance: 'Output được ghi lại 1 lần khi task hoàn tất — đây là RPC unary, không phải live tail.',
  },
}

export const dictionaries = { en, vi }

export type Dict = typeof en

function getPath(obj: unknown, path: string): unknown {
  return path.split('.').reduce<unknown>((acc, key) => {
    if (acc && typeof acc === 'object' && key in acc) {
      return (acc as Record<string, unknown>)[key]
    }
    return undefined
  }, obj)
}

function interpolate(template: string, vars?: Record<string, string | number>): string {
  if (!vars) return template
  return template.replace(/\{(\w+)\}/g, (match, key: string) =>
    key in vars ? String(vars[key]) : match,
  )
}

export function translate(lang: 'en' | 'vi', key: string, vars?: Record<string, string | number>): string {
  const value = getPath(dictionaries[lang], key)
  if (typeof value !== 'string') return key
  return interpolate(value, vars)
}

export function useT() {
  const lang = useSettingsStore((state) => state.lang)
  return (key: string, vars?: Record<string, string | number>) => translate(lang, key, vars)
}

/** Locale-aware "5s ago" / "vừa xong" formatting for a Unix-seconds timestamp. */
export function useRelativeTime() {
  const t = useT()
  return (unixSeconds: number): string => {
    if (!unixSeconds) return '—'
    const deltaSeconds = Math.round((Date.now() - unixSeconds * 1000) / 1000)
    if (deltaSeconds < 5) return t('common.justNow')
    if (deltaSeconds < 60) return t('common.secondsAgo', { n: deltaSeconds })
    const minutes = Math.round(deltaSeconds / 60)
    if (minutes < 60) return t('common.minutesAgo', { n: minutes })
    const hours = Math.round(minutes / 60)
    if (hours < 24) return t('common.hoursAgo', { n: hours })
    return t('common.daysAgo', { n: Math.round(hours / 24) })
  }
}
