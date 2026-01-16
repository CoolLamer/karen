import Foundation

enum APIError: LocalizedError {
    case invalidURL
    case invalidResponse
    case httpError(Int, String)
    case decodingError(Error)
    case networkError(Error)
    case unauthorized

    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "Neplatná URL"
        case .invalidResponse:
            return "Neplatná odpověď serveru"
        case .httpError(let code, let message):
            return "Chyba \(code): \(message)"
        case .decodingError(let error):
            return "Chyba při zpracování dat: \(error.localizedDescription)"
        case .networkError(let error):
            return "Chyba sítě: \(error.localizedDescription)"
        case .unauthorized:
            return "Neplatné přihlášení"
        }
    }
}

actor APIClient {
    static let shared = APIClient()

    private let baseURL: String
    private let session: URLSession
    private let decoder: JSONDecoder

    private init() {
        self.baseURL = AppConfig.apiBaseURL
        self.session = URLSession.shared
        self.decoder = JSONDecoder()
    }

    // MARK: - Token Management

    private var authToken: String? {
        KeychainManager.shared.getToken()
    }

    func setAuthToken(_ token: String?) {
        if let token = token {
            try? KeychainManager.shared.saveToken(token)
        } else {
            KeychainManager.shared.deleteToken()
        }
    }

    // MARK: - Request Methods

    func get<T: Decodable>(_ path: String) async throws -> T {
        try await request(path, method: "GET", body: nil as Empty?)
    }

    func post<T: Decodable, B: Encodable>(_ path: String, body: B?) async throws -> T {
        try await request(path, method: "POST", body: body)
    }

    func patch<T: Decodable, B: Encodable>(_ path: String, body: B?) async throws -> T {
        try await request(path, method: "PATCH", body: body)
    }

    func delete<T: Decodable>(_ path: String) async throws -> T {
        try await request(path, method: "DELETE", body: nil as Empty?)
    }

    func postEmpty<T: Decodable>(_ path: String) async throws -> T {
        try await request(path, method: "POST", body: nil as Empty?)
    }

    func patchEmpty<T: Decodable>(_ path: String) async throws -> T {
        try await request(path, method: "PATCH", body: nil as Empty?)
    }

    func deleteEmpty(_ path: String) async throws {
        let _: Empty = try await request(path, method: "DELETE", body: nil as Empty?)
    }

    // MARK: - Private Request Implementation

    private func request<T: Decodable, B: Encodable>(
        _ path: String,
        method: String,
        body: B?
    ) async throws -> T {
        guard let url = URL(string: baseURL + path) else {
            throw APIError.invalidURL
        }

        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if let token = authToken {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        if let body = body {
            let encoder = JSONEncoder()
            request.httpBody = try encoder.encode(body)
        }

        do {
            let (data, response) = try await session.data(for: request)

            guard let httpResponse = response as? HTTPURLResponse else {
                throw APIError.invalidResponse
            }

            // Handle 401 Unauthorized
            if httpResponse.statusCode == 401 {
                setAuthToken(nil)
                throw APIError.unauthorized
            }

            // Handle non-success status codes
            guard (200...299).contains(httpResponse.statusCode) else {
                let message = String(data: data, encoding: .utf8) ?? "Unknown error"
                throw APIError.httpError(httpResponse.statusCode, message)
            }

            // Handle 204 No Content
            if httpResponse.statusCode == 204 || data.isEmpty {
                // Return empty response for void endpoints
                if T.self == Empty.self, let result = Empty() as? T {
                    return result
                }
            }

            do {
                return try decoder.decode(T.self, from: data)
            } catch {
                throw APIError.decodingError(error)
            }
        } catch let error as APIError {
            throw error
        } catch {
            throw APIError.networkError(error)
        }
    }
}

// Empty type for requests/responses without body
struct Empty: Codable {}
