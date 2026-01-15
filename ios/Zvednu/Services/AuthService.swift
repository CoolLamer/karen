import Foundation

actor AuthService {
    static let shared = AuthService()

    private let apiClient = APIClient.shared

    private init() {}

    // MARK: - Authentication

    func sendCode(phone: String) async throws -> Bool {
        struct SendCodeRequest: Codable {
            let phone: String
        }

        let response: SuccessResponse = try await apiClient.post(
            "/auth/send-code",
            body: SendCodeRequest(phone: phone)
        )
        return response.success
    }

    func verifyCode(phone: String, code: String) async throws -> AuthResponse {
        struct VerifyCodeRequest: Codable {
            let phone: String
            let code: String
        }

        let response: AuthResponse = try await apiClient.post(
            "/auth/verify-code",
            body: VerifyCodeRequest(phone: phone, code: code)
        )

        // Store token
        await apiClient.setAuthToken(response.token)

        return response
    }

    func refreshToken() async throws -> AuthResponse {
        let response: AuthResponse = try await apiClient.postEmpty("/auth/refresh")
        await apiClient.setAuthToken(response.token)
        return response
    }

    func logout() async throws {
        do {
            let _: Empty = try await apiClient.postEmpty("/auth/logout")
        } catch {
            // Ignore logout errors, clear token anyway
        }
        await apiClient.setAuthToken(nil)
    }

    // MARK: - User Info

    func getMe() async throws -> MeResponse {
        try await apiClient.get("/api/me")
    }

    // MARK: - Onboarding

    func completeOnboarding(name: String, greetingText: String) async throws -> OnboardingResponse {
        struct OnboardingRequest: Codable {
            let name: String
            let greetingText: String

            enum CodingKeys: String, CodingKey {
                case name
                case greetingText = "greeting_text"
            }
        }

        let response: OnboardingResponse = try await apiClient.post(
            "/api/onboarding/complete",
            body: OnboardingRequest(name: name, greetingText: greetingText)
        )

        // Update token
        await apiClient.setAuthToken(response.token)

        return response
    }
}
