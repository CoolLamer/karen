import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { MantineProvider } from "@mantine/core";
import { vi, describe, it, expect, beforeEach, afterEach } from "vitest";
import { VoicePickerModal, useVoiceName } from "./VoicePickerModal";
import { api, Voice } from "../api";
import { renderHook } from "@testing-library/react";

// Mock the API module
vi.mock("../api", () => ({
  api: {
    getVoices: vi.fn(),
    previewVoice: vi.fn(),
    updateTenant: vi.fn(),
  },
}));

const mockVoices: Voice[] = [
  { id: "voice-1", name: "Rachel", description: "Přátelský hlas", gender: "female" },
  { id: "voice-2", name: "Adam", description: "Profesionální hlas", gender: "male" },
  { id: "voice-3", name: "Sarah", description: "Uklidňující hlas", gender: "female" },
];

const renderWithMantine = (ui: React.ReactElement) => {
  return render(<MantineProvider>{ui}</MantineProvider>);
};

describe("VoicePickerModal", () => {
  const mockOnClose = vi.fn();
  const mockOnSelect = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    (api.getVoices as ReturnType<typeof vi.fn>).mockResolvedValue({ voices: mockVoices });
    (api.updateTenant as ReturnType<typeof vi.fn>).mockResolvedValue({ tenant: {} });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("renders loading state initially", async () => {
    renderWithMantine(
      <VoicePickerModal
        opened={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
      />
    );

    // Should show loading spinner initially (Mantine Loader)
    expect(document.querySelector(".mantine-Loader-root")).toBeInTheDocument();
  });

  it("renders voices after loading", async () => {
    renderWithMantine(
      <VoicePickerModal
        opened={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
      />
    );

    await waitFor(() => {
      expect(screen.getByText("Rachel")).toBeInTheDocument();
    });

    expect(screen.getByText("Adam")).toBeInTheDocument();
    expect(screen.getByText("Sarah")).toBeInTheDocument();
  });

  it("renders voice descriptions", async () => {
    renderWithMantine(
      <VoicePickerModal
        opened={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
      />
    );

    await waitFor(() => {
      expect(screen.getByText("Přátelský hlas")).toBeInTheDocument();
    });

    expect(screen.getByText("Profesionální hlas")).toBeInTheDocument();
    expect(screen.getByText("Uklidňující hlas")).toBeInTheDocument();
  });

  it("shows error when loading fails", async () => {
    (api.getVoices as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("Network error"));

    renderWithMantine(
      <VoicePickerModal
        opened={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
      />
    );

    await waitFor(() => {
      expect(screen.getByText("Nepodařilo se načíst hlasy")).toBeInTheDocument();
    });
  });

  it("pre-selects current voice", async () => {
    renderWithMantine(
      <VoicePickerModal
        opened={true}
        onClose={mockOnClose}
        currentVoiceId="voice-2"
        onSelect={mockOnSelect}
      />
    );

    await waitFor(() => {
      expect(screen.getByText("Adam")).toBeInTheDocument();
    });

    // The radio for voice-2 should be checked
    const radios = screen.getAllByRole("radio");
    const adamRadio = radios.find((r) => r.getAttribute("value") === "voice-2");
    expect(adamRadio).toBeChecked();
  });

  it("selects first voice when no current voice is set", async () => {
    renderWithMantine(
      <VoicePickerModal
        opened={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
      />
    );

    await waitFor(() => {
      expect(screen.getByText("Rachel")).toBeInTheDocument();
    });

    // The first voice should be selected
    const radios = screen.getAllByRole("radio");
    const rachelRadio = radios.find((r) => r.getAttribute("value") === "voice-1");
    expect(rachelRadio).toBeChecked();
  });

  it("calls onClose when cancel button is clicked", async () => {
    renderWithMantine(
      <VoicePickerModal
        opened={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
      />
    );

    await waitFor(() => {
      expect(screen.getByText("Rachel")).toBeInTheDocument();
    });

    const cancelButton = screen.getByRole("button", { name: /zrušit/i });
    fireEvent.click(cancelButton);

    expect(mockOnClose).toHaveBeenCalled();
  });

  it("calls API and callbacks on confirm", async () => {
    renderWithMantine(
      <VoicePickerModal
        opened={true}
        onClose={mockOnClose}
        currentVoiceId="voice-1"
        onSelect={mockOnSelect}
      />
    );

    await waitFor(() => {
      expect(screen.getByText("Rachel")).toBeInTheDocument();
    });

    const confirmButton = screen.getByRole("button", { name: /potvrdit/i });
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(api.updateTenant).toHaveBeenCalledWith({ voice_id: "voice-1" });
    });

    expect(mockOnSelect).toHaveBeenCalledWith("voice-1");
    expect(mockOnClose).toHaveBeenCalled();
  });

  it("shows error when save fails", async () => {
    (api.updateTenant as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("Save failed"));

    renderWithMantine(
      <VoicePickerModal
        opened={true}
        onClose={mockOnClose}
        currentVoiceId="voice-1"
        onSelect={mockOnSelect}
      />
    );

    await waitFor(() => {
      expect(screen.getByText("Rachel")).toBeInTheDocument();
    });

    const confirmButton = screen.getByRole("button", { name: /potvrdit/i });
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(screen.getByText("Nepodařilo se uložit hlas")).toBeInTheDocument();
    });
  });

  it("renders modal title", async () => {
    renderWithMantine(
      <VoicePickerModal
        opened={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
      />
    );

    expect(screen.getByText("Vyberte hlas pro Karen")).toBeInTheDocument();
  });

  it("renders footer text", async () => {
    renderWithMantine(
      <VoicePickerModal
        opened={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
      />
    );

    await waitFor(() => {
      expect(screen.getByText("Rachel")).toBeInTheDocument();
    });

    expect(screen.getByText("Vybraný hlas se použije při příštím hovoru.")).toBeInTheDocument();
  });

  it("disables confirm button when loading", async () => {
    // Make getVoices hang
    (api.getVoices as ReturnType<typeof vi.fn>).mockImplementation(() => new Promise(() => {}));

    renderWithMantine(
      <VoicePickerModal
        opened={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
      />
    );

    const confirmButton = screen.getByRole("button", { name: /potvrdit/i });
    expect(confirmButton).toBeDisabled();
  });
});

describe("useVoiceName", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (api.getVoices as ReturnType<typeof vi.fn>).mockResolvedValue({ voices: mockVoices });
  });

  it("returns 'Výchozí' when voiceId is undefined", () => {
    const { result } = renderHook(() => useVoiceName(undefined));
    expect(result.current).toBe("Výchozí");
  });

  it("returns 'Výchozí' when voiceId is empty string", () => {
    const { result } = renderHook(() => useVoiceName(""));
    expect(result.current).toBe("Výchozí");
  });

  it("returns voice name when found", async () => {
    const { result } = renderHook(() => useVoiceName("voice-1"));

    await waitFor(() => {
      expect(result.current).toBe("Rachel");
    });
  });

  it("returns 'Výchozí' when voice not found", async () => {
    const { result } = renderHook(() => useVoiceName("non-existent-voice"));

    // Initially returns Výchozí while loading
    expect(result.current).toBe("Výchozí");

    // After loading, still returns Výchozí because voice not found
    await waitFor(() => {
      expect(api.getVoices).toHaveBeenCalled();
    });

    // Should still be Výchozí since the voice wasn't found
    expect(result.current).toBe("Výchozí");
  });
});
