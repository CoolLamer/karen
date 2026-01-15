import XCTest
@testable import Zvednu

final class ModelDecodingTests: XCTestCase {

    // MARK: - CallListItem Tests

    func testCallListItemDecoding() throws {
        let json = """
        {
            "provider": "twilio",
            "provider_call_id": "CA123456",
            "from_number": "+420123456789",
            "to_number": "+420987654321",
            "status": "completed",
            "started_at": "2024-01-15T10:30:00Z",
            "ended_at": "2024-01-15T10:35:00Z",
            "ended_by": "caller",
            "first_viewed_at": null,
            "resolved_at": null,
            "resolved_by": null,
            "screening": null
        }
        """.data(using: .utf8)!

        let decoder = JSONDecoder()
        let call = try decoder.decode(CallListItem.self, from: json)

        XCTAssertEqual(call.provider, "twilio")
        XCTAssertEqual(call.providerCallId, "CA123456")
        XCTAssertEqual(call.fromNumber, "+420123456789")
        XCTAssertEqual(call.toNumber, "+420987654321")
        XCTAssertEqual(call.status, "completed")
        XCTAssertEqual(call.startedAt, "2024-01-15T10:30:00Z")
        XCTAssertEqual(call.endedAt, "2024-01-15T10:35:00Z")
        XCTAssertEqual(call.endedBy, "caller")
        XCTAssertNil(call.firstViewedAt)
        XCTAssertNil(call.resolvedAt)
        XCTAssertFalse(call.isResolved)
        XCTAssertFalse(call.isViewed)
    }

    func testCallListItemWithScreening() throws {
        let json = """
        {
            "provider": "twilio",
            "provider_call_id": "CA789",
            "from_number": "+420111222333",
            "to_number": "+420444555666",
            "status": "completed",
            "started_at": "2024-01-15T10:30:00Z",
            "ended_at": null,
            "ended_by": null,
            "first_viewed_at": "2024-01-15T11:00:00Z",
            "resolved_at": "2024-01-15T12:00:00Z",
            "resolved_by": "user",
            "screening": {
                "legitimacy_label": "legitimate",
                "legitimacy_confidence": 0.95,
                "lead_label": "hot",
                "intent_category": "inquiry",
                "intent_text": "Asking about service availability",
                "entities_json": {},
                "created_at": "2024-01-15T10:35:00Z"
            }
        }
        """.data(using: .utf8)!

        let decoder = JSONDecoder()
        let call = try decoder.decode(CallListItem.self, from: json)

        XCTAssertTrue(call.isResolved)
        XCTAssertTrue(call.isViewed)
        XCTAssertNotNil(call.screening)
        XCTAssertEqual(call.screening?.legitimacyLabel, "legitimate")
        XCTAssertEqual(call.screening?.legitimacyConfidence, 0.95)
        XCTAssertEqual(call.screening?.leadLabel, "hot")
    }

    // MARK: - CallDetail Tests

    func testCallDetailDecoding() throws {
        let json = """
        {
            "provider": "twilio",
            "provider_call_id": "CA456",
            "from_number": "+420123456789",
            "to_number": "+420987654321",
            "status": "completed",
            "started_at": "2024-01-15T10:30:00Z",
            "ended_at": "2024-01-15T10:35:00Z",
            "ended_by": "agent",
            "first_viewed_at": null,
            "resolved_at": null,
            "resolved_by": null,
            "screening": null,
            "utterances": [
                {
                    "speaker": "agent",
                    "text": "Hello, how can I help you?",
                    "sequence": 1,
                    "started_at": "2024-01-15T10:30:05Z",
                    "ended_at": "2024-01-15T10:30:10Z",
                    "stt_confidence": 0.98,
                    "interrupted": false
                },
                {
                    "speaker": "caller",
                    "text": "I need information about your services.",
                    "sequence": 2,
                    "started_at": "2024-01-15T10:30:12Z",
                    "ended_at": "2024-01-15T10:30:18Z",
                    "stt_confidence": 0.92,
                    "interrupted": false
                }
            ]
        }
        """.data(using: .utf8)!

        let decoder = JSONDecoder()
        let call = try decoder.decode(CallDetail.self, from: json)

        XCTAssertEqual(call.providerCallId, "CA456")
        XCTAssertEqual(call.utterances.count, 2)
        XCTAssertEqual(call.utterances[0].speaker, "agent")
        XCTAssertEqual(call.utterances[0].text, "Hello, how can I help you?")
        XCTAssertTrue(call.utterances[0].isAgent)
        XCTAssertEqual(call.utterances[1].speaker, "caller")
        XCTAssertFalse(call.utterances[1].isAgent)
    }

    // MARK: - User Tests

