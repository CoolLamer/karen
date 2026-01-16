import SwiftUI

struct SettingsView: View {
    @EnvironmentObject var authViewModel: AuthViewModel
    @StateObject private var viewModel = SettingsViewModel()
    @State private var showLogoutConfirmation = false
    @State private var newVipName = ""

    var body: some View {
        Group {
            if viewModel.isLoading && viewModel.tenant == nil {
                LoadingView(message: "Nacitam nastaveni...")
            } else {
                settingsContent
            }
        }
        .navigationTitle("Nastaveni")
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
        .alert("Odhlasit se?", isPresented: $showLogoutConfirmation) {
            Button("Zrusit", role: .cancel) {}
            Button("Odhlasit", role: .destructive) {
                Task {
                    await authViewModel.logout()
                }
            }
        } message: {
            Text("Opravdu se chces odhlasit?")
        }
        .sheet(isPresented: $viewModel.showUpgradeSheet) {
            UpgradeSheetView(viewModel: viewModel)
        }
    }

    private var settingsContent: some View {
        Form {
            // Phone Number Section
            if let phoneNumber = viewModel.primaryPhoneNumber {
                Section("Karen cislo") {
                    HStack {
                        Text(phoneNumber.formattedPhoneNumber())
                            .font(.headline)
                        Spacer()
                        Button {
                            UIPasteboard.general.string = phoneNumber.replacingOccurrences(of: " ", with: "")
                        } label: {
                            Image(systemName: "doc.on.doc")
                        }
                    }
                }
            }

            // Name Section
            Section("Jmeno") {
                TextField("Jmeno", text: $viewModel.name)
            }

            // Greeting Section
            Section {
                TextEditor(text: $viewModel.greetingText)
                    .frame(minHeight: 80)
            } header: {
                Text("Pozdrav")
            } footer: {
                Text("Text, kterym Karen zacina hovor.")
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
                    TextField("Pridej VIP jmeno", text: $newVipName)
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
                Text("Kdyz se volajici predstavi jednim z techto jmen, Karen ho okamzite prepoji.")
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
                Text("Pokud je vyplneno, Karen nabidne tento email marketingovym volajicim.")
            }

            // Save Button
            Section {
                Button {
                    Task {
                        await viewModel.saveChanges()
                    }
                } label: {
                    HStack {
                        Spacer()
                        if viewModel.isSaving {
                            ProgressView()
                        } else if viewModel.showSavedConfirmation {
                            Label("Ulozeno", systemImage: "checkmark.circle.fill")
                                .foregroundStyle(.green)
                        } else {
                            Text("Ulozit zmeny")
                        }
                        Spacer()
                    }
                }
                .disabled(viewModel.isSaving)
            }

            // Forwarding Instructions Section
            if let phoneNumber = viewModel.primaryPhoneNumber {
                Section("Presmerovani hovoru") {
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Vytocte tyto kody pro aktivaci presmerovani:")
                            .font(.caption)
                            .foregroundStyle(.secondary)

                        forwardingCodeRow(
                            title: "Kdyz nezvednes",
                            code: "**61*\(phoneNumber.replacingOccurrences(of: " ", with: ""))#"
                        )

                        forwardingCodeRow(
                            title: "Kdyz mas obsazeno",
                            code: "**67*\(phoneNumber.replacingOccurrences(of: " ", with: ""))#"
                        )

                        forwardingCodeRow(
                            title: "Kdyz jsi nedostupny",
                            code: "**62*\(phoneNumber.replacingOccurrences(of: " ", with: ""))#"
                        )
                    }
                }
            }

            // Subscription Section
            Section("Predplatne") {
                // Current Plan
                HStack {
                    Text("Plan")
                    Spacer()
                    Text(planLabel)
                        .foregroundStyle(.secondary)
                    if viewModel.billing?.status == "past_due" {
                        Text("Nezaplaceno")
                            .font(.caption)
                            .padding(.horizontal, 6)
                            .padding(.vertical, 2)
                            .background(Color.red)
                            .foregroundStyle(.white)
                            .clipShape(Capsule())
                    }
                }

                // Trial Info
                if viewModel.isTrial, let callStatus = viewModel.billing?.callStatus {
                    VStack(alignment: .leading, spacing: 4) {
                        HStack {
                            if let trialCallsLeft = callStatus.trialCallsLeft {
                                Text("Zbyva \(trialCallsLeft) hovoru")
                            }
                            if let trialDaysLeft = callStatus.trialDaysLeft {
                                Text("• \(trialDaysLeft) dni")
                            }
                        }
                        .font(.caption)
                        .foregroundStyle(.secondary)

                        ProgressView(value: min(viewModel.usagePercentage, 100), total: 100)
                            .tint(progressColor)
                    }
                }

                // Time Saved
                if let billing = viewModel.billing, billing.totalTimeSaved > 0 {
                    HStack {
                        Image(systemName: "clock.fill")
                            .foregroundStyle(.teal)
                        VStack(alignment: .leading) {
                            Text("Karen ti usetrila")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                            Text(viewModel.formattedTimeSaved)
                                .font(.headline)
                                .foregroundStyle(.teal)
                        }
                        Spacer()
                    }
                    .padding(.vertical, 4)
                }

                // Upgrade/Manage Button
                if viewModel.isTrial {
                    Button {
                        viewModel.showUpgradeSheet = true
                    } label: {
                        HStack {
                            Spacer()
                            if viewModel.isUpgrading {
                                ProgressView()
                            } else {
                                Text("Upgradovat")
                            }
                            Spacer()
                        }
                    }
                } else {
                    Button {
                        Task {
                            await viewModel.openManageSubscription()
                        }
                    } label: {
                        HStack {
                            Spacer()
                            if viewModel.isUpgrading {
                                ProgressView()
                            } else {
                                HStack {
                                    Text("Spravovat predplatne")
                                    Image(systemName: "arrow.up.forward")
                                        .font(.caption)
                                }
                            }
                            Spacer()
                        }
                    }
                }
            }

            // Trial Expired Warning
            if viewModel.isTrialExpired {
                Section {
                    VStack(alignment: .leading, spacing: 8) {
                        HStack {
                            Image(systemName: "exclamationmark.triangle.fill")
                                .foregroundStyle(.red)
                            Text("Trial vyprsel")
                                .font(.headline)
                                .foregroundStyle(.red)
                        }
                        Text(viewModel.billing?.callStatus.reason == "limit_exceeded"
                             ? "Dosahli jste limitu hovoru. Karen nebude prijimat nove hovory."
                             : "Vas trial skoncil. Karen nebude prijimat nove hovory.")
                            .font(.caption)
                            .foregroundStyle(.secondary)

                        Button {
                            viewModel.showUpgradeSheet = true
                        } label: {
                            HStack {
                                Spacer()
                                Text("Upgradovat nyni")
                                Spacer()
                            }
                        }
                        .buttonStyle(.borderedProminent)
                    }
                    .padding(.vertical, 4)
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
                    Text("Ochrana osobnich udaju")
                }

                Link(destination: URL(string: "https://zvednu.cz/obchodni-podminky")!) {
                    Text("Obchodni podminky")
                }
            }

            // Logout Section
            Section {
                Button(role: .destructive) {
                    showLogoutConfirmation = true
                } label: {
                    HStack {
                        Spacer()
                        Text("Odhlasit se")
                        Spacer()
                    }
                }
            }
        }
    }

    private var planLabel: String {
        switch viewModel.billing?.plan ?? "trial" {
        case "basic": return "Zaklad"
        case "pro": return "Pro"
        default: return "Trial"
        }
    }

    private var progressColor: Color {
        if viewModel.usagePercentage >= 100 {
            return .red
        } else if viewModel.usagePercentage >= 80 {
            return .yellow
        }
        return .blue
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
                    Text("Vytocit")
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

// MARK: - Upgrade Sheet View

struct UpgradeSheetView: View {
    @ObservedObject var viewModel: SettingsViewModel
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 20) {
                    // Basic Plan Card
                    PlanCardView(
                        name: "Zaklad",
                        description: "Pro OSVC a male firmy",
                        monthlyPrice: "199 Kc",
                        yearlyPrice: "159 Kc",
                        features: [
                            "50 hovoru mesicne",
                            "Kompletni prepisy",
                            "SMS notifikace"
                        ],
                        accentColor: .blue,
                        isPopular: false,
                        isUpgrading: viewModel.isUpgrading,
                        onMonthly: {
                            Task {
                                await viewModel.openUpgrade(plan: "basic", interval: "monthly")
                                dismiss()
                            }
                        },
                        onYearly: {
                            Task {
                                await viewModel.openUpgrade(plan: "basic", interval: "annual")
                                dismiss()
                            }
                        }
                    )

                    // Pro Plan Card
                    PlanCardView(
                        name: "Pro",
                        description: "Pro profesionaly",
                        monthlyPrice: "499 Kc",
                        yearlyPrice: "399 Kc",
                        features: [
                            "Neomezene hovory",
                            "VIP prepojovani",
                            "Vlastni hlas",
                            "Prioritni podpora"
                        ],
                        accentColor: .purple,
                        isPopular: true,
                        isUpgrading: viewModel.isUpgrading,
                        onMonthly: {
                            Task {
                                await viewModel.openUpgrade(plan: "pro", interval: "monthly")
                                dismiss()
                            }
                        },
                        onYearly: {
                            Task {
                                await viewModel.openUpgrade(plan: "pro", interval: "annual")
                                dismiss()
                            }
                        }
                    )

                    Text("Platbu zpracovava Stripe. Predplatne muzete kdykoli zrusit.")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                        .padding(.top)
                }
                .padding()
            }
            .navigationTitle("Vyberte plan")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("Zavrit") {
                        dismiss()
                    }
                }
            }
        }
    }
}

