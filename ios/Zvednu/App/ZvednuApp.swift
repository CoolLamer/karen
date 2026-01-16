import SwiftUI

@main
struct ZvednuApp: App {
    @UIApplicationDelegateAdaptor(AppDelegate.self) var appDelegate
    @StateObject private var authViewModel = AuthViewModel()
    private let contactsManager = ContactsManager.shared

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environmentObject(authViewModel)
                .task {
                    await contactsManager.initializeOnAppLaunch()
                }
        }
    }
}
