import '@xyflow/react/dist/style.css'
import { useNavigate } from '@tanstack/react-router'
import {
  Background,
  BackgroundVariant,
  BaseEdge,
  EdgeLabelRenderer,
  Handle,
  Position,
  ReactFlow,
  type Edge,
  type EdgeProps,
  type Node,
  type NodeProps,
} from '@xyflow/react'
import { GitBranch, Layers, Radio } from 'lucide-react'
import { useCallback, useMemo, useState, useSyncExternalStore } from 'react'
import { useT } from '@/lib/i18n'
import { archLabel, osLabel, powerTone } from '@/lib/machine'
import {
  NODE_SIZE,
  UNGROUPED,
  layoutTopology,
  orthoPath,
  routeLength,
  type Topology,
  type TopologyWorker,
  type XY,
} from '@/lib/topology'
import { cn, formatDuration, tailEllipsis, taskLabel } from '@/lib/utils'
import { useSettingsStore } from '@/store/useSettingsStore'
import { MachineIllustration } from './MachineIllustration'

/** File chips per worker card (a 320px card fits two plus "+N"). */
const MAX_CHIPS = 2
const CHIP_CHARS = 8
const BUILD_ID_CHARS = 18
/** Dots drawn per edge; more in-flight compiles thicken the edge instead. */
const MAX_DOTS = 4
/** Dot speed along a route, so long routes to lower rows are not faster. */
const DOT_PX_PER_S = 260
const PAD = 24

type BuildNodeData = { kind: 'build' | 'ungrouped' | 'others' | 'none'; id: string; running: number; count?: number }
type CoordinatorNodeData = { running: number; queued: number; healthy: number; total: number }
type WorkerNodeData = { entry: TopologyWorker; now: number }
type FlowEdgeData = {
  points: XY[]
  active: number
  tone: 'running' | 'idle' | 'failed'
  reducedMotion: boolean
  showCount: boolean
}

type BuildFlowNode = Node<BuildNodeData, 'build'>
type CoordinatorFlowNode = Node<CoordinatorNodeData, 'coordinator'>
type WorkerFlowNode = Node<WorkerNodeData, 'worker'>
type FlowNode = BuildFlowNode | CoordinatorFlowNode | WorkerFlowNode
type FlowEdge = Edge<FlowEdgeData, 'flow'>

interface WorkerFlowProps {
  topology: Topology
  queued: number
  now: number
}

/**
 * The fleet as a live top-down flow (React Flow): running builds on top
 * feed the coordinator, which fans compiles out to a grid of workers
 * below — the grid grows downward as machines join. Edges are
 * org-chart routes through the gaps; each in-flight compile is a dot
 * travelling along its route, and worker cards list the files they are
 * compiling with a running clock.
 */
