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
                    Text("Hotovo! Karen je pripravena.")
                        .font(.title2)
                        .fontWeight(.bold)

                    Text("Kdyz nekdo zavola a ty nezvednes, Karen to vyridi za tebe.")
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
                            Text("**S presmerovanim:** Zavolej na sve cislo z jineho telefonu a nech vyzvanet 20 sekund. Karen zvedne.")
                                .font(.subheadline)

                            Text("**Primo:** Zavolej na \(viewModel.primaryPhoneNumber ?? "") - Karen zvedne okamzite.")
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
                    Text("Co dal:")
                        .font(.headline)

                    VStack(alignment: .leading, spacing: 8) {
                        nextStepRow(icon: "list.bullet", text: "Prehled vsech hovoru najdes v aplikaci")
                        nextStepRow(icon: "gearshape", text: "Nastaveni muzes kdykoliv zmenit v sekci Nastaveni")

                        if viewModel.vipNames.isEmpty {
                            nextStepRow(icon: "star", text: "Pridej VIP kontakty, ktere ma Karen vzdy prepojit")
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
                        Text("Jit do prehledu hovoru")
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
