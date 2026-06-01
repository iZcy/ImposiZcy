import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { templatesApi, fontsApi, type Template, type TemplateVariable, type Font } from '../../services/api'
import FieldRect from './FieldRect'
import FieldPropertiesPanel from './FieldPropertiesPanel'
import FontManagerModal from './FontManagerModal'

const newField = (i: number): TemplateVariable => ({
  name: `field_${i + 1}`,
  type: 'text',
  position: { x: 50, y: 50 + i * 40, width: 200, height: 32, font_size: 16, alignment: 'left', color: '#000000' },
  order_index: i,
})

export default function TemplateCanvasEditor() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [template, setTemplate] = useState<Template | null>(null)
  const [fields, setFields] = useState<TemplateVariable[]>([])
  const [selected, setSelected] = useState<number>(-1)
  const [fonts, setFonts] = useState<Font[]>([])
  const [bgUrl, setBgUrl] = useState<string>('')
  const [naturalSize, setNaturalSize] = useState<{ w: number; h: number }>({ w: 0, h: 0 })
  const [containerWidth, setContainerWidth] = useState(800)
  const [fontModal, setFontModal] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const canvasWrapRef = useRef<HTMLDivElement | null>(null)
  const bgInputRef = useRef<HTMLInputElement | null>(null)

  useEffect(() => {
    if (!id) return
    ;(async () => {
      try {
        const [t, f] = await Promise.all([templatesApi.get(id), fontsApi.list()])
        setTemplate(t.data)
        setFields(t.data.variables ?? [])
        setFonts(f.data)
        if (t.data.background_image) setBgUrl('/uploads/' + t.data.background_image)
      } catch (e: any) {
        setError(e.message ?? String(e))
      }
    })()
  }, [id])

  useEffect(() => {
    const onResize = () => {
      if (canvasWrapRef.current) setContainerWidth(canvasWrapRef.current.clientWidth)
    }
    onResize()
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  const scale = useMemo(() => {
    if (!naturalSize.w) return 1
    const max = Math.max(400, containerWidth - 24)
    return Math.min(1, max / naturalSize.w)
  }, [naturalSize.w, containerWidth])

  const handleField = (index: number, next: TemplateVariable) => {
    setFields(prev => prev.map((v, i) => (i === index ? next : v)))
  }

  const addField = () => {
    setFields(prev => {
      const next = [...prev, newField(prev.length)]
      setSelected(next.length - 1)
      return next
    })
  }

  const deleteField = (index: number) => {
    setFields(prev => prev.filter((_, i) => i !== index))
    setSelected(-1)
  }

  const uploadBackground = async (file: File) => {
    setError(null)
    const fd = new FormData()
    fd.append('file', file)
    const res = await fetch('/api/v1/upload/background', { method: 'POST', body: fd })
    const data = await res.json()
    if (!res.ok || data.success === false) {
      setError(data.error ?? 'Upload failed')
      return
    }
    const rel = data.data.path as string
    setBgUrl('/uploads/' + rel)
    if (template) setTemplate({ ...template, background_image: rel, render_engine: 'native' })
  }

  const save = async () => {
    if (!template || !id) return
    setSaving(true)
    setError(null)
    try {
      await templatesApi.update(id, {
        variables: fields,
        background_image: template.background_image ?? '',
        render_engine: 'native',
        default_font: template.default_font ?? '',
        width: naturalSize.w || template.width,
        height: naturalSize.h || template.height,
      })
    } catch (e: any) {
      setError(e.message ?? String(e))
    } finally {
      setSaving(false)
    }
  }

  if (error) return <div className="p-6 text-red-600">{error}</div>
  if (!template) return <div className="p-6">Loading…</div>

  return (
    <div className="flex h-[calc(100vh-4rem)]">
      <div className="flex-1 flex flex-col">
        <div className="flex items-center gap-2 p-3 border-b bg-white">
          <button onClick={() => navigate(-1)} className="text-sm text-gray-600 hover:underline">← Back</button>
          <h1 className="font-semibold flex-1 truncate">{template.name} — Layout editor</h1>
          <button onClick={() => bgInputRef.current?.click()} className="text-sm px-3 py-1 border rounded hover:bg-gray-50">
            {bgUrl ? 'Replace background' : 'Upload background'}
          </button>
          <input ref={bgInputRef} type="file" accept="image/png,image/jpeg" className="hidden"
            onChange={e => { const f = e.target.files?.[0]; if (f) uploadBackground(f) }} />
          <button onClick={addField} className="text-sm bg-green-600 hover:bg-green-700 text-white px-3 py-1 rounded">+ Add field</button>
          <button onClick={() => setFontModal(true)} className="text-sm px-3 py-1 border rounded hover:bg-gray-50">Fonts</button>
          <button onClick={save} disabled={saving} className="text-sm bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white px-3 py-1 rounded">
            {saving ? 'Saving…' : 'Save'}
          </button>
        </div>
        <div ref={canvasWrapRef} className="flex-1 overflow-auto bg-gray-100 p-3">
          {bgUrl ? (
            <div
              style={{
                position: 'relative',
                width: naturalSize.w * scale,
                height: naturalSize.h * scale,
                background: '#fff',
                margin: '0 auto',
                boxShadow: '0 2px 8px rgba(0,0,0,0.15)',
              }}
              onClick={() => setSelected(-1)}
            >
              <img
                src={bgUrl}
                alt=""
                draggable={false}
                onLoad={e => {
                  const img = e.target as HTMLImageElement
                  setNaturalSize({ w: img.naturalWidth, h: img.naturalHeight })
                }}
                style={{ width: '100%', height: '100%', userSelect: 'none', pointerEvents: 'none' }}
              />
              {fields.map((f, i) => (
                <FieldRect key={i} field={f} index={i} selected={i === selected} scale={scale}
                  onSelect={setSelected} onChange={handleField} />
              ))}
            </div>
          ) : (
            <div className="text-center text-gray-500 mt-20">
              No background uploaded. Click <span className="font-medium">Upload background</span> to start.
            </div>
          )}
        </div>
      </div>
      <div className="w-80 border-l bg-white overflow-y-auto">
        <FieldPropertiesPanel
          field={selected >= 0 ? fields[selected] : null}
          index={selected}
          fonts={fonts}
          onChange={handleField}
          onDelete={deleteField}
        />
      </div>
      <FontManagerModal open={fontModal} onClose={() => setFontModal(false)} onChanged={async () => setFonts((await fontsApi.list()).data)} />
    </div>
  )
}
