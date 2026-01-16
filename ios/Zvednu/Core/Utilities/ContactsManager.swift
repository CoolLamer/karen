import Contacts
import Foundation

@MainActor
final class ContactsManager: ObservableObject {
    static let shared = ContactsManager()

    @Published private(set) var isEnabled: Bool
    @Published private(set) var authorizationStatus: CNAuthorizationStatus
    @Published private(set) var isLoading = false

    private var localCache: [String: String] = [:]
    private let contactsService = ContactsService.shared

    private let enabledKey = "contacts_access_enabled"

    private init() {
        isEnabled = UserDefaults.standard.bool(forKey: enabledKey)
        authorizationStatus = CNContactStore.authorizationStatus(for: .contacts)
    }

    // MARK: - Permission and Preference

    var canRequestAccess: Bool {
        authorizationStatus == .notDetermined
    }

    var isAuthorized: Bool {
        if #available(iOS 18.0, *) {
            return authorizationStatus == .authorized || authorizationStatus == .limited
        } else {
            return authorizationStatus == .authorized
        }
    }

    func enableContactsAccess() async {
        guard !isLoading else { return }

        isLoading = true

        do {
            let granted = try await contactsService.requestAccess()
            authorizationStatus = await contactsService.authorizationStatus

            if granted {
                UserDefaults.standard.set(true, forKey: enabledKey)
                isEnabled = true
                try await contactsService.buildCache()
                localCache = await contactsService.getCacheSnapshot()
            }
        } catch {
            print("Failed to enable contacts access: \(error)")
        }

        isLoading = false
    }

    func disableContactsAccess() {
        UserDefaults.standard.set(false, forKey: enabledKey)
        isEnabled = false
        localCache = [:]
        Task {
            await contactsService.invalidateCache()
        }
    }

    func refreshCache() async {
        guard isEnabled && isAuthorized else { return }

        isLoading = true

        do {
            await contactsService.invalidateCache()
            try await contactsService.buildCache()
            localCache = await contactsService.getCacheSnapshot()
        } catch {
            print("Failed to refresh contacts cache: \(error)")
        }

        isLoading = false
    }

    // MARK: - Name Lookup (synchronous for UI)

    func contactName(for phoneNumber: String) -> String? {
        guard isEnabled && isAuthorized else { return nil }
        let normalized = phoneNumber.normalizedForLookup()
        return localCache[normalized]
    }

    // MARK: - App Lifecycle

    func initializeOnAppLaunch() async {
        guard isEnabled && isAuthorized else { return }

        do {
            try await contactsService.buildCache()
            localCache = await contactsService.getCacheSnapshot()
        } catch {
            print("Failed to build contacts cache on launch: \(error)")
        }
    }
}
