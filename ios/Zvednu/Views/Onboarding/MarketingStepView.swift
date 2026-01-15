import SwiftUI

struct MarketingStepView: View {
    @ObservedObject var viewModel: OnboardingViewModel

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                VStack(spacing: 8) {
                    Text("Jak nakládat s marketingem?")
                        .font(.title2)
                        .fontWeight(.bold)

                    Text("Karen automaticky rozpozná marketingové a obchodní hovory.")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                }
                .padding(.top, 20)

                // Options
                VStack(spacing: 16) {
                    marketingOption(
                        option: .reject,
                        title: "Zdvořile odmítne a ukončí hovor",
                        description: "Standardní nastavení pro většinu uživatelů"
                    )

                    marketingOption(
                        option: .email,
                        title: "Odmítne, ale nabídne můj email pro písemné nabídky",
                        description: "Užitečné, pokud občas chceš vidět nabídky"
                    )
                }

                // Email input (shown when email option selected)
                if viewModel.marketingOption == .email {
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Email pro marketingové nabídky")
                            .font(.headline)

                        TextField("nabidky@email.cz", text: $viewModel.marketingEmail)
                            .keyboardType(.emailAddress)
                            .textContentType(.emailAddress)
                            .autocapitalization(.none)
                            .padding()
                            .background(Color(.systemBackground))
                            .clipShape(RoundedRectangle(cornerRadius: 12))
                    }
                }

                Spacer(minLength: 20)

                HStack(spacing: 12) {
                    Button {
                        Task {
                            await viewModel.saveConfiguration()
                        }
                    } label: {
                        Text("Přeskočit")
                            .frame(maxWidth: .infinity)
                            .padding()
                            .foregroundStyle(.secondary)
                    }

                    Button {
                        Task {
                            await viewModel.saveConfiguration()
                        }
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
            .padding()
        }
    }

    private func marketingOption(option: OnboardingViewModel.MarketingOption, title: String, description: String) -> some View {
        Button {
            viewModel.marketingOption = option
        } label: {
            HStack(alignment: .top, spacing: 12) {
                Image(systemName: viewModel.marketingOption == option ? "checkmark.circle.fill" : "circle")
                    .font(.title2)
                    .foregroundStyle(viewModel.marketingOption == option ? Color.accentColor : .secondary)

                VStack(alignment: .leading, spacing: 4) {
                    Text(title)
                        .font(.subheadline)
                        .fontWeight(.medium)
                        .foregroundStyle(.primary)
                        .multilineTextAlignment(.leading)

                    Text(description)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.leading)
                }

                Spacer()
            }
            .padding()
            .background(Color(.systemBackground))
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .overlay(
                RoundedRectangle(cornerRadius: 12)
                    .stroke(viewModel.marketingOption == option ? Color.accentColor : Color.clear, lineWidth: 2)
            )
        }
    }
}

#Preview {
    MarketingStepView(viewModel: OnboardingViewModel())
}
