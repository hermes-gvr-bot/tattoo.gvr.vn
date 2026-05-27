import { useEffect, useState } from 'react';
import { getToken, parseToken, logout } from '../lib/auth';
import { useNavigate, createFileRoute } from '@tanstack/react-router';

function DashboardPage() {
  const [user, setUser] = useState<ReturnType<typeof parseToken>>(null);
  const navigate = useNavigate();

  useEffect(() => {
    const token = getToken();
    if (!token) {
      navigate({ to: '/login' });
      return;
    }
    setUser(parseToken(token));
  }, [navigate]);

  return (
    <div className="min-h-screen bg-[#121316] flex">
      {/* Sidebar */}
      <aside className="sidebar w-[220px] shrink-0 flex flex-col">
        <div className="px-4 pt-5 pb-4">
          <span className="text-[14px] font-semibold text-[#e5e7eb] tracking-tight">
            Tattoo
          </span>
        </div>
        <nav className="flex-1 px-2 space-y-px">
          <a href="/dashboard" className="flex items-center h-9 px-3 text-[13px] font-medium text-[#e5e7eb] rounded-md bg-[rgba(255,255,255,0.04)]">
            Dashboard
          </a>
          <a href="/consultations/new" className="flex items-center h-9 px-3 text-[13px] font-medium text-[#9ca3af] hover:text-[#d1d5db] rounded-md hover:bg-[rgba(255,255,255,0.03)]">
            New Consultation
          </a>
          {user?.role === 'admin' && (
            <a href="/admin" className="flex items-center h-9 px-3 text-[13px] font-medium text-[#9ca3af] hover:text-[#d1d5db] rounded-md hover:bg-[rgba(255,255,255,0.03)]">
              Admin
            </a>
          )}
        </nav>
        <div className="px-4 pb-4">
          <p className="text-[11px] text-[#6b7280]">{user?.email}</p>
          <button onClick={logout} className="text-[11px] text-[#6b7280] hover:text-[#9ca3af] mt-1">
            Sign out
          </button>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1">
        <header className="h-12 flex items-center px-6 border-b border-[rgba(255,255,255,0.06)]">
          <span className="text-[12px] text-[#9ca3af]">
            Welcome, {user?.email}
          </span>
        </header>
        <div className="p-6">
          <h2 className="text-[13px] font-semibold text-[#d1d5db] mb-4">My Consultations</h2>
          <div className="text-[12px] text-[#6b7280] border border-[rgba(255,255,255,0.06)] rounded-lg p-8 text-center">
            No consultations yet.{' '}
            <a href="/consultations/new" className="text-[#9ca3af] hover:text-[#d1d5db] underline underline-offset-2">
              Create your first design
            </a>
          </div>
        </div>
      </main>
    </div>
  );
}

export const Route = createFileRoute('/dashboard')({
  component: DashboardPage,
});
