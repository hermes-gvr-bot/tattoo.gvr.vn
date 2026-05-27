import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { isAuthenticated } from '../lib/auth'
import { useEffect, useState } from 'react'

function LandingPage() {
  const navigate = useNavigate()
  const [loggedIn, setLoggedIn] = useState(false)

  useEffect(() => {
    setLoggedIn(isAuthenticated())
  }, [])

  return (
    <div className="min-h-screen bg-[#121316] text-[#d1d5db]">
      {/* Nav */}
      <nav className="h-14 flex items-center justify-between px-6 border-b border-[rgba(255,255,255,0.06)] max-w-6xl mx-auto">
        <span className="text-[14px] font-semibold text-[#e5e7eb] tracking-tight">
          Tattoo<span className="text-[#9ca3af] font-normal">AI</span>
        </span>
        <div className="flex items-center gap-3">
          {loggedIn ? (
            <button onClick={() => navigate({ to: '/dashboard' })} className="h-9 px-4 rounded-md bg-[#d1d5db] text-[#121316] text-[12px] font-semibold hover:bg-[#e5e7eb] transition-colors">
              Dashboard
            </button>
          ) : (
            <>
              <button onClick={() => navigate({ to: '/login' })} className="text-[12px] text-[#9ca3af] hover:text-[#d1d5db] transition-colors">
                Sign in
              </button>
              <button onClick={() => navigate({ to: '/signup' })} className="h-9 px-4 rounded-md bg-[#d1d5db] text-[#121316] text-[12px] font-semibold hover:bg-[#e5e7eb] transition-colors">
                Get started
              </button>
            </>
          )}
        </div>
      </nav>

      {/* Hero */}
      <section className="max-w-4xl mx-auto px-6 pt-24 pb-16 text-center">
        <h1 className="text-[36px] font-bold text-[#e5e7eb] leading-tight mb-4 tracking-tight">
          Your tattoo, <span className="text-[#9ca3af]">designed by AI</span>
        </h1>
        <p className="text-[15px] text-[#6b7280] max-w-lg mx-auto leading-relaxed mb-8">
          Upload a photo of your body, describe your idea, and get 3 unique tattoo designs
          in minutes — each crafted by AI specifically for your body and style.
        </p>
        <button
          onClick={() => navigate({ to: loggedIn ? '/consultations/new' : '/signup' })}
          className="h-11 px-8 rounded-md bg-[#d1d5db] text-[#121316] text-[14px] font-semibold hover:bg-[#e5e7eb] transition-colors"
        >
          {loggedIn ? 'Start new design →' : 'Get your first design →'}
        </button>
        <div className="mt-12 aspect-[3/1] bg-[rgba(255,255,255,0.02)] rounded-xl border border-[rgba(255,255,255,0.04)] flex items-center justify-center">
          <div className="flex gap-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="w-40 h-40 bg-[rgba(255,255,255,0.03)] rounded-lg flex items-center justify-center">
                <span className="text-[11px] text-[#4b5563]">Sample {i}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* How it works */}
      <section className="max-w-4xl mx-auto px-6 py-16 border-t border-[rgba(255,255,255,0.04)]">
        <h2 className="text-[13px] font-semibold text-[#9ca3af] uppercase tracking-[0.05em] mb-8 text-center">
          How it works
        </h2>
        <div className="grid grid-cols-3 gap-8">
          {[
            { step: '1', title: 'Upload a photo', desc: 'Take a clear photo of the body area where you want the tattoo. Our AI analyzes your skin tone, muscle structure, and proportions.' },
            { step: '2', title: 'Describe your idea', desc: 'Tell us what you want — style, elements, mood. Traditional Irezumi dragon? Minimalist geometric wolf? Watercolor cherry blossoms?' },
            { step: '3', title: 'Get 3 AI designs', desc: 'Our AI generates 3 unique tattoo designs in different styles. Compare, choose your favorite, and take it to your artist.' },
          ].map((item) => (
            <div key={item.step} className="text-center">
              <div className="w-10 h-10 rounded-full bg-[rgba(255,255,255,0.04)] flex items-center justify-center mx-auto mb-4">
                <span className="text-[13px] font-semibold text-[#9ca3af]">{item.step}</span>
              </div>
              <h3 className="text-[14px] font-medium text-[#d1d5db] mb-2">{item.title}</h3>
              <p className="text-[12px] text-[#6b7280] leading-relaxed">{item.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Pricing */}
      <section className="max-w-4xl mx-auto px-6 py-16 border-t border-[rgba(255,255,255,0.04)]">
        <h2 className="text-[13px] font-semibold text-[#9ca3af] uppercase tracking-[0.05em] mb-8 text-center">
          Pricing
        </h2>
        <div className="max-w-sm mx-auto">
          <div className="border border-[rgba(255,255,255,0.08)] rounded-xl p-8 text-center bg-[rgba(255,255,255,0.02)]">
            <p className="text-[11px] text-[#9ca3af] uppercase tracking-[0.05em] mb-3">Per design</p>
            <p className="text-[36px] font-bold text-[#e5e7eb] mb-1">$5</p>
            <p className="text-[13px] text-[#6b7280] mb-6">per consultation</p>
            <ul className="space-y-2 mb-8 text-left">
              {[
                'AI body analysis',
                '3 unique design variants',
                'Multiple artistic styles',
                'High-resolution downloads',
                'Take to any tattoo artist',
              ].map((f) => (
                <li key={f} className="text-[12px] text-[#9ca3af] flex items-center gap-2">
                  <span className="text-[#4b5563]">✓</span> {f}
                </li>
              ))}
            </ul>
            <button
              onClick={() => navigate({ to: loggedIn ? '/consultations/new' : '/signup' })}
              className="h-10 w-full rounded-md bg-[#d1d5db] text-[#121316] text-[13px] font-semibold hover:bg-[#e5e7eb] transition-colors"
            >
              {loggedIn ? 'Start designing →' : 'Get started →'}
            </button>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-[rgba(255,255,255,0.04)] py-8 text-center">
        <p className="text-[11px] text-[#4b5563]">
          TattooAI · AI-powered tattoo design consultation · © 2026
        </p>
      </footer>
    </div>
  )
}

export const Route = createFileRoute('/')({
  component: LandingPage,
})
