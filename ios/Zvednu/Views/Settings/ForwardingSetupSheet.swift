import SwiftUI

struct ForwardingSetupSheet: View {
    @Environment(\.dismiss) private var dismiss
    let karenNumber: String

    enum SetupMode: String, CaseIterable {
        case wizard
        case quick

        var label: String {
            switch self {
            case .wizard: return "Průvodce krok za krokem"
            case .quick: return "Rychlé nastavení"
            }
        }
    }

    @State private var mode: SetupMode = .wizard

    var body: some View {
        NavigationStack {
            VStack(spacing: 16) {
                // Mode picker
                Picker("Režim", selection: $mode) {
                    ForEach(SetupMode.allCases, id: \.self) { setupMode in
                        Text(setupMode.label).tag(setupMode)
                    }
                }
                .pickerStyle(.segmented)
                .padding(.horizontal)

                // Content based on mode
                if mode == .wizard {
                    ScrollView {
                        RedirectWizardView(
                            karenNumber: karenNumber,
                            onComplete: { dismiss() }
                        )
                        .padding()
                    }
                } else {
                    RedirectQuickSetupView(karenNumber: karenNumber)
                }

                Spacer()
            }
            .navigationTitle("Nastavení přesměrování")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Zavřít") {
                        dismiss()
                    }
                }
            }
        }
    }
}

#Preview("Wizard Mode") {
    ForwardingSetupSheet(karenNumber: "+420 123 456 789")
}

#Preview("Quick Mode") {
    ForwardingSetupSheet(karenNumber: "+420 123 456 789")
}
