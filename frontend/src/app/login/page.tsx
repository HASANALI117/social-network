 'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import Link from 'next/link'
import { Fieldset, Field, ErrorMessage } from '@/components/ui/fieldset'

const formSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  rememberMe: z.boolean().optional()
})

type FormValues = z.infer<typeof formSchema>

export default function LoginPage() {
  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
    setValue,
  } = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      email: '',
      password: '',
      rememberMe: false
    }
  })

  return (
    <div className="max-w-md mx-auto my-12 p-6 bg-white rounded-lg shadow-md dark:bg-zinc-900">
      <h1 className="text-2xl font-bold mb-6 text-center">Sign In</h1>
      <form onSubmit={handleSubmit((data: FormValues) => console.log(data))}>
        <Fieldset className="space-y-6">
          <Field>
            <label className="block text-sm font-medium mb-1" htmlFor="email">
              Email *
            </label>
            <Input
              id="email"
              type="email"
              placeholder="john@example.com"
              {...register('email')}
            />
            {errors.email?.message && (
              <ErrorMessage>{errors.email.message}</ErrorMessage>
            )}
          </Field>

          <Field>
            <label className="block text-sm font-medium mb-1" htmlFor="password">
              Password *
            </label>
            <Input
              id="password"
              type="password"
              placeholder="••••••••"
              {...register('password')}
            />
            {errors.password?.message && (
              <ErrorMessage>{errors.password.message}</ErrorMessage>
            )}
          </Field>

          <Field>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="rememberMe"
                  {...register('rememberMe')}
                  className="hidden"
                />
                <div 
                  onClick={() => setValue('rememberMe', !watch('rememberMe'))}
                  className={`w-5 h-5 rounded border cursor-pointer ${
                    watch('rememberMe') 
                      ? 'bg-blue-600 border-blue-600' 
                      : 'border-gray-300 dark:border-gray-600'
                  }`}
                >
                  {watch('rememberMe') && (
                    <svg className="w-4 h-4 text-white mx-auto" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>
                  )}
                </div>
                <label htmlFor="rememberMe" className="text-sm">
                  Remember me
                </label>
              </div>
              <Link href="/forgot-password" className="text-sm text-blue-600 hover:underline">
                Forgot password?
              </Link>
            </div>
          </Field>
        </Fieldset>

        <div className="mt-6">
          <Button 
            type="submit"
            className="w-full bg-blue-600 hover:bg-blue-700"
          >
            Sign In
          </Button>
        </div>
      </form>

      <p className="mt-4 text-center text-sm">
        Don't have an account? <Link href="/register" className="text-blue-600 hover:underline">Register</Link>
      </p>
    </div>
  )
}
