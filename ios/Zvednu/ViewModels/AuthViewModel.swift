import Foundation
import SwiftUI
import Combine

@MainActor
class AuthViewModel: ObservableObject {
    @Published var isAuthenticated = false
    @Published var isLoading = true
    @Published var user: User?
    @Published var tenant: Tenant?
    @Published var needsOnboarding = false
    @Published var error: String?

    // Login state
    @Published var phoneNumber = ""
    @Published var verificationCode = ""
    @Published var isCodeSent = false
    @Published var isSendingCode = false
    @Published var isVerifyingCode = false

    private let authService = AuthService.shared

    init() {
        Task {
            await checkAuthentication()
        }
    }

    // MARK: - Authentication Check

    func checkAuthentication() async {
        isLoading = true

        // Check if we have a stored token
        guard KeychainManager.shared.getToken() != nil else {
            isLoading = false
            isAuthenticated = false
            return
        }

        do {
            let response = try await authService.getMe()
            user = response.user
            tenant = response.tenant
            isAuthenticated = true
            needsOnboarding = response.user.tenantId == nil

            // Re-register push token if available
            await PushNotificationService.shared.reregisterStoredToken()
        } catch {
            // Token is invalid
            await APIClient.shared.setAuthToken(nil)
            isAuthenticated = false
        }

        isLoading = false
    }

    // MARK: - Login Flow

    func sendCode() async {
        guard !phoneNumber.isEmpty else {
            error = "Zadej telefonní číslo"
            return
        }

        let normalizedPhone = phoneNumber.normalizedPhoneNumber()

        isSendingCode = true
        error = nil

        do {
            _ = try await authService.sendCode(phone: normalizedPhone)
            isCodeSent = true
        } catch {
            self.error = error.localizedDescription
        }

        isSendingCode = false
    }

    func verifyCode() async {
        guard !verificationCode.isEmpty else {
            error = "Zadej overovaci kod"
            return
        }

        let normalizedPhone = phoneNumber.normalizedPhoneNumber()

        isVerifyingCode = true
        error = nil

        do {
            let response = try await authService.verifyCode(phone: normalizedPhone, code: verificationCode)
            user = response.user
            isAuthenticated = true
            needsOnboarding = response.user.tenantId == nil

            // Request notification permission after login
            if let appDelegate = UIApplication.shared.delegate as? AppDelegate {
                appDelegate.requestNotificationPermission()
            }

            // Re-register push token
            await PushNotificationService.shared.reregisterStoredToken()
        } catch let apiError as APIError {
            switch apiError {
            case .httpError(400, _):
                self.error = "Neplatny kod"
            default:
                self.error = apiError.localizedDescription
            }
        } catch {
            self.error = error.localizedDescription
        }

        isVerifyingCode = false
    }

    func resetLoginState() {
        phoneNumber = ""
        verificationCode = ""
        isCodeSent = false
        error = nil
    }

    // MARK: - Logout

    func logout() async {
        do {
            try await PushNotificationService.shared.unregisterDeviceToken()
        } catch {
            print("Failed to unregister push token: \(error)")
        }

        do {
            try await authService.logout()
        } catch {
            print("Logout error: \(error)")
        }

        user = nil
        tenant = nil
        isAuthenticated = false
        needsOnboarding = false
        resetLoginState()
    }

    // MARK: - Onboarding

    func completeOnboarding(name: String, greetingText: String) async throws -> OnboardingResponse {
        let response = try await authService.completeOnboarding(name: name, greetingText: greetingText)
        tenant = response.tenant
        return response
    }

    func finishOnboarding() async {
        needsOnboarding = false
        await refreshUser()
    }

    // MARK: - Refresh User Data

    func refreshUser() async {
        do {
            let response = try await authService.getMe()
            user = response.user
            tenant = response.tenant
            needsOnboarding = response.user.tenantId == nil
        } catch {
            print("Failed to refresh user: \(error)")
        }
    }

    func updateTenant(_ newTenant: Tenant) {
        tenant = newTenant
    }
}
