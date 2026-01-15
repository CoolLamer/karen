import SwiftUI

struct OnboardingContainerView: View {
    @EnvironmentObject var authViewModel: AuthViewModel
    @StateObject private var viewModel = OnboardingViewModel()

    var body: some View {
        VStack(spacing: 0) {
            // Progress indicator (not shown on welcome/complete)
            if viewModel.currentStep != .welcome && viewModel.currentStep != .complete {
                progressHeader
            }

            // Step content
            Group {
                switch viewModel.currentStep {
                case .welcome:
                    WelcomeStepView(viewModel: viewModel)
                case .name:
                    NameStepView(viewModel: viewModel)
                case .vipContacts:
                    VIPContactsStepView(viewModel: viewModel)
                case .marketing:
                    MarketingStepView(viewModel: viewModel)
                case .phoneSetup:
                    PhoneSetupStepView(viewModel: viewModel)
                case .complete:
                    CompleteStepView(viewModel: viewModel)
                }
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
        .background(Color(.systemGroupedBackground))
        .onAppear {
            viewModel.authViewModel = authViewModel
        }
    }

    private var progressHeader: some View {
        VStack(spacing: 8) {
            HStack {
                Text("Krok \(viewModel.currentStep.rawValue) z 5")
                    .font(.caption)
                    .foregroundStyle(.secondary)
                Spacer()
                Text("\(Int(viewModel.progress * 100))%")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            ProgressView(value: viewModel.progress)
                .tint(.accentColor)
        }
        .padding()
        .background(Color(.systemBackground))
    }
}

// MARK: - Welcome Step

struct WelcomeStepView: View {
    @ObservedObject var viewModel: OnboardingViewModel

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                Spacer(minLength: 40)

                Image(systemName: "phone.badge.checkmark.fill")
                    .font(.system(size: 80))
                    .foregroundStyle(Color.accentColor)

                VStack(spacing: 8) {
                    Text("Vitej! Jsem Karen")
                        .font(.title)
                        .fontWeight(.bold)

                    Text("Jsem tvoje AI telefonni asistentka. Kdyz nezvednes telefon, prebiram hovory za tebe.")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                }

                // Features list
                VStack(alignment: .leading, spacing: 12) {
                    Text("Co pro tebe udelam:")
                        .font(.headline)

                    featureRow(icon: "phone.arrow.down.left", text: "Zvednu hovory, kdyz budes zaneprazdneny")
                    featureRow(icon: "person.crop.circle.badge.questionmark", text: "Zjistim, kdo vola a co potrebuje")
                    featureRow(icon: "xmark.circle", text: "Odmitnu marketing a spam")
                    featureRow(icon: "arrow.right.arrow.left", text: "Okamzite prepojim dulezite kontakty")
                    featureRow(icon: "list.bullet.clipboard", text: "Poslu ti prehled hovoru v aplikaci")
                }
                .padding()
                .background(Color(.systemBackground))
                .clipShape(RoundedRectangle(cornerRadius: 12))

                Text("Nastaveni zabere asi 3 minuty.")
                    .font(.caption)
                    .foregroundStyle(.secondary)

                Button {
                    viewModel.goToNext()
                } label: {
                    HStack {
                        Text("Pojdme na to")
                        Image(systemName: "arrow.right")
                    }
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(Color.accentColor)
                    .foregroundStyle(.white)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                }

                Spacer(minLength: 20)
            }
            .padding()
        }
    }

    private func featureRow(icon: String, text: String) -> some View {
        HStack(spacing: 12) {
            Image(systemName: icon)
                .foregroundStyle(Color.accentColor)
                .frame(width: 24)
            Text(text)
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
    }
}

#Preview {
    OnboardingContainerView()
        .environmentObject(AuthViewModel())
}
