import SwiftUI

struct CompleteStepView: View {
    @ObservedObject var viewModel: OnboardingViewModel

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                Spacer(minLength: 40)

                Image(systemName: "party.popper.fill")
                    .font(.system(size: 80))
                    .foregroundStyle(.green)

                VStack(spacing: 8) {
                    Text("Hotovo! Karen je připravena.")
                        .font(.title2)
                        .fontWeight(.bold)

                    Text("Když někdo zavolá a ty nezvedneš, Karen to vyřídí za tebe.")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                }

                // Test instructions
                if viewModel.hasPhoneNumber {
                    VStack(alignment: .leading, spacing: 12) {
                        Text("Vyzkoušej Karen!")
                            .font(.headline)
                            .foregroundStyle(Color.accentColor)

                        VStack(alignment: .leading, spacing: 8) {
                            Text("**S přesměrováním:** Zavolej na své číslo z jiného telefonu a nech vyzvánět 20 sekund. Karen zvedne.")
                                .font(.subheadline)

                            Text("**Přímo:** Zavolej na \(viewModel.primaryPhoneNumber ?? "") - Karen zvedne okamžitě.")
                                .font(.subheadline)
                        }
                        .foregroundStyle(.secondary)
                    }
                    .padding()
                    .background(Color.accentColor.opacity(0.1))
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                }

                // Next steps
                VStack(alignment: .leading, spacing: 12) {
                    Text("Co dál:")
                        .font(.headline)

                    VStack(alignment: .leading, spacing: 8) {
                        nextStepRow(icon: "list.bullet", text: "Přehled všech hovorů najdeš v aplikaci")
                        nextStepRow(icon: "gearshape", text: "Nastavení můžeš kdykoliv změnit v sekci Nastavení")

                        if viewModel.vipNames.isEmpty {
                            nextStepRow(icon: "star", text: "Přidej VIP kontakty, které má Karen vždy přepojit")
                        }
                    }
                }
                .padding()
                .background(Color(.systemBackground))
                .clipShape(RoundedRectangle(cornerRadius: 12))

                Spacer(minLength: 20)

                Button {
                    Task {
                        await viewModel.finish()
                    }
                } label: {
                    HStack {
                        Text("Jít do přehledu hovorů")
                        Image(systemName: "arrow.right")
                    }
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(Color.accentColor)
                    .foregroundStyle(.white)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                }
            }
            .padding()
        }
    }

    private func nextStepRow(icon: String, text: String) -> some View {
        HStack(spacing: 12) {
            Image(systemName: icon)
                .foregroundStyle(.secondary)
                .frame(width: 20)
            Text(text)
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
    }
}

#Preview {
    CompleteStepView(viewModel: {
        let vm = OnboardingViewModel()
        vm.phoneNumbers = [TenantPhoneNumber(id: "1", twilioNumber: "+420 123 456 789", isPrimary: true)]
        return vm
    }())
}
