import SwiftUI

struct LegitimacyBadge: View {
    let label: LegitimacyLabel

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
        LegitimacyBadge(label: .legitimate)
        LegitimacyBadge(label: .spam)
        LegitimacyBadge(label: .marketing)
        LegitimacyBadge(label: .scam)
        LegitimacyBadge(label: .unknown)
    }
    .padding()
}
