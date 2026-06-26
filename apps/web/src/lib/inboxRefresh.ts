const inboxRefreshEvent = 'hangout:inbox-refresh'

export function emitInboxRefresh() {
  if (typeof window === 'undefined') {
    return
  }
  window.dispatchEvent(new Event(inboxRefreshEvent))
}

export function subscribeInboxRefresh(listener: () => void): () => void {
  if (typeof window === 'undefined') {
    return () => {}
  }
  window.addEventListener(inboxRefreshEvent, listener)
  return () => window.removeEventListener(inboxRefreshEvent, listener)
}
