import XCTest
@testable import Zvednu

@MainActor
final class OnboardingViewModelTests: XCTestCase {

    // MARK: - Initial State Tests

    func testInitialState() {
        let viewModel = OnboardingViewModel()

        XCTAssertEqual(viewModel.currentStep, .welcome)
        XCTAssertEqual(viewModel.name, "")
        XCTAssertEqual(viewModel.greetingText, "")
        XCTAssertFalse(viewModel.greetingGenerated)
        XCTAssertTrue(viewModel.vipNames.isEmpty)
        XCTAssertEqual(viewModel.marketingOption, .reject)
        XCTAssertEqual(viewModel.marketingEmail, "")
        XCTAssertTrue(viewModel.phoneNumbers.isEmpty)
        XCTAssertFalse(viewModel.isLoading)
        XCTAssertNil(viewModel.error)
    }

    // MARK: - Step Navigation Tests

    func testGoToNext() {
        let viewModel = OnboardingViewModel()

        XCTAssertEqual(viewModel.currentStep, .welcome)

        viewModel.goToNext()
        XCTAssertEqual(viewModel.currentStep, .name)

        viewModel.goToNext()
        XCTAssertEqual(viewModel.currentStep, .vipContacts)

        viewModel.goToNext()
        XCTAssertEqual(viewModel.currentStep, .marketing)

        viewModel.goToNext()
        XCTAssertEqual(viewModel.currentStep, .phoneSetup)

        viewModel.goToNext()
        XCTAssertEqual(viewModel.currentStep, .complete)
    }

    func testGoToPrevious() {
        let viewModel = OnboardingViewModel()
        viewModel.currentStep = .marketing

        viewModel.goToPrevious()
        XCTAssertEqual(viewModel.currentStep, .vipContacts)

        viewModel.goToPrevious()
        XCTAssertEqual(viewModel.currentStep, .name)

        viewModel.goToPrevious()
        XCTAssertEqual(viewModel.currentStep, .welcome)
    }

    func testGoToPreviousAtWelcomeStaysAtWelcome() {
        let viewModel = OnboardingViewModel()
        XCTAssertEqual(viewModel.currentStep, .welcome)

        viewModel.goToPrevious()
        XCTAssertEqual(viewModel.currentStep, .welcome)
    }

    func testGoToNextAtCompleteStaysAtComplete() {
        let viewModel = OnboardingViewModel()
        viewModel.currentStep = .complete

        viewModel.goToNext()
        XCTAssertEqual(viewModel.currentStep, .complete)
    }

    // MARK: - VIP Name Management Tests

    func testAddVipName() {
        let viewModel = OnboardingViewModel()

        viewModel.addVipName("Mom")
        XCTAssertEqual(viewModel.vipNames, ["Mom"])

        viewModel.addVipName("Dad")
        XCTAssertEqual(viewModel.vipNames, ["Mom", "Dad"])
    }

    func testAddVipNameTrimsWhitespace() {
        let viewModel = OnboardingViewModel()

        viewModel.addVipName("  Boss  ")
        XCTAssertEqual(viewModel.vipNames, ["Boss"])
    }

    func testAddVipNameIgnoresEmpty() {
        let viewModel = OnboardingViewModel()

        viewModel.addVipName("")
        XCTAssertTrue(viewModel.vipNames.isEmpty)

        viewModel.addVipName("   ")
        XCTAssertTrue(viewModel.vipNames.isEmpty)
    }

    func testAddVipNameIgnoresDuplicates() {
        let viewModel = OnboardingViewModel()

        viewModel.addVipName("Mom")
        viewModel.addVipName("Mom")
        XCTAssertEqual(viewModel.vipNames, ["Mom"])
    }

    func testRemoveVipName() {
        let viewModel = OnboardingViewModel()

        viewModel.addVipName("Mom")
        viewModel.addVipName("Dad")
        viewModel.addVipName("Boss")

        viewModel.removeVipName(at: 1)
        XCTAssertEqual(viewModel.vipNames, ["Mom", "Boss"])
    }

