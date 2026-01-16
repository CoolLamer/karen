import XCTest
@testable import Zvednu

@MainActor
final class SettingsViewModelTests: XCTestCase {

    // MARK: - Initial State Tests

    func testInitialState() {
        let viewModel = SettingsViewModel()

        XCTAssertNil(viewModel.tenant)
        XCTAssertNil(viewModel.billing)
        XCTAssertTrue(viewModel.phoneNumbers.isEmpty)
        XCTAssertFalse(viewModel.isLoading)
        XCTAssertFalse(viewModel.isSaving)
        XCTAssertFalse(viewModel.isUpgrading)
        XCTAssertNil(viewModel.error)
        XCTAssertFalse(viewModel.showSavedConfirmation)
        XCTAssertFalse(viewModel.showUpgradeSheet)
        // Voice-related initial state
        XCTAssertTrue(viewModel.voices.isEmpty)
        XCTAssertFalse(viewModel.isLoadingVoices)
        XCTAssertFalse(viewModel.showVoiceSheet)
    }

    // MARK: - Billing Info Helpers Tests

    func testFormattedTimeSavedZero() {
        let viewModel = SettingsViewModel()
        viewModel.billing = nil

        XCTAssertEqual(viewModel.formattedTimeSaved, "0min")
    }

    func testFormattedTimeSavedMinutesOnly() {
        let viewModel = SettingsViewModel()
        viewModel.billing = BillingInfo(
            plan: "trial",
            status: "active",
            callStatus: CallStatus(
                canReceive: true,
                reason: "ok",
                callsUsed: 5,
                callsLimit: 20,
                trialDaysLeft: 10,
                trialCallsLeft: 15
            ),
            trialEndsAt: nil,
            totalTimeSaved: 1800, // 30 minutes
            currentUsage: nil
        )

        XCTAssertEqual(viewModel.formattedTimeSaved, "30min")
    }

    func testFormattedTimeSavedHoursAndMinutes() {
        let viewModel = SettingsViewModel()
        viewModel.billing = BillingInfo(
            plan: "trial",
            status: "active",
            callStatus: CallStatus(
                canReceive: true,
                reason: "ok",
                callsUsed: 5,
                callsLimit: 20,
                trialDaysLeft: 10,
                trialCallsLeft: 15
            ),
            trialEndsAt: nil,
            totalTimeSaved: 5400, // 1 hour 30 minutes
            currentUsage: nil
        )

        XCTAssertEqual(viewModel.formattedTimeSaved, "1h 30min")
    }

    func testUsagePercentageWithNoBilling() {
        let viewModel = SettingsViewModel()
        viewModel.billing = nil

        XCTAssertEqual(viewModel.usagePercentage, 0)
    }

    func testUsagePercentageWithBilling() {
        let viewModel = SettingsViewModel()
        viewModel.billing = BillingInfo(
            plan: "trial",
            status: "active",
            callStatus: CallStatus(
                canReceive: true,
                reason: "ok",
                callsUsed: 10,
                callsLimit: 20,
                trialDaysLeft: nil,
                trialCallsLeft: nil
            ),
            trialEndsAt: nil,
            totalTimeSaved: 0,
            currentUsage: nil
        )

        XCTAssertEqual(viewModel.usagePercentage, 50.0)
    }

    func testUsagePercentageAtLimit() {
        let viewModel = SettingsViewModel()
        viewModel.billing = BillingInfo(
            plan: "trial",
            status: "active",
            callStatus: CallStatus(
                canReceive: false,
                reason: "limit_exceeded",
                callsUsed: 20,
                callsLimit: 20,
                trialDaysLeft: nil,
                trialCallsLeft: nil
            ),
            trialEndsAt: nil,
            totalTimeSaved: 0,
            currentUsage: nil
        )

        XCTAssertEqual(viewModel.usagePercentage, 100.0)
    }

    // MARK: - Trial Status Tests

    func testIsTrialWithTrialPlan() {
        let viewModel = SettingsViewModel()
        viewModel.billing = BillingInfo(
            plan: "trial",
            status: "active",
            callStatus: CallStatus(
                canReceive: true,
                reason: "ok",
                callsUsed: 5,
                callsLimit: 20,
                trialDaysLeft: 10,
                trialCallsLeft: 15
            ),
            trialEndsAt: nil,
            totalTimeSaved: 0,
            currentUsage: nil
        )

        XCTAssertTrue(viewModel.isTrial)
    }

    func testIsTrialWithBasicPlan() {
        let viewModel = SettingsViewModel()
        viewModel.billing = BillingInfo(
            plan: "basic",
            status: "active",
            callStatus: CallStatus(
                canReceive: true,
                reason: "ok",
                callsUsed: 5,
                callsLimit: 50,
                trialDaysLeft: nil,
                trialCallsLeft: nil
            ),
            trialEndsAt: nil,
            totalTimeSaved: 0,
            currentUsage: nil
        )

        XCTAssertFalse(viewModel.isTrial)
    }

    func testIsTrialExpiredWhenCannotReceive() {
        let viewModel = SettingsViewModel()
        viewModel.billing = BillingInfo(
            plan: "trial",
            status: "active",
            callStatus: CallStatus(
                canReceive: false,
                reason: "trial_expired",
                callsUsed: 5,
                callsLimit: 20,
                trialDaysLeft: 0,
                trialCallsLeft: 0
            ),
            trialEndsAt: nil,
            totalTimeSaved: 0,
            currentUsage: nil
        )

        XCTAssertTrue(viewModel.isTrialExpired)
    }

