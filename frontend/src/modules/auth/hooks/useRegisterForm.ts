// frontend/src/modules/auth/hooks/useRegisterForm.ts

'use client'

import { useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { ApiError } from '@/shared/api-client'
import { authApi } from '../api'
import { RegisterRequest } from '../types'

export interface RegisterFormFields {
  display_name: string
  username: string
  email: string
  password: string
  confirm_password: string
}

export interface RegisterFormErrors {
  display_name?: string
  username?: string
  email?: string
  password?: string
  confirm_password?: string
  general?: string
}

export interface UseRegisterFormReturn {
  fields: RegisterFormFields
  errors: RegisterFormErrors
  isLoading: boolean
  handleChange: (e: React.ChangeEvent<HTMLInputElement>) => void
  handleSubmit: (e: React.FormEvent<HTMLFormElement>) => Promise<void>
}

function validateRegister(fields: RegisterFormFields): RegisterFormErrors {
  const errors: RegisterFormErrors = {}

  if (!fields.display_name.trim()) {
    errors.display_name = 'Display name is required'
  } else if (fields.display_name.trim().length < 2 || fields.display_name.trim().length > 50) {
    errors.display_name = 'Display name must be 2–50 characters'
  }

  if (!fields.username.trim()) {
    errors.username = 'Username is required'
  } else if (fields.username.length < 3 || fields.username.length > 30) {
    errors.username = 'Username must be 3–30 characters'
  } else if (!/^[a-z0-9_]+$/i.test(fields.username)) {
    errors.username = 'Username may only contain letters, numbers, and underscores'
  }

  if (!fields.email.trim()) {
    errors.email = 'Email is required'
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(fields.email)) {
    errors.email = 'Enter a valid email address'
  }

  if (!fields.password) {
    errors.password = 'Password is required'
  } else if (fields.password.length < 8) {
    errors.password = 'Password must be at least 8 characters'
  }

  if (!fields.confirm_password) {
    errors.confirm_password = 'Please confirm your password'
  } else if (fields.confirm_password !== fields.password) {
    errors.confirm_password = 'Passwords do not match'
  }

  return errors
}

export function useRegisterForm(): UseRegisterFormReturn {
  const router = useRouter()
  const [fields, setFields] = useState<RegisterFormFields>({
    display_name: '',
    username: '',
    email: '',
    password: '',
    confirm_password: '',
  })
  const [errors, setErrors] = useState<RegisterFormErrors>({})
  const [isLoading, setIsLoading] = useState(false)

  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setFields(prev => ({ ...prev, [name]: value }))
  }, [])

  const handleSubmit = useCallback(async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setErrors({})

    const clientErrors = validateRegister(fields)
    if (Object.keys(clientErrors).length > 0) {
      setErrors(clientErrors)
      return
    }

    setIsLoading(true)
    try {
      const body: RegisterRequest = {
        display_name: fields.display_name,
        username: fields.username,
        email: fields.email,
        password: fields.password,
      }
      const res = await authApi.register(body)
      localStorage.setItem('ritualx_access_token', res.access_token)
      router.replace('/')
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.hasFieldErrors && err.fields) {
          const mapped: RegisterFormErrors = {}
          if (err.fields.display_name) mapped.display_name = err.fields.display_name[0]
          if (err.fields.username) mapped.username = err.fields.username[0]
          if (err.fields.email) mapped.email = err.fields.email[0]
          if (err.fields.password) mapped.password = err.fields.password[0]
          setErrors(mapped)
        } else {
          setErrors({ general: err.message || 'Registration failed. Please try again.' })
        }
      } else {
        setErrors({ general: 'An unexpected error occurred.' })
      }
    } finally {
      setIsLoading(false)
    }
  }, [fields, router])

  return { fields, errors, isLoading, handleChange, handleSubmit }
}
