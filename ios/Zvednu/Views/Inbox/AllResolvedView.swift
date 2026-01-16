import SwiftUI

struct AllResolvedView: View {
    var body: some View {
        VStack(spacing: 20) {
            Image(systemName: "checkmark.circle")
                .font(.system(size: 60))
                .foregroundStyle(.teal)

            VStack(spacing: 8) {
                Text("Všechny hovory vyřešené")
                    .font(.title2)
                    .fontWeight(.semibold)

                Text("Vypni filtr pro zobrazení všech hovorů.")
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
    AllResolvedView()
}
