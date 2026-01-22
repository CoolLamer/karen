import { render, screen, fireEvent } from "@testing-library/react";
import { MantineProvider } from "@mantine/core";
import { vi, describe, it, expect, beforeEach } from "vitest";
import { ForwardingSetupModal } from "./ForwardingSetupModal";

const renderWithMantine = (ui: React.ReactElement) => {
  return render(<MantineProvider>{ui}</MantineProvider>);
};

describe("ForwardingSetupModal", () => {
  const mockOnClose = vi.fn();
  const testKarenNumber = "+420 123 456 789";

  beforeEach(() => {
    vi.clearAllMocks();
    // Mock window.location.href
    Object.defineProperty(window, "location", {
      value: { href: "" },
      writable: true,
    });
  });

  it("renders modal with title", () => {
    renderWithMantine(
      <ForwardingSetupModal
        opened={true}
        onClose={mockOnClose}
        karenNumber={testKarenNumber}
      />
    );

    expect(screen.getByText("Jak nastavit přesměrování")).toBeInTheDocument();
  });

  it("shows mode toggle with wizard as default", () => {
    renderWithMantine(
      <ForwardingSetupModal
        opened={true}
        onClose={mockOnClose}
        karenNumber={testKarenNumber}
      />
    );

    expect(screen.getByText("Průvodce krok za krokem")).toBeInTheDocument();
    expect(screen.getByText("Rychlé nastavení")).toBeInTheDocument();
  });

  it("shows wizard mode by default", () => {
    renderWithMantine(
      <ForwardingSetupModal
        opened={true}
        onClose={mockOnClose}
        karenNumber={testKarenNumber}
      />
    );

    // Wizard intro step should be visible
    expect(screen.getByText("Nastavíme přesměrování hovorů")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Začít/i })).toBeInTheDocument();
  });

  it("switches to quick mode when clicking quick setup", () => {
    renderWithMantine(
      <ForwardingSetupModal
        opened={true}
        onClose={mockOnClose}
        karenNumber={testKarenNumber}
      />
    );

    // Click on quick setup
    fireEvent.click(screen.getByText("Rychlé nastavení"));

    // Quick mode content should be visible (accordion)
    expect(screen.getByText("Když nezvednu")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Zavřít/i })).toBeInTheDocument();
  });

  it("shows close button in quick mode", () => {
    renderWithMantine(
      <ForwardingSetupModal
        opened={true}
        onClose={mockOnClose}
        karenNumber={testKarenNumber}
      />
    );

    // Switch to quick mode
    fireEvent.click(screen.getByText("Rychlé nastavení"));

    const closeButton = screen.getByRole("button", { name: /Zavřít/i });
    expect(closeButton).toBeInTheDocument();

    fireEvent.click(closeButton);
    expect(mockOnClose).toHaveBeenCalled();
  });

  it("calls onClose when wizard completes", () => {
    renderWithMantine(
      <ForwardingSetupModal
        opened={true}
        onClose={mockOnClose}
        karenNumber={testKarenNumber}
      />
    );

    // Complete wizard by skipping all steps
    fireEvent.click(screen.getByRole("button", { name: /Začít/i }));
    fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
    fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
    fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
    fireEvent.click(screen.getByRole("button", { name: /Přeskočit tento krok/i }));
    fireEvent.click(screen.getByRole("button", { name: /Pokračovat/i }));

    expect(mockOnClose).toHaveBeenCalled();
  });

  it("does not render when closed", () => {
    renderWithMantine(
      <ForwardingSetupModal
        opened={false}
        onClose={mockOnClose}
        karenNumber={testKarenNumber}
      />
    );

    expect(screen.queryByText("Jak nastavit přesměrování")).not.toBeInTheDocument();
  });

  it("handles undefined karen number", () => {
    renderWithMantine(
      <ForwardingSetupModal
        opened={true}
        onClose={mockOnClose}
        karenNumber={undefined}
      />
    );

    // Should still render without crashing
    expect(screen.getByText("Nastavíme přesměrování hovorů")).toBeInTheDocument();
  });
});