    func testUserDecoding() throws {
        let json = """
        {
            "id": "user-123",
            "phone": "+420123456789",
            "name": "Jan Novak",
            "tenant_id": "tenant-456"
        }
        """.data(using: .utf8)!

        let decoder = JSONDecoder()
        let user = try decoder.decode(User.self, from: json)

        XCTAssertEqual(user.id, "user-123")
        XCTAssertEqual(user.phone, "+420123456789")
        XCTAssertEqual(user.name, "Jan Novak")
        XCTAssertEqual(user.tenantId, "tenant-456")
    }

    func testUserDecodingWithOptionalFields() throws {
        let json = """
        {
            "id": "user-789",
            "phone": "+420999888777",
            "name": null,
            "tenant_id": null
        }
        """.data(using: .utf8)!

        let decoder = JSONDecoder()
        let user = try decoder.decode(User.self, from: json)

        XCTAssertEqual(user.id, "user-789")
        XCTAssertNil(user.name)
        XCTAssertNil(user.tenantId)
    }

    // MARK: - Tenant Tests

    func testTenantDecoding() throws {
        let json = """
        {
            "id": "tenant-123",
            "name": "Test Company",
            "system_prompt": "You are a helpful assistant",
            "greeting_text": "Hello, how can I help?",
            "language": "cs",
            "vip_names": ["Boss", "Wife"],
            "marketing_email": "info@test.com",
            "plan": "free",
            "status": "active"
        }
        """.data(using: .utf8)!

        let decoder = JSONDecoder()
        let tenant = try decoder.decode(Tenant.self, from: json)

        XCTAssertEqual(tenant.id, "tenant-123")
        XCTAssertEqual(tenant.name, "Test Company")
        XCTAssertEqual(tenant.vipNames, ["Boss", "Wife"])
        XCTAssertEqual(tenant.marketingEmail, "info@test.com")
        XCTAssertEqual(tenant.language, "cs")
        XCTAssertEqual(tenant.plan, "free")
        XCTAssertEqual(tenant.status, "active")
    }

    // MARK: - TenantPhoneNumber Tests

    func testTenantPhoneNumberDecoding() throws {
        let json = """
        {
            "id": "phone-1",
            "twilio_number": "+420123456789",
            "is_primary": true
        }
        """.data(using: .utf8)!

        let decoder = JSONDecoder()
        let phoneNumber = try decoder.decode(TenantPhoneNumber.self, from: json)

        XCTAssertEqual(phoneNumber.id, "phone-1")
        XCTAssertEqual(phoneNumber.twilioNumber, "+420123456789")
        XCTAssertTrue(phoneNumber.isPrimary)
    }

    // MARK: - ScreeningResult Tests

    func testScreeningResultDecoding() throws {
        let json = """
        {
            "legitimacy_label": "suspicious",
            "legitimacy_confidence": 0.75,
            "lead_label": "cold",
            "intent_category": "sales",
            "intent_text": "Trying to sell insurance",
            "entities_json": {"company": "Insurance Corp"},
            "created_at": "2024-01-15T10:35:00Z"
        }
        """.data(using: .utf8)!

        let decoder = JSONDecoder()
        let screening = try decoder.decode(ScreeningResult.self, from: json)

        XCTAssertEqual(screening.legitimacyLabel, "suspicious")
        XCTAssertEqual(screening.legitimacyConfidence, 0.75)
        XCTAssertEqual(screening.leadLabel, "cold")
        XCTAssertEqual(screening.intentCategory, "sales")
        XCTAssertEqual(screening.intentText, "Trying to sell insurance")
    }

    // MARK: - Utterance Tests

    func testUtteranceDecoding() throws {
        let json = """
        {
            "speaker": "caller",
            "text": "Hello, is this the right number?",
            "sequence": 1,
            "started_at": "2024-01-15T10:30:00Z",
            "ended_at": "2024-01-15T10:30:05Z",
            "stt_confidence": 0.95,
            "interrupted": true
        }
        """.data(using: .utf8)!

        let decoder = JSONDecoder()
        let utterance = try decoder.decode(Utterance.self, from: json)

        XCTAssertEqual(utterance.speaker, "caller")
        XCTAssertEqual(utterance.text, "Hello, is this the right number?")
        XCTAssertEqual(utterance.sequence, 1)
        XCTAssertEqual(utterance.sttConfidence, 0.95)
        XCTAssertTrue(utterance.interrupted)
        XCTAssertFalse(utterance.isAgent)
        XCTAssertEqual(utterance.speakerDisplayName, "Volajici")
    }

    func testUtteranceAgentSpeaker() throws {
        let json = """
        {
            "speaker": "agent",
            "text": "How can I help you?",
            "sequence": 2,
            "started_at": null,
            "ended_at": null,
            "stt_confidence": null,
            "interrupted": false
        }
        """.data(using: .utf8)!

        let decoder = JSONDecoder()
        let utterance = try decoder.decode(Utterance.self, from: json)

        XCTAssertTrue(utterance.isAgent)
        XCTAssertEqual(utterance.speakerDisplayName, "Karen")
        XCTAssertNil(utterance.startedAt)
        XCTAssertNil(utterance.sttConfidence)
    }
}
