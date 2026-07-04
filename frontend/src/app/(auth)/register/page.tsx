// frontend/src/app/(auth)/register/page.tsx

import type { Metadata } from 'next'
import { RegisterForm } from '@/modules/auth/components/RegisterForm'

export const metadata: Metadata = {
  title: 'Create account — RitualX',
  description: 'Join RitualX and start building powerful habits today.',
}

export default function RegisterPage() {
  return <RegisterForm />
}
