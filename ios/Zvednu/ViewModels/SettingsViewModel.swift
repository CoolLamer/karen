import Foundation
import SwiftUI

@MainActor
class SettingsViewModel: ObservableObject {
    @Published var tenant: Tenant?
    @Published var phoneNumbers: [TenantPhoneNumber] = []
    @Published var isLoading = false
    @Published var isSaving = false
    @Published var error: String?
    @Published var showSavedConfirmation = false

    // Editable fields
    @Published var name = ""
    @Published var greetingText = ""
    @Published var vipNames: [String] = []
    @Published var marketingEmail = ""

    private let tenantService = TenantService.shared
    weak var authViewModel: AuthViewModel?

    var primaryPhoneNumber: String? {
        phoneNumbers.first(where: { $0.isPrimary })?.twilioNumber
    }

    // MARK: - Load Data

    func loadTenant() async {
        guard !isLoading else { return }

        isLoading = true
        error = nil

        do {
            let response = try await tenantService.getTenant()
            tenant = response.tenant
            phoneNumbers = response.phoneNumbers

            // Populate editable fields
            name = response.tenant.name
            greetingText = response.tenant.greetingText ?? ""
            vipNames = response.tenant.vipNames ?? []
            marketingEmail = response.tenant.marketingEmail ?? ""
        } catch {
            self.error = error.localizedDescription
        }

        isLoading = false
    }

    // MARK: - Save Changes

    func saveChanges() async {
        guard !isSaving else { return }

        isSaving = true
        error = nil

        var update = TenantUpdateRequest()
        var hasChanges = false

        if name != tenant?.name {
            update.name = name
            hasChanges = true
        }

        if greetingText != (tenant?.greetingText ?? "") {
            update.greetingText = greetingText
            hasChanges = true
        }

        if vipNames != (tenant?.vipNames ?? []) {
            update.vipNames = vipNames
            hasChanges = true
        }

        if marketingEmail != (tenant?.marketingEmail ?? "") {
            update.marketingEmail = marketingEmail.isEmpty ? nil : marketingEmail
            hasChanges = true
        }

        guard hasChanges else {
            isSaving = false
            return
        }

        do {
            let updatedTenant = try await tenantService.updateTenant(update)
            tenant = updatedTenant
            authViewModel?.updateTenant(updatedTenant)
            showSavedConfirmation = true

            // Hide confirmation after delay
            Task {
                try? await Task.sleep(nanoseconds: 2_000_000_000)
                showSavedConfirmation = false
            }
        } catch {
            self.error = error.localizedDescription
        }

        isSaving = false
    }

    // MARK: - VIP Name Management

    func addVipName(_ name: String) {
        let trimmed = name.trimmingCharacters(in: .whitespaces)
        guard !trimmed.isEmpty && !vipNames.contains(trimmed) else { return }
        vipNames.append(trimmed)
    }

    func removeVipName(at index: Int) {
        guard vipNames.indices.contains(index) else { return }
        vipNames.remove(at: index)
    }
}
