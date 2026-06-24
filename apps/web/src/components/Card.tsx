import type { ReactNode } from 'react'

type CardProps = {
  children: ReactNode
  className?: string
}

export function Card({ children, className = '' }: CardProps) {
  return (
    <div
      className={`rounded-3xl border border-hairline bg-surface p-5 shadow-card ${className}`}
    >
      {children}
    </div>
  )
}
