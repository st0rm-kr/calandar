export type ApiEnvelope<T> = {
  data: T | null
  error: { code: string; message: string } | null
}

export async function apiGet<T>(path: string): Promise<T> {
  const response = await fetch(path)
  const body = (await response.json()) as ApiEnvelope<T>
  if (!response.ok || body.error) {
    throw new Error(body.error?.message ?? `Request failed: ${response.status}`)
  }
  if (body.data === null) {
    throw new Error('Response data is empty')
  }
  return body.data
}
