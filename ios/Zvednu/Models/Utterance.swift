import Foundation

struct Utterance: Codable, Identifiable, Equatable, Sendable {
    var id: Int { sequence }

    let speaker: String
    let text: String
    let sequence: Int
    let startedAt: String?
    let endedAt: String?
    let sttConfidence: Double?
    let interrupted: Bool

    enum CodingKeys: String, CodingKey {
        case speaker
        case text
        case sequence
        case startedAt = "started_at"
        case endedAt = "ended_at"
        case sttConfidence = "stt_confidence"
        case interrupted
    }

    var isAgent: Bool {
        speaker == "agent"
    }

    var speakerDisplayName: String {
        switch speaker {
        case "agent": return "Karen"
        case "caller": return "Volajici"
        default: return speaker
        }
    }
}
