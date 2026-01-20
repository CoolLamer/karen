import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { setAuthToken, getAuthToken, isAuthenticated } from './api'

// We'll test the utility functions since the api object relies on fetch

describe('api utilities', () => {
  beforeEach(() => {
    // Clear localStorage before each test
    localStorage.clear()
    // Reset auth token state
    setAuthToken(null)
  })

  afterEach(() => {
    localStorage.clear()
  })

  describe('setAuthToken', () => {
    it('should store token in localStorage', () => {
      setAuthToken('test-token')
      expect(localStorage.getItem('karen_token')).toBe('test-token')
    })

    it('should remove token from localStorage when null', () => {
      localStorage.setItem('karen_token', 'existing-token')
      setAuthToken(null)
      expect(localStorage.getItem('karen_token')).toBeNull()
    })

    it('should update the in-memory token', () => {
      setAuthToken('new-token')
      expect(getAuthToken()).toBe('new-token')
    })
  })

  describe('getAuthToken', () => {
    it('should return null when no token is set', () => {
      expect(getAuthToken()).toBeNull()
    })

    it('should return the current token', () => {
      setAuthToken('my-token')
      expect(getAuthToken()).toBe('my-token')
    })
  })

  describe('isAuthenticated', () => {
    it('should return false when no token', () => {
      setAuthToken(null)
      expect(isAuthenticated()).toBe(false)
    })

    it('should return true when token exists', () => {
      setAuthToken('valid-token')
      expect(isAuthenticated()).toBe(true)
    })
  })
})

describe('api HTTP client', () => {
  const originalFetch = global.fetch

  beforeEach(() => {
    localStorage.clear()
    setAuthToken(null)
  })

  afterEach(() => {
    global.fetch = originalFetch
    localStorage.clear()
  })

  it('should clear token on 401 response', async () => {
    setAuthToken('expired-token')

    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 401,
      text: () => Promise.resolve('Unauthorized'),
    })

    // Import fresh api to use mocked fetch
    const { api } = await import('./api')

    await expect(api.getMe()).rejects.toThrow()
    expect(getAuthToken()).toBeNull()
  })

  it('should include Authorization header when token exists', async () => {
    setAuthToken('my-auth-token')

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ user: { id: '1', phone: '+420123' } }),
    })

    const { api } = await import('./api')
    await api.getMe()

    expect(global.fetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: 'Bearer my-auth-token',
        }),
      })
    )
  })

  it('should handle 204 No Content responses', async () => {
    setAuthToken('token')

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 204,
    })

    const { api } = await import('./api')
    const result = await api.logout()

    expect(result).toBeUndefined()
  })
})

describe('api type exports', () => {
  it('should export type definitions', async () => {
    // Verify types are properly exported by importing them
    const apiModule = await import('./api')

    expect(apiModule.api).toBeDefined()
    expect(apiModule.setAuthToken).toBeDefined()
    expect(apiModule.getAuthToken).toBeDefined()
    expect(apiModule.isAuthenticated).toBeDefined()
  })
})

describe('token refresh on 401', () => {
  const originalFetch = global.fetch

  beforeEach(() => {
    localStorage.clear()
    // Reset module to get fresh state
    vi.resetModules()
  })

  afterEach(() => {
    global.fetch = originalFetch
    localStorage.clear()
  })

  it('should attempt refresh on 401 and retry request', async () => {
    let apiCallCount = 0
    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (url.includes('/auth/refresh')) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: () => Promise.resolve({
            token: 'new-token',
            expires_at: new Date().toISOString(),
            user: { id: '1', phone: '+420123' }
          }),
        })
      }
      apiCallCount++
      if (apiCallCount === 1) {
        // First call returns 401
        return Promise.resolve({ ok: false, status: 401, text: () => Promise.resolve('Unauthorized') })
      }
      // Retry succeeds
      return Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve({ user: { id: '1', phone: '+420123' } }),
      })
    })

    // Import fresh module and set token in that module
    const apiModule = await import('./api')
    apiModule.setAuthToken('expired-token')

    const result = await apiModule.api.getMe()

    expect(result.user.id).toBe('1')
    expect(apiModule.getAuthToken()).toBe('new-token')
  })

  it('should clear token when refresh fails', async () => {
    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (url.includes('/auth/refresh')) {
        return Promise.resolve({ ok: false, status: 401, text: () => Promise.resolve('Invalid') })
      }
      return Promise.resolve({ ok: false, status: 401, text: () => Promise.resolve('Unauthorized') })
    })

    // Import fresh module and set token in that module
    const apiModule = await import('./api')
    apiModule.setAuthToken('expired-token')

    await expect(apiModule.api.getMe()).rejects.toThrow()
    expect(apiModule.getAuthToken()).toBeNull()
  })

  it('should not attempt refresh when no token exists', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 401,
      text: () => Promise.resolve('Unauthorized'),
    })

    // Import fresh module without token
    const apiModule = await import('./api')
    apiModule.setAuthToken(null)

    await expect(apiModule.api.getMe()).rejects.toThrow()

    // Should not have called refresh endpoint
    expect(global.fetch).toHaveBeenCalledTimes(1)
    expect(global.fetch).not.toHaveBeenCalledWith(
      expect.stringContaining('/auth/refresh'),
      expect.anything()
    )
  })
})
