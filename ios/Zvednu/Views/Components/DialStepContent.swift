import SwiftUI

struct DialStepContent: View {
    @ObservedObject var viewModel: RedirectWizardViewModel
    let step: RedirectWizardViewModel.WizardStep
    let karenNumber: String?
    let dialCode: String
    var showTimingControl: Bool = false

    var body: some View {
        VStack(spacing: 16) {
            // Title and description
            VStack(spacing: 8) {
                Text(step.title)
                    .font(.headline)

                Text(viewModel.getDescription(for: step))
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
            }

            // Karen number display (if applicable)
            if let number = karenNumber {
                karenNumberCard(number: number)
            }

            // Dial code display with copy button
            dialCodeCard

            // Timing control (for noAnswer step)
            if showTimingControl {
                timingSelector
            }

            // Action buttons
            if !viewModel.hasDialed {
                primaryActions
            } else {
                confirmationActions
            }
        }
    }

    private func karenNumberCard(number: String) -> some View {
        VStack(spacing: 4) {
            Text("Karen číslo")
                .font(.caption)
                .foregroundStyle(.secondary)

            Text(number)
                .font(.title2)
                .fontWeight(.bold)
                .foregroundStyle(Color.accentColor)
        }
        .padding()
        .frame(maxWidth: .infinity)
        .background(Color.accentColor.opacity(0.1))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var dialCodeCard: some View {
        VStack(spacing: 4) {
            Text("Kód k vytočení")
                .font(.caption)
                .foregroundStyle(.secondary)

            HStack(spacing: 8) {
                Text(dialCode)
                    .font(.system(.body, design: .monospaced))
                    .fontWeight(.medium)

                Button {
                    UIPasteboard.general.string = dialCode
                } label: {
                    Image(systemName: "doc.on.doc")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }
        }
        .padding()
        .frame(maxWidth: .infinity)
        .background(Color(.systemGray6))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var timingSelector: some View {
        VStack(alignment: .leading, spacing: 8) {
            if !viewModel.showTimingSelector {
                Button {
                    viewModel.toggleTimingSelector()
                } label: {
                    Text("Změnit časování (\(viewModel.noAnswerTime)s)")
                        .font(.caption)
                        .foregroundStyle(Color.accentColor)
                }
            } else {
                Text("Po kolika sekundách přesměrovat?")
                    .font(.caption)
                    .fontWeight(.medium)

                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 8) {
                        ForEach(RedirectCodes.noAnswerTimeOptions, id: \.self) { time in
                            Button {
                                viewModel.noAnswerTime = time
                            } label: {
                                Text("\(time)s")
                                    .font(.caption)
                                    .padding(.horizontal, 12)
                                    .padding(.vertical, 8)
                                    .background(
                                        viewModel.noAnswerTime == time
                                            ? Color.accentColor
                                            : Color(.systemGray5)
                                    )
                                    .foregroundStyle(
                                        viewModel.noAnswerTime == time
                                            ? .white
                                            : .primary
                                    )
                                    .clipShape(Capsule())
                            }
                        }
                    }
                }
            }
        }
    }

    private var primaryActions: some View {
        VStack(spacing: 12) {
            // Primary dial button
            if let url = URL(string: "tel:\(dialCode)") {
                Link(destination: url) {
                    HStack {
                        Image(systemName: "phone.fill")
                        Text("Vytočit kód")
                    }
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(Color.accentColor)
                    .foregroundStyle(.white)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                }
                .simultaneousGesture(TapGesture().onEnded {
                    viewModel.markDialed()
                })
            }

            // Skip button
            Button {
                viewModel.skipStep()
            } label: {
                Text("Přeskočit tento krok")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }
        }
    }

    private var confirmationActions: some View {
        VStack(spacing: 12) {
            Text("Viděl/a jsi potvrzení od operátora?")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            // Confirm button
            Button {
                viewModel.confirmStep()
            } label: {
                HStack {
                    Image(systemName: "checkmark")
                    Text("Ano, aktivováno")
                }
                .frame(maxWidth: .infinity)
                .padding()
                .background(Color.green)
                .foregroundStyle(.white)
                .clipShape(RoundedRectangle(cornerRadius: 12))
            }

            // Secondary actions
            HStack(spacing: 12) {
                Button {
                    viewModel.retryDial()
                } label: {
                    HStack {
                        Image(systemName: "arrow.clockwise")
                        Text("Zkusit znovu")
                    }
                    .font(.subheadline)
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(Color(.systemGray5))
                    .foregroundStyle(.primary)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                }

                Button {
                    viewModel.skipStep()
                } label: {
                    HStack {
                        Image(systemName: "forward.fill")
                        Text("Přeskočit")
                    }
                    .font(.subheadline)
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(Color(.systemGray5))
                    .foregroundStyle(.primary)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                }
            }
        }
    }
}

#Preview("Dial Step - Clear") {
    DialStepContent(
        viewModel: RedirectWizardViewModel(),
        step: .clear,
        karenNumber: nil,
        dialCode: "##002#"
    )
    .padding()
}

#Preview("Dial Step - No Answer") {
    DialStepContent(
        viewModel: RedirectWizardViewModel(),
        step: .noAnswer,
        karenNumber: "+420 123 456 789",
        dialCode: "**61*+420123456789**10#",
        showTimingControl: true
    )
    .padding()
}
