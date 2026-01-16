import SwiftUI

struct CallRowView: View {
    let call: CallListItem
    @ObservedObject private var contactsManager = ContactsManager.shared

    private var legitimacyLabel: LegitimacyLabel {
        LegitimacyLabel(from: call.screening?.legitimacyLabel ?? "unknown")
    }

    private var leadLabel: LeadLabel {
        LeadLabel(from: call.screening?.leadLabel ?? "unknown")
    }

    private var displayName: String {
        if let contactName = contactsManager.contactName(for: call.fromNumber) {
            return contactName
        }
        return call.fromNumber.formattedPhoneNumber()
    }

    private var showPhoneSubtitle: Bool {
        contactsManager.contactName(for: call.fromNumber) != nil
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            // Contact name/phone number and time
            HStack {
                HStack(spacing: 6) {
                    // Unread indicator
                    if !call.isViewed {
                        Circle()
                            .fill(Color.accentColor)
                            .frame(width: 8, height: 8)
                    }
                    VStack(alignment: .leading, spacing: 2) {
                        Text(displayName)
                            .font(.headline)
                            .fontWeight(call.isViewed ? .regular : .semibold)

                        if showPhoneSubtitle {
                            Text(call.fromNumber.formattedPhoneNumber())
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                    }
                }

                Spacer()

                Text(call.startDate?.smartFormat() ?? "")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            // Intent text or status
            if let intentText = call.screening?.intentText, !intentText.isEmpty {
                Text(intentText)
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .lineLimit(2)
            }

            // Badges
            HStack(spacing: 8) {
                LegitimacyBadge(label: legitimacyLabel)
                LeadBadge(label: leadLabel)

                Spacer()

                if call.isResolved {
                    Image(systemName: "checkmark.circle.fill")
                        .font(.caption)
                        .foregroundStyle(.green)
                }
            }
        }
        .padding(.vertical, 8)
        .contentShape(Rectangle())
    }
}

#Preview {
    List {
        CallRowView(call: CallListItem(
            provider: "twilio",
            providerCallId: "CA123",
            fromNumber: "+420123456789",
            toNumber: "+420987654321",
            status: "completed",
            startedAt: "2024-01-15T10:30:00Z",
            endedAt: nil,
            endedBy: nil,
            firstViewedAt: nil,
            resolvedAt: nil,
            resolvedBy: nil,
            screening: ScreeningResult(
                legitimacyLabel: "legitimni",
                legitimacyConfidence: 0.95,
                leadLabel: "hot_lead",
                intentCategory: "inquiry",
                intentText: "Dotaz na cenu sluzby",
                entitiesJson: nil,
                createdAt: "2024-01-15T10:35:00Z"
            )
        ))
    }
    .listStyle(.plain)
}
