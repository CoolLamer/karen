import Foundation

extension Date {
    /// Format date for display in Czech locale
    func formattedCzech() -> String {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "cs_CZ")
        formatter.dateStyle = .medium
        formatter.timeStyle = .short
        return formatter.string(from: self)
    }

    /// Relative time string (e.g., "pred 5 minutami")
    func relativeString() -> String {
        let formatter = RelativeDateTimeFormatter()
        formatter.locale = Locale(identifier: "cs_CZ")
        formatter.unitsStyle = .full
        return formatter.localizedString(for: self, relativeTo: Date())
    }

    /// Check if date is today
    var isToday: Bool {
        Calendar.current.isDateInToday(self)
    }

    /// Check if date is yesterday
    var isYesterday: Bool {
        Calendar.current.isDateInYesterday(self)
    }

    /// Smart format: shows time for today, "vcera" for yesterday, date otherwise
    func smartFormat() -> String {
        if isToday {
            let formatter = DateFormatter()
            formatter.locale = Locale(identifier: "cs_CZ")
            formatter.timeStyle = .short
            return formatter.string(from: self)
        } else if isYesterday {
            let formatter = DateFormatter()
            formatter.locale = Locale(identifier: "cs_CZ")
            formatter.timeStyle = .short
            return "Vcera \(formatter.string(from: self))"
        } else {
            return formattedCzech()
        }
    }
}

extension ISO8601DateFormatter {
    nonisolated(unsafe) static let apiFormatter: ISO8601DateFormatter = {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return formatter
    }()

    nonisolated(unsafe) static let apiFormatterWithoutFractional: ISO8601DateFormatter = {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime]
        return formatter
    }()
}

extension String {
    /// Parse ISO8601 date string from API
    func toDate() -> Date? {
        ISO8601DateFormatter.apiFormatter.date(from: self)
            ?? ISO8601DateFormatter.apiFormatterWithoutFractional.date(from: self)
    }
}