    func testRemoveVipNameInvalidIndex() {
        let viewModel = OnboardingViewModel()

        viewModel.addVipName("Mom")

        // Should not crash with invalid index
        viewModel.removeVipName(at: 5)
        viewModel.removeVipName(at: -1)
        XCTAssertEqual(viewModel.vipNames, ["Mom"])
    }

    // MARK: - Greeting Generation Tests

    func testGenerateGreeting() {
        let viewModel = OnboardingViewModel()
        viewModel.name = "Jan"

        viewModel.generateGreeting()

        XCTAssertTrue(viewModel.greetingGenerated)
        XCTAssertTrue(viewModel.greetingText.contains("Jan"))
        XCTAssertTrue(viewModel.greetingText.contains("Karen"))
    }

    func testGenerateGreetingWithEmptyName() {
        let viewModel = OnboardingViewModel()
        viewModel.name = ""

        viewModel.generateGreeting()

        // Should not generate if name is empty
        XCTAssertFalse(viewModel.greetingGenerated)
        XCTAssertEqual(viewModel.greetingText, "")
    }

    func testGenerateGreetingOnlyOnce() {
        let viewModel = OnboardingViewModel()
        viewModel.name = "Jan"

        viewModel.generateGreeting()
        let firstGreeting = viewModel.greetingText

        viewModel.name = "Petr"
        viewModel.generateGreeting()

        // Should not regenerate
        XCTAssertEqual(viewModel.greetingText, firstGreeting)
    }

    // MARK: - Progress Tests

    func testProgressAtWelcome() {
        let viewModel = OnboardingViewModel()
        viewModel.currentStep = .welcome

        XCTAssertEqual(viewModel.progress, 0)
    }

    func testProgressAtComplete() {
        let viewModel = OnboardingViewModel()
        viewModel.currentStep = .complete

        XCTAssertEqual(viewModel.progress, 0)
    }

    func testProgressAtMiddleStep() {
        let viewModel = OnboardingViewModel()
        viewModel.currentStep = .vipContacts // Step 2

        // Progress should be 2/5 = 0.4
        XCTAssertEqual(viewModel.progress, 0.4, accuracy: 0.01)
    }

    // MARK: - Phone Number Tests

    func testPrimaryPhoneNumber() {
        let viewModel = OnboardingViewModel()

        viewModel.phoneNumbers = [
            TenantPhoneNumber(id: "1", twilioNumber: "+420123456789", isPrimary: false),
            TenantPhoneNumber(id: "2", twilioNumber: "+420987654321", isPrimary: true)
        ]

        XCTAssertEqual(viewModel.primaryPhoneNumber, "+420987654321")
    }

    func testPrimaryPhoneNumberNone() {
        let viewModel = OnboardingViewModel()

        viewModel.phoneNumbers = [
            TenantPhoneNumber(id: "1", twilioNumber: "+420123456789", isPrimary: false)
        ]

        XCTAssertNil(viewModel.primaryPhoneNumber)
    }

    func testHasPhoneNumber() {
        let viewModel = OnboardingViewModel()

        XCTAssertFalse(viewModel.hasPhoneNumber)

        viewModel.phoneNumbers = [
            TenantPhoneNumber(id: "1", twilioNumber: "+420123456789", isPrimary: true)
        ]

        XCTAssertTrue(viewModel.hasPhoneNumber)
    }

    // MARK: - Marketing Option Tests

    func testMarketingOptionDefault() {
        let viewModel = OnboardingViewModel()

        XCTAssertEqual(viewModel.marketingOption, .reject)
    }

    func testMarketingOptionEmail() {
        let viewModel = OnboardingViewModel()

        viewModel.marketingOption = .email
        viewModel.marketingEmail = "test@example.com"

        XCTAssertEqual(viewModel.marketingOption, .email)
        XCTAssertEqual(viewModel.marketingEmail, "test@example.com")
    }
}
