import { describe, expect, it, vi } from 'vitest'
import { bindExclusiveHorizontalWheel } from './exclusiveWheel'

describe('bindExclusiveHorizontalWheel', () => {
  it('keeps wheel movement inside the horizontal scroller', () => {
    const scroller = document.createElement('div')
    const onScroll = vi.fn()

    const cleanup = bindExclusiveHorizontalWheel(scroller, onScroll)
    const event = new WheelEvent('wheel', {
      cancelable: true,
      deltaY: 96,
    })

    const propagated = scroller.dispatchEvent(event)

    expect(propagated).toBe(false)
    expect(event.defaultPrevented).toBe(true)
    expect(scroller.scrollLeft).toBe(96)
    expect(onScroll).toHaveBeenCalledWith(scroller)

    cleanup()
  })
})
