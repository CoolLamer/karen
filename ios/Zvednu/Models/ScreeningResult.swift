import Foundation
import SwiftUI

struct ScreeningResult: Codable, Equatable, Sendable {
    let legitimacyLabel: String
    let legitimacyConfidence: Double
    let leadLabel: String
    let intentCategory: String
    let intentText: String
    let entitiesJson: AnyCodable?
    let createdAt: String

    enum CodingKeys: String, CodingKey {
        case legitimacyLabel = "legitimacy_label"
        case legitimacyConfidence = "legitimacy_confidence"
        case leadLabel = "lead_label"
        case intentCategory = "intent_category"
        case intentText = "intent_text"
        case entitiesJson = "entities_json"
        case createdAt = "created_at"
    }
}

// Helper for decoding arbitrary JSON
struct AnyCodable: Codable, Equatable, @unchecked Sendable {
    let value: Any

    init(_ value: Any) {
        self.value = value
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()

        if container.decodeNil() {
            self.value = NSNull()
        } else if let bool = try? container.decode(Bool.self) {
            self.value = bool
        } else if let int = try? container.decode(Int.self) {
            self.value = int
        } else if let double = try? container.decode(Double.self) {
            self.value = double
        } else if let string = try? container.decode(String.self) {
            self.value = string
        } else if let array = try? container.decode([AnyCodable].self) {
            self.value = array.map { $0.value }
        } else if let dictionary = try? container.decode([String: AnyCodable].self) {
            self.value = dictionary.mapValues { $0.value }
        } else {
            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Cannot decode value")
        }
    }

    func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()

        switch value {
        case is NSNull:
            try container.encodeNil()
        case let bool as Bool:
            try container.encode(bool)
        case let int as Int:
            try container.encode(int)
        case let double as Double:
            try container.encode(double)
        case let string as String:
            try container.encode(string)
        case let array as [Any]:
            try container.encode(array.map { AnyCodable($0) })
        case let dictionary as [String: Any]:
            try container.encode(dictionary.mapValues { AnyCodable($0) })
        default:
            let context = EncodingError.Context(codingPath: container.codingPath, debugDescription: "Cannot encode value")
            throw EncodingError.invalidValue(value, context)
        }
    }

    static func == (lhs: AnyCodable, rhs: AnyCodable) -> Bool {
        // Simple equality check - compare JSON representations
        let encoder = JSONEncoder()
        guard let lhsData = try? encoder.encode(lhs),
              let rhsData = try? encoder.encode(rhs) else {
            return false
        }
        return lhsData == rhsData
    }
}

// MARK: - Legitimacy Label Helpers

enum LegitimacyLabel: String {
    case legitimate = "legitimni"
    case spam = "spam"
    case marketing = "marketing"
    case scam = "podvod"
    case unknown = "neznamy"

    var displayText: String {
        switch self {
        case .legitimate: return "Legitimni"
        case .spam: return "Spam"
        case .marketing: return "Marketing"
        case .scam: return "Podvod"
        case .unknown: return "Neznamy"
        }
    }

    var color: Color {
        switch self {
        case .legitimate: return .green
        case .spam: return .red
        case .marketing: return .orange
        case .scam: return .red
        case .unknown: return .gray
        }
    }

    var iconName: String {
        switch self {
        case .legitimate: return "checkmark.shield.fill"
        case .spam: return "xmark.shield.fill"
        case .marketing: return "megaphone.fill"
        case .scam: return "exclamationmark.triangle.fill"
        case .unknown: return "questionmark.circle.fill"
        }
    }

    init(from string: String) {
        self = LegitimacyLabel(rawValue: string) ?? .unknown
    }
}

// MARK: - Lead Label Helpers

enum LeadLabel: String {
    case hotLead = "hot_lead"
    case warmLead = "warm_lead"
    case coldLead = "cold_lead"
    case notALead = "not_a_lead"
    case unknown = "unknown"

    var displayText: String {
        switch self {
        case .hotLead: return "Horky lead"
        case .warmLead: return "Teply lead"
        case .coldLead: return "Studeny lead"
        case .notALead: return "Neni lead"
        case .unknown: return "Neznamy"
        }
    }

    var color: Color {
        switch self {
        case .hotLead: return .red
        case .warmLead: return .orange
        case .coldLead: return .blue
        case .notALead: return .gray
        case .unknown: return .gray
        }
    }

    var iconName: String {
        switch self {
        case .hotLead: return "flame.fill"
        case .warmLead: return "sun.max.fill"
        case .coldLead: return "snowflake"
        case .notALead: return "minus.circle.fill"
        case .unknown: return "questionmark.circle.fill"
        }
    }

    init(from string: String) {
        self = LeadLabel(rawValue: string) ?? .unknown
    }
}
