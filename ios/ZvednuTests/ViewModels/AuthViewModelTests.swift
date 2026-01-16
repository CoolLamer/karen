import XCTest
@testable import Zvednu

@MainActor
final class AuthViewModelTests: XCTestCase {

    // MARK: - Initial State Tests

    func testInitialState() {
        let viewModel = AuthViewModel()

        // Initially should be loading (checking authentication)
        XCTAssertTrue(viewModel.isLoading)
        XCTAssertFalse(viewModel.isAuthenticated)
        XCTAssertNil(viewModel.user)
        XCTAssertNil(viewModel.tenant)
        XCTAssertFalse(viewModel.needsOnboarding)
        XCTAssertNil(viewModel.error)
    }

    func testLoginStateInitial() {
        let viewModel = AuthViewModel()

        XCTAssertEqual(viewModel.phoneNumber, "")
        XCTAssertEqual(viewModel.verificationCode, "")
        XCTAssertFalse(viewModel.isCodeSent)
        XCTAssertFalse(viewModel.isSendingCode)
        XCTAssertFalse(viewModel.isVerifyingCode)
    }

    // MARK: - Reset Login State Tests

    func testResetLoginState() {
        let viewModel = AuthViewModel()

        // Set some state
        viewModel.phoneNumber = "+420123456789"
        viewModel.verificationCode = "123456"
        viewModel.error = "Some error"

        // Reset
        viewModel.resetLoginState()

        XCTAssertEqual(viewModel.phoneNumber, "")
        XCTAssertEqual(viewModel.verificationCode, "")
        XCTAssertFalse(viewModel.isCodeSent)
        XCTAssertNil(viewModel.error)
    }

    // MARK: - Validation Tests

    func testSendCodeWithEmptyPhoneShowsError() async {
        let viewModel = AuthViewModel()
        viewModel.phoneNumber = ""

        await viewModel.sendCode()

        XCTAssertEqual(viewModel.error, "Zadej telefonní číslo")
        XCTAssertFalse(viewModel.isCodeSent)
    }

    func testVerifyCodeWithEmptyCodeShowsError() async {
        let viewModel = AuthViewModel()
        viewModel.phoneNumber = "+420123456789"
        viewModel.verificationCode = ""

        await viewModel.verifyCode()

        XCTAssertEqual(viewModel.error, "Zadej overovaci kod")
    }

    // MARK: - Update Tenant Tests

    func testUpdateTenant() {
        let viewModel = AuthViewModel()
        XCTAssertNil(viewModel.tenant)

        let tenant = Tenant(
            id: "tenant-123",
            name: "Test",
            systemPrompt: "prompt",
            greetingText: "hello",
            voiceId: nil,
            language: "cs",
            vipNames: [],
            marketingEmail: nil,
            forwardNumber: nil,
            maxTurnTimeoutMs: nil,
            plan: "free",
            status: "active"
        )

        viewModel.updateTenant(tenant)

        XCTAssertNotNil(viewModel.tenant)
        XCTAssertEqual(viewModel.tenant?.id, "tenant-123")
    }

    // MARK: - State Transitions Tests

    func testLogoutClearsState() async {
        let viewModel = AuthViewModel()

        // Simulate authenticated state
        viewModel.phoneNumber = "+420123456789"
        viewModel.verificationCode = "123456"

        await viewModel.logout()

        XCTAssertNil(viewModel.user)
        XCTAssertNil(viewModel.tenant)
        XCTAssertFalse(viewModel.isAuthenticated)
        XCTAssertFalse(viewModel.needsOnboarding)
        XCTAssertEqual(viewModel.phoneNumber, "")
        XCTAssertEqual(viewModel.verificationCode, "")
    }
}
