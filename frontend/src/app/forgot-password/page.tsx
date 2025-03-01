'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import Link from 'next/link'
import { Fieldset, Field, ErrorMessage } from '@/components/ui/fieldset'

const formSchema = z.object({
  email: z.string().email('Invalid email address')
})

type FormValues = z.infer<typeof formSchema>

export default function ForgotPasswordPage() {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      email: '',
    }
  })

  return (
    <div className="max-w-md mx-auto my-12 p-6 bg-white rounded-lg shadow-md dark:bg-zinc-900">
      <h1 className="text-2xl font-bold mb-6 text-center">Reset Password</h1>
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
        </Fieldset>

        <div className="mt-6">
          <Button 
            type="submit"
            className="w-full bg-blue-600 hover:bg-blue-700"
          >
            Send Reset Link
          </Button>
        </div>
      </form>

      <p className="mt-4 text-center text-sm">
        Remember your password? <Link href="/login" className="text-blue-600 hover:underline">Log in</Link>
      </p>
    </div>
  )
}
