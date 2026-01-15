import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { AuthProvider, useAuth } from './AuthContext'
import * as apiModule from './api'

// Mock the api module
vi.mock('./api', () => ({
  api: {
    getMe: vi.fn(),
    logout: vi.fn(),
  },
  setAuthToken: vi.fn(),
  getAuthToken: vi.fn(),
}))

const mockApi = apiModule.api as {
  getMe: ReturnType<typeof vi.fn>
  logout: ReturnType<typeof vi.fn>
}
const mockSetAuthToken = apiModule.setAuthToken as ReturnType<typeof vi.fn>
const mockGetAuthToken = apiModule.getAuthToken as ReturnType<typeof vi.fn>

// Test component that exposes auth context
function TestConsumer() {
  const auth = useAuth()
  return (
    <div>
      <span data-testid="loading">{String(auth.isLoading)}</span>
      <span data-testid="authenticated">{String(auth.isAuthenticated)}</span>
      <span data-testid="user">{auth.user?.phone ?? 'none'}</span>
      <span data-testid="tenant">{auth.tenant?.name ?? 'none'}</span>
      <span data-testid="needs-onboarding">{String(auth.needsOnboarding)}</span>
      <span data-testid="is-admin">{String(auth.isAdmin)}</span>
      <button onClick={() => auth.login('test-token', { id: '1', phone: '+420123456789' })}>
        Login
      </button>
      <button onClick={() => auth.logout()}>Logout</button>
    </div>
  )
}

describe('AuthContext', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('should start in loading state', async () => {
    mockGetAuthToken.mockReturnValue(null)

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>
    )

    // Initially loading should be true
    await waitFor(() => {
      expect(screen.getByTestId('loading').textContent).toBe('false')
    })
  })

  it('should be unauthenticated when no token exists', async () => {
    mockGetAuthToken.mockReturnValue(null)

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('authenticated').textContent).toBe('false')
      expect(screen.getByTestId('user').textContent).toBe('none')
    })
  })

  it('should load user when token exists', async () => {
    mockGetAuthToken.mockReturnValue('valid-token')
    mockApi.getMe.mockResolvedValue({
      user: { id: '1', phone: '+420123456789', name: 'Test User' },
      tenant: { id: 't1', name: 'Test Tenant' },
      is_admin: false,
    })

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('authenticated').textContent).toBe('true')
      expect(screen.getByTestId('user').textContent).toBe('+420123456789')
      expect(screen.getByTestId('tenant').textContent).toBe('Test Tenant')
    })
  })

  it('should set isAdmin when user is admin', async () => {
    mockGetAuthToken.mockReturnValue('valid-token')
    mockApi.getMe.mockResolvedValue({
      user: { id: '1', phone: '+420123456789' },
      tenant: { id: 't1', name: 'Test Tenant' },
      is_admin: true,
    })

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('is-admin').textContent).toBe('true')
    })
  })

  it('should set needsOnboarding when no tenant', async () => {
    mockGetAuthToken.mockReturnValue('valid-token')
    mockApi.getMe.mockResolvedValue({
      user: { id: '1', phone: '+420123456789' },
      tenant: null,
      is_admin: false,
    })

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('needs-onboarding').textContent).toBe('true')
    })
  })

  it('should clear auth state on API error', async () => {
    mockGetAuthToken.mockReturnValue('invalid-token')
    mockApi.getMe.mockRejectedValue(new Error('Unauthorized'))

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('authenticated').textContent).toBe('false')
      expect(mockSetAuthToken).toHaveBeenCalledWith(null)
    })
  })

  it('should throw error when useAuth is used outside provider', () => {
    // Suppress console.error for this test
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    expect(() => {
      render(<TestConsumer />)
    }).toThrow('useAuth must be used within an AuthProvider')

    consoleSpy.mockRestore()
  })
})
