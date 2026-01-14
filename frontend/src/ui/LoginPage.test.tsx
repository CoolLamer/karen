import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { BrowserRouter } from 'react-router-dom'
import { MantineProvider } from '@mantine/core'
import { LoginPage } from './LoginPage'
import * as apiModule from '../api'
import * as authModule from '../AuthContext'

// Mock the api module
vi.mock('../api', () => ({
  api: {
    sendCode: vi.fn(),
    verifyCode: vi.fn(),
  },
}))

// Mock useAuth
vi.mock('../AuthContext', () => ({
  useAuth: vi.fn(),
}))

// Mock useNavigate
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

const mockApi = apiModule.api as {
  sendCode: ReturnType<typeof vi.fn>
  verifyCode: ReturnType<typeof vi.fn>
}

const mockUseAuth = authModule.useAuth as ReturnType<typeof vi.fn>

function renderLoginPage() {
  return render(
    <MantineProvider>
      <BrowserRouter>
        <LoginPage />
      </BrowserRouter>
    </MantineProvider>
  )
}

describe('LoginPage', () => {
  const mockLogin = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    mockUseAuth.mockReturnValue({
      login: mockLogin,
      isLoading: false,
      isAuthenticated: false,
    })
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('Phone step', () => {
    it('should render phone input initially', () => {
      renderLoginPage()

      expect(screen.getByPlaceholderText('+420 777 123 456')).toBeInTheDocument()
      expect(screen.getByText('Poslat ověřovací kód')).toBeInTheDocument()
    })

    it('should format phone number with + prefix', async () => {
      const user = userEvent.setup()
      renderLoginPage()

      const input = screen.getByPlaceholderText('+420 777 123 456')
      await user.type(input, '420123456789')

      expect(input).toHaveValue('+420123456789')
    })

    it('should show error for short phone number', async () => {
      const user = userEvent.setup()
      renderLoginPage()

      const input = screen.getByPlaceholderText('+420 777 123 456')
      await user.type(input, '123')

      const submitButton = screen.getByText('Poslat ověřovací kód')
      await user.click(submitButton)

      expect(screen.getByText('Zadej platné telefonní číslo')).toBeInTheDocument()
    })

    it('should call sendCode API and transition to OTP step', async () => {
      mockApi.sendCode.mockResolvedValue({ success: true })
      const user = userEvent.setup()
      renderLoginPage()

      const input = screen.getByPlaceholderText('+420 777 123 456')
      await user.type(input, '+420777123456')

      const submitButton = screen.getByText('Poslat ověřovací kód')
      await user.click(submitButton)

      await waitFor(() => {
        expect(mockApi.sendCode).toHaveBeenCalledWith('+420777123456')
        expect(screen.getByText('Ověřovací kód')).toBeInTheDocument()
      })
    })

    it('should show error on sendCode failure', async () => {
      mockApi.sendCode.mockRejectedValue(new Error('Network error'))
      const user = userEvent.setup()
      renderLoginPage()

      const input = screen.getByPlaceholderText('+420 777 123 456')
      await user.type(input, '+420777123456')

      const submitButton = screen.getByText('Poslat ověřovací kód')
      await user.click(submitButton)

      await waitFor(() => {
        expect(
          screen.getByText('Nepodařilo se odeslat kód. Zkontroluj číslo a zkus to znovu.')
        ).toBeInTheDocument()
      })
    })
  })

  describe('OTP step', () => {
    beforeEach(async () => {
      mockApi.sendCode.mockResolvedValue({ success: true })
    })

    async function goToOtpStep() {
      const user = userEvent.setup()
      renderLoginPage()

      const input = screen.getByPlaceholderText('+420 777 123 456')
      await user.type(input, '+420777123456')

      const submitButton = screen.getByText('Poslat ověřovací kód')
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('Ověřovací kód')).toBeInTheDocument()
      })

      return user
    }

    it('should show phone number in OTP step', async () => {
      await goToOtpStep()

      expect(screen.getByText('+420777123456')).toBeInTheDocument()
    })

    it('should render PIN input in OTP step', async () => {
      await goToOtpStep()

      // Verify PIN inputs are rendered (Mantine PinInput renders 6 inputs)
      const pinInputs = document.querySelectorAll('input')
      expect(pinInputs.length).toBeGreaterThanOrEqual(6)
    })

    it('should show resend countdown', async () => {
      await goToOtpStep()

      // Should show countdown text
      expect(screen.getByText(/Poslat znovu za/)).toBeInTheDocument()
    })

    it('should allow going back to phone step', async () => {
      const user = await goToOtpStep()

      const backButton = screen.getByText('Zpět')
      await user.click(backButton)

      expect(screen.getByPlaceholderText('+420 777 123 456')).toBeInTheDocument()
    })
  })
})
