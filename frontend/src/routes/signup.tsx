import { useState } from 'react';
import { useNavigate, createFileRoute } from '@tanstack/react-router';
import { register } from '../lib/auth';

function SignupPage() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await register(email, password, name);
      navigate({ to: '/dashboard' });
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-[#121316]">
      <div className="w-full max-w-[360px]">
        <div className="text-center mb-8">
          <h1 className="text-[14px] font-semibold text-[#e5e7eb] tracking-tight">
            Create Account
          </h1>
          <p className="text-[12px] text-[#6b7280] mt-1">Start your tattoo design journey</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="text-[12px] text-[#ef4444] bg-[rgba(239,68,68,0.08)] border border-[rgba(239,68,68,0.15)] rounded-md px-3 py-2">
              {error}
            </div>
          )}

          <div>
            <label className="block text-[11px] font-medium text-[#9ca3af] mb-1.5">Name</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="input-field w-full"
              placeholder="Your name"
              required
              autoFocus
            />
          </div>

          <div>
            <label className="block text-[11px] font-medium text-[#9ca3af] mb-1.5">Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="input-field w-full"
              placeholder="you@example.com"
              required
            />
          </div>

          <div>
            <label className="block text-[11px] font-medium text-[#9ca3af] mb-1.5">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="input-field w-full"
              placeholder="••••••••"
              required
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="btn-primary w-full flex items-center justify-center gap-2"
          >
            {loading ? (
              <span className="inline-block w-4 h-4 border-2 border-[#121316] border-t-transparent rounded-full animate-spin" />
            ) : null}
            Create Account
          </button>
        </form>

        <p className="text-center text-[12px] text-[#6b7280] mt-6">
          Already have an account?{' '}
          <a href="/login" className="text-[#9ca3af] hover:text-[#d1d5db] underline underline-offset-2">
            Sign in
          </a>
        </p>
      </div>
    </div>
  );
}

export const Route = createFileRoute('/signup')({
  component: SignupPage,
});
