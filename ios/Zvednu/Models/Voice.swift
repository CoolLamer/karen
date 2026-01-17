import Foundation

struct Voice: Codable, Identifiable, Equatable {
    let id: String
    let name: String
    let description: String
    let gender: String
}

struct VoicesResponse: Codable {
    let voices: [Voice]
}