export function WorkerFlow({ topology, queued, now }: WorkerFlowProps) {
  const theme = useSettingsStore((state) => state.theme)
  const navigate = useNavigate()
  const reducedMotion = useMediaQuery('(prefers-reduced-motion: reduce)')
  const [containerRef, containerWidth] = useElementWidth<HTMLDivElement>()

  const buildNodes = useMemo<BuildNodeData[]>(() => {
    const list: BuildNodeData[] = topology.builds.map((b) => ({
      kind: b.id === UNGROUPED ? 'ungrouped' : 'build',
      id: b.id,
      running: b.running,
    }))
    if (topology.otherBuilds.builds > 0) {
      list.push({ kind: 'others', id: '__others__', running: topology.otherBuilds.running, count: topology.otherBuilds.builds })
    }
    if (list.length === 0) list.push({ kind: 'none', id: '__none__', running: 0 })
    return list
  }, [topology])

  const available = Math.max(0, containerWidth - PAD * 2)
  const layout = useMemo(
    () => layoutTopology(buildNodes.length, topology.workers.length, available),
    [buildNodes.length, topology.workers.length, available],
  )

  const { nodes, edges } = useMemo(() => {
    const healthy = topology.workers.filter((w) => w.worker?.healthy).length
    const nodes: FlowNode[] = [
      ...buildNodes.map<BuildFlowNode>((data, i) => ({
        id: `b:${data.id}`,
        type: 'build',
        position: layout.builds[i],
        data,
        style: { width: NODE_SIZE.build.w, height: NODE_SIZE.build.h },
      })),
      {
        id: 'coordinator',
        type: 'coordinator',
        position: layout.coordinator,
        data: { running: topology.running, queued, healthy, total: topology.workers.length },
        style: { width: NODE_SIZE.coordinator.w, height: NODE_SIZE.coordinator.h },
      },
      ...topology.workers.map<WorkerFlowNode>((entry, i) => ({
        id: `w:${entry.id}`,
        type: 'worker',
        position: layout.workers[i],
        data: { entry, now },
        style: { width: NODE_SIZE.worker.w, height: NODE_SIZE.worker.h },
      })),
    ]

    const edges: FlowEdge[] = [
      ...buildNodes.flatMap<FlowEdge>((b, i) =>
        b.kind === 'none'
          ? []
          : [
              {
                id: `e:b:${b.id}`,
                source: `b:${b.id}`,
                target: 'coordinator',
                type: 'flow',
                data: {
                  points: layout.routes.builds[i],
                  active: b.running,
                  tone: b.running > 0 ? 'running' : 'idle',
                  reducedMotion,
                  showCount: true,
                },
              },
            ],
      ),
      ...topology.workers.map<FlowEdge>((entry, i) => {
        const unhealthy = entry.worker ? powerTone(entry.worker) === 'failed' : true
        return {
          id: `e:w:${entry.id}`,
          source: 'coordinator',
          target: `w:${entry.id}`,
          type: 'flow',
          data: {
            points: layout.routes.workers[i],
            active: entry.running.length,
            tone: unhealthy ? 'failed' : entry.running.length > 0 ? 'running' : 'idle',
            reducedMotion,
            // The card already shows busy/max; counts on many routes that
            // share the trunk would only pile up on top of each other.
            showCount: false,
          },
        }
      }),
    ]
    return { nodes, edges }
  }, [buildNodes, layout, topology, queued, now, reducedMotion])

  // Shown at 1:1 like page content; only a map wider than the panel
  // (e.g. many builds on a phone) is scaled down, and the canvas height
  // follows that scale so no empty band is left.
  const scale = available > 0 ? Math.min(1, available / layout.width) : 1
  const height = Math.round(layout.height * scale + PAD * 2)
  // Remount (and so re-fit) only when the graph's shape changes, not on every tick.
  const shapeKey = `${layout.cols}|${scale.toFixed(2)}|${buildNodes.map((b) => b.id).join(',')}|${topology.workers.map((w) => w.id).join(',')}`

  return (
    <div ref={containerRef} className="worker-flow" style={{ height }}>
      {containerWidth > 0 && (
        <ReactFlow<FlowNode, FlowEdge>
          key={shapeKey}
          nodes={nodes}
          edges={edges}
          nodeTypes={nodeTypes}
          edgeTypes={edgeTypes}
          colorMode={theme}
          fitView
          fitViewOptions={{ padding: PAD, minZoom: scale, maxZoom: scale }}
          minZoom={scale}
          maxZoom={scale}
          nodesDraggable={false}
          nodesConnectable={false}
          elementsSelectable={false}
          panOnDrag={false}
          zoomOnScroll={false}
          zoomOnPinch={false}
          zoomOnDoubleClick={false}
          preventScrolling={false}
          onNodeClick={(_, node) => {
            if (node.id.startsWith('w:')) navigate({ to: '/workers/$workerId', params: { workerId: node.id.slice(2) } })
            else if (node.type === 'build' && (node.data as BuildNodeData).kind === 'build') {
              navigate({ to: '/builds/$buildId', params: { buildId: (node.data as BuildNodeData).id } })
            }
          }}
        >
          <Background variant={BackgroundVariant.Dots} gap={18} size={1} />
        </ReactFlow>
      )}
    </div>
  )
}

/** Width of an element, kept current with a ResizeObserver. */
function useElementWidth<T extends HTMLElement>(): [(node: T | null) => void, number] {
  const [width, setWidth] = useState(0)
  const ref = useCallback((node: T | null) => {
    if (!node) return
    setWidth(Math.round(node.getBoundingClientRect().width))
    const observer = new ResizeObserver(([entry]) => setWidth(Math.round(entry.contentRect.width)))
    observer.observe(node)
    return () => observer.disconnect()
  }, [])
  return [ref, width]
}

function useMediaQuery(query: string): boolean {
  return useSyncExternalStore(
    (onChange) => {
      const list = window.matchMedia(query)
      list.addEventListener('change', onChange)
      return () => list.removeEventListener('change', onChange)
    },
    () => window.matchMedia(query).matches,
    () => false,
  )
}

const hiddenHandle = 'pointer-events-none !h-1.5 !w-1.5 !min-w-0 !border-0 !bg-(--color-hairline-strong)'

