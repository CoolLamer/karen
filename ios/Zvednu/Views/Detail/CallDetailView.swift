import SwiftUI

struct CallDetailView: View {
    let providerCallId: String
    @StateObject private var viewModel: CallDetailViewModel
    @ObservedObject private var contactsManager = ContactsManager.shared
    @Environment(\.dismiss) private var dismiss

    init(providerCallId: String) {
        self.providerCallId = providerCallId
        self._viewModel = StateObject(wrappedValue: CallDetailViewModel(providerCallId: providerCallId))
    }

    var body: some View {
        Group {
            if let error = viewModel.error {
                errorView(error)
            } else if let call = viewModel.call {
                callDetailContent(call)
            } else {
                LoadingView(message: "Načítám detail hovoru...")
            }
        }
        .navigationTitle("Detail hovoru")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .navigationBarTrailing) {
                resolveButton
            }
        }
        .task {
            await viewModel.loadCall()
        }
    }

    // MARK: - Resolve Button

    private var resolveButton: some View {
        Button {
            Task {
                await viewModel.toggleResolved()
            }
        } label: {
            HStack(spacing: 4) {
                if viewModel.isTogglingResolved {
                    ProgressView()
                        .scaleEffect(0.8)
                } else {
                    Image(systemName: viewModel.isResolved ? "checkmark.circle.fill" : "circle")
                }
                Text(viewModel.isResolved ? "Vyřešeno" : "Označit jako vyřešené")
                    .font(.subheadline)
            }
            .foregroundStyle(viewModel.isResolved ? .gray : Color.accentColor)
        }
        .disabled(viewModel.isTogglingResolved)
    }

    // MARK: - Error View

    private func errorView(_ error: String) -> some View {
        VStack(spacing: 16) {
            Image(systemName: "exclamationmark.triangle")
                .font(.system(size: 40))
                .foregroundStyle(.red)

            Text("Chyba: \(error)")
                .foregroundStyle(.secondary)

            Button("Zkusit znovu") {
                Task {
                    await viewModel.loadCall()
                }
            }
        }
        .padding()
    }

    // MARK: - Call Detail Content

    private func callDetailContent(_ call: CallDetail) -> some View {
        ScrollView {
            VStack(spacing: 16) {
                // Call info header
                callInfoCard(call)

                // Transcript
                transcriptCard(call)
            }
            .padding()
        }
    }

    // MARK: - Call Info Card

    private func callInfoCard(_ call: CallDetail) -> some View {
        VStack(alignment: .leading, spacing: 16) {
            // Phone numbers and time
            HStack(alignment: .top) {
                VStack(alignment: .leading, spacing: 4) {
                    if let contactName = contactsManager.contactName(for: call.fromNumber) {
                        Text(contactName)
                            .font(.title2)
                            .fontWeight(.bold)

                        Text(call.fromNumber.formattedPhoneNumber())
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    } else {
                        Text(call.fromNumber.formattedPhoneNumber())
                            .font(.title2)
                            .fontWeight(.bold)
                    }

                    Text("na \(call.toNumber.formattedPhoneNumber())")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)

                    Text(call.startDate?.formattedCzech() ?? "")
                        .font(.caption)
                        .foregroundStyle(.secondary)

                    // Status badges
                    HStack(spacing: 8) {
                        Text(viewModel.formattedStatus)
                            .font(.caption)
                            .padding(.horizontal, 8)
                            .padding(.vertical, 4)
                            .background(Color(.systemGray5))
                            .clipShape(Capsule())

                        if viewModel.isResolved {
                            HStack(spacing: 4) {
                                Image(systemName: "checkmark.circle.fill")
                                    .font(.caption2)
                                Text("Vyřešeno")
                                    .font(.caption)
                            }
                            .padding(.horizontal, 8)
                            .padding(.vertical, 4)
                            .background(Color.green.opacity(0.15))
                            .foregroundStyle(.green)
                            .clipShape(Capsule())
                        }
                    }
                }

                Spacer()

                // Legitimacy and lead badges
                VStack(alignment: .trailing, spacing: 8) {
                    LegitimacyBadge(label: viewModel.legitimacyLabel)
                    LeadBadge(label: viewModel.leadLabel)

                    // Confidence
                    if let confidence = call.screening?.legitimacyConfidence {
                        VStack(alignment: .trailing, spacing: 4) {
                            Text("Spolehlivost: \(Int(confidence * 100))%")
                                .font(.caption2)
                                .foregroundStyle(.secondary)

                            ProgressView(value: confidence)
                                .tint(viewModel.legitimacyLabel.color)
                                .frame(width: 80)
                        }
                    }
                }
            }

            // Intent text
            if let intentText = call.screening?.intentText, !intentText.isEmpty {
                VStack(alignment: .leading, spacing: 4) {
                    Text("Účel hovoru:")
                        .font(.caption)
                        .fontWeight(.medium)

                    Text(intentText)
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }
                .padding()
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(Color(.systemGray6))
                .clipShape(RoundedRectangle(cornerRadius: 8))
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .shadow(color: .black.opacity(0.05), radius: 5, y: 2)
    }

    // MARK: - Transcript Card

    private func transcriptCard(_ call: CallDetail) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Přepis hovoru")
                .font(.headline)

            if call.utterances.isEmpty {
                Text("Přepis zatím není k dispozici.")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity, alignment: .center)
                    .padding()
            } else {
                VStack(spacing: 12) {
                    ForEach(call.utterances) { utterance in
                        TranscriptBubbleView(utterance: utterance)
                    }

                    // Ended by text
                    if let endedByText = viewModel.endedByText {
                        Text("- \(endedByText) -")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                            .italic()
                            .frame(maxWidth: .infinity)
                    }
                }
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .shadow(color: .black.opacity(0.05), radius: 5, y: 2)
    }
}

#Preview {
    NavigationStack {
        CallDetailView(providerCallId: "CA123")
    }
}