struct PlanCardView: View {
    let name: String
    let description: String
    let monthlyPrice: String
    let yearlyPrice: String
    let features: [String]
    let accentColor: Color
    let isPopular: Bool
    let isUpgrading: Bool
    let onMonthly: () -> Void
    let onYearly: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            // Header
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    Text(name)
                        .font(.title2)
                        .fontWeight(.bold)
                    Text(description)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
                Spacer()
                if isPopular {
                    Text("Popularni")
                        .font(.caption2)
                        .fontWeight(.semibold)
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(accentColor)
                        .foregroundStyle(.white)
                        .clipShape(Capsule())
                }
            }

            // Pricing
            VStack(alignment: .leading, spacing: 2) {
                HStack(alignment: .lastTextBaseline, spacing: 4) {
                    Text(monthlyPrice)
                        .font(.title)
                        .fontWeight(.bold)
                    Text("/mesic")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
                Text("nebo \(yearlyPrice)/mesic rocne")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            // Features
            VStack(alignment: .leading, spacing: 6) {
                ForEach(features, id: \.self) { feature in
                    HStack(spacing: 8) {
                        Image(systemName: "checkmark")
                            .foregroundStyle(accentColor)
                            .font(.caption)
                        Text(feature)
                            .font(.subheadline)
                    }
                }
            }

            // Buttons
            VStack(spacing: 8) {
                Button {
                    onMonthly()
                } label: {
                    HStack {
                        Spacer()
                        if isUpgrading {
                            ProgressView()
                                .tint(.white)
                        } else {
                            Text("Mesicni platba")
                        }
                        Spacer()
                    }
                    .padding(.vertical, 12)
                }
                .buttonStyle(.borderedProminent)
                .tint(accentColor)
                .disabled(isUpgrading)

                Button {
                    onYearly()
                } label: {
                    HStack {
                        Spacer()
                        Text("Rocni platba (usetri 20%)")
                        Spacer()
                    }
                    .padding(.vertical, 12)
                }
                .buttonStyle(.bordered)
                .tint(accentColor)
                .disabled(isUpgrading)
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 16))
        .overlay(
            RoundedRectangle(cornerRadius: 16)
                .stroke(isPopular ? accentColor : Color(.separator), lineWidth: isPopular ? 2 : 1)
        )
    }
}

#Preview {
    NavigationStack {
        SettingsView()
    }
    .environmentObject(AuthViewModel())
}
