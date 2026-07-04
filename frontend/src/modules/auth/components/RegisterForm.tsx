// frontend/src/modules/auth/components/RegisterForm.tsx

'use client'

import Link from 'next/link'
import { Input } from '@/shared/components/ui/Input'
import { Button } from '@/shared/components/ui/Button'
import { useRegisterForm } from '../hooks/useRegisterForm'

export function RegisterForm() {
  const { fields, errors, isLoading, handleChange, handleSubmit } = useRegisterForm()

  return (
    <div className="w-full max-w-[420px] px-4">
      {/* Logo */}
      <div className="mb-8 text-center">
        <h1 className="font-heading text-3xl font-bold tracking-tight text-white">
          Ritual<span className="text-emerald">X</span>
        </h1>
        <p className="mt-2 text-sm text-zinc-400">Your ritual starts here.</p>
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
            id="register-display-name"
            name="display_name"
            type="text"
            label="Display Name"
            placeholder="Your Name"
            autoComplete="name"
            value={fields.display_name}
            onChange={handleChange}
            error={errors.display_name}
          />

          <Input
            id="register-username"
            name="username"
            type="text"
            label="Username"
            placeholder="your_handle"
            autoComplete="username"
            value={fields.username}
            onChange={handleChange}
            error={errors.username}
          />

          <Input
            id="register-email"
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
            id="register-password"
            name="password"
            type="password"
            label="Password"
            placeholder="Min. 8 characters"
            autoComplete="new-password"
            value={fields.password}
            onChange={handleChange}
            error={errors.password}
          />

          <Input
            id="register-confirm-password"
            name="confirm_password"
            type="password"
            label="Confirm Password"
            placeholder="Repeat password"
            autoComplete="new-password"
            value={fields.confirm_password}
            onChange={handleChange}
            error={errors.confirm_password}
          />

          <Button
            id="register-submit"
            type="submit"
            isLoading={isLoading}
            className="mt-2"
          >
            {isLoading ? 'Creating account…' : 'Create account'}
          </Button>
        </form>
      </div>

      {/* Switch link */}
      <p className="mt-6 text-center text-sm text-zinc-400">
        Already have an account?{' '}
        <Link href="/login" className="text-emerald hover:text-emerald-300 font-medium transition-colors">
          Sign in
        </Link>
      </p>
    </div>
  )
}
