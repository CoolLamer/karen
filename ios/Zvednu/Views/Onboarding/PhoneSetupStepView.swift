import SwiftUI

struct PhoneSetupStepView: View {
    @ObservedObject var viewModel: OnboardingViewModel

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                VStack(spacing: 8) {
                    Text(viewModel.hasPhoneNumber ? "Tvoje Karen cislo" : "Skoro hotovo!")
                        .font(.title2)
                        .fontWeight(.bold)

                    Text(viewModel.hasPhoneNumber
                         ? "Toto je cislo, na ktere presmerujes hovory"
                         : "Cislo ti pridelime co nejdrive")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                }
                .padding(.top, 20)

                if viewModel.hasPhoneNumber {
                    phoneNumberSection
                } else {
                    noPhoneSection
                }

                Spacer(minLength: 20)

                Button {
                    viewModel.goToNext()
                } label: {
                    HStack {
                        Text("Pokracovat")
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

    private var phoneNumberSection: some View {
        VStack(spacing: 16) {
            // Phone number display
            VStack(spacing: 8) {
                Text(viewModel.primaryPhoneNumber ?? "")
                    .font(.title)
                    .fontWeight(.bold)
                    .foregroundStyle(Color.accentColor)

                Button {
                    UIPasteboard.general.string = viewModel.primaryPhoneNumber?.replacingOccurrences(of: " ", with: "")
                } label: {
                    Label("Kopirovat", systemImage: "doc.on.doc")
                        .font(.caption)
                }
            }
            .padding()
            .frame(maxWidth: .infinity)
            .background(Color.accentColor.opacity(0.1))
            .clipShape(RoundedRectangle(cornerRadius: 12))

            // Ready notice
            HStack(spacing: 8) {
                Image(systemName: "checkmark.circle.fill")
                    .foregroundStyle(.green)
                Text("Karen je pripravena! Muzes ji hned vyzkouset zavolanim na cislo vyse z jineho telefonu.")
                    .font(.subheadline)
            }
            .padding()
            .background(Color.green.opacity(0.1))
            .clipShape(RoundedRectangle(cornerRadius: 12))

            // Usage explanation
            VStack(alignment: .leading, spacing: 12) {
                Text("Jak Karen pouzivat dlouhodobe?")
                    .font(.headline)
                    .foregroundStyle(.blue)

                VStack(alignment: .leading, spacing: 8) {
                    Text("**Varianta A:** Nastav presmerovani hovoru (doporuceno) - kdyz nezvednes, hovor se automaticky prepoji na Karen.")
                        .font(.subheadline)

                    Text("**Varianta B:** Zavolej primo na Karen cislo - idealni pro rychle vyzkouseni.")
                        .font(.subheadline)
                }
                .foregroundStyle(.secondary)
            }
            .padding()
            .background(Color.blue.opacity(0.1))
            .clipShape(RoundedRectangle(cornerRadius: 12))

            // Forwarding instructions
            forwardingInstructions
        }
    }

    private var forwardingInstructions: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Nastaveni presmerovani")
                .font(.headline)

            Text("Presmerovani se nastavuje vytocenim specialniho kodu na telefonu. Otevre tuto stranku na mobilu a klikni na tlacitko.")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            // Forwarding code examples
            VStack(spacing: 8) {
                forwardingCodeRow(
                    title: "Kdyz nezvednes",
                    code: "**61*\(viewModel.primaryPhoneNumber?.replacingOccurrences(of: " ", with: "") ?? "")#",
                    description: "Aktivuje presmerovani pri neprijeti"
                )

                forwardingCodeRow(
                    title: "Kdyz mas obsazeno",
                    code: "**67*\(viewModel.primaryPhoneNumber?.replacingOccurrences(of: " ", with: "") ?? "")#",
                    description: "Aktivuje presmerovani pri obsazeni"
                )

                forwardingCodeRow(
                    title: "Kdyz jsi nedostupny",
                    code: "**62*\(viewModel.primaryPhoneNumber?.replacingOccurrences(of: " ", with: "") ?? "")#",
                    description: "Aktivuje presmerovani pri nedostupnosti"
                )
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private func forwardingCodeRow(title: String, code: String, description: String) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(title)
                .font(.subheadline)
                .fontWeight(.medium)

            HStack {
                Text(code)
                    .font(.system(.caption, design: .monospaced))
                    .foregroundStyle(Color.accentColor)

                Spacer()

                if let url = URL(string: "tel:\(code)") {
                    Link(destination: url) {
                        Text("Vytocit")
                            .font(.caption)
                            .padding(.horizontal, 12)
                            .padding(.vertical, 6)
                            .background(Color.accentColor)
                            .foregroundStyle(.white)
                            .clipShape(Capsule())
                    }
                }
            }

            Text(description)
                .font(.caption2)
                .foregroundStyle(.secondary)
        }
        .padding()
        .background(Color(.systemGray6))
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }

    private var noPhoneSection: some View {
        VStack(spacing: 16) {
            Image(systemName: "exclamationmark.triangle")
                .font(.system(size: 40))
                .foregroundStyle(.orange)

            Text("Momentalne nemame volne cislo. Jakmile bude dostupne, pridelime ti ho a oznamime ti to. Presmerovani nastavis v nastaveni.")
                .font(.subheadline)
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)
        }
        .padding()
        .background(Color.orange.opacity(0.1))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

#Preview {
    PhoneSetupStepView(viewModel: {
        let vm = OnboardingViewModel()
        vm.phoneNumbers = [TenantPhoneNumber(id: "1", twilioNumber: "+420 123 456 789", isPrimary: true)]
        return vm
    }())
}
