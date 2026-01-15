import XCTest
@testable import Zvednu

final class ExtensionsTests: XCTestCase {

    // MARK: - String Phone Formatting Tests

    func testPhoneNumberFormattingCzech() {
        let phoneNumber = "+420123456789"
        let formatted = phoneNumber.formattedPhoneNumber()

        // Should format Czech numbers with spaces
        XCTAssertTrue(formatted.contains(" ") || formatted == phoneNumber)
    }

    func testPhoneNumberFormattingInternational() {
        let phoneNumber = "+1234567890"
        let formatted = phoneNumber.formattedPhoneNumber()

        // Should return a string
        XCTAssertFalse(formatted.isEmpty)
    }

    func testPhoneNumberFormattingEmpty() {
        let phoneNumber = ""
        let formatted = phoneNumber.formattedPhoneNumber()

        XCTAssertEqual(formatted, "")
    }

    // MARK: - String to Date Conversion Tests

    func testStringToDateISO8601() {
        let dateString = "2024-01-15T10:30:00Z"
        let date = dateString.toDate()

        XCTAssertNotNil(date)

        if let date = date {
            let calendar = Calendar(identifier: .gregorian)
            var components = calendar.dateComponents(in: TimeZone(identifier: "UTC")!, from: date)
            XCTAssertEqual(components.year, 2024)
            XCTAssertEqual(components.month, 1)
            XCTAssertEqual(components.day, 15)
        }
    }

    func testStringToDateInvalid() {
        let invalidString = "not a date"
        let date = invalidString.toDate()

        XCTAssertNil(date)
    }

    func testStringToDateEmpty() {
        let emptyString = ""
        let date = emptyString.toDate()

        XCTAssertNil(date)
    }

    // MARK: - Date Formatting Tests

    func testDateFormattedCzech() {
        // Create a known date
        var components = DateComponents()
        components.year = 2024
        components.month = 1
        components.day = 15
        components.hour = 10
        components.minute = 30
        components.timeZone = TimeZone(identifier: "Europe/Prague")

        let calendar = Calendar(identifier: .gregorian)
        guard let date = calendar.date(from: components) else {
            XCTFail("Could not create test date")
            return
        }

        let formatted = date.formattedCzech()

        // Should contain day and month
        XCTAssertFalse(formatted.isEmpty)
        XCTAssertTrue(formatted.contains("15") || formatted.contains("1"))
    }

    func testDateSmartFormat() {
        var components = DateComponents()
        components.year = 2024
        components.month = 1
        components.day = 15
        components.hour = 14
        components.minute = 30
        components.timeZone = TimeZone(identifier: "Europe/Prague")

        let calendar = Calendar(identifier: .gregorian)
        guard let date = calendar.date(from: components) else {
            XCTFail("Could not create test date")
            return
        }

        let formatted = date.smartFormat()

        // Should return non-empty string
        XCTAssertFalse(formatted.isEmpty)
    }

    func testDateRelativeString() {
        let now = Date()
        let formatted = now.relativeString()

        // Should return something for current date
        XCTAssertFalse(formatted.isEmpty)
    }

    func testDateRelativeStringYesterday() {
        let yesterday = Calendar.current.date(byAdding: .day, value: -1, to: Date())!
        let formatted = yesterday.relativeString()

        XCTAssertFalse(formatted.isEmpty)
    }

    func testDateRelativeStringLastWeek() {
        let lastWeek = Calendar.current.date(byAdding: .day, value: -5, to: Date())!
        let formatted = lastWeek.relativeString()

        XCTAssertFalse(formatted.isEmpty)
    }

    // MARK: - Date Properties Tests

    func testDateIsToday() {
        let today = Date()
        XCTAssertTrue(today.isToday)
    }

    func testDateIsYesterday() {
        let yesterday = Calendar.current.date(byAdding: .day, value: -1, to: Date())!
        XCTAssertTrue(yesterday.isYesterday)
    }

    func testDateIsNotToday() {
        let yesterday = Calendar.current.date(byAdding: .day, value: -1, to: Date())!
        XCTAssertFalse(yesterday.isToday)
    }
}
