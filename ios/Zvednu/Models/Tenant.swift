import Foundation

struct Tenant: Codable, Identifiable, Equatable {
    let id: String
    var name: String
    var systemPrompt: String
    var greetingText: String?
    var voiceId: String?
    var language: String
    var vipNames: [String]?
    var marketingEmail: String?
    var forwardNumber: String?
    var maxTurnTimeoutMs: Int?
    var plan: String
    var status: String

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case systemPrompt = "system_prompt"
        case greetingText = "greeting_text"
        case voiceId = "voice_id"
        case language
        case vipNames = "vip_names"
        case marketingEmail = "marketing_email"
        case forwardNumber = "forward_number"
        case maxTurnTimeoutMs = "max_turn_timeout_ms"
        case plan
        case status
    }
}

struct TenantPhoneNumber: Codable, Identifiable, Equatable {
    let id: String
    let twilioNumber: String
    let isPrimary: Bool

    enum CodingKeys: String, CodingKey {
        case id
        case twilioNumber = "twilio_number"
        case isPrimary = "is_primary"
    }
}

struct TenantUpdateRequest: Codable {
    var name: String?
    var greetingText: String?
    var vipNames: [String]?
    var marketingEmail: String?

    enum CodingKeys: String, CodingKey {
        case name
        case greetingText = "greeting_text"
        case vipNames = "vip_names"
        case marketingEmail = "marketing_email"
    }
}
