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
