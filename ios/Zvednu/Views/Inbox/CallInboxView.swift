import SwiftUI

struct CallInboxView: View {
    @StateObject private var viewModel = CallInboxViewModel()
    @Binding var selectedCallId: String?
    @Binding var showCallDetail: Bool

    var body: some View {
        Group {
            if viewModel.isLoading && viewModel.calls.isEmpty {
                LoadingView(message: "Nacitam hovory...")
            } else if viewModel.calls.isEmpty {
                VStack(spacing: 16) {
                    if let billing = viewModel.billing {
                        BillingStatusView(billing: billing)
                            .padding(.horizontal)
                    }
                    EmptyInboxView()
                }
            } else {
                callsListWithBilling
            }
        }
        .navigationTitle("Hovory")
        .toolbar {
            ToolbarItem(placement: .navigationBarTrailing) {
                if viewModel.unresolvedCount > 0 {
                    Text("\(viewModel.unresolvedCount) nevyresenych")
                        .font(.caption)
                        .foregroundStyle(.white)
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(Color.accentColor)
                        .clipShape(Capsule())
                }
            }
        }
        .refreshable {
            await viewModel.refreshCalls()
        }
        .task {
            await viewModel.loadCalls()
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
    }

    private var callsListWithBilling: some View {
        List {
            // Billing widget at the top
            if let billing = viewModel.billing {
                Section {
                    BillingStatusView(billing: billing)
                }
                .listRowInsets(EdgeInsets())
                .listRowBackground(Color.clear)
            }

            // Calls list
            Section {
                ForEach(viewModel.calls) { call in
                    Button {
                        selectedCallId = call.providerCallId
                        showCallDetail = true
                    } label: {
                        CallRowView(call: call)
                    }
                    .buttonStyle(.plain)
                    .swipeActions(edge: .trailing) {
                        if !call.isResolved {
                            Button {
                                Task {
                                    await viewModel.markAsResolved(call)
                                }
                            } label: {
                                Label("Vyreseno", systemImage: "checkmark.circle")
                            }
                            .tint(.green)
                        }
                    }
                }
            }
        }
        .listStyle(.plain)
    }
}

// MARK: - Billing Status View

struct BillingStatusView: View {
    let billing: BillingInfo

    var body: some View {
        VStack(spacing: 12) {
            // Time Saved Card
            HStack(spacing: 12) {
                Image(systemName: "clock.fill")
                    .font(.title2)
                    .foregroundStyle(.teal)
                    .frame(width: 44, height: 44)
                    .background(Color.teal.opacity(0.15))
                    .clipShape(Circle())

                VStack(alignment: .leading, spacing: 2) {
                    Text("Karen ti usetřila")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                    Text(billing.formattedTimeSaved)
                        .font(.title2)
                        .fontWeight(.bold)
                        .foregroundStyle(.teal)
                    Text("tento mesic (\(billing.currentUsage?.callsCount ?? 0) hovoru)")
                        .font(.caption2)
                        .foregroundStyle(.secondary)
                }

                Spacer()
            }
            .padding()
            .background(Color(.systemBackground))
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .shadow(color: .black.opacity(0.05), radius: 2, y: 1)

            // Trial Status Card (only for trial plan)
            if billing.isTrial {
                HStack(spacing: 12) {
                    Image(systemName: billing.callStatus.canReceive ? "phone.fill" : "phone.down.fill")
                        .font(.title2)
                        .foregroundStyle(billing.callStatus.canReceive ? .blue : .red)
                        .frame(width: 44, height: 44)
                        .background((billing.callStatus.canReceive ? Color.blue : Color.red).opacity(0.15))
                        .clipShape(Circle())

                    VStack(alignment: .leading, spacing: 4) {
                        Text("Trial status")
                            .font(.caption)
                            .foregroundStyle(.secondary)

                        if billing.callStatus.canReceive {
                            Text("\(billing.callStatus.trialCallsLeft ?? 0) hovoru zbyva")
                                .font(.headline)
                            Text("\(billing.callStatus.trialDaysLeft ?? 0) dni do konce trialu")
                                .font(.caption2)
                                .foregroundStyle(.secondary)
                        } else {
                            Text("Trial vypršel")
                                .font(.headline)
                                .foregroundStyle(.red)
                            Text(billing.callStatus.reason == "limit_exceeded"
                                 ? "Dosahli jste limitu hovoru"
                                 : "Trial skoncil")
                                .font(.caption2)
                                .foregroundStyle(.secondary)
                        }

                        // Progress bar
                        if billing.callStatus.callsLimit > 0 {
                            ProgressView(value: billing.usagePercentage, total: 100)
                                .tint(progressColor)
                        }
                    }

                    Spacer()
                }
                .padding()
                .background(Color(.systemBackground))
                .clipShape(RoundedRectangle(cornerRadius: 12))
                .shadow(color: .black.opacity(0.05), radius: 2, y: 1)
            }

            // Trial expired alert
            if billing.isTrialExpired {
                HStack {
                    Image(systemName: "exclamationmark.triangle.fill")
                        .foregroundStyle(.red)
                    Text(billing.callStatus.reason == "trial_expired"
                         ? "Tvuj trial vypršel. Karen nebude prijimat hovory."
                         : "Dosahli jste limitu hovoru. Karen nebude prijimat hovory.")
                        .font(.subheadline)
                    Spacer()
                }
                .padding()
                .background(Color.red.opacity(0.1))
                .clipShape(RoundedRectangle(cornerRadius: 12))
            }
        }
        .padding(.horizontal)
        .padding(.vertical, 8)
    }

    private var progressColor: Color {
        if billing.usagePercentage >= 100 {
            return .red
        } else if billing.usagePercentage >= 80 {
            return .yellow
        }
        return .blue
    }
}

#Preview {
    NavigationStack {
        CallInboxView(selectedCallId: .constant(nil), showCallDetail: .constant(false))
    }
}
