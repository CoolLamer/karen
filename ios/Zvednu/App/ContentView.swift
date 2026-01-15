import SwiftUI

struct ContentView: View {
    @EnvironmentObject var authViewModel: AuthViewModel
    @State private var selectedCallId: String?
    @State private var showCallDetail = false

    var body: some View {
        Group {
            if authViewModel.isLoading {
                LoadingView(message: "Načítám...")
            } else if !authViewModel.isAuthenticated {
                LoginView()
            } else if authViewModel.needsOnboarding {
                OnboardingContainerView()
            } else {
                mainTabView
            }
        }
        .onReceive(NotificationCenter.default.publisher(for: .openCallDetail)) { notification in
            if let callId = notification.userInfo?["callId"] as? String {
                selectedCallId = callId
                showCallDetail = true
            }
        }
    }

    private var mainTabView: some View {
        TabView {
            NavigationStack {
                CallInboxView(selectedCallId: $selectedCallId, showCallDetail: $showCallDetail)
                    .navigationDestination(isPresented: $showCallDetail) {
                        if let callId = selectedCallId {
                            CallDetailView(providerCallId: callId)
                        }
                    }
            }
            .tabItem {
                Label("Hovory", systemImage: "phone.fill")
            }

            NavigationStack {
                SettingsView()
            }
            .tabItem {
                Label("Nastavení", systemImage: "gearshape.fill")
            }
        }
        .tint(.accentColor)
    }
}

#Preview {
    ContentView()
        .environmentObject(AuthViewModel())
}
