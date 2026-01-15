import Foundation

extension String {
    /// Format phone number for display
    func formattedPhoneNumber() -> String {
        // Remove all non-digits
        let digits = self.filter { $0.isNumber || $0 == "+" }

        // If starts with +420, format as Czech number
        if digits.hasPrefix("+420") && digits.count == 13 {
            let index1 = digits.index(digits.startIndex, offsetBy: 4)
            let index2 = digits.index(digits.startIndex, offsetBy: 7)
            let index3 = digits.index(digits.startIndex, offsetBy: 10)
            return "+420 \(digits[index1..<index2]) \(digits[index2..<index3]) \(digits[index3...])"
        }

        return self
    }

    /// Check if string is a valid Czech phone number
    var isValidCzechPhone: Bool {
        let pattern = #"^(\+420)?\s?\d{3}\s?\d{3}\s?\d{3}$"#
        return self.range(of: pattern, options: .regularExpression) != nil
    }

    /// Normalize phone number (remove spaces, add +420 if needed)
    func normalizedPhoneNumber() -> String {
        var digits = self.filter { $0.isNumber || $0 == "+" }

        // If it starts with just digits (no +), prepend +420
        if !digits.hasPrefix("+") {
            if digits.hasPrefix("420") {
                digits = "+" + digits
            } else {
                digits = "+420" + digits
            }
        }

        return digits
    }

    /// Truncate string with ellipsis
    func truncated(to length: Int) -> String {
        if self.count <= length {
            return self
        }
        return String(self.prefix(length)) + "..."
    }
}
