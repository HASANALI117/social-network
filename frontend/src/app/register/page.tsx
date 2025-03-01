'use client'

import { useForm } from 'react-hook-form'  
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import Link from 'next/link'
import { Fieldset, Field, ErrorMessage } from '@/components/ui/fieldset'

const formSchema = z.object({
  firstName: z.string().min(2, 'First name must be at least 2 characters'),
  lastName: z.string().min(2, 'Last name must be at least 2 characters'),
  email: z.string().email('Invalid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  dob: z.date().refine(date => date <= new Date(), 'Date of birth cannot be in the future'),
  avatar: z.instanceof(FileList).optional(),
  nickname: z.string().optional(),
  bio: z.string().optional(),
  terms: z.literal<boolean>(true, {
    errorMap: () => ({ message: 'You must accept the terms and conditions' })
  })
})

type FormValues = z.infer<typeof formSchema>

export default function RegisterPage() {
  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
    setValue,
  } = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      firstName: '',
      lastName: '',
      email: '',
      password: '',
      terms: false
    }
  })
  return (
    <div className="max-w-md mx-auto my-12 p-6 bg-white rounded-lg shadow-md dark:bg-zinc-900">
      <h1 className="text-2xl font-bold mb-6 text-center">Create Account</h1>
      <form onSubmit={handleSubmit((data: FormValues) => console.log(data))}>
        <Fieldset className="space-y-6">
          <div className="grid grid-cols-2 gap-4">
          <Field>
            <label className="block text-sm font-medium mb-1" htmlFor="first-name">
              First Name *
            </label>
            <Input
              id="first-name"
              placeholder="John"
              {...register('firstName')}
            />
            {errors.firstName?.message && (
              <ErrorMessage>{errors.firstName.message}</ErrorMessage>
            )}
          </Field>
          <Field>
            <label className="block text-sm font-medium mb-1" htmlFor="last-name">
              Last Name *
            </label>
            <Input
              id="last-name"
              placeholder="Doe"
              {...register('lastName')}
            />
            {errors.lastName?.message && (
              <ErrorMessage>{errors.lastName.message}</ErrorMessage>
            )}
          </Field>
        </div>

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
          <label className="block text-sm font-medium mb-1" htmlFor="dob">
            Date of Birth *
          </label>
          <Input
            id="dob"
            type="date"
            max={new Date().toISOString().split('T')[0]}
            {...register('dob', {
              valueAsDate: true
            })}
          />
          {errors.dob?.message && (
            <ErrorMessage>{errors.dob.message}</ErrorMessage>
          )}
        </Field>

        <Field>
          <label className="block text-sm font-medium mb-1" htmlFor="avatar">
            Profile Picture (Optional)
          </label>
          <Input
            id="avatar"
            type="file"
            accept="image/*"
            {...register('avatar')}
          />
        </Field>

        <Field>
          <label className="block text-sm font-medium mb-1" htmlFor="nickname">
            Nickname (Optional)
          </label>
          <Input
            id="nickname"
            placeholder="CoolDude123"
            {...register('nickname')}
          />
        </Field>

        <Field>
          <label className="block text-sm font-medium mb-1" htmlFor="bio">
            About Me (Optional)
          </label>
          <textarea
            id="bio"
            className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent dark:bg-zinc-800 dark:border-zinc-700"
            rows={3}
            placeholder="Tell us something about yourself..."
            {...register('bio')}
          />
        </Field>

        <Field>
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="terms"
                {...register('terms')}
                className="hidden"
              />
              <div 
                onClick={() => setValue('terms', !watch('terms'))}
                className={`w-5 h-5 rounded border cursor-pointer ${
                  watch('terms') 
                    ? 'bg-blue-600 border-blue-600' 
                    : 'border-gray-300 dark:border-gray-600'
                }`}
              >
                {watch('terms') && (
                  <svg className="w-4 h-4 text-white mx-auto" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                )}
              </div>
              <label htmlFor="terms" className="text-sm">
                I agree to the <Link href="/terms" className="text-blue-600 hover:underline">Terms of Service</Link>
              </label>
            </div>
          </div>
          {errors.terms?.message && (
            <ErrorMessage>{errors.terms.message}</ErrorMessage>
          )}
        </Field>

        </Fieldset>
        <div className="mt-6">
          <Button 
            type="submit"
            className="w-full bg-blue-600 hover:bg-blue-700"
          >
            Create Account
          </Button>
        </div>
      </form>

      <p className="mt-4 text-center text-sm">
        Already have an account? <Link href="/login" className="text-blue-600 hover:underline">Log in</Link>
      </p>
    </div>
  )
}
