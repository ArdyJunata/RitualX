// frontend/src/app/(auth)/login/page.tsx

import type { Metadata } from 'next'
import { LoginForm } from '@/modules/auth/components/LoginForm'

export const metadata: Metadata = {
  title: 'Sign in — RitualX',
  description: 'Sign in to RitualX and continue your habit journey.',
}

export default function LoginPage() {
  return <LoginForm />
}
