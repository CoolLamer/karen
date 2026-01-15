import SwiftUI

struct CallRowView: View {
    let call: CallListItem

    private var legitimacyLabel: LegitimacyLabel {
        LegitimacyLabel(from: call.screening?.legitimacyLabel ?? "unknown")
    }

    private var leadLabel: LeadLabel {
        LeadLabel(from: call.screening?.leadLabel ?? "unknown")
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            // Phone number and time
            HStack {
                HStack(spacing: 6) {
                    // Unread indicator
                    if !call.isViewed {
                        Circle()
                            .fill(Color.accentColor)
                            .frame(width: 8, height: 8)
                    }
                    Text(call.fromNumber.formattedPhoneNumber())
                        .font(.headline)
                        .fontWeight(call.isViewed ? .regular : .semibold)
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
