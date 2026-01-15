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
                EmptyInboxView()
            } else {
                callsList
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

    private var callsList: some View {
        List {
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
                .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 0, trailing: 16))
            }
        }
        .listStyle(.plain)
    }
}

#Preview {
    NavigationStack {
        CallInboxView(selectedCallId: .constant(nil), showCallDetail: .constant(false))
    }
}
