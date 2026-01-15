import XCTest
@testable import Zvednu

@MainActor
final class CallInboxViewModelTests: XCTestCase {

    // MARK: - Initial State Tests

    func testInitialState() {
        let viewModel = CallInboxViewModel()

        XCTAssertTrue(viewModel.calls.isEmpty)
        XCTAssertEqual(viewModel.unresolvedCount, 0)
        XCTAssertFalse(viewModel.isLoading)
        XCTAssertFalse(viewModel.isRefreshing)
        XCTAssertNil(viewModel.error)
    }

    // MARK: - Loading State Tests

    func testLoadCallsPreventsDuplicateLoading() async {
        let viewModel = CallInboxViewModel()

        // Start loading
        viewModel.isLoading = true

        // Try to load again - should return immediately
        let expectation = XCTestExpectation(description: "Load should return quickly")

        Task {
            await viewModel.loadCalls()
            expectation.fulfill()
        }

        // Should fulfill quickly since isLoading is already true
        await fulfillment(of: [expectation], timeout: 1.0)

        // isLoading should still be true (not reset because guard returned)
        XCTAssertTrue(viewModel.isLoading)
    }

    func testRefreshCallsPreventsDuplicateRefreshing() async {
        let viewModel = CallInboxViewModel()

        // Start refreshing
        viewModel.isRefreshing = true

        // Try to refresh again - should return immediately
        let expectation = XCTestExpectation(description: "Refresh should return quickly")

        Task {
            await viewModel.refreshCalls()
            expectation.fulfill()
        }

        await fulfillment(of: [expectation], timeout: 1.0)

        // isRefreshing should still be true
        XCTAssertTrue(viewModel.isRefreshing)
    }

    // MARK: - Call List Item Tests

    func testCallListItemIsResolved() {
        let resolvedCall = CallListItem(
            provider: "twilio",
            providerCallId: "CA123",
            fromNumber: "+420123456789",
            toNumber: "+420987654321",
            status: "completed",
            startedAt: "2024-01-15T10:30:00Z",
            endedAt: "2024-01-15T10:35:00Z",
            endedBy: "caller",
            firstViewedAt: nil,
            resolvedAt: "2024-01-15T11:00:00Z",
            resolvedBy: "user",
            screening: nil
        )

        XCTAssertTrue(resolvedCall.isResolved)
    }

    func testCallListItemIsNotResolved() {
        let unresolvedCall = CallListItem(
            provider: "twilio",
            providerCallId: "CA456",
            fromNumber: "+420123456789",
            toNumber: "+420987654321",
            status: "completed",
            startedAt: "2024-01-15T10:30:00Z",
            endedAt: "2024-01-15T10:35:00Z",
            endedBy: "caller",
            firstViewedAt: nil,
            resolvedAt: nil,
            resolvedBy: nil,
            screening: nil
        )

        XCTAssertFalse(unresolvedCall.isResolved)
    }

    func testCallListItemIsViewed() {
        let viewedCall = CallListItem(
            provider: "twilio",
            providerCallId: "CA789",
            fromNumber: "+420123456789",
            toNumber: "+420987654321",
            status: "completed",
            startedAt: "2024-01-15T10:30:00Z",
            endedAt: nil,
            endedBy: nil,
            firstViewedAt: "2024-01-15T10:40:00Z",
            resolvedAt: nil,
            resolvedBy: nil,
            screening: nil
        )

        XCTAssertTrue(viewedCall.isViewed)
    }

    func testCallListItemStartDate() {
        let call = CallListItem(
            provider: "twilio",
            providerCallId: "CA999",
            fromNumber: "+420123456789",
            toNumber: "+420987654321",
            status: "completed",
            startedAt: "2024-01-15T10:30:00Z",
            endedAt: nil,
            endedBy: nil,
            firstViewedAt: nil,
            resolvedAt: nil,
            resolvedBy: nil,
            screening: nil
        )

        XCTAssertNotNil(call.startDate)
    }
}
