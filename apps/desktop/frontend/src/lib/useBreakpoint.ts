import { useEffect, useState } from 'react';

/**
 * Tracks the window width and derives desktop layout breakpoints. Used by the
 * app shell to auto-collapse panels so the topology canvas stays primary on
 * small/narrow windows and panels never permanently cover the map.
 *
 * Breakpoints (logical CSS px, so they already account for Windows scaling):
 *   - compact: < 1024  → collapse the bottom status/log row, hide the sidebar
 *   - narrow:  < 1280  → the details panel overlays the canvas instead of
 *                        taking its own column (keeps the map full-width)
 *   - wide:    ≥ 1680  → room for sidebar + canvas + details side-by-side
 */
export interface Breakpoint {
  width: number;
  isCompact: boolean;
  isNarrow: boolean;
  isWide: boolean;
}

function read(): number {
  return typeof window === 'undefined' ? 1920 : window.innerWidth;
}

export function useBreakpoint(): Breakpoint {
  const [width, setWidth] = useState<number>(read);

  useEffect(() => {
    let raf = 0;
    const onResize = () => {
      // Coalesce resize bursts to a single state update per frame.
      cancelAnimationFrame(raf);
      raf = requestAnimationFrame(() => setWidth(window.innerWidth));
    };
    window.addEventListener('resize', onResize);
    return () => {
      window.removeEventListener('resize', onResize);
      cancelAnimationFrame(raf);
    };
  }, []);

  return {
    width,
    isCompact: width < 1024,
    isNarrow: width < 1280,
    isWide: width >= 1680,
  };
}
