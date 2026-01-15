import SwiftUI

struct LeadBadge: View {
    let label: LeadLabel

    var body: some View {
        HStack(spacing: 4) {
            Image(systemName: label.iconName)
                .font(.caption2)
            Text(label.displayText)
                .font(.caption2)
                .fontWeight(.medium)
        }
        .padding(.horizontal, 8)
        .padding(.vertical, 4)
        .background(label.color.opacity(0.15))
        .foregroundStyle(label.color)
        .clipShape(Capsule())
    }
}

#Preview {
    VStack(spacing: 8) {
        LeadBadge(label: .hotLead)
        LeadBadge(label: .warmLead)
        LeadBadge(label: .coldLead)
        LeadBadge(label: .notALead)
        LeadBadge(label: .unknown)
    }
    .padding()
}
