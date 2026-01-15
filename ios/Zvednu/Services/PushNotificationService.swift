import Foundation

actor PushNotificationService {
    static let shared = PushNotificationService()

    private let apiClient = APIClient.shared

    private init() {}

    // MARK: - Device Token Registration

    func registerDeviceToken(_ token: String) async throws {
        struct RegisterRequest: Codable {
            let token: String
            let platform: String
        }

        // Store token locally for later use
        try? KeychainManager.shared.saveDeviceToken(token)

        let _: SuccessResponse = try await apiClient.post(
            "/api/push/register",
            body: RegisterRequest(token: token, platform: "ios")
        )
    }

    func unregisterDeviceToken() async throws {
        guard let token = KeychainManager.shared.getDeviceToken() else {
            return
        }

        struct UnregisterRequest: Codable {
            let token: String
        }

        let _: SuccessResponse = try await apiClient.post(
            "/api/push/unregister",
            body: UnregisterRequest(token: token)
        )

        KeychainManager.shared.deleteDeviceToken()
    }

    // Re-register stored token (useful after login)
    func reregisterStoredToken() async {
        guard let token = KeychainManager.shared.getDeviceToken() else {
            return
        }

        do {
            try await registerDeviceToken(token)
        } catch {
            print("Failed to re-register push token: \(error)")
        }
    }
}
