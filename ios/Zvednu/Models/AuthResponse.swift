import Foundation

struct AuthResponse: Codable {
    let token: String
    let expiresAt: String
    let user: User

    enum CodingKeys: String, CodingKey {
        case token
        case expiresAt = "expires_at"
        case user
    }
}

struct OnboardingResponse: Codable {
    let tenant: Tenant
    let token: String
    let expiresAt: String
    let phoneNumber: TenantPhoneNumber?

    enum CodingKeys: String, CodingKey {
        case tenant
        case token
        case expiresAt = "expires_at"
        case phoneNumber = "phone_number"
    }
}

struct MeResponse: Codable {
    let user: User
    let tenant: Tenant?
    let isAdmin: Bool?

    enum CodingKeys: String, CodingKey {
        case user
        case tenant
        case isAdmin = "is_admin"
    }
}

struct TenantResponse: Codable {
    let tenant: Tenant
    let phoneNumbers: [TenantPhoneNumber]

    enum CodingKeys: String, CodingKey {
        case tenant
        case phoneNumbers = "phone_numbers"
    }
}

struct SuccessResponse: Codable {
    let success: Bool
}

struct UnresolvedCountResponse: Codable {
    let count: Int
}
