import { useCallback, useEffect, useState } from 'react'

export type AsyncState<T> = {
  data: T | null
  error: string | null
  loading: boolean
  reload: () => void
}

export function useAsync<T>(
  fn: () => Promise<T>,
  deps: unknown[] = [],
): AsyncState<T> {
  const [data, setData] = useState<T | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [nonce, setNonce] = useState(0)

  const reload = useCallback(() => setNonce((value) => value + 1), [])

  useEffect(() => {
    let active = true
    Promise.resolve().then(() => {
      if (active) {
        setLoading(true)
      }
    })
    fn()
      .then((result) => {
        if (!active) {
          return
        }
        setData(result)
        setError(null)
      })
      .catch((err: unknown) => {
        if (!active) {
          return
        }
        setData(null)
        setError(err instanceof Error ? err.message : '加载失败')
      })
      .finally(() => {
        if (active) {
          setLoading(false)
        }
      })
    return () => {
      active = false
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...deps, nonce])

  return { data, error, loading, reload }
}
