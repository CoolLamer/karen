import Foundation
import SwiftUI

@MainActor
class SettingsViewModel: ObservableObject {
    @Published var tenant: Tenant?
    @Published var phoneNumbers: [TenantPhoneNumber] = []
    @Published var billing: BillingInfo?
    @Published var isLoading = false
    @Published var isSaving = false
    @Published var isUpgrading = false
    @Published var error: String?
    @Published var showSavedConfirmation = false
    @Published var showUpgradeSheet = false
    @Published var showVoiceSheet = false

    // Voice selection
    @Published var voices: [Voice] = []
    @Published var isLoadingVoices = false

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

    var currentVoiceName: String {
        guard let voiceId = tenant?.voiceId else { return "Výchozí" }
        return voices.first(where: { $0.id == voiceId })?.name ?? "Výchozí"
    }

    // MARK: - Load Data

    func loadTenant() async {
        guard !isLoading else { return }

        isLoading = true
        error = nil

        do {
            async let tenantResponse = tenantService.getTenant()
            async let billingResponse = tenantService.getBilling()

            let response = try await tenantResponse
            tenant = response.tenant
            phoneNumbers = response.phoneNumbers

            // Populate editable fields
            name = response.tenant.name
            greetingText = response.tenant.greetingText ?? ""
            vipNames = response.tenant.vipNames ?? []
            marketingEmail = response.tenant.marketingEmail ?? ""

            // Load billing (non-fatal if fails)
            billing = try? await billingResponse
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

    // MARK: - Billing

    var formattedTimeSaved: String {
        guard let seconds = billing?.totalTimeSaved, seconds > 0 else {
            return "0min"
        }
        let hours = seconds / 3600
        let minutes = (seconds % 3600) / 60
        if hours > 0 {
            return "\(hours)h \(minutes)min"
        }
        return "\(minutes)min"
    }

    var usagePercentage: Double {
        guard let callStatus = billing?.callStatus else { return 0 }
        guard callStatus.callsLimit > 0 else { return 0 }
        return Double(callStatus.callsUsed) / Double(callStatus.callsLimit) * 100
    }

    var isTrial: Bool {
        billing?.plan == "trial"
    }

    var isTrialExpired: Bool {
        guard let callStatus = billing?.callStatus else { return false }
        return !callStatus.canReceive
    }

    func openUpgrade(plan: String, interval: String) async {
        isUpgrading = true
        error = nil

        do {
            let response = try await tenantService.createCheckout(plan: plan, interval: interval)
            if let url = URL(string: response.checkoutUrl) {
                await MainActor.run {
                    UIApplication.shared.open(url)
                }
            }
        } catch {
            self.error = "Nepodařilo se otevřít platební bránu"
        }

        isUpgrading = false
    }

    func openManageSubscription() async {
        isUpgrading = true
        error = nil

        do {
            let response = try await tenantService.createPortal()
            if let url = URL(string: response.portalUrl) {
                await MainActor.run {
                    UIApplication.shared.open(url)
                }
            }
        } catch {
            self.error = "Nepodařilo se otevřít správu předplatného"
        }

        isUpgrading = false
    }

    // MARK: - Voice Selection

    func loadVoices() async {
        guard !isLoadingVoices else { return }

        isLoadingVoices = true
        do {
            voices = try await tenantService.getVoices()
        } catch {
            self.error = "Nepodařilo se načíst hlasy"
        }
        isLoadingVoices = false
    }

    func selectVoice(_ voiceId: String) async {
        isSaving = true
        error = nil

        var update = TenantUpdateRequest()
        update.voiceId = voiceId

        do {
            let updatedTenant = try await tenantService.updateTenant(update)
            tenant = updatedTenant
            authViewModel?.updateTenant(updatedTenant)
            showVoiceSheet = false
            showSavedConfirmation = true

            Task {
                try? await Task.sleep(nanoseconds: 2_000_000_000)
                showSavedConfirmation = false
            }
        } catch {
            self.error = "Nepodařilo se uložit hlas"
        }

        isSaving = false
    }

    func previewVoice(_ voiceId: String) async -> Data? {
        do {
            return try await tenantService.previewVoice(voiceId: voiceId)
        } catch {
            self.error = "Nepodařilo se přehrát ukázku"
            return nil
        }
    }
}
