import { useMemo } from 'react'
import { buildWavePoints, toSvgPath } from '@/lib/seededWave'
import { cn } from '@/lib/utils'

const WIDTH = 240
const HEIGHT = 40

const CIRCUIT_COLOR: Record<string, string> = {
  CLOSED: 'var(--color-circuit-closed)',
  HALF_OPEN: 'var(--color-circuit-half-open)',
  OPEN: 'var(--color-circuit-open)',
}

interface WorkerHeartbeatProps {
  workerId: string
  circuitState: string
  loadRatio: number
  className?: string
}

/**
 * The dashboard's signature element: a live oscilloscope-style trace
 * per worker instead of a static progress bar, echoing the system's
 * real heartbeat/TTL mechanic. Amplitude tracks load, color tracks the
 * circuit breaker state, and an OPEN breaker visibly flatlines.
 */
export function WorkerHeartbeat({ workerId, circuitState, loadRatio, className }: WorkerHeartbeatProps) {
  const color = CIRCUIT_COLOR[circuitState] ?? 'var(--color-ink-faint)'
  const isOpen = circuitState === 'OPEN'
  const intensity = isOpen ? 0 : Math.min(1, Math.max(0.08, loadRatio))

  const path = useMemo(() => {
    const points = buildWavePoints(workerId, 40, intensity)
    return toSvgPath(points, WIDTH, HEIGHT)
  }, [workerId, intensity])

  const duration = isOpen ? 0 : 5.5 - intensity * 3.2

  return (
    <div className={cn('relative h-10 w-full overflow-hidden', className)}>
      <svg
        viewBox={`0 0 ${WIDTH * 2} ${HEIGHT}`}
        preserveAspectRatio="none"
        className="h-full w-full"
        aria-hidden
      >
        <g
          style={
            isOpen
              ? undefined
              : {
                  animation: `hg-trace-scroll ${duration}s linear infinite`,
                }
          }
        >
          <path d={path} fill="none" stroke={color} strokeWidth={1.4} strokeLinecap="round" opacity={0.9} />
          <path
            d={path}
            fill="none"
            stroke={color}
            strokeWidth={1.4}
            strokeLinecap="round"
            opacity={0.9}
            transform={`translate(${WIDTH}, 0)`}
          />
        </g>
        {isOpen && (
          <line x1={0} y1={HEIGHT / 2} x2={WIDTH * 2} y2={HEIGHT / 2} stroke={color} strokeWidth={1.4} opacity={0.6} />
        )}
      </svg>
      <style>{`
        @keyframes hg-trace-scroll {
          from { transform: translateX(0); }
          to { transform: translateX(-${WIDTH}px); }
        }
      `}</style>
    </div>
  )
}
