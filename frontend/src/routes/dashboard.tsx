import { useEffect, useState } from 'react'
import { getToken, parseToken, logout } from '../lib/auth'
import { useNavigate, createFileRoute } from '@tanstack/react-router'

interface ConsultationRow {
  id: string
  body_photo_path: string
  idea_text: string
  status: string
  body_part: string | null
  skin_tone: string | null
  created_at: string
  variant_count: number
}

function DashboardPage() {
  const [user, setUser] = useState<ReturnType<typeof parseToken>>(null)
  const [consultations, setConsultations] = useState<ConsultationRow[]>([])
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()

  useEffect(() => {
    const token = getToken()
    if (!token) { navigate({ to: '/login' }); return }
    setUser(parseToken(token))

    fetch('/api/consultations', { headers: { Authorization: `Bearer ${token}` } })
      .then((r) => r.json())
      .then((data) => setConsultations(data || []))
      .finally(() => setLoading(false))
  }, [navigate])

  const stats = {
    total: consultations.length,
    completed: consultations.filter((c) => c.status === 'completed').length,
    generating: consultations.filter((c) => c.status === 'generating').length,
    pending: consultations.filter((c) => c.status === 'pending_payment').length,
  }

  const statusBadge = (s: string) => {
    const map: Record<string, string> = {
      pending_payment: 'status-pending',
      payment_confirmed: 'text-[#3b82f6]',
      generating: 'text-[#3b82f6]',
      completed: 'status-completed',
      delivered: 'status-delivered',
      rejected: 'status-rejected',
    }
    return map[s] || 'text-[#6b7280]'
  }

  const statusText = (s: string) => {
    const map: Record<string, string> = {
      pending_payment: 'Pending',
      payment_confirmed: 'Paid',
      generating: 'Generating',
      completed: 'Complete',
      delivered: 'Delivered',
      rejected: 'Rejected',
    }
    return map[s] || s
  }

  const formatDate = (d: string) => {
    try { return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }) }
    catch { return d }
  }

  const imgSrc = (path: string) => {
    if (!path) return null
    return path.startsWith('http') ? path : `/uploads/${path.replace(/^\/?uploads\//, '')}`
  }

  return (
    <div className="min-h-screen bg-[#121316] flex">
      {/* Sidebar */}
      <aside className="sidebar w-[200px] shrink-0 flex flex-col">
        <div className="px-4 pt-5 pb-4">
          <a href="/" className="text-[13px] font-semibold text-[#e5e7eb] tracking-tight">TattooAI</a>
        </div>
        <nav className="flex-1 px-2 space-y-px">
          <a href="/dashboard" className="flex items-center h-9 px-3 text-[13px] font-medium text-[#e5e7eb] rounded-md bg-[rgba(255,255,255,0.04)]">Dashboard</a>
          <a href="/consultations/new" className="flex items-center h-9 px-3 text-[13px] font-medium text-[#9ca3af] hover:text-[#d1d5db] rounded-md">New Design</a>
        </nav>
        <div className="px-4 pb-4 border-t border-[rgba(255,255,255,0.04)] pt-3">
          <p className="text-[11px] text-[#6b7280]">{user?.email}</p>
          <button onClick={logout} className="text-[11px] text-[#4b5563] hover:text-[#9ca3af] mt-1">Sign out</button>
        </div>
      </aside>

      <main className="flex-1">
        <header className="h-12 flex items-center justify-between px-6 border-b border-[rgba(255,255,255,0.06)]">
          <span className="text-[12px] text-[#9ca3af]">My Designs</span>
          <a href="/consultations/new" className="h-8 px-3 rounded-md bg-[#d1d5db] text-[#121316] text-[11px] font-semibold hover:bg-[#e5e7eb] transition-colors flex items-center">
            + New Design
          </a>
        </header>

        <div className="p-6">
          {/* Stats */}
          {!loading && consultations.length > 0 && (
            <div className="grid grid-cols-4 gap-3 mb-6">
              {[
                { label: 'Total', value: stats.total },
                { label: 'Completed', value: stats.completed, color: 'text-[#22c55e]' },
                { label: 'In Progress', value: stats.generating, color: 'text-[#3b82f6]' },
                { label: 'Pending', value: stats.pending, color: 'text-[#f59e0b]' },
              ].map((stat) => (
                <div key={stat.label} className="border border-[rgba(255,255,255,0.04)] rounded-lg p-3 bg-[rgba(255,255,255,0.01)]">
                  <p className="text-[10px] text-[#6b7280] uppercase tracking-[0.05em] mb-1">{stat.label}</p>
                  <p className={`text-[20px] font-bold ${stat.color || 'text-[#d1d5db]'}`}>{stat.value}</p>
                </div>
              ))}
            </div>
          )}

          {loading && <div className="text-[12px] text-[#6b7280]">Loading...</div>}

          {!loading && consultations.length === 0 && (
            <div className="border border-[rgba(255,255,255,0.06)] rounded-xl p-12 text-center max-w-md mx-auto mt-12">
              <div className="w-12 h-12 rounded-full bg-[rgba(255,255,255,0.03)] flex items-center justify-center mx-auto mb-4">
                <span className="text-xl">🎨</span>
              </div>
              <h3 className="text-[14px] font-semibold text-[#d1d5db] mb-2">No designs yet</h3>
              <p className="text-[12px] text-[#6b7280] mb-6">Upload a photo and describe your idea to get AI-generated tattoo designs.</p>
              <a href="/consultations/new" className="h-9 px-5 rounded-md bg-[#d1d5db] text-[#121316] text-[12px] font-semibold hover:bg-[#e5e7eb] transition-colors inline-flex items-center">
                Create your first design →
              </a>
            </div>
          )}

          {/* Consultation cards */}
          {!loading && consultations.length > 0 && (
            <div className="grid grid-cols-1 gap-2">
              {consultations.map((c) => (
                <div
                  key={c.id}
                  className="flex items-center gap-4 border border-[rgba(255,255,255,0.04)] rounded-lg p-3 hover:bg-[rgba(255,255,255,0.015)] cursor-pointer transition-colors"
                  onClick={() => navigate({ to: '/consultations/$id', params: { id: c.id } })}
                >
                  {/* Thumbnail */}
                  <div className="w-12 h-12 shrink-0 rounded-md overflow-hidden bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.04)]">
                    {imgSrc(c.body_photo_path) ? (
                      <img src={imgSrc(c.body_photo_path)!} alt="" className="w-full h-full object-cover" />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center text-[#4b5563] text-[16px]">📷</div>
                    )}
                  </div>

                  {/* Info */}
                  <div className="flex-1 min-w-0">
                    <p className="text-[13px] text-[#d1d5db] leading-snug truncate">
                      {c.idea_text.length > 80 ? c.idea_text.slice(0, 80) + '...' : c.idea_text}
                    </p>
                    <div className="flex items-center gap-3 mt-1">
                      <span className={`tag text-[10px] ${statusBadge(c.status)}`}>{statusText(c.status)}</span>
                      {c.variant_count > 0 && (
                        <span className="text-[10px] text-[#4b5563]">{c.variant_count} variant{c.variant_count > 1 ? 's' : ''}</span>
                      )}
                      {c.body_part && (
                        <span className="text-[10px] text-[#4b5563]">📍 {c.body_part}</span>
                      )}
                    </div>
                  </div>

                  {/* Date + arrow */}
                  <div className="text-right shrink-0">
                    <p className="text-[10px] text-[#4b5563] font-mono">{formatDate(c.created_at)}</p>
                    <p className="text-[11px] text-[#6b7280] mt-1">View →</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </main>
    </div>
  )
}

export const Route = createFileRoute('/dashboard')({
  component: DashboardPage,
})
