import { useId } from 'react'
import { MAX_DRAWN_SLOTS, machineKind, rackUnits, slotGrid, type PowerTone, type SlotCell } from '@/lib/machine'
import { cn } from '@/lib/utils'

const POWER_COLOR: Record<PowerTone, string> = {
  success: 'var(--color-status-success)',
  warning: 'var(--color-status-timeout)',
  failed: 'var(--color-status-failed)',
}

const BODY = 'var(--color-machine-body)'
const DECK = 'var(--color-machine-deck)'
const EDGE = 'var(--color-machine-edge)'
const SCREEN = 'var(--color-machine-screen)'
const BUSY = 'var(--color-status-running)'

interface MachineIllustrationProps {
  os: string
  /** Executor slots the worker advertises, and how many are busy. */
  slots: number
  busy: number
  power: PowerTone
  /** Short status line rendered on the screen, e.g. "3/12 busy". */
  caption: string
  /** Accessible description of the whole drawing. */
  label: string
  className?: string
}

/**
 * A worker drawn as the machine it is — laptop (macOS), desktop
 * (Windows) or rack server (everything else) — with its executor slots
 * as lit cells on the screen or LEDs on the rack, and a power LED for
 * health. Pure SVG on theme tokens, so it follows light/dark.
 */
export function MachineIllustration({ os, slots, busy, power, caption, label, className }: MachineIllustrationProps) {
  const kind = machineKind(os)
  const dimmed = power === 'failed'
  const overflow = Math.max(slots, busy) - MAX_DRAWN_SLOTS

  return (
    <svg viewBox="0 0 240 150" role="img" aria-label={label} className={cn('h-auto w-full', className)}>
      <ellipse cx="120" cy="138" rx="96" ry="5" fill="var(--color-ink)" opacity="0.06" />
      {kind === 'laptop' && <Laptop slots={slots} busy={busy} power={power} caption={caption} dimmed={dimmed} overflow={overflow} />}
      {kind === 'desktop' && <Desktop slots={slots} busy={busy} power={power} caption={caption} dimmed={dimmed} overflow={overflow} />}
      {kind === 'server' && <Rack slots={slots} busy={busy} power={power} dimmed={dimmed} overflow={overflow} />}
    </svg>
  )
}

interface PartProps {
  slots: number
  busy: number
  power: PowerTone
  dimmed: boolean
  overflow: number
}

/** Screen glass with the slot grid and a one-line terminal-style status. */
function Screen({
  x,
  y,
  w,
  h,
  slots,
  busy,
  power,
  caption,
  dimmed,
  overflow,
}: PartProps & { x: number; y: number; w: number; h: number; caption: string }) {
  const glowId = useId()
  const cells = slotGrid(slots, busy, { x: x + 8, y: y + 8, w: w - 16, h: h - 24 })
  return (
    <g>
      <rect x={x} y={y} width={w} height={h} rx="4" fill={SCREEN} />
      <defs>
        <radialGradient id={glowId}>
          <stop offset="0%" stopColor={BUSY} stopOpacity="0.55" />
          <stop offset="100%" stopColor={BUSY} stopOpacity="0" />
        </radialGradient>
      </defs>
      <g opacity={dimmed ? 0.35 : 1}>
        {cells.map((cell, i) => (
          <ScreenCell key={i} cell={cell} glowId={glowId} />
        ))}
      </g>
      <circle cx={x + 10} cy={y + h - 8} r="2" fill={POWER_COLOR[power]} />
      <text
        x={x + 16}
        y={y + h - 5.6}
        fill="var(--color-machine-screen-ink)"
        fontFamily="var(--font-mono)"
        fontSize="7"
        letterSpacing="0.2"
      >
        {caption}
        {overflow > 0 ? ` · +${overflow}` : ''}
      </text>
    </g>
  )
}

function ScreenCell({ cell, glowId }: { cell: SlotCell; glowId: string }) {
  const r = Math.min(2, cell.size * 0.22)
  if (!cell.busy) {
    return <rect x={cell.x} y={cell.y} width={cell.size} height={cell.size} rx={r} fill="var(--color-machine-screen-cell)" />
  }
  const pad = cell.size * 0.6
  return (
    <g className="machine-led-busy">
      <rect x={cell.x - pad} y={cell.y - pad} width={cell.size + pad * 2} height={cell.size + pad * 2} fill={`url(#${glowId})`} />
      <rect x={cell.x} y={cell.y} width={cell.size} height={cell.size} rx={r} fill={BUSY} />
    </g>
  )
}

