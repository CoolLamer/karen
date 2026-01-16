import Contacts
import Foundation

enum ContactsError: LocalizedError {
    case accessDenied
    case accessRestricted
    case fetchFailed(Error)

    var errorDescription: String? {
        switch self {
        case .accessDenied:
            return "Pristup ke kontaktum byl zamitnut"
        case .accessRestricted:
            return "Pristup ke kontaktum je omezen"
        case .fetchFailed(let error):
            return "Chyba pri nacitani kontaktu: \(error.localizedDescription)"
        }
    }
}

actor ContactsService {
    static let shared = ContactsService()

    private let store = CNContactStore()
    private var phoneToNameCache: [String: String] = [:]
    private var cacheBuilt = false

    private init() {}

    // MARK: - Permission Management

    var authorizationStatus: CNAuthorizationStatus {
        CNContactStore.authorizationStatus(for: .contacts)
    }

    func requestAccess() async throws -> Bool {
        let status = authorizationStatus

        switch status {
        case .authorized:
            return true
        case .notDetermined:
            return try await store.requestAccess(for: .contacts)
        case .denied:
            throw ContactsError.accessDenied
        case .restricted:
            throw ContactsError.accessRestricted
        @unknown default:
            return false
        }
    }

    // MARK: - Contact Lookup

    func buildCache() async throws {
        guard !cacheBuilt else { return }

        let keysToFetch: [CNKeyDescriptor] = [
            CNContactGivenNameKey as CNKeyDescriptor,
            CNContactFamilyNameKey as CNKeyDescriptor,
            CNContactPhoneNumbersKey as CNKeyDescriptor,
        ]

        let request = CNContactFetchRequest(keysToFetch: keysToFetch)
        var newCache: [String: String] = [:]

        try store.enumerateContacts(with: request) { contact, _ in
            let displayName =
                [contact.givenName, contact.familyName]
                    .filter { !$0.isEmpty }
                    .joined(separator: " ")

            guard !displayName.isEmpty else { return }

            for phoneNumber in contact.phoneNumbers {
                let normalized = phoneNumber.value.stringValue.normalizedForLookup()
                newCache[normalized] = displayName
            }
        }

        phoneToNameCache = newCache
        cacheBuilt = true
    }

    func lookupName(for phoneNumber: String) -> String? {
        let normalized = phoneNumber.normalizedForLookup()
        return phoneToNameCache[normalized]
    }

    func invalidateCache() {
        phoneToNameCache = [:]
        cacheBuilt = false
    }

    func getCacheSnapshot() -> [String: String] {
        phoneToNameCache
    }
}