function BuildNode({ data }: NodeProps<BuildFlowNode>) {
  const t = useT()
  const active = data.running > 0
  const title =
    data.kind === 'build'
      ? tailEllipsis(data.id, BUILD_ID_CHARS)
      : data.kind === 'ungrouped'
        ? t('workers.flowUngrouped')
        : data.kind === 'others'
          ? t('workers.flowOtherBuilds', { n: data.count ?? 0 })
          : t('workers.flowNoBuild')
  const Icon = data.kind === 'others' || data.kind === 'ungrouped' ? Layers : GitBranch
  return (
    <div
      className={cn(
        'flex h-full items-center gap-2.5 rounded-(--radius-lg) border bg-(--color-surface) px-3',
        active ? 'border-[color-mix(in_oklch,var(--color-status-running)_40%,transparent)]' : 'border-(--color-hairline)',
        data.kind === 'build' && 'cursor-pointer hover:border-(--color-status-running)',
        data.kind === 'none' && 'border-dashed text-(--color-ink-muted)',
      )}
      title={data.kind === 'build' ? data.id : undefined}
    >
      <Icon className={cn('h-4 w-4 shrink-0', active ? 'text-(--color-status-running)' : 'text-(--color-ink-faint)')} />
      <div className="min-w-0">
        <div className={cn('text-[12px] font-medium', data.kind === 'none' ? 'line-clamp-2' : 'truncate', data.kind === 'build' && 'font-mono')}>
          {title}
        </div>
        {data.kind !== 'none' && (
          <div className="font-tabular text-[11px] text-(--color-ink-muted)">{t('workers.flowRunning', { n: data.running })}</div>
        )}
      </div>
      <Handle type="source" position={Position.Bottom} className={hiddenHandle} isConnectable={false} />
    </div>
  )
}

function CoordinatorNode({ data }: NodeProps<CoordinatorFlowNode>) {
  const t = useT()
  const active = data.running > 0
  return (
    <div className="flex h-full flex-col justify-between rounded-(--radius-lg) border border-(--color-hairline-strong) bg-(--color-surface) p-3.5 shadow-(--shadow-popover)">
      <Handle type="target" position={Position.Top} className={hiddenHandle} isConnectable={false} />
      <div className="flex items-center gap-2">
        <span className="relative flex h-7 w-7 items-center justify-center rounded-(--radius-md) bg-(--color-signal) text-(--color-on-signal)">
          <Radio className="h-3.5 w-3.5" strokeWidth={2.25} />
          {active && (
            <span className="absolute inset-0 rounded-(--radius-md) border-2 border-(--color-status-running) animate-(--animate-pulse-ring)" />
          )}
        </span>
        <span className="text-[13px] font-semibold text-(--color-ink)">{t('workers.flowCoordinator')}</span>
      </div>
      <div>
        <div
          className="font-tabular text-[26px] font-semibold leading-none"
          style={{ color: active ? 'var(--color-status-running)' : 'var(--color-ink)' }}
        >
          {data.running}
        </div>
        <div className="mt-1 text-[12px] text-(--color-ink-muted)">{t('workers.flowRunningLabel')}</div>
      </div>
      <div className="flex flex-wrap items-center justify-between gap-x-2 font-tabular text-[11px] text-(--color-ink-muted)">
        <span className={data.queued > 0 ? 'text-(--color-status-timeout)' : undefined}>{t('workers.flowQueued', { n: data.queued })}</span>
        <span>{t('workers.flowHealthy', { ok: data.healthy, total: data.total })}</span>
      </div>
      <Handle type="source" position={Position.Bottom} className={hiddenHandle} isConnectable={false} />
    </div>
  )
}

