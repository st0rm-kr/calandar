import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { supabase } from './lib/supabase'

type AuthStatus = 'idle' | 'loading' | 'success' | 'error'

export default function LoginPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const redirect = searchParams.get('redirect') || '/'
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [status, setStatus] = useState<AuthStatus>('idle')
  const [message, setMessage] = useState('')

  async function runAuthAction(
    action: () => Promise<{ error: { message: string } | null }>,
    success: string,
    onSuccess?: () => void,
  ) {
    setStatus('loading')
    setMessage('')
    const { error } = await action()
    if (error) {
      setStatus('error')
      setMessage(error.message)
      return
    }
    setStatus('success')
    setMessage(success)
    onSuccess?.()
  }

  async function handleLogin(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    await runAuthAction(
      () => supabase.auth.signInWithPassword({ email, password }),
      '登录成功',
      () => navigate(redirect, { replace: true }),
    )
  }

  async function handleRegister() {
    await runAuthAction(() => supabase.auth.signUp({ email, password }), '注册邮件已发送')
  }

  async function handleResetPassword() {
    await runAuthAction(() => supabase.auth.resetPasswordForEmail(email), '重置密码邮件已发送')
  }

  async function handleSignOut() {
    await runAuthAction(() => supabase.auth.signOut(), '已退出登录')
  }

  return (
    <main className="flex min-h-screen items-center bg-canvas px-5 py-10 text-ink">
      <section className="mx-auto w-full max-w-md rounded-4xl border border-hairline bg-surface p-7 shadow-soft">
        <Link className="text-sm font-semibold text-brand" to="/">
          Hangout
        </Link>
        <h1 className="mt-3 text-3xl font-bold tracking-tight">登录 Hangout</h1>
        <p className="mt-1 text-sm text-muted">登录后约局、报名、自动进日历。</p>
        <form className="mt-6 space-y-4" onSubmit={handleLogin}>
          <label className="block text-sm font-semibold" htmlFor="email">
            邮箱
          </label>
          <input
            className="w-full rounded-2xl border border-hairline bg-canvas px-4 py-3 text-ink outline-none transition-colors duration-200 focus:border-brand"
            id="email"
            onChange={(event) => setEmail(event.target.value)}
            required
            type="email"
            value={email}
          />

          <label className="block text-sm font-semibold" htmlFor="password">
            密码
          </label>
          <input
            className="w-full rounded-2xl border border-hairline bg-canvas px-4 py-3 text-ink outline-none transition-colors duration-200 focus:border-brand"
            id="password"
            minLength={6}
            onChange={(event) => setPassword(event.target.value)}
            required
            type="password"
            value={password}
          />

          <button
            className="w-full cursor-pointer rounded-2xl bg-brand px-4 py-3 font-bold text-white transition-colors duration-200 hover:bg-brand-ink disabled:cursor-not-allowed disabled:opacity-50"
            disabled={status === 'loading'}
            type="submit"
          >
            登录
          </button>
        </form>

        <div className="mt-4 grid gap-3 sm:grid-cols-3">
          <button
            className="cursor-pointer rounded-2xl border border-hairline px-4 py-3 font-semibold text-muted transition-colors duration-200 hover:text-ink disabled:opacity-50"
            disabled={status === 'loading'}
            onClick={handleRegister}
            type="button"
          >
            注册
          </button>
          <button
            className="cursor-pointer rounded-2xl border border-hairline px-4 py-3 font-semibold text-muted transition-colors duration-200 hover:text-ink disabled:opacity-50"
            disabled={status === 'loading' || email === ''}
            onClick={handleResetPassword}
            type="button"
          >
            重置密码
          </button>
          <button
            className="cursor-pointer rounded-2xl border border-hairline px-4 py-3 font-semibold text-muted transition-colors duration-200 hover:text-ink disabled:opacity-50"
            disabled={status === 'loading'}
            onClick={handleSignOut}
            type="button"
          >
            退出
          </button>
        </div>

        {message !== '' ? (
          <p
            className={
              status === 'error'
                ? 'mt-4 text-sm font-semibold text-rose'
                : 'mt-4 text-sm font-semibold text-grass'
            }
          >
            {message}
          </p>
        ) : null}
      </section>
    </main>
  )
}
