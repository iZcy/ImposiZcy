import type { Font, TemplateVariable } from '../../services/api'

interface Props {
  field: TemplateVariable | null
  index: number
  fonts: Font[]
  onChange: (index: number, next: TemplateVariable) => void
  onDelete: (index: number) => void
}

const labelCls = 'block text-xs font-medium text-gray-600 mb-1'
const inputCls = 'w-full px-2 py-1 text-sm border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500'

export default function FieldPropertiesPanel({ field, index, fonts, onChange, onDelete }: Props) {
  if (!field) {
    return (
      <div className="p-4 text-sm text-gray-500">
        Select a field on the canvas to edit its properties, or click <span className="font-medium">Add Field</span> on the toolbar.
      </div>
    )
  }
  const set = <K extends keyof TemplateVariable>(k: K, v: TemplateVariable[K]) =>
    onChange(index, { ...field, [k]: v })
  const setPos = (patch: Partial<NonNullable<TemplateVariable['position']>>) =>
    onChange(index, { ...field, position: { ...(field.position ?? {}), ...patch } })

  const pos = field.position ?? {}

  return (
    <div className="p-4 space-y-3 overflow-y-auto">
      <div className="flex justify-between items-center">
        <h3 className="font-semibold text-sm">Field properties</h3>
        <button onClick={() => onDelete(index)} className="text-xs text-red-600 hover:underline">Delete</button>
      </div>

      <div>
        <label className={labelCls}>Name</label>
        <input className={inputCls} value={field.name} onChange={e => set('name', e.target.value)} />
      </div>
      <div>
        <label className={labelCls}>Type</label>
        <select className={inputCls} value={field.type} onChange={e => set('type', e.target.value as TemplateVariable['type'])}>
          <option value="text">Text</option>
          <option value="barcode">Barcode</option>
          <option value="image">Image</option>
        </select>
      </div>
      {field.type === 'barcode' && (
        <div>
          <label className={labelCls}>Barcode format</label>
          <select className={inputCls} value={field.barcode_format ?? 'code128'} onChange={e => set('barcode_format', e.target.value as TemplateVariable['barcode_format'])}>
            <option value="code128">Code 128</option>
            <option value="qr">QR Code</option>
            <option value="ean13">EAN-13</option>
            <option value="code39">Code 39</option>
          </select>
        </div>
      )}

      <div className="grid grid-cols-2 gap-2">
        <div><label className={labelCls}>X</label><input className={inputCls} type="number" value={pos.x ?? 0} onChange={e => setPos({ x: Number(e.target.value) })} /></div>
        <div><label className={labelCls}>Y</label><input className={inputCls} type="number" value={pos.y ?? 0} onChange={e => setPos({ y: Number(e.target.value) })} /></div>
        <div><label className={labelCls}>Width</label><input className={inputCls} type="number" value={pos.width ?? 100} onChange={e => setPos({ width: Number(e.target.value) })} /></div>
        <div><label className={labelCls}>Height</label><input className={inputCls} type="number" value={pos.height ?? 30} onChange={e => setPos({ height: Number(e.target.value) })} /></div>
      </div>

      {field.type === 'text' && (
        <>
          <div>
            <label className={labelCls}>Font family</label>
            <select className={inputCls} value={field.font_family ?? ''} onChange={e => set('font_family', e.target.value)}>
              <option value="">(template default / built-in fallback)</option>
              {fonts.map(f => <option key={f.id} value={f.family}>{f.family}</option>)}
            </select>
          </div>
          <div className="grid grid-cols-2 gap-2">
            <div><label className={labelCls}>Font size</label><input className={inputCls} type="number" value={pos.font_size ?? 14} onChange={e => setPos({ font_size: Number(e.target.value) })} /></div>
            <div>
              <label className={labelCls}>Color</label>
              <input className={inputCls} type="color" value={pos.color ?? '#000000'} onChange={e => setPos({ color: e.target.value })} />
            </div>
          </div>
          <div>
            <label className={labelCls}>Alignment</label>
            <select className={inputCls} value={pos.alignment ?? 'left'} onChange={e => setPos({ alignment: e.target.value as 'left' | 'center' | 'right' })}>
              <option value="left">Left</option>
              <option value="center">Center</option>
              <option value="right">Right</option>
            </select>
          </div>
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={!!pos.bold} onChange={e => setPos({ bold: e.target.checked })} />
            Bold
          </label>
        </>
      )}

      <div className="grid grid-cols-2 gap-2">
        <div><label className={labelCls}>Source field</label><input className={inputCls} value={field.source_field ?? ''} onChange={e => set('source_field', e.target.value)} placeholder="ticket.code" /></div>
        <div><label className={labelCls}>Order index</label><input className={inputCls} type="number" value={field.order_index ?? 0} onChange={e => set('order_index', Number(e.target.value))} /></div>
      </div>
      <div>
        <label className={labelCls}>Default value</label>
        <input className={inputCls} value={field.default_value ?? ''} onChange={e => set('default_value', e.target.value)} />
      </div>
    </div>
  )
}