function WorkerNode({ data }: NodeProps<WorkerFlowNode>) {
  const t = useT()
  const { entry, now } = data
  const worker = entry.worker
  const busy = entry.running.length
  const max = worker?.max_parallel_tasks ?? busy
  const power = worker ? powerTone(worker) : 'failed'
  const platform = worker ? [osLabel(worker.os), archLabel(worker.architecture)].filter(Boolean).join(' · ') : t('workers.flowOffline')
  const shown = entry.running.slice(0, MAX_CHIPS)

  return (
    <div
      className={cn(
        'flex h-full cursor-pointer flex-col gap-2 rounded-(--radius-lg) border bg-(--color-surface) p-2.5 transition-colors',
        power === 'failed'
          ? 'border-[color-mix(in_oklch,var(--color-status-failed)_45%,transparent)]'
          : busy > 0
            ? 'border-[color-mix(in_oklch,var(--color-status-running)_40%,transparent)] hover:border-(--color-status-running)'
            : 'border-(--color-hairline) hover:border-(--color-hairline-strong)',
      )}
      title={entry.id}
    >
      <Handle type="target" position={Position.Top} className={hiddenHandle} isConnectable={false} />
      <div className="flex items-center gap-2.5">
        <MachineIllustration os={worker?.os ?? ''} slots={max} busy={busy} power={power} caption="" label={entry.name} className="w-[52px] shrink-0" />
        <div className="min-w-0 flex-1">
          <div className="truncate font-mono text-[12px] font-semibold text-(--color-ink)">{entry.name}</div>
          <div className="truncate text-[11px] text-(--color-ink-muted)">{platform}</div>
          <SlotBar busy={busy} max={max} />
        </div>
        <span className="self-start font-tabular text-[12px] font-medium text-(--color-ink-secondary)">
          {busy}/{max}
        </span>
      </div>
      <div className="flex min-w-0 items-center gap-1 overflow-hidden">
        {shown.map((task) => (
          <span
            key={task.id}
            title={taskLabel(task)}
            className="dispatch-enter flex shrink-0 items-center gap-1 rounded-full border border-[color-mix(in_oklch,var(--color-status-running)_35%,transparent)] px-1.5 py-0.5 font-mono text-[10px] text-(--color-ink)"
          >
            <span className="h-1.5 w-1.5 rounded-full bg-(--color-status-running) animate-(--animate-heartbeat)" />
            {tailEllipsis(taskLabel(task), CHIP_CHARS)}
            {task.started_at_ms > 0 && (
              <span className="font-tabular text-(--color-status-running)">{formatDuration(Math.max(0, now - task.started_at_ms))}</span>
            )}
          </span>
        ))}
        {busy > shown.length && (
          <span className="shrink-0 font-tabular text-[10px] text-(--color-status-running)">{t('workers.flowMore', { n: busy - shown.length })}</span>
        )}
        {busy === 0 && <span className="text-[11px] text-(--color-ink-faint)">{t('workers.flowIdleWorker')}</span>}
      </div>
    </div>
  )
}

function SlotBar({ busy, max }: { busy: number; max: number }) {
  const cells = Math.max(max, busy, 1)
  return (
    <div className="mt-1 flex gap-0.5">
      {Array.from({ length: Math.min(cells, 32) }, (_, i) => (
        <span
          key={i}
          className={cn('h-1.5 flex-1 rounded-full', i < busy ? 'bg-(--color-status-running)' : 'bg-(--color-surface-inset)')}
        />
      ))}
    </div>
  )
}

const EDGE_COLOR: Record<FlowEdgeData['tone'], string> = {
  running: 'var(--color-status-running)',
  idle: 'var(--color-hairline-strong)',
  failed: 'var(--color-status-failed)',
}

/**
 * An org-chart edge drawn from the route the layout computed (React
 * Flow's handle positions are not used), with in-flight compiles as dots
 * moving along it at a constant speed.
 */
function FlowEdgeView({ id, data }: EdgeProps<FlowEdge>) {
  const points = data?.points ?? []
  const path = orthoPath(points)
  const active = data?.active ?? 0
  const tone = data?.tone ?? 'idle'
  const color = EDGE_COLOR[tone]
  const dots = data?.reducedMotion ? 0 : Math.min(active, MAX_DOTS)
  const duration = Math.max(0.8, routeLength(points) / DOT_PX_PER_S)
  // Count sits on the edge's own first leg (a build's drop line), before
  // routes merge onto a shared bus where labels would stack on each other.
  const label = points.length >= 2 ? { x: (points[0].x + points[1].x) / 2, y: (points[0].y + points[1].y) / 2 } : undefined

  return (
    <>
      <BaseEdge
        id={id}
        path={path}
        style={{
          stroke: color,
          strokeWidth: tone === 'running' ? Math.min(1.5 + active * 0.35, 3.5) : 1.25,
          strokeDasharray: tone === 'failed' ? '5 5' : undefined,
          opacity: tone === 'idle' ? 0.8 : 1,
        }}
      />
      {Array.from({ length: dots }, (_, i) => (
        // Keyed by the dot count: when it changes the circles are recreated.
        // Updating `begin` on a live <animateMotion> does not restart it in
        // Chrome, which left a dot parked at the canvas origin. The negative
        // begin starts each dot part-way along the route.
        <circle key={`${dots}:${i}`} r="3.5" fill={color}>
          <animateMotion dur={`${duration}s`} begin={`-${(i * duration) / dots}s`} repeatCount="indefinite" path={path} />
        </circle>
      ))}
      {active > 0 && data?.showCount && label && (
        <EdgeLabelRenderer>
          <div
            className="pointer-events-none absolute rounded-full border border-(--color-hairline) bg-(--color-surface) px-1.5 font-tabular text-[10px] font-medium"
            style={{ transform: `translate(-50%, -50%) translate(${label.x}px, ${label.y}px)`, color }}
          >
            {active}
          </div>
        </EdgeLabelRenderer>
      )}
    </>
  )
}

const nodeTypes = { build: BuildNode, coordinator: CoordinatorNode, worker: WorkerNode }
const edgeTypes = { flow: FlowEdgeView }
