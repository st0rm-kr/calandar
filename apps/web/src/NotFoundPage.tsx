import { Link } from 'react-router-dom'

export default function NotFoundPage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-4 bg-canvas px-5 text-center text-ink">
      <p className="font-display text-6xl font-bold text-brand">404</p>
      <p className="font-semibold text-muted">页面不存在</p>
      <Link
        className="rounded-full bg-brand px-5 py-2.5 text-sm font-bold text-white transition-colors duration-200 hover:bg-brand-ink"
        to="/"
      >
        返回首页
      </Link>
    </main>
  )
}
