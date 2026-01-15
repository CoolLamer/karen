import SwiftUI

struct LoginView: View {
    @EnvironmentObject var authViewModel: AuthViewModel
    @State private var showError = false

    var body: some View {
        NavigationStack {
            VStack(spacing: 24) {
                Spacer()

                // Logo/Header
                VStack(spacing: 12) {
                    Image(systemName: "phone.badge.checkmark.fill")
                        .font(.system(size: 60))
                        .foregroundStyle(Color.accentColor)

                    Text("Zvednu")
                        .font(.largeTitle)
                        .fontWeight(.bold)

                    Text("AI telefonní asistentka")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }

                Spacer()

                // Login Form
                VStack(spacing: 16) {
                    if !authViewModel.isCodeSent {
                        phoneInputSection
                    } else {
                        codeInputSection
                    }
                }
                .padding(.horizontal)

                Spacer()
            }
            .padding()
            .alert("Chyba", isPresented: $showError) {
                Button("OK") {
                    authViewModel.error = nil
                }
            } message: {
                if let error = authViewModel.error {
                    Text(error)
                }
            }
            .onChange(of: authViewModel.error) { _, newValue in
                showError = newValue != nil
            }
        }
    }

    // MARK: - Phone Input Section

    private var phoneInputSection: some View {
        VStack(spacing: 16) {
            Text("Přihlaste se telefonním číslem")
                .font(.headline)

            TextField("Telefonní číslo", text: $authViewModel.phoneNumber)
                .keyboardType(.phonePad)
                .textContentType(.telephoneNumber)
                .padding()
                .background(Color(.systemGray6))
                .clipShape(RoundedRectangle(cornerRadius: 12))

            Button {
                Task {
                    await authViewModel.sendCode()
                }
            } label: {
                HStack {
                    if authViewModel.isSendingCode {
                        ProgressView()
                            .tint(.white)
                    }
                    Text("Odeslat kód")
                }
                .frame(maxWidth: .infinity)
                .padding()
                .background(Color.accentColor)
                .foregroundStyle(.white)
                .clipShape(RoundedRectangle(cornerRadius: 12))
            }
            .disabled(authViewModel.phoneNumber.isEmpty || authViewModel.isSendingCode)
        }
    }

    // MARK: - Code Input Section

    private var codeInputSection: some View {
        VStack(spacing: 16) {
            Text("Zadejte ověřovací kód")
                .font(.headline)

            Text("Kód byl odeslán na \(authViewModel.phoneNumber)")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            TextField("Ověřovací kód", text: $authViewModel.verificationCode)
                .keyboardType(.numberPad)
                .textContentType(.oneTimeCode)
                .multilineTextAlignment(.center)
                .font(.title2.monospacedDigit())
                .padding()
                .background(Color(.systemGray6))
                .clipShape(RoundedRectangle(cornerRadius: 12))

            Button {
                Task {
                    await authViewModel.verifyCode()
                }
            } label: {
                HStack {
                    if authViewModel.isVerifyingCode {
                        ProgressView()
                            .tint(.white)
                    }
                    Text("Ověřit")
                }
                .frame(maxWidth: .infinity)
                .padding()
                .background(Color.accentColor)
                .foregroundStyle(.white)
                .clipShape(RoundedRectangle(cornerRadius: 12))
            }
            .disabled(authViewModel.verificationCode.isEmpty || authViewModel.isVerifyingCode)

            Button("Změnit číslo") {
                authViewModel.resetLoginState()
            }
            .font(.subheadline)
            .foregroundStyle(.secondary)
        }
    }
}

#Preview {
    LoginView()
        .environmentObject(AuthViewModel())
}
