import { Link } from 'react-router-dom'

export default function NotFoundPage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-4 bg-neutral-950 px-5 text-center text-white">
      <p className="text-5xl font-semibold">404</p>
      <p className="text-white/60">页面不存在</p>
      <Link
        className="rounded-full bg-white px-5 py-2 text-sm font-medium text-neutral-950"
        to="/"
      >
        返回首页
      </Link>
    </main>
  )
}
