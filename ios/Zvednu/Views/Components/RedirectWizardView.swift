import SwiftUI

struct RedirectWizardView: View {
    @StateObject private var viewModel = RedirectWizardViewModel()
    let karenNumber: String
    let onComplete: () -> Void

    private var cleanKarenNumber: String {
        karenNumber.replacingOccurrences(of: " ", with: "")
    }

    var body: some View {
        // Error state when Karen number is not available
        if cleanKarenNumber.isEmpty {
            emptyNumberState
        } else {
            wizardContent
        }
    }

    private var emptyNumberState: some View {
        VStack(spacing: 16) {
            HStack(spacing: 8) {
                Image(systemName: "exclamationmark.triangle.fill")
                    .foregroundStyle(.orange)
                Text("Číslo Karen ještě není přiděleno. Přesměrování budeš moct nastavit později v Nastavení.")
                    .font(.subheadline)
            }
            .padding()
            .background(Color.orange.opacity(0.1))
            .clipShape(RoundedRectangle(cornerRadius: 12))

            Button {
                onComplete()
            } label: {
                Text("Pokračovat bez nastavení")
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(Color.accentColor.opacity(0.1))
                    .foregroundStyle(Color.accentColor)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
            }
        }
    }

    private var wizardContent: some View {
        VStack(spacing: 20) {
            // Progress bar (hidden on intro/complete)
            if viewModel.currentStep != .intro && viewModel.currentStep != .complete {
                WizardProgressView(viewModel: viewModel)
            }

            // Step content
            switch viewModel.currentStep {
            case .intro:
                IntroStepContent(onStart: viewModel.goToNextStep)
            case .clear:
                DialStepContent(
                    viewModel: viewModel,
                    step: .clear,
                    karenNumber: nil,
                    dialCode: viewModel.getDialCode(for: .clear, phoneNumber: "")
                )
            case .noAnswer:
                DialStepContent(
                    viewModel: viewModel,
                    step: .noAnswer,
                    karenNumber: karenNumber,
                    dialCode: viewModel.getDialCode(for: .noAnswer, phoneNumber: cleanKarenNumber),
                    showTimingControl: true
                )
            case .busy:
                DialStepContent(
                    viewModel: viewModel,
                    step: .busy,
                    karenNumber: karenNumber,
                    dialCode: viewModel.getDialCode(for: .busy, phoneNumber: cleanKarenNumber)
                )
            case .unreachable:
                DialStepContent(
                    viewModel: viewModel,
                    step: .unreachable,
                    karenNumber: karenNumber,
                    dialCode: viewModel.getDialCode(for: .unreachable, phoneNumber: cleanKarenNumber)
                )
            case .complete:
                CompleteStepContent(viewModel: viewModel, onFinish: onComplete)
            }
        }
    }
}

// MARK: - Progress View

struct WizardProgressView: View {
    @ObservedObject var viewModel: RedirectWizardViewModel

    private let steps: [(step: RedirectWizardViewModel.WizardStep, label: String)] = [
        (.clear, "Vymazat"),
        (.noAnswer, "Nezvedám"),
        (.busy, "Obsazeno"),
        (.unreachable, "Nedostupný"),
    ]

    var body: some View {
        VStack(spacing: 8) {
            ProgressView(value: viewModel.progressPercent, total: 100)
                .tint(.accentColor)

            HStack(spacing: 4) {
                ForEach(steps, id: \.step) { item in
                    stepIndicator(for: item.step, label: item.label)
                    if item.step != .unreachable {
                        Spacer()
                    }
                }
            }
        }
    }

    @ViewBuilder
    private func stepIndicator(for step: RedirectWizardViewModel.WizardStep, label: String) -> some View {
        let status = viewModel.status(for: step)
        let isCurrent = viewModel.currentStep == step

        HStack(spacing: 4) {
            stepIcon(status: status, isCurrent: isCurrent)
                .font(.caption)

            Text(label)
                .font(.caption)
                .foregroundStyle(isCurrent ? Color.accentColor : status == .completed ? Color.green : Color.secondary)
                .fontWeight(isCurrent ? .medium : .regular)
        }
    }