    func testIsTrialExpiredWhenCanReceive() {
        let viewModel = SettingsViewModel()
        viewModel.billing = BillingInfo(
            plan: "trial",
            status: "active",
            callStatus: CallStatus(
                canReceive: true,
                reason: "ok",
                callsUsed: 5,
                callsLimit: 20,
                trialDaysLeft: 10,
                trialCallsLeft: 15
            ),
            trialEndsAt: nil,
            totalTimeSaved: 0,
            currentUsage: nil
        )

        XCTAssertFalse(viewModel.isTrialExpired)
    }

    // MARK: - VIP Name Management Tests

    func testAddVipName() {
        let viewModel = SettingsViewModel()
        XCTAssertTrue(viewModel.vipNames.isEmpty)

        viewModel.addVipName("John Doe")

        XCTAssertEqual(viewModel.vipNames.count, 1)
        XCTAssertEqual(viewModel.vipNames.first, "John Doe")
    }

    func testAddVipNameTrimsWhitespace() {
        let viewModel = SettingsViewModel()

        viewModel.addVipName("  Jane Doe  ")

        XCTAssertEqual(viewModel.vipNames.first, "Jane Doe")
    }

    func testAddVipNameIgnoresEmpty() {
        let viewModel = SettingsViewModel()

        viewModel.addVipName("")
        viewModel.addVipName("   ")

        XCTAssertTrue(viewModel.vipNames.isEmpty)
    }

    func testAddVipNameIgnoresDuplicate() {
        let viewModel = SettingsViewModel()

        viewModel.addVipName("John Doe")
        viewModel.addVipName("John Doe")

        XCTAssertEqual(viewModel.vipNames.count, 1)
    }

    func testRemoveVipName() {
        let viewModel = SettingsViewModel()
        viewModel.vipNames = ["John", "Jane", "Bob"]

        viewModel.removeVipName(at: 1)

        XCTAssertEqual(viewModel.vipNames, ["John", "Bob"])
    }

    func testRemoveVipNameInvalidIndex() {
        let viewModel = SettingsViewModel()
        viewModel.vipNames = ["John"]

        viewModel.removeVipName(at: 5)

        XCTAssertEqual(viewModel.vipNames.count, 1)
    }

    // MARK: - Primary Phone Number Tests

    func testPrimaryPhoneNumberWithNone() {
        let viewModel = SettingsViewModel()
        viewModel.phoneNumbers = []

        XCTAssertNil(viewModel.primaryPhoneNumber)
    }

    func testPrimaryPhoneNumberWithPrimary() {
        let viewModel = SettingsViewModel()
        viewModel.phoneNumbers = [
            TenantPhoneNumber(id: "1", twilioNumber: "+420123456789", isPrimary: false),
            TenantPhoneNumber(id: "2", twilioNumber: "+420987654321", isPrimary: true)
        ]

        XCTAssertEqual(viewModel.primaryPhoneNumber, "+420987654321")
    }

    // MARK: - Voice Selection Tests

    func testCurrentVoiceNameWithNoTenant() {
        let viewModel = SettingsViewModel()
        viewModel.tenant = nil

        XCTAssertEqual(viewModel.currentVoiceName, "Výchozí")
    }

    func testCurrentVoiceNameWithNoVoiceId() {
        let viewModel = SettingsViewModel()
        viewModel.tenant = Tenant(
            id: "tenant-1",
            name: "Test",
            systemPrompt: "Test prompt",
            greetingText: nil,
            voiceId: nil,
            language: "cs",
            vipNames: nil,
            marketingEmail: nil,
            forwardNumber: nil,
            maxTurnTimeoutMs: nil,
            plan: "trial",
            status: "active"
        )

        XCTAssertEqual(viewModel.currentVoiceName, "Výchozí")
    }

    func testCurrentVoiceNameWithMatchingVoice() {
        let viewModel = SettingsViewModel()
        viewModel.tenant = Tenant(
            id: "tenant-1",
            name: "Test",
            systemPrompt: "Test prompt",
            greetingText: nil,
            voiceId: "voice-1",
            language: "cs",
            vipNames: nil,
            marketingEmail: nil,
            forwardNumber: nil,
            maxTurnTimeoutMs: nil,
            plan: "trial",
            status: "active"
        )
        viewModel.voices = [
            Voice(id: "voice-1", name: "Rachel", description: "Friendly", gender: "female"),
            Voice(id: "voice-2", name: "Adam", description: "Professional", gender: "male")
        ]

        XCTAssertEqual(viewModel.currentVoiceName, "Rachel")
    }

    func testCurrentVoiceNameWithNonMatchingVoice() {
        let viewModel = SettingsViewModel()
        viewModel.tenant = Tenant(
            id: "tenant-1",
            name: "Test",
            systemPrompt: "Test prompt",
            greetingText: nil,
            voiceId: "non-existent-voice",
            language: "cs",
            vipNames: nil,
            marketingEmail: nil,
            forwardNumber: nil,
            maxTurnTimeoutMs: nil,
            plan: "trial",
            status: "active"
        )
        viewModel.voices = [
            Voice(id: "voice-1", name: "Rachel", description: "Friendly", gender: "female")
        ]

        XCTAssertEqual(viewModel.currentVoiceName, "Výchozí")
    }

    func testCurrentVoiceNameWithEmptyVoicesList() {
        let viewModel = SettingsViewModel()
        viewModel.tenant = Tenant(
            id: "tenant-1",
            name: "Test",
            systemPrompt: "Test prompt",
            greetingText: nil,
            voiceId: "voice-1",
            language: "cs",
            vipNames: nil,
            marketingEmail: nil,
            forwardNumber: nil,
            maxTurnTimeoutMs: nil,
            plan: "trial",
            status: "active"
        )
        viewModel.voices = []

        XCTAssertEqual(viewModel.currentVoiceName, "Výchozí")
    }
}
