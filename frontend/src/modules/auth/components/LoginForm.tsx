// frontend/src/modules/auth/components/LoginForm.tsx

'use client'

import Link from 'next/link'
import { Input } from '@/shared/components/ui/Input'
import { Button } from '@/shared/components/ui/Button'
import { useLoginForm } from '../hooks/useLoginForm'

export function LoginForm() {
  const { fields, errors, isLoading, handleChange, handleSubmit } = useLoginForm()

  return (
    <div className="w-full max-w-[420px] px-4">
      {/* Logo */}
      <div className="mb-8 text-center">
        <h1 className="font-heading text-3xl font-bold tracking-tight text-white">
          Ritual<span className="text-emerald">X</span>
        </h1>
        <p className="mt-2 text-sm text-zinc-400">Track. Level up. Dominate.</p>
      </div>

      {/* Card */}
      <div className="glass-card p-6">
        <form onSubmit={handleSubmit} noValidate className="flex flex-col gap-4">
          {errors.general && (
            <div role="alert" className="rounded-lg bg-red/10 border border-red/30 px-4 py-3 text-sm text-red">
              {errors.general}
            </div>
          )}

          <Input
            id="login-email"
            name="email"
            type="email"
            label="Email"
            placeholder="you@example.com"
            autoComplete="email"
            value={fields.email}
            onChange={handleChange}
            error={errors.email}
          />

          <Input
            id="login-password"
            name="password"
            type="password"
            label="Password"
            placeholder="••••••••"
            autoComplete="current-password"
            value={fields.password}
            onChange={handleChange}
            error={errors.password}
          />

          <Button
            id="login-submit"
            type="submit"
            isLoading={isLoading}
            className="mt-2"
          >
            {isLoading ? 'Signing in…' : 'Sign in'}
          </Button>
        </form>
      </div>

      {/* Switch link */}
      <p className="mt-6 text-center text-sm text-zinc-400">
        Don&apos;t have an account?{' '}
        <Link href="/register" className="text-emerald hover:text-emerald-300 font-medium transition-colors">
          Register
        </Link>
      </p>
    </div>
  )
}
