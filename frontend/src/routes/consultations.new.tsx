import { useState, useRef } from 'react'
import { useNavigate, createFileRoute } from '@tanstack/react-router'
import { getToken } from '../lib/auth'

const STYLE_SUGGESTIONS = [
  'Traditional Japanese Irezumi',
  'Black & grey realism',
  'Neo-traditional',
  'Watercolor',
  'Geometric / Mandala',
  'Minimalist line art',
  'Tribal / Polynesian',
  'Old school traditional',
  'New school',
  'Fine line / Micro',
]

const CONCERN_SUGGESTIONS = [
  'Lines are blown out / faded',
  'Colors have faded unevenly',
  'Design no longer matches my style',
  'Want to extend / add elements',
  'Complete cover-up with new design',
  'Touch-up and refresh colors',
  'Poor original quality, needs fix',
  'Sun damage / aging visible',
]

function NewConsultationPage() {
  const [step, setStep] = useState(1)
  const [consultationType, setConsultationType] = useState<'new_tattoo' | 'makeup_enhance'>('new_tattoo')
  const [ideaText, setIdeaText] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const [preview, setPreview] = useState<string | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const navigate = useNavigate()

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0]
    if (f) {
      setFile(f)
      setPreview(URL.createObjectURL(f))
      setError('')
    }
  }

  const addSuggestion = (text: string) => {
    setIdeaText((prev) => {
      const hasText = prev.toLowerCase().includes(text.toLowerCase())
      return hasText ? prev : (prev ? `${prev}. ${text}` : text)
    })
  }

  const handleSubmit = async () => {
    if (!file) { setError('Please upload a photo'); return }
    if (!ideaText.trim()) { setError('Please describe what you want'); return }

    setError('')
    setLoading(true)
    try {
      const token = getToken()
      const formData = new FormData()
      formData.append('body_photo', file)
      formData.append('idea_text', ideaText)
      formData.append('consultation_type', consultationType)

      const res = await fetch('/api/consultations', {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: formData,
      })
      if (!res.ok) {
        const err = await res.json()
        throw new Error(err.error || 'Failed to create consultation')
      }
      const data = await res.json()
      navigate({ to: '/consultations/$id', params: { id: data.id } })
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const totalSteps = 4

  return (
    <div className="min-h-screen bg-[#121316] flex">
      {/* Sidebar */}
      <aside className="sidebar w-[200px] shrink-0 flex flex-col">
        <div className="px-4 pt-5 pb-4">
          <a href="/" className="text-[13px] font-semibold text-[#e5e7eb] tracking-tight">TattooAI</a>
        </div>
        <nav className="flex-1 px-2 space-y-px">
          <a href="/dashboard" className="flex items-center h-9 px-3 text-[13px] font-medium text-[#9ca3af] hover:text-[#d1d5db] rounded-md">Dashboard</a>
          <a href="/consultations/new" className="flex items-center h-9 px-3 text-[13px] font-medium text-[#e5e7eb] rounded-md bg-[rgba(255,255,255,0.04)]">New</a>
        </nav>
      </aside>

      <main className="flex-1">
        <header className="h-12 flex items-center px-6 border-b border-[rgba(255,255,255,0.06)]">
          <a href="/dashboard" className="text-[12px] text-[#6b7280] hover:text-[#9ca3af] mr-2">← Dashboard</a>
          <span className="text-[12px] text-[#4b5563]">/</span>
          <span className="text-[12px] text-[#9ca3af] ml-2">
            {consultationType === 'makeup_enhance' ? 'Tattoo Makeup' : 'New design'}
          </span>
        </header>

        <div className="max-w-2xl mx-auto px-6 py-10">
          {/* Progress steps */}
          <div className="flex items-center justify-center gap-2 mb-10">
            {[1, 2, 3, 4].map((s) => (
              <div key={s} className="flex items-center gap-2">
                <div className={`w-8 h-8 rounded-full flex items-center justify-center text-[11px] font-semibold transition-colors ${
                  s === step
                    ? 'bg-[#d1d5db] text-[#121316]'
                    : s < step
                    ? 'bg-[rgba(255,255,255,0.1)] text-[#9ca3af]'
                    : 'bg-[rgba(255,255,255,0.03)] text-[#4b5563]'
                }`}>
                  {s < step ? '✓' : s}
                </div>
                {s < totalSteps && <div className={`w-10 h-px ${s < step ? 'bg-[rgba(255,255,255,0.15)]' : 'bg-[rgba(255,255,255,0.04)]'}`} />}
              </div>
            ))}
          </div>

          {error && (
            <div className="text-[12px] text-[#ef4444] bg-[rgba(239,68,68,0.08)] border border-[rgba(239,68,68,0.15)] rounded-md px-3 py-2 mb-4">{error}</div>
          )}

          {/* Step 1: Choose type */}
          {step === 1 && (
            <div>
              <h2 className="text-[16px] font-semibold text-[#d1d5db] mb-2">What do you need?</h2>
              <p className="text-[12px] text-[#6b7280] mb-6">Choose the type of consultation. This determines how our AI analyzes your photo and generates designs.</p>

              <div className="grid grid-cols-2 gap-4">
                <button
                  onClick={() => { setConsultationType('new_tattoo'); setStep(2); }}
                  className={`border rounded-xl p-6 text-left transition-all ${
                    consultationType === 'new_tattoo'
                      ? 'border-[#d1d5db] bg-[rgba(255,255,255,0.04)]'
                      : 'border-[rgba(255,255,255,0.06)] hover:border-[rgba(255,255,255,0.12)]'
                  }`}
                >
                  <div className="text-2xl mb-2">✨</div>
                  <h3 className="text-[14px] font-semibold text-[#e5e7eb] mb-1">New Tattoo Design</h3>
                  <p className="text-[11px] text-[#6b7280] leading-relaxed">
                    Design a brand new tattoo from scratch. AI analyzes your body and creates original artwork.
                  </p>
                </button>

                <button
                  onClick={() => { setConsultationType('makeup_enhance'); setStep(2); }}
                  className={`border rounded-xl p-6 text-left transition-all ${
                    consultationType === 'makeup_enhance'
                      ? 'border-[#d1d5db] bg-[rgba(255,255,255,0.04)]'
                      : 'border-[rgba(255,255,255,0.06)] hover:border-[rgba(255,255,255,0.12)]'
                  }`}
                >
                  <div className="text-2xl mb-2">🎨</div>
                  <h3 className="text-[14px] font-semibold text-[#e5e7eb] mb-1">Tattoo Makeup</h3>
                  <p className="text-[11px] text-[#6b7280] leading-relaxed">
                    Improve, enhance, or cover up an existing tattoo. AI deeply analyzes your current ink.
                  </p>
                </button>
              </div>
            </div>
          )}

          {/* Step 2: Upload photo */}
          {step === 2 && (
            <div>
              <h2 className="text-[16px] font-semibold text-[#d1d5db] mb-2">
                {consultationType === 'makeup_enhance' ? 'Upload your tattoo photo' : 'Upload body photo'}
              </h2>
              <p className="text-[12px] text-[#6b7280] mb-6">
                {consultationType === 'makeup_enhance'
                  ? 'Take a clear, well-lit photo of your existing tattoo. Show the full tattoo clearly — our AI will perform a detailed forensic analysis.'
                  : 'Take a clear photo of the body area where you want the tattoo. Good lighting helps our AI analyze your skin tone and proportions accurately.'}
              </p>

              <div
                className="border-2 border-dashed border-[rgba(255,255,255,0.08)] rounded-xl p-10 text-center hover:border-[rgba(255,255,255,0.15)] transition-colors cursor-pointer"
                onClick={() => fileInputRef.current?.click()}
              >
                {preview ? (
                  <div>
                    <img src={preview} alt="Preview" className="max-h-64 mx-auto rounded-lg mb-3" />
                    <p className="text-[11px] text-[#6b7280]">Click to change photo</p>
                  </div>
                ) : (
                  <div>
                    <div className="w-12 h-12 rounded-full bg-[rgba(255,255,255,0.04)] flex items-center justify-center mx-auto mb-4">
                      <span className="text-xl text-[#4b5563]">📷</span>
                    </div>
                    <p className="text-[14px] text-[#9ca3af] mb-1">Click to upload photo</p>
                    <p className="text-[11px] text-[#6b7280]">JPG, PNG, WebP — max 10MB</p>
                  </div>
                )}
                <input ref={fileInputRef} type="file" accept="image/*" onChange={handleFileChange} className="hidden" />
              </div>

              {file && (
                <p className="text-[11px] text-[#6b7280] mt-2 text-center">{file.name} ({(file.size / 1024).toFixed(1)}KB)</p>
              )}

              <div className="flex justify-between mt-8">
                <button onClick={() => setStep(1)} className="text-[12px] text-[#6b7280] hover:text-[#9ca3af] transition-colors">
                  ← Back
                </button>
                <button
                  onClick={() => file ? setStep(3) : setError('Please upload a photo first')}
                  className="h-10 px-6 rounded-md bg-[#d1d5db] text-[#121316] text-[13px] font-semibold hover:bg-[#e5e7eb] transition-colors"
                >
                  Continue →
                </button>
              </div>
            </div>
          )}

          {/* Step 3: Describe */}
          {step === 3 && (
            <div>
              <h2 className="text-[16px] font-semibold text-[#d1d5db] mb-2">
                {consultationType === 'makeup_enhance' ? 'What bothers you about this tattoo?' : 'Describe your idea'}
              </h2>
              <p className="text-[12px] text-[#6b7280] mb-6">
                {consultationType === 'makeup_enhance'
                  ? 'Tell us what you want to fix or improve. What do you not like? What style do you prefer? The more detail, the better our AI can plan the perfect approach.'
                  : 'Tell us what you want. The more detail, the better the AI can match your vision.'}
              </p>

              <textarea
                value={ideaText}
                onChange={(e) => setIdeaText(e.target.value)}
                className="w-full min-h-[140px] rounded-lg bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] px-4 py-3 text-[13px] text-[#d1d5db] placeholder-[#6b7280] focus:border-[rgba(255,255,255,0.15)] outline-none resize-y"
                placeholder={
                  consultationType === 'makeup_enhance'
                    ? 'Example: I got this tattoo 5 years ago. The lines are starting to blur and the colors have faded. I want to refresh the black ink, maybe add some floral elements around it, or consider a full cover-up with a Japanese style design...'
                    : 'Example: A black and grey realistic wolf on my left forearm, with pine trees and moon background, 15cm tall, fine detail shading...'
                }
                autoFocus
              />

              <div className="mt-4">
                <p className="text-[11px] text-[#6b7280] mb-2">
                  {consultationType === 'makeup_enhance' ? 'Quick add concern:' : 'Quick add style:'}
                </p>
                <div className="flex flex-wrap gap-1.5">
                  {(consultationType === 'makeup_enhance' ? CONCERN_SUGGESTIONS : STYLE_SUGGESTIONS).map((s) => (
                    <button
                      key={s}
                      type="button"
                      onClick={() => addSuggestion(s)}
                      className="text-[10px] px-2.5 py-1 rounded-md border border-[rgba(255,255,255,0.06)] text-[#9ca3af] hover:border-[rgba(255,255,255,0.12)] hover:text-[#d1d5db] transition-colors"
                    >
                      {s}
                    </button>
                  ))}
                </div>
              </div>

              <div className="flex justify-between mt-8">
                <button onClick={() => setStep(2)} className="text-[12px] text-[#6b7280] hover:text-[#9ca3af] transition-colors">
                  ← Back
                </button>
                <button
                  onClick={() => ideaText.trim() ? setStep(4) : setError(
                    consultationType === 'makeup_enhance'
                      ? 'Please describe what bothers you about the tattoo'
                      : 'Please describe your idea'
                  )}
                  className="h-10 px-6 rounded-md bg-[#d1d5db] text-[#121316] text-[13px] font-semibold hover:bg-[#e5e7eb] transition-colors"
                >
                  Review →
                </button>
              </div>
            </div>
          )}

          {/* Step 4: Review & Submit */}
          {step === 4 && (
            <div>
              <h2 className="text-[16px] font-semibold text-[#d1d5db] mb-6">Review your design request</h2>

              <div className="space-y-4">
                {/* Type */}
                <div className="border border-[rgba(255,255,255,0.06)] rounded-xl p-4">
                  <p className="text-[10px] text-[#6b7280] uppercase tracking-[0.05em] mb-2">Consultation Type</p>
                  <p className="text-[13px] text-[#d1d5db]">
                    {consultationType === 'makeup_enhance' ? '🎨 Tattoo Makeup (Enhance / Cover-up)' : '✨ New Tattoo Design'}
                  </p>
                  <button onClick={() => setStep(1)} className="text-[10px] text-[#6b7280] hover:text-[#9ca3af] mt-2">Change</button>
                </div>

                {/* Photo preview */}
                <div className="border border-[rgba(255,255,255,0.06)] rounded-xl p-4">
                  <p className="text-[10px] text-[#6b7280] uppercase tracking-[0.05em] mb-2">
                    {consultationType === 'makeup_enhance' ? 'Tattoo Photo' : 'Body Photo'}
                  </p>
                  {preview && (
                    <img src={preview} alt="Preview" className="max-h-48 rounded-lg" />
                  )}
                  <button onClick={() => setStep(2)} className="text-[10px] text-[#6b7280] hover:text-[#9ca3af] mt-2">Change</button>
                </div>

                {/* Idea */}
                <div className="border border-[rgba(255,255,255,0.06)] rounded-xl p-4">
                  <p className="text-[10px] text-[#6b7280] uppercase tracking-[0.05em] mb-2">
                    {consultationType === 'makeup_enhance' ? 'Your Concern' : 'Your Idea'}
                  </p>
                  <p className="text-[13px] text-[#d1d5db] leading-relaxed">{ideaText}</p>
                  <button onClick={() => setStep(3)} className="text-[10px] text-[#6b7280] hover:text-[#9ca3af] mt-2">Edit</button>
                </div>

                {/* Summary */}
                <div className="border border-[rgba(255,255,255,0.06)] rounded-xl p-4 bg-[rgba(255,255,255,0.01)]">
                  <p className="text-[10px] text-[#6b7280] uppercase tracking-[0.05em] mb-2">Summary</p>
                  <ul className="text-[12px] text-[#9ca3af] space-y-1">
                    {consultationType === 'makeup_enhance' ? (
                      <>
                        <li>• Deep forensic analysis of your tattoo</li>
                        <li>• 3 improvement approaches</li>
                        <li>• Touch-up, Enhance & Cover-up options</li>
                        <li>• High-resolution previews</li>
                      </>
                    ) : (
                      <>
                        <li>• AI body analysis included</li>
                        <li>• 3 unique design variants</li>
                        <li>• Different artistic styles</li>
                        <li>• High-resolution downloads</li>
                      </>
                    )}
                  </ul>
                </div>
              </div>

              <div className="flex justify-between mt-8">
                <button onClick={() => setStep(3)} className="text-[12px] text-[#6b7280] hover:text-[#9ca3af] transition-colors">
                  ← Back
                </button>
                <button
                  onClick={handleSubmit}
                  disabled={loading}
                  className="h-10 px-8 rounded-md bg-[#d1d5db] text-[#121316] text-[13px] font-semibold hover:bg-[#e5e7eb] transition-colors disabled:opacity-50 flex items-center gap-2"
                >
                  {loading && <span className="w-4 h-4 border-2 border-[#121316] border-t-transparent rounded-full animate-spin" />}
                  {loading ? 'Submitting...' : 'Submit for design →'}
                </button>
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  )
}

export const Route = createFileRoute('/consultations/new')({
  component: NewConsultationPage,
})
