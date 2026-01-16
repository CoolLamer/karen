import Foundation
import SwiftUI

@MainActor
class CallInboxViewModel: ObservableObject {
    @Published var calls: [CallListItem] = []
    @Published var unresolvedCount = 0
    @Published var billing: BillingInfo?
    @Published var isLoading = false
    @Published var isRefreshing = false
    @Published var error: String?
    @AppStorage("hideResolvedCalls") var hideResolved = false

    var filteredCalls: [CallListItem] {
        hideResolved ? calls.filter { !$0.isResolved } : calls
    }

    private let callService = CallService.shared
    private let tenantService = TenantService.shared

    // MARK: - Load Calls

    func loadCalls() async {
        guard !isLoading else { return }

        isLoading = true
        error = nil

        do {
            async let callsTask = callService.listCalls()
            async let countTask = callService.getUnresolvedCount()
            async let billingTask = tenantService.getBilling()

            let (fetchedCalls, count, billingInfo) = try await (callsTask, countTask, billingTask)

            calls = fetchedCalls.sorted { ($0.startDate ?? .distantPast) > ($1.startDate ?? .distantPast) }
            unresolvedCount = count
            billing = billingInfo
        } catch {
            self.error = error.localizedDescription
        }

        isLoading = false
    }

    func refreshCalls() async {
        guard !isRefreshing else { return }

        isRefreshing = true

        do {
            async let callsTask = callService.listCalls()
            async let countTask = callService.getUnresolvedCount()
            async let billingTask = tenantService.getBilling()

            let (fetchedCalls, count, billingInfo) = try await (callsTask, countTask, billingTask)

            calls = fetchedCalls.sorted { ($0.startDate ?? .distantPast) > ($1.startDate ?? .distantPast) }
            unresolvedCount = count
            billing = billingInfo
        } catch {
            print("Refresh error: \(error)")
        }

        isRefreshing = false
    }

    // MARK: - Call Actions

    func markAsResolved(_ call: CallListItem) async {
        do {
            try await callService.markCallResolved(providerCallId: call.providerCallId)
            await refreshCalls()
        } catch {
            print("Failed to mark as resolved: \(error)")
        }
    }
}
