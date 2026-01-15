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
        try await apiClient.patch("/api/tenant", body: update)
    }
}
