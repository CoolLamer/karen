import SwiftUI

struct EmptyInboxView: View {
    var body: some View {
        VStack(spacing: 20) {
            Image(systemName: "phone.badge.checkmark")
                .font(.system(size: 60))
                .foregroundStyle(.secondary)

            VStack(spacing: 8) {
                Text("Žádné hovory")
                    .font(.title2)
                    .fontWeight(.semibold)

                Text("Když někdo zavolá a Karen zvedne, hovor se objeví zde.")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
            }
        }
        .padding()
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

#Preview {
    EmptyInboxView()
}
