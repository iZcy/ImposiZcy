import { useState, useEffect, useRef } from 'react'
import { fontsApi, type Font } from '../../services/api'

interface Props {
  open: boolean
  onClose: () => void
  onChanged?: () => void
}

export default function FontManagerModal({ open, onClose, onChanged }: Props) {
  const [fonts, setFonts] = useState<Font[]>([])
  const [family, setFamily] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const fileRef = useRef<HTMLInputElement>(null)

  const load = async () => {
    try {
      const res = await fontsApi.list()
      setFonts(res.data)
    } catch (e: any) {
      setError(e.message ?? String(e))
    }
  }

  useEffect(() => {
    if (open) load()
  }, [open])

  const upload = async () => {
    const file = fileRef.current?.files?.[0]
    if (!file || !family.trim()) {
      setError('Select a .ttf/.otf file and enter a family name')
      return
    }
    setBusy(true)
    setError(null)
    try {
      await fontsApi.upload(file, family.trim())
      setFamily('')
      if (fileRef.current) fileRef.current.value = ''
      await load()
      onChanged?.()
    } catch (e: any) {
      setError(e.message ?? String(e))
    } finally {
      setBusy(false)
    }
  }

  const remove = async (id: string) => {
    if (!confirm('Delete font?')) return
    try {
      await fontsApi.delete(id)
      await load()
      onChanged?.()
    } catch (e: any) {
      setError(e.message ?? String(e))
    }
  }

  if (!open) return null

  return (
    <div className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded shadow-lg w-full max-w-lg p-4">
        <div className="flex justify-between items-center mb-3">
          <h2 className="text-lg font-semibold">Fonts</h2>
          <button onClick={onClose} className="text-gray-500 hover:text-gray-800">✕</button>
        </div>
        {error && <div className="mb-2 text-sm text-red-600">{error}</div>}

        <div className="border rounded p-3 mb-3 space-y-2">
          <div className="text-sm font-medium">Upload new font</div>
          <input ref={fileRef} type="file" accept=".ttf,.otf" className="text-sm" />
          <input className="w-full px-2 py-1 text-sm border rounded" placeholder="Family name (e.g. Roboto)" value={family} onChange={e => setFamily(e.target.value)} />
          <button disabled={busy} onClick={upload} className="bg-blue-600 hover:bg-blue-700 text-white text-sm px-3 py-1 rounded disabled:opacity-50">
            {busy ? 'Uploading…' : 'Upload'}
          </button>
        </div>

        <div className="max-h-64 overflow-y-auto border rounded">
          <table className="w-full text-sm">
            <thead className="bg-gray-50">
              <tr><th className="text-left p-2">Family</th><th className="text-left p-2">Format</th><th className="text-left p-2">Scope</th><th /></tr>
            </thead>
            <tbody>
              {fonts.map(f => (
                <tr key={f.id} className="border-t">
                  <td className="p-2 font-medium">{f.family}</td>
                  <td className="p-2 uppercase">{f.format}</td>
                  <td className="p-2">{f.scope}</td>
                  <td className="p-2 text-right">
                    <button onClick={() => remove(f.id)} className="text-red-600 text-xs hover:underline">Delete</button>
                  </td>
                </tr>
              ))}
              {fonts.length === 0 && <tr><td colSpan={4} className="p-4 text-center text-gray-500">No fonts uploaded yet.</td></tr>}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
