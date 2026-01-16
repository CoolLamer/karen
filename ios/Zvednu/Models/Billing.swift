import Foundation

struct CallStatus: Codable, Equatable {
    let canReceive: Bool
    let reason: String // "ok", "trial_expired", "limit_exceeded"
    let callsUsed: Int
    let callsLimit: Int // -1 = unlimited
    let trialDaysLeft: Int?
    let trialCallsLeft: Int?

    enum CodingKeys: String, CodingKey {
        case canReceive = "can_receive"
        case reason
        case callsUsed = "calls_used"
        case callsLimit = "calls_limit"
        case trialDaysLeft = "trial_days_left"
        case trialCallsLeft = "trial_calls_left"
    }
}

struct CurrentUsage: Codable, Equatable {
    let callsCount: Int
    let minutesUsed: Int
    let timeSavedSeconds: Int
    let spamCallsBlocked: Int
    let periodStart: String
    let periodEnd: String

    enum CodingKeys: String, CodingKey {
        case callsCount = "calls_count"
        case minutesUsed = "minutes_used"
        case timeSavedSeconds = "time_saved_seconds"
        case spamCallsBlocked = "spam_calls_blocked"
        case periodStart = "period_start"
        case periodEnd = "period_end"
    }
}

struct BillingInfo: Codable, Equatable {
    let plan: String
    let status: String
    let callStatus: CallStatus
    let trialEndsAt: String?
    let totalTimeSaved: Int // seconds
    let currentUsage: CurrentUsage?

    enum CodingKeys: String, CodingKey {
        case plan
        case status
        case callStatus = "call_status"
        case trialEndsAt = "trial_ends_at"
        case totalTimeSaved = "total_time_saved"
        case currentUsage = "current_usage"
    }

    // Helper to format time saved as a human-readable string (Czech)
    var formattedTimeSaved: String {
        let seconds = currentUsage?.timeSavedSeconds ?? 0
        if seconds < 60 { return "\(seconds) sekund" }
        let minutes = seconds / 60
        if minutes < 60 { return "\(minutes) minut" }
        let hours = minutes / 60
        let remainingMins = minutes % 60
        if remainingMins == 0 { return "\(hours)h" }
        return "\(hours)h \(remainingMins)min"
    }

    var formattedTotalTimeSaved: String {
        if totalTimeSaved < 60 { return "\(totalTimeSaved) sekund" }
        let minutes = totalTimeSaved / 60
        if minutes < 60 { return "\(minutes) minut" }
        let hours = minutes / 60
        let remainingMins = minutes % 60
        if remainingMins == 0 { return "\(hours)h" }
        return "\(hours)h \(remainingMins)min"
    }

    var isTrial: Bool {
        plan == "trial"
    }

    var isTrialExpired: Bool {
        !callStatus.canReceive && (callStatus.reason == "trial_expired" || callStatus.reason == "limit_exceeded")
    }

    var usagePercentage: Double {
        guard callStatus.callsLimit > 0 else { return 0 }
        return Double(callStatus.callsUsed) / Double(callStatus.callsLimit) * 100
    }
}
