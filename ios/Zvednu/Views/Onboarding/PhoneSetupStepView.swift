import SwiftUI

struct PhoneSetupStepView: View {
    @ObservedObject var viewModel: OnboardingViewModel

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                VStack(spacing: 8) {
                    Text(viewModel.hasPhoneNumber ? "Tvoje Karen číslo" : "Skoro hotovo!")
                        .font(.title2)
                        .fontWeight(.bold)

                    Text(viewModel.hasPhoneNumber
                         ? "Toto je číslo, na které přesměruješ hovory"
                         : "Číslo ti přidělíme co nejdříve")
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
                        Text("Pokračovat")
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

                HStack(spacing: 16) {
                    Button {
                        UIPasteboard.general.string = viewModel.primaryPhoneNumber?.replacingOccurrences(of: " ", with: "")
                    } label: {
                        Label("Kopírovat", systemImage: "doc.on.doc")
                            .font(.caption)
                    }

                    if let phoneNumber = viewModel.primaryPhoneNumber,
                       let url = URL(string: "tel:\(phoneNumber.replacingOccurrences(of: " ", with: ""))") {
                        Link(destination: url) {
                            Label("Zavolat", systemImage: "phone.fill")
                                .font(.caption)
                        }
                    }
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
                Text("Karen je připravena! Můžeš ji hned vyzkoušet zavoláním na číslo výše.")
                    .font(.subheadline)
            }
            .padding()
            .background(Color.green.opacity(0.1))
            .clipShape(RoundedRectangle(cornerRadius: 12))

            // Usage explanation
            VStack(alignment: .leading, spacing: 12) {
                Text("Jak Karen používat dlouhodobě?")
                    .font(.headline)
                    .foregroundStyle(.blue)

                VStack(alignment: .leading, spacing: 8) {
                    Text("**Varianta A:** Nastav přesměrování hovoru (doporučeno) - když nezvedneš, hovor se automaticky přepojí na Karen.")
                        .font(.subheadline)

                    Text("**Varianta B:** Zavolej přímo na Karen číslo - ideální pro rychlé vyzkoušení.")
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
            Text("Nastavení přesměrování")
                .font(.headline)

            Text("Přesměrování se nastavuje vytočením speciálního kódu na telefonu. Otevři tuto stránku na mobilu a klikni na tlačítko.")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            // Warning about existing forwarding
            HStack(spacing: 8) {
                Image(systemName: "exclamationmark.triangle.fill")
                    .foregroundStyle(.orange)
                Text("**Zobrazuje se chyba?** Pokud máš již nastavené přesměrování na jiné číslo, musíš ho nejdřív zrušit kódem ##61#, ##67# nebo ##62#.")
                    .font(.caption)
            }
            .padding(10)
            .background(Color.orange.opacity(0.1))
            .clipShape(RoundedRectangle(cornerRadius: 8))

            // Forwarding code examples
            VStack(spacing: 8) {
                forwardingCodeRow(
                    title: "Když nezvedneš",
                    code: "**61*\(viewModel.primaryPhoneNumber?.replacingOccurrences(of: " ", with: "") ?? "")#",
                    description: "Aktivuje přesměrování při nepřijetí"
                )

                forwardingCodeRow(
                    title: "Když máš obsazeno",
                    code: "**67*\(viewModel.primaryPhoneNumber?.replacingOccurrences(of: " ", with: "") ?? "")#",
                    description: "Aktivuje přesměrování při obsazení"
                )

                forwardingCodeRow(
                    title: "Když jsi nedostupný",
                    code: "**62*\(viewModel.primaryPhoneNumber?.replacingOccurrences(of: " ", with: "") ?? "")#",
                    description: "Aktivuje přesměrování při nedostupnosti"
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
                        Text("Vytočit")
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

            Text("Momentálně nemáme volné číslo. Jakmile bude dostupné, přidělíme ti ho a oznámíme ti to. Přesměrování nastavíš v nastavení.")
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
