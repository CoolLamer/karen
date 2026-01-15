import SwiftUI

struct NameStepView: View {
    @ObservedObject var viewModel: OnboardingViewModel
    @FocusState private var isNameFocused: Bool

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                VStack(spacing: 8) {
                    Text("Jak se jmenuješ?")
                        .font(.title2)
                        .fontWeight(.bold)

                    Text("Karen bude oslovovat volající tvým jménem")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }
                .padding(.top, 20)

                if let error = viewModel.error {
                    Text(error)
                        .font(.subheadline)
                        .foregroundStyle(.red)
                        .padding()
                        .frame(maxWidth: .infinity)
                        .background(Color.red.opacity(0.1))
                        .clipShape(RoundedRectangle(cornerRadius: 8))
                }

                TextField("Lukas", text: $viewModel.name)
                    .font(.title3)
                    .padding()
                    .background(Color(.systemBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                    .focused($isNameFocused)
                    .onSubmit {
                        viewModel.generateGreeting()
                    }

                if viewModel.greetingGenerated {
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Pozdrav")
                            .font(.headline)

                        Text("Text, kterým Karen začíná hovor. Můžeš ho upravit.")
                            .font(.caption)
                            .foregroundStyle(.secondary)

                        TextEditor(text: $viewModel.greetingText)
                            .frame(minHeight: 100)
                            .padding(8)
                            .background(Color(.systemBackground))
                            .clipShape(RoundedRectangle(cornerRadius: 12))
                    }
                }

                Spacer(minLength: 20)

                Button {
                    Task {
                        await viewModel.completeNameStep()
                    }
                } label: {
                    HStack {
                        if viewModel.isLoading {
                            ProgressView()
                                .tint(.white)
                        }
                        Text("Pokračovat")
                        Image(systemName: "arrow.right")
                    }
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(Color.accentColor)
                    .foregroundStyle(.white)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                }
                .disabled(viewModel.name.isEmpty || viewModel.isLoading)
            }
            .padding()
        }
        .onAppear {
            isNameFocused = true
        }
    }
}

#Preview {
    NameStepView(viewModel: OnboardingViewModel())
}
