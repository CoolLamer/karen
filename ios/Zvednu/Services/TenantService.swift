import Foundation

actor TenantService {
    static let shared = TenantService()

    private let apiClient = APIClient.shared

    private init() {}

    // MARK: - Tenant Data

    func getTenant() async throws -> TenantResponse {
        try await apiClient.get("/api/tenant")
    }

    func updateTenant(_ update: TenantUpdateRequest) async throws -> Tenant {
        struct UpdateResponse: Codable {
            let tenant: Tenant
        }

        let response: UpdateResponse = try await apiClient.patch("/api/tenant", body: update)
        return response.tenant
    }

    // MARK: - Billing

    func getBilling() async throws -> BillingInfo {
        try await apiClient.get("/api/billing")
    }

    func createCheckout(plan: String, interval: String) async throws -> CheckoutResponse {
        struct CheckoutRequest: Codable {
            let plan: String
            let interval: String
        }

        return try await apiClient.post(
            "/api/billing/checkout",
            body: CheckoutRequest(plan: plan, interval: interval)
        )
    }

    func createPortal() async throws -> PortalResponse {
        try await apiClient.post("/api/billing/portal", body: EmptyBody())
    }
}

// MARK: - Response Types

struct CheckoutResponse: Codable {
    let checkoutUrl: String
    let sessionId: String

    enum CodingKeys: String, CodingKey {
        case checkoutUrl = "checkout_url"
        case sessionId = "session_id"
    }
}

struct PortalResponse: Codable {
    let portalUrl: String

    enum CodingKeys: String, CodingKey {
        case portalUrl = "portal_url"
    }
}

struct EmptyBody: Codable {}
