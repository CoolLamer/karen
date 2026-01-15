import Foundation
import SwiftUI

@MainActor
class OnboardingViewModel: ObservableObject {
    @Published var currentStep: OnboardingStep = .welcome
    @Published var name = ""
    @Published var greetingText = ""
    @Published var greetingGenerated = false
    @Published var vipNames: [String] = []
    @Published var marketingOption: MarketingOption = .reject
    @Published var marketingEmail = ""
    @Published var phoneNumbers: [TenantPhoneNumber] = []
    @Published var isLoading = false
    @Published var error: String?

    private let authService = AuthService.shared
    private let tenantService = TenantService.shared
    weak var authViewModel: AuthViewModel?

    enum OnboardingStep: Int, CaseIterable {
        case welcome = 0
        case name = 1
        case vipContacts = 2
        case marketing = 3
        case phoneSetup = 4
        case complete = 5
    }

    enum MarketingOption: String {
        case reject
        case email
    }

    var primaryPhoneNumber: String? {
        phoneNumbers.first(where: { $0.isPrimary })?.twilioNumber
    }

    var hasPhoneNumber: Bool {
        primaryPhoneNumber != nil
    }

    var progress: Double {
        guard currentStep != .welcome && currentStep != .complete else { return 0 }
        return Double(currentStep.rawValue) / 5.0
    }

    // MARK: - Greeting Generation

    func generateGreeting() {
        guard !name.isEmpty && !greetingGenerated else { return }
        greetingText = "Dobry den, tady asistentka Karen. \(name.trimmingCharacters(in: .whitespaces)) ted nemuze prijmout hovor, ale muzu vam pro nej zanechat vzkaz - co od nej potrebujete?"
        greetingGenerated = true
    }

    // MARK: - Step Navigation

    func goToNext() {
        guard let nextStep = OnboardingStep(rawValue: currentStep.rawValue + 1) else { return }
        currentStep = nextStep
    }

    func goToPrevious() {
        guard let prevStep = OnboardingStep(rawValue: currentStep.rawValue - 1) else { return }
        currentStep = prevStep
    }

    // MARK: - Onboarding Actions

    func completeNameStep() async {
        guard !name.trimmingCharacters(in: .whitespaces).isEmpty else {
            error = "Zadej sve jmeno"
            return
        }

        // Generate greeting if not done yet
        if !greetingGenerated {
            generateGreeting()
        }

        // Use default greeting if user cleared it
        var finalGreeting = greetingText.trimmingCharacters(in: .whitespaces)
        if finalGreeting.isEmpty {
            generateGreeting()
            finalGreeting = greetingText
        }

        isLoading = true
        error = nil

        do {
            let response = try await authViewModel?.completeOnboarding(
                name: name.trimmingCharacters(in: .whitespaces),
                greetingText: finalGreeting
            )

            if let phoneNumber = response?.phoneNumber {
                phoneNumbers = [phoneNumber]
            } else {
                // Try to fetch phone numbers
                do {
                    let tenantData = try await tenantService.getTenant()
                    phoneNumbers = tenantData.phoneNumbers
                } catch {
                    // Phone numbers might not be assigned yet
                }
            }

            goToNext()
        } catch {
            self.error = "Nepodarilo se dokoncit registraci. Zkus to znovu."
        }

        isLoading = false
    }

    func saveConfiguration() async {
        // Save VIP names and marketing email if configured
        if !vipNames.isEmpty || (marketingOption == .email && !marketingEmail.isEmpty) {
            var update = TenantUpdateRequest()

            if !vipNames.isEmpty {
                update.vipNames = vipNames
            }

            if marketingOption == .email && !marketingEmail.isEmpty {
                update.marketingEmail = marketingEmail
            }

            do {
                _ = try await tenantService.updateTenant(update)
            } catch {
                // Non-critical, continue anyway
                print("Failed to save configuration: \(error)")
            }
        }

        goToNext()
    }

    func finish() async {
        await authViewModel?.finishOnboarding()
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
