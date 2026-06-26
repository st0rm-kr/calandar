type HorizontalWheelCallback = (node: HTMLElement) => void

export function bindExclusiveHorizontalWheel(
  node: HTMLElement,
  onScroll?: HorizontalWheelCallback,
): () => void {
  const listener = (event: WheelEvent) => {
    const delta = event.deltaY || event.deltaX
    if (delta === 0) {
      return
    }

    event.preventDefault()
    event.stopPropagation()
    node.scrollLeft += delta
    onScroll?.(node)
  }

  node.addEventListener('wheel', listener, { passive: false })
  return () => node.removeEventListener('wheel', listener)
}
