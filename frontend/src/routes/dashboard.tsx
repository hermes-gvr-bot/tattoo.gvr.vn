import { useEffect, useState } from 'react';
import { getToken, parseToken, logout } from '../lib/auth';
import { useNavigate, createFileRoute } from '@tanstack/react-router';

interface ConsultationRow {
  id: string;
  body_photo_path: string;
  idea_text: string;
  status: string;
  body_part: string | null;
  skin_tone: string | null;
  created_at: string;
}

function DashboardPage() {
  const [user, setUser] = useState<ReturnType<typeof parseToken>>(null);
  const [consultations, setConsultations] = useState<ConsultationRow[]>([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    const token = getToken();
    if (!token) { navigate({ to: '/login' }); return; }
    setUser(parseToken(token));

    fetch('/api/consultations', { headers: { Authorization: `Bearer ${token}` } })
      .then((r) => r.json())
      .then((data) => setConsultations(data || []))
      .finally(() => setLoading(false));
  }, [navigate]);

  const statusBadge = (s: string) => {
    const map: Record<string, string> = {
      pending_payment: 'status-pending',
      payment_confirmed: 'text-[#3b82f6]',
      generating: 'text-[#3b82f6]',
      completed: 'status-completed',
      delivered: 'status-delivered',
      rejected: 'status-rejected',
    };
    return map[s] || 'text-[#6b7280]';
  };

  const statusText = (s: string) => {
    const map: Record<string, string> = {
      pending_payment: 'Pending',
      payment_confirmed: 'Paid',
      generating: 'Generating',
      completed: 'Complete',
      delivered: 'Delivered',
      rejected: 'Rejected',
    };
    return map[s] || s;
  };

  const formatDate = (d: string) => {
    try { return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }); }
    catch { return d; }
  };

  return (
    <div className="min-h-screen bg-[#121316] flex">
      <aside className="sidebar w-[220px] shrink-0 flex flex-col">
        <div className="px-4 pt-5 pb-4">
          <span className="text-[14px] font-semibold text-[#e5e7eb] tracking-tight">Tattoo</span>
        </div>
        <nav className="flex-1 px-2 space-y-px">
          <a href="/dashboard" className="flex items-center h-9 px-3 text-[13px] font-medium text-[#e5e7eb] rounded-md bg-[rgba(255,255,255,0.04)]">Dashboard</a>
          <a href="/consultations/new" className="flex items-center h-9 px-3 text-[13px] font-medium text-[#9ca3af] hover:text-[#d1d5db] rounded-md">New Consultation</a>
          {user?.role === 'admin' && (
            <a href="/admin" className="flex items-center h-9 px-3 text-[13px] font-medium text-[#9ca3af] hover:text-[#d1d5db] rounded-md">Admin</a>
          )}
        </nav>
        <div className="px-4 pb-4">
          <p className="text-[11px] text-[#6b7280]">{user?.email}</p>
          <button onClick={logout} className="text-[11px] text-[#6b7280] hover:text-[#9ca3af] mt-1">Sign out</button>
        </div>
      </aside>

      <main className="flex-1">
        <header className="h-12 flex items-center justify-between px-6 border-b border-[rgba(255,255,255,0.06)]">
          <span className="text-[12px] text-[#9ca3af]">Welcome, {user?.email}</span>
          <a href="/consultations/new" className="btn-primary text-[12px]">+ New Consultation</a>
        </header>
        <div className="p-6">
          <h2 className="text-[13px] font-semibold text-[#d1d5db] mb-4">My Consultations</h2>

          {loading && <div className="text-[12px] text-[#6b7280]">Loading...</div>}

          {!loading && consultations.length === 0 && (
            <div className="text-[12px] text-[#6b7280] border border-[rgba(255,255,255,0.06)] rounded-lg p-8 text-center">
              No consultations yet.{' '}
              <a href="/consultations/new" className="text-[#9ca3af] hover:text-[#d1d5db] underline underline-offset-2">Create your first design</a>
            </div>
          )}

          {!loading && consultations.length > 0 && (
            <table className="w-full">
              <thead>
                <tr>
                  <th className="table-header text-left px-6 h-10">IDEA</th>
                  <th className="table-header text-left px-5 h-10">STATUS</th>
                  <th className="table-header text-left px-5 h-10">CREATED</th>
                  <th className="table-header text-right px-6 h-10"></th>
                </tr>
              </thead>
              <tbody>
                {consultations.map((c) => (
                  <tr key={c.id} className="table-row cursor-pointer hover:bg-[rgba(255,255,255,0.015)]"
                      onClick={() => navigate({ to: '/consultations/$id', params: { id: c.id } })}>
                    <td className="px-6 py-3.5">
                      <span className="text-[13px] text-[#d1d5db] leading-snug line-clamp-1">
                        {c.idea_text.length > 60 ? c.idea_text.slice(0, 60) + '...' : c.idea_text}
                      </span>
                    </td>
                    <td className="px-5 py-3.5">
                      <span className={`tag text-[10px] ${statusBadge(c.status)}`}>{statusText(c.status)}</span>
                    </td>
                    <td className="px-5 py-3.5">
                      <span className="text-[11px] text-[#6b7280] font-mono">{formatDate(c.created_at)}</span>
                    </td>
                    <td className="px-6 py-3.5 text-right">
                      <span className="text-[11px] text-[#6b7280]">View →</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </main>
    </div>
  );
}

export const Route = createFileRoute('/dashboard')({
  component: DashboardPage,
});
