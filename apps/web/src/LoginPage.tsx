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
    <main className="min-h-screen bg-neutral-950 px-5 py-8 text-white">
      <section className="mx-auto max-w-md rounded-3xl bg-white/10 p-6 shadow-xl">
        <Link className="text-sm text-white/60" to="/">
          Hangout
        </Link>
        <h1 className="mt-3 text-3xl font-semibold">登录 Hangout</h1>
        <form className="mt-6 space-y-4" onSubmit={handleLogin}>
          <label className="block text-sm font-medium" htmlFor="email">
            邮箱
          </label>
          <input
            className="w-full rounded-2xl border border-white/15 bg-white/10 px-4 py-3 text-white outline-none focus:border-white"
            id="email"
            onChange={(event) => setEmail(event.target.value)}
            required
            type="email"
            value={email}
          />

          <label className="block text-sm font-medium" htmlFor="password">
            密码
          </label>
          <input
            className="w-full rounded-2xl border border-white/15 bg-white/10 px-4 py-3 text-white outline-none focus:border-white"
            id="password"
            minLength={6}
            onChange={(event) => setPassword(event.target.value)}
            required
            type="password"
            value={password}
          />

          <button className="w-full rounded-2xl bg-white px-4 py-3 font-semibold text-neutral-950" disabled={status === 'loading'} type="submit">
            登录
          </button>
        </form>

        <div className="mt-4 grid gap-3 sm:grid-cols-3">
          <button className="rounded-2xl border border-white/20 px-4 py-3" disabled={status === 'loading'} onClick={handleRegister} type="button">
            注册
          </button>
          <button className="rounded-2xl border border-white/20 px-4 py-3" disabled={status === 'loading' || email === ''} onClick={handleResetPassword} type="button">
            重置密码
          </button>
          <button className="rounded-2xl border border-white/20 px-4 py-3" disabled={status === 'loading'} onClick={handleSignOut} type="button">
            退出
          </button>
        </div>

        {message !== '' ? (
          <p className={status === 'error' ? 'mt-4 text-sm text-red-300' : 'mt-4 text-sm text-emerald-300'}>
            {message}
          </p>
        ) : null}
      </section>
    </main>
  )
}
