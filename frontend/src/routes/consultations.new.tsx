import { useState } from 'react';
import { useNavigate, createFileRoute } from '@tanstack/react-router';
import { getToken } from '../lib/auth';

function NewConsultationPage() {
  const [ideaText, setIdeaText] = useState('');
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0];
    if (f) {
      setFile(f);
      setPreview(URL.createObjectURL(f));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) { setError('Please select a body photo'); return; }
    if (!ideaText.trim()) { setError('Please describe your tattoo idea'); return; }

    setError('');
    setLoading(true);
    try {
      const token = getToken();
      const formData = new FormData();
      formData.append('body_photo', file);
      formData.append('idea_text', ideaText);

      const res = await fetch('/api/consultations', {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: formData,
      });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || 'Failed to create consultation');
      }
      const data = await res.json();
      navigate({ to: '/consultations/$id', params: { id: data.id } });
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
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
          <a href="/consultations/new" className="flex items-center h-9 px-3 text-[13px] font-medium text-[#e5e7eb] rounded-md bg-[rgba(255,255,255,0.04)]">New Consultation</a>
        </nav>
      </aside>
      <main className="flex-1">
        <header className="h-12 flex items-center px-6 border-b border-[rgba(255,255,255,0.06)]">
          <span className="text-[12px] text-[#9ca3af]">New Consultation</span>
        </header>
        <div className="p-6 max-w-2xl">
          {error && (
            <div className="text-[12px] text-[#ef4444] bg-[rgba(239,68,68,0.08)] border border-[rgba(239,68,68,0.15)] rounded-md px-3 py-2 mb-4">{error}</div>
          )}
          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label className="block text-[12px] font-medium text-[#9ca3af] mb-2">Body Photo</label>
              <div className="border-2 border-dashed border-[rgba(255,255,255,0.08)] rounded-lg p-8 text-center hover:border-[rgba(255,255,255,0.12)] transition-colors cursor-pointer"
                   onClick={() => document.getElementById('photo-upload')?.click()}>
                {preview ? (
                  <img src={preview} alt="Preview" className="max-h-48 mx-auto rounded-md" />
                ) : (
                  <div>
                    <p className="text-[13px] text-[#9ca3af]">Click to upload body photo</p>
                    <p className="text-[11px] text-[#6b7280] mt-1">JPG, PNG, WebP — max 10MB</p>
                  </div>
                )}
                <input id="photo-upload" type="file" accept="image/*" onChange={handleFileChange} className="hidden" />
              </div>
              {file && <p className="text-[11px] text-[#6b7280] mt-1">{file.name}</p>}
            </div>

            <div>
              <label className="block text-[12px] font-medium text-[#9ca3af] mb-2">Your Tattoo Idea</label>
              <textarea
                value={ideaText}
                onChange={(e) => setIdeaText(e.target.value)}
                className="input-field w-full min-h-[120px] resize-y"
                placeholder="Describe your tattoo idea — style, elements, placement, size, colors, any reference inspiration..."
                required
              />
              <p className="text-[11px] text-[#6b7280] mt-1">Be specific: include style, elements, colors, placement on body</p>
            </div>

            <button type="submit" disabled={loading} className="btn-primary w-full flex items-center justify-center gap-2">
              {loading ? <span className="inline-block w-4 h-4 border-2 border-[#121316] border-t-transparent rounded-full animate-spin" /> : null}
              Submit for Design
            </button>
          </form>
        </div>
      </main>
    </div>
  );
}

export const Route = createFileRoute('/consultations/new')({
  component: NewConsultationPage,
});
