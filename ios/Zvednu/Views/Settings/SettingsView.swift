import SwiftUI

struct SettingsView: View {
    @EnvironmentObject var authViewModel: AuthViewModel
    @StateObject private var viewModel = SettingsViewModel()
    @State private var showLogoutConfirmation = false
    @State private var newVipName = ""

    var body: some View {
        Group {
            if viewModel.isLoading && viewModel.tenant == nil {
                LoadingView(message: "Načítám nastavení...")
            } else {
                settingsContent
            }
        }
        .navigationTitle("Nastavení")
        .toolbar {
            ToolbarItem(placement: .navigationBarTrailing) {
                Button {
                    Task {
                        await viewModel.saveChanges()
                    }
                } label: {
                    if viewModel.isSaving {
                        ProgressView()
                    } else if viewModel.showSavedConfirmation {
                        Image(systemName: "checkmark.circle.fill")
                            .foregroundStyle(.green)
                    } else {
                        Text("Uložit")
                    }
                }
                .disabled(viewModel.isSaving)
            }
        }
        .task {
            viewModel.authViewModel = authViewModel
            await viewModel.loadTenant()
        }
        .alert("Chyba", isPresented: .constant(viewModel.error != nil)) {
            Button("OK") {
                viewModel.error = nil
            }
        } message: {
            if let error = viewModel.error {
                Text(error)
            }
        }
        .alert("Odhlásit se?", isPresented: $showLogoutConfirmation) {
            Button("Zrušit", role: .cancel) {}
            Button("Odhlásit", role: .destructive) {
                Task {
                    await authViewModel.logout()
                }
            }
        } message: {
            Text("Opravdu se chceš odhlásit?")
        }
    }

    private var settingsContent: some View {
        Form {
            // Phone Number Section
            if let phoneNumber = viewModel.primaryPhoneNumber {
                Section("Karen číslo") {
                    HStack {
                        Text(phoneNumber.formattedPhoneNumber())
                            .font(.headline)
                        Spacer()
                        Button {
                            UIPasteboard.general.string = phoneNumber.replacingOccurrences(of: " ", with: "")
                        } label: {
                            Image(systemName: "doc.on.doc")
                        }
                        if let url = URL(string: "tel:\(phoneNumber.replacingOccurrences(of: " ", with: ""))") {
                            Link(destination: url) {
                                Image(systemName: "phone.fill")
                            }
                        }
                    }
                }
            }

            // Name Section
            Section("Jméno") {
                TextField("Jméno", text: $viewModel.name)
            }

            // Greeting Section
            Section {
                TextEditor(text: $viewModel.greetingText)
                    .frame(minHeight: 80)
            } header: {
                Text("Pozdrav")
            } footer: {
                Text("Text, kterým Karen začíná hovor.")
            }

            // VIP Names Section
            Section {
                ForEach(Array(viewModel.vipNames.enumerated()), id: \.offset) { index, name in
                    HStack {
                        Text(name)
                        Spacer()
                        Button {
                            viewModel.removeVipName(at: index)
                        } label: {
                            Image(systemName: "xmark.circle.fill")
                                .foregroundStyle(.secondary)
                        }
                    }
                }

                HStack {
                    TextField("Přidej VIP jméno", text: $newVipName)
                        .onSubmit {
                            viewModel.addVipName(newVipName)
                            newVipName = ""
                        }

                    Button {
                        viewModel.addVipName(newVipName)
                        newVipName = ""
                    } label: {
                        Image(systemName: "plus.circle.fill")
                    }
                    .disabled(newVipName.isEmpty)
                }
            } header: {
                Text("VIP kontakty")
            } footer: {
                Text("Když se volající představí jedním z těchto jmen, Karen ho okamžitě přepojí.")
            }

            // Marketing Email Section
            Section {
                TextField("Email pro marketing", text: $viewModel.marketingEmail)
                    .keyboardType(.emailAddress)
                    .textContentType(.emailAddress)
                    .textInputAutocapitalization(.never)
            } header: {
                Text("Marketing")
            } footer: {
                Text("Pokud je vyplněno, Karen nabídne tento email marketingovým volajícím.")
            }

            // Forwarding Instructions Section
            if let phoneNumber = viewModel.primaryPhoneNumber {
                Section("Přesměrování hovoru") {
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Vytočte tyto kódy pro aktivaci přesměrování:")
                            .font(.caption)
                            .foregroundStyle(.secondary)

                        forwardingCodeRow(
                            title: "Když nezvedneš",
                            code: "**61*\(phoneNumber.replacingOccurrences(of: " ", with: ""))#"
                        )

                        forwardingCodeRow(
                            title: "Když máš obsazeno",
                            code: "**67*\(phoneNumber.replacingOccurrences(of: " ", with: ""))#"
                        )

                        forwardingCodeRow(
                            title: "Když jsi nedostupný",
                            code: "**62*\(phoneNumber.replacingOccurrences(of: " ", with: ""))#"
                        )
                    }
                }
            }

            // About Section
            Section("O aplikaci") {
                HStack {
                    Text("Verze")
                    Spacer()
                    Text("\(AppConfig.appVersion) (\(AppConfig.buildNumber))")
                        .foregroundStyle(.secondary)
                }

                Link(destination: URL(string: "https://zvednu.cz/ochrana-osobnich-udaju")!) {
                    Text("Ochrana osobních údajů")
                }

                Link(destination: URL(string: "https://zvednu.cz/obchodni-podminky")!) {
                    Text("Obchodní podmínky")
                }
            }

            // Logout Section
            Section {
                Button(role: .destructive) {
                    showLogoutConfirmation = true
                } label: {
                    HStack {
                        Spacer()
                        Text("Odhlásit se")
                        Spacer()
                    }
                }
            }
        }
    }

    private func forwardingCodeRow(title: String, code: String) -> some View {
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .font(.caption)
                    .fontWeight(.medium)
                Text(code)
                    .font(.caption2.monospaced())
                    .foregroundStyle(Color.accentColor)
            }

            Spacer()

            if let url = URL(string: "tel:\(code)") {
                Link(destination: url) {
                    Text("Vytočit")
                        .font(.caption2)
                        .padding(.horizontal, 10)
                        .padding(.vertical, 6)
                        .background(Color.accentColor)
                        .foregroundStyle(.white)
                        .clipShape(Capsule())
                }
            }
        }
        .padding(.vertical, 4)
    }
}

#Preview {
    NavigationStack {
        SettingsView()
    }
    .environmentObject(AuthViewModel())
}
