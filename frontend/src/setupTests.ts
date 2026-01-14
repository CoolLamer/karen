import '@testing-library/jest-dom'

// Mock import.meta.env for tests
Object.defineProperty(import.meta, 'env', {
  value: {
    VITE_API_BASE_URL: 'http://localhost:8080',
  },
})

// Mock window.matchMedia for Mantine
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => false,
  }),
})

// Mock ResizeObserver for Mantine
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}