function Laptop({ caption, ...props }: PartProps & { caption: string }) {
  return (
    <g>
      {/* Lid */}
      <rect x="50" y="10" width="140" height="96" rx="9" fill={BODY} stroke={EDGE} strokeWidth="1.25" />
      <circle cx="120" cy="13.8" r="1.1" fill={EDGE} />
      <Screen x={57} y={18} w={126} h={81} caption={caption} {...props} />
      {/* Deck */}
      <path d="M36 106 H204 L219 119 Q221 125 214 125 H26 Q19 125 21 119 Z" fill={DECK} stroke={EDGE} strokeWidth="1.25" strokeLinejoin="round" />
      <rect x="104" y="106" width="32" height="3.2" rx="1.6" fill={EDGE} opacity="0.7" />
    </g>
  )
}

function Desktop({ caption, ...props }: PartProps & { caption: string }) {
  return (
    <g>
      {/* Stand */}
      <path d="M108 108 H132 L137 127 H103 Z" fill={DECK} stroke={EDGE} strokeWidth="1.25" strokeLinejoin="round" />
      <rect x="80" y="126" width="80" height="6" rx="3" fill={DECK} stroke={EDGE} strokeWidth="1.25" />
      {/* Monitor */}
      <rect x="36" y="6" width="168" height="104" rx="8" fill={BODY} stroke={EDGE} strokeWidth="1.25" />
      <Screen x={43} y={13} w={154} h={84} caption={caption} {...props} />
      <circle cx="194" cy="103.5" r="1.6" fill={POWER_COLOR[props.power]} />
    </g>
  )
}

const RACK_BAYS = 4

function Rack({ slots, busy, power, dimmed, overflow }: PartProps) {
  const units = rackUnits(Math.max(slots, busy))
  const top = 14
  const unitH = 24
  const gap = 4
  // Index of each unit's first slot, so busy slots fill units front to back.
  const firstSlot = units.map((_, i) => units.slice(0, i).reduce((sum, n) => sum + n, 0))

  return (
    <g>
      {/* Cabinet */}
      <rect x="54" y="4" width="132" height="130" rx="7" fill={DECK} stroke={EDGE} strokeWidth="1.25" />
      <line x1="62" y1="10" x2="62" y2="128" stroke={EDGE} strokeWidth="1" opacity="0.6" />
      <line x1="178" y1="10" x2="178" y2="128" stroke={EDGE} strokeWidth="1" opacity="0.6" />

      {Array.from({ length: RACK_BAYS }, (_, bay) => {
        const y = top + bay * (unitH + gap)
        const count = units[bay]
        if (count === undefined) {
          // Empty bay: a blank plate so the cabinet reads as a rack.
          return (
            <rect key={bay} x="66" y={y} width="108" height={unitH} rx="3" fill="none" stroke={EDGE} strokeWidth="1" strokeDasharray="3 3" opacity="0.6" />
          )
        }
        const first = firstSlot[bay]
        return (
          <RackUnit
            key={bay}
            y={y}
            h={unitH}
            count={count}
            busy={Math.max(0, Math.min(count, busy - first))}
            power={bay === 0 ? power : undefined}
            dimmed={dimmed}
          />
        )
      })}

      {overflow > 0 && (
        <text x="120" y="130.5" textAnchor="middle" fill="var(--color-ink-muted)" fontFamily="var(--font-mono)" fontSize="6.5">
          +{overflow}
        </text>
      )}
    </g>
  )
}

function RackUnit({
  y,
  h,
  count,
  busy,
  power,
  dimmed,
}: {
  y: number
  h: number
  count: number
  busy: number
  power?: PowerTone
  dimmed: boolean
}) {
  const ledArea = { x: 104, w: 64 }
  const pitch = count > 0 ? Math.min(8, ledArea.w / count) : 8
  const size = Math.min(4.5, pitch * 0.68)
  const startX = ledArea.x + ledArea.w - count * pitch + (pitch - size) / 2
  const cy = y + h / 2

  return (
    <g>
      <rect x="66" y={y} width="108" height={h} rx="3" fill="var(--color-machine-body)" stroke={EDGE} strokeWidth="1" />
      {/* Drive bays */}
      <rect x="71" y={y + 5} width="12" height={h - 10} rx="1.5" fill={SCREEN} opacity="0.85" />
      <rect x="85" y={y + 5} width="12" height={h - 10} rx="1.5" fill={SCREEN} opacity="0.85" />
      {/* Power LED on the first unit */}
      {power && <circle cx="77" cy={y + h - 3} r="1.4" fill={POWER_COLOR[power]} />}
      {/* Slot LEDs */}
      <g opacity={dimmed ? 0.4 : 1}>
        {Array.from({ length: count }, (_, i) => {
          const isBusy = i < busy
          return (
            <rect
              key={i}
              x={startX + i * pitch}
              y={cy - size / 2}
              width={size}
              height={size}
              rx={size / 2}
              fill={isBusy ? BUSY : 'var(--color-machine-led-off)'}
              className={isBusy ? 'machine-led-busy' : undefined}
            />
          )
        })}
      </g>
    </g>
  )
}
