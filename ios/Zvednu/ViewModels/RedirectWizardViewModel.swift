import Foundation
import SwiftUI

@MainActor
class RedirectWizardViewModel: ObservableObject {
    // MARK: - Types

    enum WizardStep: Int, CaseIterable {
        case intro = 0
        case clear = 1
        case noAnswer = 2
        case busy = 3
        case unreachable = 4
        case complete = 5

        var stepNumber: Int? {
            switch self {
            case .intro, .complete: return nil
            case .clear: return 1
            case .noAnswer: return 2
            case .busy: return 3
            case .unreachable: return 4
            }
        }

        var title: String {
            switch self {
            case .intro: return "Nastavíme přesměrování hovorů"
            case .clear: return "Krok 1: Vymazat stávající přesměrování"
            case .noAnswer: return "Krok 2: Když nezvedáš"
            case .busy: return "Krok 3: Když máš obsazeno"
            case .unreachable: return "Krok 4: Když jsi nedostupný"
            case .complete: return "Nastavení dokončeno!"
            }
        }
    }

    enum StepStatus: String {
        case pending
        case completed
        case skipped
    }

    // MARK: - Published Properties

    @Published var currentStep: WizardStep = .intro
    @Published var clearStepStatus: StepStatus = .pending
    @Published var noAnswerStepStatus: StepStatus = .pending
    @Published var busyStepStatus: StepStatus = .pending
    @Published var unreachableStepStatus: StepStatus = .pending
    @Published var noAnswerTime: Int = RedirectCodes.defaultNoAnswerTime
    @Published var hasDialed: Bool = false
    @Published var showTimingSelector: Bool = false

    // MARK: - Computed Properties

    var progressPercent: Double {
        let index = currentStep.rawValue
        let total = WizardStep.allCases.count - 1 // exclude complete from denominator
        return Double(index) / Double(total) * 100
    }

    var activatedSteps: [(key: String, label: String)] {
        stepsWithStatus(.completed)
    }

    var skippedSteps: [(key: String, label: String)] {
        stepsWithStatus(.skipped)
    }

    private func stepsWithStatus(_ targetStatus: StepStatus) -> [(key: String, label: String)] {
        let stepLabels: [(key: String, label: String, status: StepStatus)] = [
            ("clear", "Vymazání stávajících přesměrování", clearStepStatus),
            ("noAnswer", "Přesměrování když nezvedáš", noAnswerStepStatus),
            ("busy", "Přesměrování při obsazení", busyStepStatus),
            ("unreachable", "Přesměrování při nedostupnosti", unreachableStepStatus),
        ]
        return stepLabels
            .filter { $0.status == targetStatus }
            .map { (key: $0.key, label: $0.label) }
    }

    // MARK: - Actions

    func goToNextStep() {
        guard let currentIndex = WizardStep.allCases.firstIndex(of: currentStep),
              currentIndex < WizardStep.allCases.count - 1 else { return }
        currentStep = WizardStep.allCases[currentIndex + 1]
        hasDialed = false
        showTimingSelector = false
    }

    func confirmStep() {
        updateStatus(for: currentStep, status: .completed)
        goToNextStep()
    }

    func skipStep() {
        updateStatus(for: currentStep, status: .skipped)
        goToNextStep()
    }

    func retryDial() {
        hasDialed = false
    }

    func markDialed() {
        hasDialed = true
    }

    func toggleTimingSelector() {
        showTimingSelector.toggle()
    }

    func getDialCode(for step: WizardStep, phoneNumber: String) -> String {
        switch step {
        case .clear:
            return RedirectCodes.clearAllRedirectsCode
        case .noAnswer:
            return RedirectCodes.getDialCode(type: .noAnswer, phoneNumber: phoneNumber, time: noAnswerTime)
        case .busy:
            return RedirectCodes.getDialCode(type: .busy, phoneNumber: phoneNumber)
        case .unreachable:
            return RedirectCodes.getDialCode(type: .unreachable, phoneNumber: phoneNumber)
        default:
            return ""
        }
    }

    func getDescription(for step: WizardStep) -> String {
        switch step {
        case .intro:
            return "Provedeme tě 4 krátkými kroky. Na mobilu stačí klikat na tlačítka."
        case .clear:
            return "Nejdřív vymažeme případná existující přesměrování, aby nedošlo ke konfliktu."
        case .noAnswer:
            return "Když nezvedneš do \(noAnswerTime) sekund, hovor se přesměruje na Karen."
        case .busy:
            return "Když máš obsazeno nebo odmítneš hovor, přesměruje se na Karen."
        case .unreachable:
            return "Když nemáš signál nebo máš vypnutý telefon, hovor jde na Karen."
        case .complete:
            return "Nastavení můžeš kdykoliv změnit v sekci Nastavení."
        }
    }

    func status(for step: WizardStep) -> StepStatus {
        switch step {
        case .clear: return clearStepStatus
        case .noAnswer: return noAnswerStepStatus
        case .busy: return busyStepStatus
        case .unreachable: return unreachableStepStatus
        default: return .pending
        }
    }

    // MARK: - Private

    private func updateStatus(for step: WizardStep, status: StepStatus) {
        switch step {
        case .clear:
            clearStepStatus = status
        case .noAnswer:
            noAnswerStepStatus = status
        case .busy:
            busyStepStatus = status
        case .unreachable:
            unreachableStepStatus = status
        default:
            break
        }
    }
}
