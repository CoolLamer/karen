import Foundation

actor CallService {
    static let shared = CallService()

    private let apiClient = APIClient.shared

    private init() {}

    // MARK: - Call List

    func listCalls() async throws -> [CallListItem] {
        try await apiClient.get("/api/calls")
    }

    func getUnresolvedCount() async throws -> Int {
        let response: UnresolvedCountResponse = try await apiClient.get("/api/calls/unresolved-count")
        return response.count
    }

    // MARK: - Call Detail

    func getCall(providerCallId: String) async throws -> CallDetail {
        let encodedId = providerCallId.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? providerCallId
        return try await apiClient.get("/api/calls/\(encodedId)")
    }

    // MARK: - Call Actions

    func markCallViewed(providerCallId: String) async throws {
        let encodedId = providerCallId.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? providerCallId
        let _: SuccessResponse = try await apiClient.patchEmpty("/api/calls/\(encodedId)/viewed")
    }

    func markCallResolved(providerCallId: String) async throws {
        let encodedId = providerCallId.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? providerCallId
        let _: SuccessResponse = try await apiClient.patchEmpty("/api/calls/\(encodedId)/resolve")
    }

    func markCallUnresolved(providerCallId: String) async throws {
        let encodedId = providerCallId.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? providerCallId
        try await apiClient.deleteEmpty("/api/calls/\(encodedId)/resolve")
    }
}