    @ViewBuilder
    private func stepIcon(status: RedirectWizardViewModel.StepStatus, isCurrent: Bool) -> some View {
        if status == .completed {
            Image(systemName: "checkmark.circle.fill")
                .foregroundStyle(.green)
        } else if status == .skipped {
            Image(systemName: "forward.fill")
                .foregroundStyle(.gray)
        } else if isCurrent {
            Image(systemName: "circle.fill")
                .foregroundStyle(Color.accentColor)
        } else {
            Image(systemName: "circle")
                .foregroundStyle(.gray.opacity(0.5))
        }
    }
}

// MARK: - Intro Step

struct IntroStepContent: View {
    let onStart: () -> Void

    var body: some View {
        VStack(spacing: 20) {
            VStack(spacing: 8) {
                Text("Nastavíme přesměrování hovorů")
                    .font(.title3)
                    .fontWeight(.semibold)

                Text("Provedeme tě 4 krátkými kroky. Na mobilu stačí klikat na tlačítka.")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
            }

            VStack(alignment: .leading, spacing: 8) {
                ForEach(
                    [
                        "Vymažeme stávající přesměrování",
                        "Nastavíme přesměrování, když nezvedáš",
                        "Nastavíme přesměrování, když máš obsazeno",
                        "Nastavíme přesměrování, když jsi nedostupný",
                    ], id: \.self
                ) { item in
                    HStack(spacing: 8) {
                        Image(systemName: "circle.fill")
                            .font(.system(size: 6))
                            .foregroundStyle(.secondary)
                        Text(item)
                            .font(.subheadline)
                    }
                }
            }
            .padding()
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(Color(.systemGray6))
            .clipShape(RoundedRectangle(cornerRadius: 12))

            Text("Každý krok můžeš přeskočit, pokud nechceš daný typ přesměrování nastavit.")
                .font(.caption)
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)

            Button {
                onStart()
            } label: {
                HStack {
                    Text("Začít")
                    Image(systemName: "arrow.right")
                }
                .frame(maxWidth: .infinity)
                .padding()
                .background(Color.accentColor)
                .foregroundStyle(.white)
                .clipShape(RoundedRectangle(cornerRadius: 12))
            }
        }
    }
}

// MARK: - Complete Step

struct CompleteStepContent: View {
    @ObservedObject var viewModel: RedirectWizardViewModel
    let onFinish: () -> Void

    var body: some View {
        VStack(spacing: 20) {
            VStack(spacing: 12) {
                Image(systemName: "checkmark.circle.fill")
                    .font(.system(size: 50))
                    .foregroundStyle(.green)

                Text("Nastavení dokončeno!")
                    .font(.title3)
                    .fontWeight(.semibold)
            }

            if !viewModel.activatedSteps.isEmpty {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Aktivováno (\(viewModel.activatedSteps.count))")
                        .font(.subheadline)
                        .fontWeight(.medium)
                        .foregroundStyle(.green)

                    ForEach(viewModel.activatedSteps, id: \.key) { item in
                        HStack(spacing: 8) {
                            Image(systemName: "checkmark")
                                .font(.caption)
                                .foregroundStyle(.green)
                            Text(item.label)
                                .font(.subheadline)
                        }
                    }
                }
                .padding()
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(Color.green.opacity(0.1))
                .clipShape(RoundedRectangle(cornerRadius: 12))
            }

            if !viewModel.skippedSteps.isEmpty {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Přeskočeno (\(viewModel.skippedSteps.count))")
                        .font(.subheadline)
                        .fontWeight(.medium)
                        .foregroundStyle(.secondary)

                    ForEach(viewModel.skippedSteps, id: \.key) { item in
                        HStack(spacing: 8) {
                            Image(systemName: "forward.fill")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                            Text(item.label)
                                .font(.subheadline)
                                .foregroundStyle(.secondary)
                        }
                    }
                }
                .padding()
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(Color(.systemGray6))
                .clipShape(RoundedRectangle(cornerRadius: 12))
            }

            Text("Nastavení můžeš kdykoliv změnit v sekci Nastavení.")
                .font(.caption)
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)

            Button {
                onFinish()
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
    }
}

// MARK: - Preview

#Preview("Wizard - With Number") {
    RedirectWizardView(
        karenNumber: "+420 123 456 789",
        onComplete: {}
    )
    .padding()
}

#Preview("Wizard - No Number") {
    RedirectWizardView(
        karenNumber: "",
        onComplete: {}
    )
    .padding()
}
