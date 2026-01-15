import Foundation

enum AppConfig {
    // API Base URL - change for production
    #if DEBUG
    // For local testing, use your Mac's local IP (find with `ifconfig | grep "inet "`)
    // static let apiBaseURL = "http://192.168.x.x:8080"
    static let apiBaseURL = "https://api.zvednu.cz"
    #else
    static let apiBaseURL = "https://api.zvednu.cz"
    #endif

    // Keychain service identifier
    static let keychainService = "cz.zvednu.app"

    // App version info
    static var appVersion: String {
        Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "1.0.0"
    }

    static var buildNumber: String {
        Bundle.main.infoDictionary?["CFBundleVersion"] as? String ?? "1"
    }
}
