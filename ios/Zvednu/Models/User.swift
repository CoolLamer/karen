import Foundation

struct User: Codable, Identifiable, Equatable {
    let id: String
    let phone: String
    var name: String?
    var tenantId: String?

    enum CodingKeys: String, CodingKey {
        case id
        case phone
        case name
        case tenantId = "tenant_id"
    }
}
