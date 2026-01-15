import Foundation
import SwiftUI

@MainActor
class CallDetailViewModel: ObservableObject {
    @Published var call: CallDetail?
    @Published var isLoading = false
    @Published var error: String?
    @Published var isResolved = false
    @Published var isTogglingResolved = false

    private let callService = CallService.shared
    private let providerCallId: String

    init(providerCallId: String) {
        self.providerCallId = providerCallId
    }

    // MARK: - Load Call

    func loadCall() async {
        guard !isLoading else { return }

        isLoading = true
        error = nil

        do {
            call = try await callService.getCall(providerCallId: providerCallId)
            isResolved = call?.isResolved ?? false

            // Mark as viewed (fire and forget)
            Task {
                try? await callService.markCallViewed(providerCallId: providerCallId)
            }
        } catch {
            self.error = error.localizedDescription
        }

        isLoading = false
    }

    // MARK: - Resolve Actions

    func toggleResolved() async {
        guard !isTogglingResolved else { return }

        isTogglingResolved = true
        let wasResolved = isResolved

        // Optimistic update
        isResolved = !wasResolved

        do {
            if wasResolved {
                try await callService.markCallUnresolved(providerCallId: providerCallId)
            } else {
                try await callService.markCallResolved(providerCallId: providerCallId)
            }
        } catch {
            // Revert on error
            isResolved = wasResolved
            print("Failed to toggle resolved: \(error)")
        }

        isTogglingResolved = false
    }

    // MARK: - Computed Properties

    var legitimacyLabel: LegitimacyLabel {
        LegitimacyLabel(from: call?.screening?.legitimacyLabel ?? "unknown")
    }

    var leadLabel: LeadLabel {
        LeadLabel(from: call?.screening?.leadLabel ?? "unknown")
    }

    var formattedStatus: String {
        guard let status = call?.status else { return "" }
        switch status {
        case "in_progress": return "Probiha"
        case "completed": return "Dokonceno"
        case "queued": return "Ceka"
        case "ringing": return "Vyzvan"
        default: return status
        }
    }

    var endedByText: String? {
        guard let endedBy = call?.endedBy else { return nil }
        switch endedBy {
        case "agent": return "Zavesila asistentka"
        case "caller": return "Zavesil volajici"
        default: return nil
        }
    }
}
