import { useEffect } from "react";

interface HotjarFunction {
  (...args: unknown[]): void;
  q?: unknown[][];
}

declare global {
  interface Window {
    hj?: HotjarFunction;
    _hjSettings?: { hjid: number; hjsv: number };
  }
}

export function useHotjar() {
  useEffect(() => {
    const hotjarId = import.meta.env.VITE_HOTJAR_ID;

    if (!hotjarId) {
      return;
    }

    const hjid = parseInt(hotjarId, 10);
    if (isNaN(hjid)) {
      console.warn("Invalid Hotjar ID:", hotjarId);
      return;
    }

    // Skip if already initialized
    if (window._hjSettings) {
      return;
    }

    const hj: HotjarFunction = function (...args: unknown[]) {
      (hj.q = hj.q || []).push(args);
    };
    window.hj = hj;
    window._hjSettings = { hjid, hjsv: 6 };

    const script = document.createElement("script");
    script.async = true;
    script.src = `https://static.hotjar.com/c/hotjar-${hjid}.js?sv=6`;
    document.head.appendChild(script);
  }, []);
}
