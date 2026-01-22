import { render, screen, fireEvent } from "@testing-library/react";
import { MantineProvider } from "@mantine/core";
import { vi, describe, it, expect, beforeEach } from "vitest";
import { RedirectWizard } from "./RedirectWizard";

const renderWithMantine = (ui: React.ReactElement) => {
  return render(<MantineProvider>{ui}</MantineProvider>);
};

describe("RedirectWizard", () => {
  const mockOnComplete = vi.fn();
  const testKarenNumber = "+420 123 456 789";

  beforeEach(() => {
    vi.clearAllMocks();
    // Mock window.location.href
    Object.defineProperty(window, "location", {
      value: { href: "" },
      writable: true,
    });
  });

  describe("IntroStep", () => {
    it("renders intro step initially", () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );

      expect(screen.getByText("Nastavíme přesměrování hovorů")).toBeInTheDocument();
      expect(screen.getByText(/Provedeme tě 4 krátkými kroky/)).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /Začít/i })).toBeInTheDocument();
    });

    it("shows list of steps to be done", () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );

      expect(screen.getByText("Vymažeme stávající přesměrování")).toBeInTheDocument();
      expect(screen.getByText("Nastavíme přesměrování, když nezvedáš")).toBeInTheDocument();
      expect(screen.getByText("Nastavíme přesměrování, když máš obsazeno")).toBeInTheDocument();
      expect(screen.getByText("Nastavíme přesměrování, když jsi nedostupný")).toBeInTheDocument();
    });

    it("moves to clear step when clicking Start button", () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );

      fireEvent.click(screen.getByRole("button", { name: /Začít/i }));

      expect(screen.getByText("Krok 1: Vymazat stávající přesměrování")).toBeInTheDocument();
    });
  });

  describe("ClearStep", () => {
    it("shows clear all redirects code", () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );

      // Move to clear step
      fireEvent.click(screen.getByRole("button", { name: /Začít/i }));

      expect(screen.getByText("##002#")).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /Vytočit kód/i })).toBeInTheDocument();
    });

    it("shows confirmation buttons after dialing", () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );

      // Move to clear step
      fireEvent.click(screen.getByRole("button", { name: /Začít/i }));
      // Click dial button
      fireEvent.click(screen.getByRole("button", { name: /Vytočit kód/i }));

      expect(screen.getByText("Viděl/a jsi potvrzení od operátora?")).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /Ano, aktivováno/i })).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /Zkusit znovu/i })).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /Přeskočit/i })).toBeInTheDocument();
    });

    it("moves to noAnswer step when confirming", () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );

      // Move to clear step
      fireEvent.click(screen.getByRole("button", { name: /Začít/i }));
      // Click dial button
      fireEvent.click(screen.getByRole("button", { name: /Vytočit kód/i }));
      // Confirm
      fireEvent.click(screen.getByRole("button", { name: /Ano, aktivováno/i }));

      expect(screen.getByText("Krok 2: Když nezvedáš")).toBeInTheDocument();
    });

    it("allows skipping the clear step", () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );

      // Move to clear step
      fireEvent.click(screen.getByRole("button", { name: /Začít/i }));
      // Skip
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));

      expect(screen.getByText("Krok 2: Když nezvedáš")).toBeInTheDocument();
    });
  });

  describe("NoAnswerStep", () => {
    const moveToNoAnswerStep = () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );
      // Move through steps
      fireEvent.click(screen.getByRole("button", { name: /Začít/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
    };

    it("shows noAnswer step with correct dial code", () => {
      moveToNoAnswerStep();

      expect(screen.getByText("Krok 2: Když nezvedáš")).toBeInTheDocument();
      // Default time is 10 seconds
      expect(screen.getByText("**61*+420123456789**10#")).toBeInTheDocument();
    });

    it("shows Karen number", () => {
      moveToNoAnswerStep();

      expect(screen.getByText(testKarenNumber)).toBeInTheDocument();
    });

    it("shows timing control when clicking change timing link", () => {
      moveToNoAnswerStep();

      // Find the timing link by text content (it's an UnstyledButton)
      const timingLink = screen.getByText(/Změnit časování \(10s\)/i);
      fireEvent.click(timingLink);

      expect(screen.getByText("Po kolika sekundách přesměrovat?")).toBeInTheDocument();
    });

    it("moves to busy step when confirming", () => {
      moveToNoAnswerStep();

      fireEvent.click(screen.getByRole("button", { name: /Vytočit kód/i }));
      fireEvent.click(screen.getByRole("button", { name: /Ano, aktivováno/i }));

      expect(screen.getByText("Krok 3: Když máš obsazeno")).toBeInTheDocument();
    });
  });

  describe("BusyStep", () => {
    const moveToBusyStep = () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );
      // Move through steps
      fireEvent.click(screen.getByRole("button", { name: /Začít/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
    };

    it("shows busy step with correct dial code", () => {
      moveToBusyStep();

      expect(screen.getByText("Krok 3: Když máš obsazeno")).toBeInTheDocument();
      expect(screen.getByText("**67*+420123456789#")).toBeInTheDocument();
    });

    it("moves to unreachable step when confirming", () => {
      moveToBusyStep();

      fireEvent.click(screen.getByRole("button", { name: /Vytočit kód/i }));
      fireEvent.click(screen.getByRole("button", { name: /Ano, aktivováno/i }));

      expect(screen.getByText("Krok 4: Když jsi nedostupný")).toBeInTheDocument();
    });
  });

  describe("UnreachableStep", () => {
    const moveToUnreachableStep = () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );
      // Move through steps
      fireEvent.click(screen.getByRole("button", { name: /Začít/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
    };

    it("shows unreachable step with correct dial code", () => {
      moveToUnreachableStep();

      expect(screen.getByText("Krok 4: Když jsi nedostupný")).toBeInTheDocument();
      expect(screen.getByText("**62*+420123456789#")).toBeInTheDocument();
    });

    it("moves to complete step when confirming", () => {
      moveToUnreachableStep();

      fireEvent.click(screen.getByRole("button", { name: /Vytočit kód/i }));
      fireEvent.click(screen.getByRole("button", { name: /Ano, aktivováno/i }));

      expect(screen.getByText("Nastavení dokončeno!")).toBeInTheDocument();
    });
  });

  describe("CompleteStep", () => {
    it("shows completed steps summary", () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );

      // Complete all steps
      fireEvent.click(screen.getByRole("button", { name: /Začít/i }));
      // Clear step
      fireEvent.click(screen.getByRole("button", { name: /Vytočit kód/i }));
      fireEvent.click(screen.getByRole("button", { name: /Ano, aktivováno/i }));
      // NoAnswer step
      fireEvent.click(screen.getByRole("button", { name: /Vytočit kód/i }));
      fireEvent.click(screen.getByRole("button", { name: /Ano, aktivováno/i }));
      // Busy step
      fireEvent.click(screen.getByRole("button", { name: /Vytočit kód/i }));
      fireEvent.click(screen.getByRole("button", { name: /Ano, aktivováno/i }));
      // Unreachable step
      fireEvent.click(screen.getByRole("button", { name: /Vytočit kód/i }));
      fireEvent.click(screen.getByRole("button", { name: /Ano, aktivováno/i }));

      expect(screen.getByText("Nastavení dokončeno!")).toBeInTheDocument();
      expect(screen.getByText(/Aktivováno \(4\)/)).toBeInTheDocument();
    });

    it("shows skipped steps summary", () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );

      // Skip all steps
      fireEvent.click(screen.getByRole("button", { name: /Začít/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));

      expect(screen.getByText("Nastavení dokončeno!")).toBeInTheDocument();
      expect(screen.getByText(/Přeskočeno \(4\)/)).toBeInTheDocument();
    });

    it("calls onComplete when clicking finish button", () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );

      // Skip all steps to complete
      fireEvent.click(screen.getByRole("button", { name: /Začít/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));

      // Click finish
      fireEvent.click(screen.getByRole("button", { name: /Pokračovat/i }));

      expect(mockOnComplete).toHaveBeenCalled();
    });
  });

  describe("WizardProgress", () => {
    it("shows progress bar after intro step", () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );

      // Move to clear step
      fireEvent.click(screen.getByRole("button", { name: /Začít/i }));

      // Check for progress indicator labels
      expect(screen.getByText("Vymazat")).toBeInTheDocument();
      expect(screen.getByText("Nezvedám")).toBeInTheDocument();
      expect(screen.getByText("Obsazeno")).toBeInTheDocument();
      expect(screen.getByText("Nedostupný")).toBeInTheDocument();
    });

    it("does not show progress bar on intro step", () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );

      // On intro step, these labels should not be visible
      expect(screen.queryByText("Vymazat")).not.toBeInTheDocument();
    });

    it("does not show progress bar on complete step", () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );

      // Skip all steps to complete
      fireEvent.click(screen.getByRole("button", { name: /Začít/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
      fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));

      // On complete step, these labels should not be visible
      expect(screen.queryByText("Vymazat")).not.toBeInTheDocument();
    });
  });

  describe("dial functionality", () => {
    it("triggers tel: link when clicking dial button", () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );

      // Move to clear step
      fireEvent.click(screen.getByRole("button", { name: /Začít/i }));
      // Click dial
      fireEvent.click(screen.getByRole("button", { name: /Vytočit kód/i }));

      expect(window.location.href).toBe("tel:##002#");
    });

    it("resets dial state when clicking retry", () => {
      renderWithMantine(
        <RedirectWizard karenNumber={testKarenNumber} onComplete={mockOnComplete} />
      );

      // Move to clear step
      fireEvent.click(screen.getByRole("button", { name: /Začít/i }));
      // Click dial
      fireEvent.click(screen.getByRole("button", { name: /Vytočit kód/i }));
      // Click retry
      fireEvent.click(screen.getByRole("button", { name: /Zkusit znovu/i }));

      // Should show dial button again
      expect(screen.getByRole("button", { name: /Vytočit kód/i })).toBeInTheDocument();
    });
  });
});
