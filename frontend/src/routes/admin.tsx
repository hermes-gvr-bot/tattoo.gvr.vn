import { useEffect, useState } from 'react'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { getToken, parseToken, logout } from '../lib/auth'

interface ConsultationRow {
  id: string
  user_email: string
  body_photo_path: string
  idea_text: string
  status: string
  body_part: string | null
  skin_tone: string | null
  created_at: string
  variant_count: number
}

interface Stats {
  total: number
  pending: number
  generating: number
  completed: number
  rejected: number
}

function AdminPage() {
  const [consultations, setConsultations] = useState<ConsultationRow[]>([])
  const [stats, setStats] = useState<Stats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [filter, setFilter] = useState<string>('all')
  const [generatingId, setGeneratingId] = useState<string | null>(null)
  const [toast, setToast] = useState('')
  const navigate = useNavigate()
  const user = parseToken(getToken() || '')

  const fetchData = () => {
    const token = getToken()
    if (!token) return
    Promise.all([
      fetch('/api/admin/stats', { headers: { Authorization: `Bearer ${token}` } }).then(r => r.json()),
      fetch('/api/admin/consultations', { headers: { Authorization: `Bearer ${token}` } }).then(r => r.json()),
    ])
      .then(([s, c]) => {
        setStats(s)
        setConsultations(Array.isArray(c) ? c : [])
      })
      .catch(() => setError('Failed to load'))
      .finally(() => setLoading(false))
  }

  useEffect(() => { fetchData() }, [])

  const triggerGenerate = async (id: string) => {
    const token = getToken()
    if (!token) return
    setGeneratingId(id)
    setToast('')
    try {
      const res = await fetch(`/api/admin/consultations/${id}/generate`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
      })
      const data = await res.json()
      if (res.ok) {
        setToast(`✅ Generation started for consultation`)
        fetchData()
      } else {
        setToast(`❌ ${data.error}`)
      }
    } catch {
      setToast('❌ Network error')
    } finally {
      setGeneratingId(null)
    }
  }

  const statusBadge = (s: string) => {
    const map: Record<string, string> = {
      pending_payment: 'status-pending',
      generating: 'text-[#3b82f6]',
      completed: 'status-completed',
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
      rejected: 'Rejected',
    }
    return map[s] || s
  }

  const formatDate = (d: string) => {
    try { return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }) }
    catch { return d }
  }

  const filtered = filter === 'all'
    ? consultations
    : consultations.filter(c => c.status === filter || (filter === 'needs_action' && (c.status === 'pending_payment' || c.status === 'payment_confirmed')))

  return (
    <div className="min-h-screen bg-[#121316] flex">
      {/* Sidebar */}
      <aside className="sidebar w-[200px] shrink-0 flex flex-col">
        <div className="px-4 pt-5 pb-4">
          <a href="/" className="text-[13px] font-semibold text-[#e5e7eb] tracking-tight">TattooAI</a>
        </div>
        <nav className="flex-1 px-2 space-y-px">
          <a href="/dashboard" className="flex items-center h-9 px-3 text-[13px] font-medium text-[#9ca3af] hover:text-[#d1d5db] rounded-md">Dashboard</a>
          <a href="/admin" className="flex items-center h-9 px-3 text-[13px] font-medium text-[#e5e7eb] rounded-md bg-[rgba(255,255,255,0.04)]">Admin</a>
        </nav>
        <div className="px-4 pb-4 border-t border-[rgba(255,255,255,0.04)] pt-3">
          <p className="text-[11px] text-[#6b7280]">{user?.email}</p>
          <button onClick={logout} className="text-[11px] text-[#4b5563] hover:text-[#9ca3af] mt-1">Sign out</button>
        </div>
      </aside>

      <main className="flex-1">
        <header className="h-12 flex items-center justify-between px-6 border-b border-[rgba(255,255,255,0.06)]">
          <span className="text-[12px] text-[#9ca3af]">Admin Panel</span>
        </header>

        <div className="p-6">
          {/* Toast */}
          {toast && (
            <div className={`mb-4 text-[12px] px-3 py-2 rounded-md ${
              toast.startsWith('✅') ? 'bg-[rgba(34,197,94,0.08)] border border-[rgba(34,197,94,0.15)] text-[#22c55e]' : 'bg-[rgba(239,68,68,0.08)] border border-[rgba(239,68,68,0.15)] text-[#ef4444]'
            }`}>
              {toast}
            </div>
          )}

          {/* Stats */}
          {stats && (
            <div className="grid grid-cols-5 gap-3 mb-6">
              {[
                { label: 'Total', value: stats.total },
                { label: 'Pending', value: stats.pending, color: 'text-[#f59e0b]' },
                { label: 'Generating', value: stats.generating, color: 'text-[#3b82f6]' },
                { label: 'Completed', value: stats.completed, color: 'text-[#22c55e]' },
                { label: 'Rejected', value: stats.rejected, color: 'text-[#ef4444]' },
              ].map((s) => (
                <div key={s.label} className="border border-[rgba(255,255,255,0.04)] rounded-lg p-3 bg-[rgba(255,255,255,0.01)]">
                  <p className="text-[10px] text-[#6b7280] uppercase tracking-[0.05em] mb-1">{s.label}</p>
                  <p className={`text-[20px] font-bold ${s.color || 'text-[#d1d5db]'}`}>{s.value}</p>
                </div>
              ))}
            </div>
          )}

          {loading && <div className="text-[12px] text-[#6b7280]">Loading...</div>}
          {error && <div className="text-[12px] text-[#ef4444] mb-4">{error}</div>}

          {/* Filters */}
          <div className="flex items-center gap-2 mb-4">
            <span className="text-[10px] text-[#6b7280] uppercase tracking-[0.05em]">Filter:</span>
            {[
              { key: 'all', label: 'All' },
              { key: 'needs_action', label: 'Needs Action' },
              { key: 'pending_payment', label: 'Pending' },
              { key: 'generating', label: 'Generating' },
              { key: 'completed', label: 'Completed' },
              { key: 'rejected', label: 'Rejected' },
            ].map((f) => (
              <button
                key={f.key}
                onClick={() => setFilter(f.key)}
                className={`text-[10px] px-2.5 py-1 rounded-md border transition-colors ${
                  filter === f.key
                    ? 'border-[rgba(255,255,255,0.15)] text-[#d1d5db] bg-[rgba(255,255,255,0.04)]'
                    : 'border-[rgba(255,255,255,0.04)] text-[#6b7280] hover:text-[#9ca3af]'
                }`}
              >
                {f.label}
              </button>
            ))}
          </div>

          {/* Table */}
          {!loading && filtered.length > 0 && (
            <table className="w-full">
              <thead>
                <tr>
                  <th className="table-header text-left px-4 h-10">USER</th>
                  <th className="table-header text-left px-4 h-10">IDEA</th>
                  <th className="table-header text-center px-4 h-10">STATUS</th>
                  <th className="table-header text-center px-4 h-10">VARIANTS</th>
                  <th className="table-header text-right px-4 h-10">DATE</th>
                  <th className="table-header text-right px-4 h-10 w-[120px]">ACTIONS</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((c) => (
                  <tr key={c.id} className="table-row hover:bg-[rgba(255,255,255,0.015)]">
                    <td className="px-4 py-3">
                      <span className="text-[11px] text-[#9ca3af] font-mono">{c.user_email}</span>
                    </td>
                    <td className="px-4 py-3">
                      <span className="text-[12px] text-[#d1d5db] leading-snug line-clamp-1 block max-w-[300px]">
                        {c.idea_text}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-center">
                      <span className={`tag text-[10px] ${statusBadge(c.status)}`}>{statusText(c.status)}</span>
                    </td>
                    <td className="px-4 py-3 text-center">
                      <span className="text-[11px] text-[#6b7280]">{c.variant_count}</span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <span className="text-[10px] text-[#4b5563] font-mono">{formatDate(c.created_at)}</span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-1">
                        {(c.status === 'pending_payment' || c.status === 'payment_confirmed') && (
                          <button
                            onClick={(e) => { e.stopPropagation(); triggerGenerate(c.id) }}
                            disabled={generatingId === c.id}
                            className="text-[10px] px-2 py-1 rounded bg-[#3b82f6]/10 text-[#3b82f6] hover:bg-[#3b82f6]/20 transition-colors disabled:opacity-50"
                          >
                            {generatingId === c.id ? '...' : 'Generate'}
                          </button>
                        )}
                        <button
                          onClick={() => navigate({ to: '/consultations/$id', params: { id: c.id } })}
                          className="text-[10px] px-2 py-1 rounded text-[#6b7280] hover:text-[#d1d5db] hover:bg-[rgba(255,255,255,0.04)] transition-colors"
                        >
                          View
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}

          {!loading && filtered.length === 0 && (
            <div className="text-[12px] text-[#6b7280] border border-[rgba(255,255,255,0.06)] rounded-lg p-8 text-center">
              No consultations matching filter.
            </div>
          )}
        </div>
      </main>
    </div>
  )
}

export const Route = createFileRoute('/admin')({
  component: AdminPage,
  beforeLoad: () => {
    const token = getToken()
    if (!token) throw new Error('Not authenticated')
    const user = parseToken(token)
    if (user?.role !== 'admin') throw new Error('Admin only')
  },
})
