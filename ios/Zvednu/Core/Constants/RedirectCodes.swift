import Foundation

// MARK: - Redirect Types

enum RedirectType: String, CaseIterable, Identifiable {
    case noAnswer
    case busy
    case unreachable

    var id: String { rawValue }

    var label: String {
        switch self {
        case .noAnswer: return "Když nezvedáš"
        case .busy: return "Když máš obsazeno"
        case .unreachable: return "Když jsi nedostupný"
        }
    }

    var shortLabel: String {
        switch self {
        case .noAnswer: return "Nezvedám"
        case .busy: return "Obsazeno"
        case .unreachable: return "Nedostupný"
        }
    }

    var codeTemplate: String {
        switch self {
        case .noAnswer: return "**61*{number}**{time}#"
        case .busy: return "**67*{number}#"
        case .unreachable: return "**62*{number}#"
        }
    }

    var deactivateCode: String {
        switch self {
        case .noAnswer: return "##61#"
        case .busy: return "##67#"
        case .unreachable: return "##62#"
        }
    }

    func description(time: Int = RedirectCodes.defaultNoAnswerTime) -> String {
        switch self {
        case .noAnswer:
            return "Když nezvedneš do \(time) sekund, hovor se přesměruje na Karen."
        case .busy:
            return "Když máš obsazeno nebo odmítneš hovor, přesměruje se na Karen."
        case .unreachable:
            return "Když nemáš signál nebo máš vypnutý telefon, hovor jde na Karen."
        }
    }
}

// MARK: - Constants

enum RedirectCodes {
    /// Code to clear all conditional forwarding at once
    static let clearAllRedirectsCode = "##002#"

    /// Available time options for "no answer" redirect (in seconds)
    static let noAnswerTimeOptions = [5, 10, 15, 20, 25, 30]

    /// Default time for "no answer" redirect
    static let defaultNoAnswerTime = 10

    /// Generate dial code for a specific redirect type
    static func getDialCode(type: RedirectType, phoneNumber: String, time: Int = defaultNoAnswerTime) -> String {
        let cleanNumber = phoneNumber.replacingOccurrences(of: " ", with: "")

        switch type {
        case .noAnswer:
            return "**61*\(cleanNumber)**\(time)#"
        case .busy:
            return "**67*\(cleanNumber)#"
        case .unreachable:
            return "**62*\(cleanNumber)#"
        }
    }
}
