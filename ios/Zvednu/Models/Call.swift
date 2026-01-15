import Foundation

struct CallListItem: Codable, Identifiable, Equatable, Sendable {
    var id: String { providerCallId }

    let provider: String
    let providerCallId: String
    let fromNumber: String
    let toNumber: String
    let status: String
    let startedAt: String
    let endedAt: String?
    let endedBy: String?
    let firstViewedAt: String?
    let resolvedAt: String?
    let resolvedBy: String?
    let screening: ScreeningResult?

    enum CodingKeys: String, CodingKey {
        case provider
        case providerCallId = "provider_call_id"
        case fromNumber = "from_number"
        case toNumber = "to_number"
        case status
        case startedAt = "started_at"
        case endedAt = "ended_at"
        case endedBy = "ended_by"
        case firstViewedAt = "first_viewed_at"
        case resolvedAt = "resolved_at"
        case resolvedBy = "resolved_by"
        case screening
    }

    var isResolved: Bool {
        resolvedAt != nil
    }

    var isViewed: Bool {
        firstViewedAt != nil
    }

    var startDate: Date? {
        startedAt.toDate()
    }
}

struct CallDetail: Codable, Identifiable, Equatable, Sendable {
    var id: String { providerCallId }

    let provider: String
    let providerCallId: String
    let fromNumber: String
    let toNumber: String
    let status: String
    let startedAt: String
    let endedAt: String?
    let endedBy: String?
    let firstViewedAt: String?
    let resolvedAt: String?
    let resolvedBy: String?
    let screening: ScreeningResult?
    let utterances: [Utterance]

    enum CodingKeys: String, CodingKey {
        case provider
        case providerCallId = "provider_call_id"
        case fromNumber = "from_number"
        case toNumber = "to_number"
        case status
        case startedAt = "started_at"
        case endedAt = "ended_at"
        case endedBy = "ended_by"
        case firstViewedAt = "first_viewed_at"
        case resolvedAt = "resolved_at"
        case resolvedBy = "resolved_by"
        case screening
        case utterances
    }

    var isResolved: Bool {
        resolvedAt != nil
    }

    var startDate: Date? {
        startedAt.toDate()
    }
}
