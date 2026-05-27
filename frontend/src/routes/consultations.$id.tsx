import { useEffect, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { getToken } from '../lib/auth'

interface Consultation {
  id: string
  body_photo_path: string
  idea_text: string
  status: string
  body_part: string | null
  skin_tone: string | null
  created_at: string
}

interface Variant {
  id: string
  variant_number: number
  prompt_used: string
  sketch_path: string | null
  final_path: string | null
  sketch_status: string
  final_status: string
}

function getStyleLabel(prompt: string): string {
  if (!prompt) return 'Variant'
  const lower = prompt.toLowerCase()
  if (lower.includes('watercolor') || lower.includes('neo-japanese')) return 'Neo-Japanese Watercolor'
  if (lower.includes('bold') || lower.includes('dynamic') || lower.includes('intense')) return 'Bold Irezumi'
  if (lower.includes('traditional')) return 'Traditional Irezumi'
  if (lower.includes('minimal')) return 'Minimalist'
  if (lower.includes('geometric')) return 'Geometric'
  return 'Custom Design'
}

function getMoodLabel(prompt: string): string {
  if (!prompt) return ''
  const lower = prompt.toLowerCase()
  if (lower.includes('powerful') || lower.includes('intense')) return 'Powerful & Intense'
  if (lower.includes('ethereal') || lower.includes('dream')) return 'Ethereal & Dreamy'
  if (lower.includes('harmonious') || lower.includes('balanced')) return 'Harmonious & Balanced'
  return 'Artistic'
}

function VariantCard({ v, selected, onToggle, isCompared }: {
  v: Variant
  selected: boolean
  onToggle: () => void
  isCompared: boolean
}) {
  const [zoomed, setZoomed] = useState(false)
  const style = getStyleLabel(v.prompt_used)
  const mood = getMoodLabel(v.prompt_used)

  const imgSrc = v.final_path
    ? (v.final_path.startsWith('http') ? v.final_path : `/uploads/${v.final_path}`)
    : v.sketch_path
    ? (v.sketch_path.startsWith('http') ? v.sketch_path : `/uploads/${v.sketch_path}`)
    : null

  return (
    <div className={`border rounded-xl overflow-hidden transition-all ${
      selected ? 'border-[rgba(255,255,255,0.2)] ring-1 ring-[rgba(255,255,255,0.08)]' : 'border-[rgba(255,255,255,0.06)]'
    }`}>
      {/* Image */}
      {imgSrc ? (
        <div
          className={`relative cursor-pointer overflow-hidden bg-[rgba(255,255,255,0.02)] ${
            zoomed ? 'fixed inset-0 z-50 bg-[#121316]/95 flex items-center justify-center p-8' : 'aspect-square'
          }`}
          onClick={() => setZoomed(!zoomed)}
        >
          <img
            src={imgSrc}
            alt={`${style} design`}
            className={`${zoomed ? 'max-h-[90vh] max-w-[90vw] object-contain' : 'w-full h-full object-cover hover:scale-105'} transition-transform duration-300`}
          />
          {!zoomed && (
            <div className="absolute bottom-2 right-2 bg-[#121316]/80 rounded px-2 py-1 text-[10px] text-[#9ca3af]">
              Click to zoom
            </div>
          )}
          {zoomed && (
            <button
              className="absolute top-4 right-4 w-8 h-8 rounded-full bg-[rgba(255,255,255,0.1)] flex items-center justify-center text-[#9ca3af] hover:text-[#d1d5db] text-lg"
              onClick={(e) => { e.stopPropagation(); setZoomed(false) }}
            >
              ✕
            </button>
          )}
        </div>
      ) : (
        <div className="aspect-square bg-[rgba(255,255,255,0.02)] flex items-center justify-center">
          <span className="text-[11px] text-[#4b5563]">
            {v.final_status === 'generating' ? '🎨 Generating...' : '⏳ Pending'}
          </span>
        </div>
      )}

      {/* Info */}
      <div className="p-4">
        <div className="flex items-center justify-between mb-2">
          <span className="text-[12px] font-semibold text-[#d1d5db]">#{v.variant_number} — {style}</span>
          <button
            onClick={onToggle}
            className={`text-[10px] px-2 py-1 rounded transition-colors ${
              selected
                ? 'bg-[rgba(255,255,255,0.08)] text-[#d1d5db]'
                : 'text-[#6b7280] hover:text-[#9ca3af]'
            }`}
          >
            {selected ? '✓ Selected' : isCompared ? 'Compare' : 'Select'}
          </button>
        </div>
        <p className="text-[10px] text-[#6b7280] mb-2">{mood}</p>

        {/* Prompt preview */}
        <details className="group">
          <summary className="text-[10px] text-[#6b7280] cursor-pointer hover:text-[#9ca3af] list-none flex items-center gap-1">
            <span className="group-open:hidden">▶ Show prompt</span>
            <span className="hidden group-open:inline">▼ Hide prompt</span>
          </summary>
          <p className="text-[10px] text-[#4b5563] mt-2 leading-relaxed line-clamp-6">{v.prompt_used}</p>
        </details>
      </div>
    </div>
  )
}

function ConsultationDetailPage() {
  const { id } = Route.useParams()
  const [consultation, setConsultation] = useState<Consultation | null>(null)
  const [variants, setVariants] = useState<Variant[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [compareMode, setCompareMode] = useState(false)
  const [compared, setCompared] = useState<Set<string>>(new Set())

  useEffect(() => {
    const token = getToken()
    if (!token) return
    fetch(`/api/consultations/${id}`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((r) => r.json())
      .then((data) => {
        if (data.error) { setError(data.error); return }
        setConsultation(data.consultation)
        setVariants(data.variants || [])
      })
      .catch(() => setError('Failed to load'))
      .finally(() => setLoading(false))
  }, [id])

  const toggleCompare = (vid: string) => {
    setCompared((prev) => {
      const next = new Set(prev)
      if (next.has(vid)) {
        next.delete(vid)
      } else {
        if (next.size >= 2) {
          const [first] = next
          next.delete(first)
        }
        next.add(vid)
      }
      return next
    })
  }

  const bodyImg = consultation?.body_photo_path
    ? (consultation.body_photo_path.startsWith('http')
        ? consultation.body_photo_path
        : `/uploads/${consultation.body_photo_path.replace(/^\/?uploads\//, '')}`)
    : null

  return (
    <div className="min-h-screen bg-[#121316] flex">
      {/* Sidebar */}
      <aside className="sidebar w-[200px] shrink-0 flex flex-col">
        <div className="px-4 pt-5 pb-4">
          <a href="/" className="text-[13px] font-semibold text-[#e5e7eb] tracking-tight">TattooAI</a>
        </div>
        <nav className="flex-1 px-2 space-y-px">
          <a href="/dashboard" className="flex items-center h-9 px-3 text-[13px] font-medium text-[#9ca3af] hover:text-[#d1d5db] rounded-md">Dashboard</a>
          <a href="/consultations/new" className="flex items-center h-9 px-3 text-[13px] font-medium text-[#9ca3af] hover:text-[#d1d5db] rounded-md">New</a>
        </nav>
      </aside>

      <main className="flex-1">
        <header className="h-12 flex items-center justify-between px-6 border-b border-[rgba(255,255,255,0.06)]">
          <div className="flex items-center gap-4">
            <a href="/dashboard" className="text-[12px] text-[#6b7280] hover:text-[#9ca3af]">← Dashboard</a>
            <span className="text-[12px] text-[#4b5563]">/</span>
            <span className="text-[12px] text-[#9ca3af]">Consultation</span>
          </div>
          {variants.length >= 2 && (
            <button
              onClick={() => { setCompareMode(!compareMode); setCompared(new Set()) }}
              className={`text-[11px] px-3 py-1.5 rounded-md transition-colors ${
                compareMode ? 'bg-[rgba(255,255,255,0.08)] text-[#d1d5db]' : 'text-[#6b7280] hover:text-[#9ca3af]'
              }`}
            >
              {compareMode ? 'Exit compare' : 'Compare A/B'}
            </button>
          )}
        </header>

        {loading && <div className="p-6 text-[12px] text-[#6b7280]">Loading...</div>}
        {error && <div className="p-6 text-[12px] text-[#ef4444]">{error}</div>}

        {consultation && (
          <div className="p-6 max-w-6xl">
            {/* Idea + Body photo header */}
            <div className="flex gap-6 mb-8">
              {bodyImg && (
                <div className="w-32 h-32 shrink-0 rounded-lg overflow-hidden border border-[rgba(255,255,255,0.06)]">
                  <img src={bodyImg} alt="Body photo" className="w-full h-full object-cover" />
                </div>
              )}
              <div className="flex-1 min-w-0">
                <h2 className="text-[15px] font-semibold text-[#d1d5db] mb-2 leading-snug">{consultation.idea_text}</h2>
                <div className="flex flex-wrap gap-2">
                  {consultation.body_part && (
                    <span className="text-[10px] text-[#6b7280] bg-[rgba(255,255,255,0.03)] rounded px-2 py-0.5">
                      📍 {consultation.body_part}
                    </span>
                  )}
                  {consultation.skin_tone && (
                    <span className="text-[10px] text-[#6b7280] bg-[rgba(255,255,255,0.03)] rounded px-2 py-0.5">
                      🎨 {consultation.skin_tone}
                    </span>
                  )}
                </div>
              </div>
            </div>

            {/* Variants */}
            {variants.length > 0 && (
              <div>
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-[12px] font-semibold text-[#9ca3af] uppercase tracking-[0.05em]">
                    {variants.length} Design Variants
                  </h3>
                </div>

                {compareMode && compared.size === 2 ? (
                  /* Side-by-side comparison */
                  <div className="grid grid-cols-2 gap-4">
                    {variants.filter((v) => compared.has(v.id)).map((v) => (
                      <VariantCard
                        key={v.id}
                        v={v}
                        selected={true}
                        onToggle={() => toggleCompare(v.id)}
                        isCompared={true}
                      />
                    ))}
                  </div>
                ) : (
                  /* Grid view */
                  <div className="grid grid-cols-3 gap-4">
                    {variants.map((v) => (
                      <VariantCard
                        key={v.id}
                        v={v}
                        selected={compareMode && compared.has(v.id)}
                        onToggle={() => toggleCompare(v.id)}
                        isCompared={compareMode}
                      />
                    ))}
                  </div>
                )}
              </div>
            )}

            {variants.length === 0 && consultation.status !== 'generating' && (
              <div className="border border-[rgba(255,255,255,0.06)] rounded-xl p-12 text-center">
                <p className="text-[13px] text-[#6b7280] mb-2">No designs generated yet</p>
                <p className="text-[11px] text-[#4b5563]">Designs will appear here once an admin processes your consultation.</p>
              </div>
            )}
          </div>
        )}
      </main>
    </div>
  )
}

export const Route = createFileRoute('/consultations/$id')({
  component: ConsultationDetailPage,
})
