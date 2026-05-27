import { useEffect, useState } from 'react';
import { createFileRoute } from '@tanstack/react-router';
import { getToken } from '../lib/auth';

interface Consultation {
  id: string;
  body_photo_path: string;
  idea_text: string;
  status: string;
  body_part: string | null;
  skin_tone: string | null;
  created_at: string;
}

interface Variant {
  id: string;
  variant_number: number;
  prompt_used: string;
  sketch_path: string | null;
  final_path: string | null;
  sketch_status: string;
  final_status: string;
}

function ConsultationDetailPage() {
  const { id } = Route.useParams();
  const [consultation, setConsultation] = useState<Consultation | null>(null);
  const [variants, setVariants] = useState<Variant[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    fetch(`/api/consultations/${id}`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((r) => r.json())
      .then((data) => {
        if (data.error) { setError(data.error); return; }
        setConsultation(data.consultation);
        setVariants(data.variants || []);
      })
      .catch(() => setError('Failed to load'))
      .finally(() => setLoading(false));
  }, [id]);

  const statusLabel = (s: string) => {
    switch (s) {
      case 'pending_payment': return '⏳ Awaiting Payment';
      case 'payment_confirmed': return '💰 Paid';
      case 'generating': return '🎨 Generating...';
      case 'completed': return '✅ Complete';
      case 'delivered': return '📬 Delivered';
      case 'rejected': return '❌ Rejected';
      default: return s;
    }
  };

  return (
    <div className="min-h-screen bg-[#121316] flex">
      <aside className="sidebar w-[220px] shrink-0 flex flex-col">
        <div className="px-4 pt-5 pb-4">
          <span className="text-[14px] font-semibold text-[#e5e7eb] tracking-tight">Tattoo</span>
        </div>
        <nav className="flex-1 px-2 space-y-px">
          <a href="/dashboard" className="flex items-center h-9 px-3 text-[13px] font-medium text-[#9ca3af] hover:text-[#d1d5db] rounded-md">Dashboard</a>
          <a href="/consultations/new" className="flex items-center h-9 px-3 text-[13px] font-medium text-[#9ca3af] hover:text-[#d1d5db] rounded-md">New</a>
        </nav>
      </aside>
      <main className="flex-1">
        <header className="h-12 flex items-center px-6 border-b border-[rgba(255,255,255,0.06)]">
          <a href="/dashboard" className="text-[12px] text-[#6b7280] hover:text-[#9ca3af] mr-2">← Back</a>
          <span className="text-[12px] text-[#9ca3af]">Consultation Detail</span>
        </header>

        {loading && <div className="p-6 text-[12px] text-[#6b7280]">Loading...</div>}
        {error && <div className="p-6 text-[12px] text-[#ef4444]">{error}</div>}

        {consultation && (
          <div className="p-6 max-w-4xl">
            <div className="flex gap-6 mb-6">
              <div className="w-48 shrink-0">
                <img
                  src={consultation.body_photo_path.startsWith('/uploads/') ? consultation.body_photo_path : `/uploads/${consultation.body_photo_path.replace('uploads/', '')}`}
                  alt="Body photo"
                  className="w-full rounded-lg border border-[rgba(255,255,255,0.06)]"
                />
              </div>
              <div className="flex-1">
                <h2 className="text-[14px] font-semibold text-[#d1d5db] mb-3">{statusLabel(consultation.status)}</h2>
                <p className="text-[13px] text-[#d1d5db] leading-relaxed">{consultation.idea_text}</p>
                {consultation.body_part && (
                  <p className="text-[12px] text-[#6b7280] mt-2">Body part: {consultation.body_part}</p>
                )}
              </div>
            </div>

            {variants.length > 0 && (
              <div>
                <h3 className="text-[12px] font-semibold text-[#9ca3af] uppercase tracking-[0.05em] mb-3">Design Variants</h3>
                <div className="grid grid-cols-3 gap-3">
                  {variants.map((v) => (
                    <div key={v.id} className="border border-[rgba(255,255,255,0.06)] rounded-lg p-3">
                      <p className="text-[11px] font-medium text-[#9ca3af] mb-2">Variant {v.variant_number}</p>
                      {v.final_path ? (
                        <img src={v.final_path} alt={`Variant ${v.variant_number}`} className="w-full rounded-md" />
                      ) : v.sketch_path ? (
                        <img src={v.sketch_path} alt={`Sketch ${v.variant_number}`} className="w-full rounded-md opacity-60" />
                      ) : (
                        <div className="aspect-square bg-[rgba(255,255,255,0.03)] rounded-md flex items-center justify-center">
                          <span className="text-[11px] text-[#6b7280]">
                            {v.final_status === 'generating' ? '🎨 Generating...' : '⏳ Pending'}
                          </span>
                        </div>
                      )}
                      <p className="text-[10px] text-[#6b7280] mt-1.5">
                        {v.final_status === 'completed' ? '✓ Ready' : v.final_status === 'generating' ? 'Generating...' : 'Pending'}
                      </p>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </main>
    </div>
  );
}

export const Route = createFileRoute('/consultations/$id')({
  component: ConsultationDetailPage,
});
