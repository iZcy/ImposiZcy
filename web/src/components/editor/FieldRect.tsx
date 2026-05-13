import { useRef, PointerEvent } from 'react'
import type { TemplateVariable } from '../../services/api'

interface Props {
  field: TemplateVariable
  index: number
  selected: boolean
  scale: number
  onSelect: (index: number) => void
  onChange: (index: number, next: TemplateVariable) => void
}

type DragMode = 'move' | 'resize-se' | null

export default function FieldRect({ field, index, selected, scale, onSelect, onChange }: Props) {
  const pos = field.position ?? { x: 0, y: 0, width: 100, height: 30 }
  const startRef = useRef<{ mode: DragMode; px: number; py: number; x: number; y: number; w: number; h: number } | null>(null)

  const beginDrag = (mode: DragMode) => (e: PointerEvent<HTMLDivElement>) => {
    e.stopPropagation()
    onSelect(index)
    ;(e.target as HTMLElement).setPointerCapture(e.pointerId)
    startRef.current = {
      mode,
      px: e.clientX,
      py: e.clientY,
      x: pos.x ?? 0,
      y: pos.y ?? 0,
      w: pos.width ?? 100,
      h: pos.height ?? 30,
    }
  }

  const onMove = (e: PointerEvent<HTMLDivElement>) => {
    const s = startRef.current
    if (!s) return
    const dx = (e.clientX - s.px) / scale
    const dy = (e.clientY - s.py) / scale
    if (s.mode === 'move') {
      onChange(index, { ...field, position: { ...pos, x: Math.round(s.x + dx), y: Math.round(s.y + dy) } })
    } else if (s.mode === 'resize-se') {
      onChange(index, { ...field, position: { ...pos, width: Math.max(8, Math.round(s.w + dx)), height: Math.max(8, Math.round(s.h + dy)) } })
    }
  }

  const endDrag = (e: PointerEvent<HTMLDivElement>) => {
    ;(e.target as HTMLElement).releasePointerCapture(e.pointerId)
    startRef.current = null
  }

  const label = field.type === 'barcode' ? `${field.name} (${field.barcode_format ?? 'code128'})` : field.name
  const borderColor = selected ? '#2563eb' : field.type === 'barcode' ? '#a855f7' : '#22c55e'

  return (
    <div
      style={{
        position: 'absolute',
        left: (pos.x ?? 0) * scale,
        top: (pos.y ?? 0) * scale,
        width: (pos.width ?? 100) * scale,
        height: (pos.height ?? 30) * scale,
        border: `2px solid ${borderColor}`,
        background: selected ? 'rgba(37,99,235,0.12)' : 'rgba(34,197,94,0.06)',
        cursor: 'move',
        boxSizing: 'border-box',
        touchAction: 'none',
      }}
      onPointerDown={beginDrag('move')}
      onPointerMove={onMove}
      onPointerUp={endDrag}
      onPointerCancel={endDrag}
    >
      <div style={{ fontSize: 11, color: borderColor, padding: 2, lineHeight: 1.1, pointerEvents: 'none', whiteSpace: 'nowrap', overflow: 'hidden' }}>
        {label}
      </div>
      {selected && (
        <div
          onPointerDown={beginDrag('resize-se')}
          onPointerMove={onMove}
          onPointerUp={endDrag}
          onPointerCancel={endDrag}
          style={{ position: 'absolute', right: -6, bottom: -6, width: 12, height: 12, background: '#2563eb', cursor: 'nwse-resize', touchAction: 'none' }}
        />
      )}
    </div>
  )
}
