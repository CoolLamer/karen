import React, { useState, useEffect, useRef } from "react";
import {
  Modal,
  Stack,
  Group,
  Text,
  Button,
  Paper,
  ActionIcon,
  ThemeIcon,
  Radio,
  Loader,
  Alert,
} from "@mantine/core";
import { IconPlayerPlay, IconPlayerStop, IconUser, IconAlertCircle } from "@tabler/icons-react";
import { api, Voice } from "../api";

interface VoicePickerModalProps {
  opened: boolean;
  onClose: () => void;
  currentVoiceId?: string;
  onSelect: (voiceId: string) => void;
}

export function VoicePickerModal({
  opened,
  onClose,
  currentVoiceId,
  onSelect,
}: VoicePickerModalProps) {
  const [voices, setVoices] = useState<Voice[]>([]);
  const [selectedVoiceId, setSelectedVoiceId] = useState<string>(currentVoiceId || "");
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [playingVoiceId, setPlayingVoiceId] = useState<string | null>(null);
  const [loadingVoiceId, setLoadingVoiceId] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const audioUrlRef = useRef<string | null>(null);

  useEffect(() => {
    const loadVoices = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const { voices } = await api.getVoices();
        setVoices(voices);
        // If no current voice, select the first one
        if (!currentVoiceId && voices.length > 0) {
          setSelectedVoiceId(voices[0].id);
        }
      } catch {
        setError("Nepodařilo se načíst hlasy");
      } finally {
        setIsLoading(false);
      }
    };

    if (opened) {
      loadVoices();
      setSelectedVoiceId(currentVoiceId || "");
    }
    return () => {
      stopAudio();
    };
  }, [opened, currentVoiceId]);

  const stopAudio = () => {
    if (audioRef.current) {
      audioRef.current.pause();
      audioRef.current = null;
    }
    if (audioUrlRef.current) {
      URL.revokeObjectURL(audioUrlRef.current);
      audioUrlRef.current = null;
    }
    setPlayingVoiceId(null);
  };

  const handlePreview = async (voiceId: string) => {
    // Stop any currently playing audio
    stopAudio();

    // If clicking the same voice that was playing, just stop
    if (playingVoiceId === voiceId) {
      return;
    }

    setLoadingVoiceId(voiceId);
    try {
      const audioBlob = await api.previewVoice(voiceId);
      const audioUrl = URL.createObjectURL(audioBlob);
      audioUrlRef.current = audioUrl;

      const audio = new Audio(audioUrl);
      audioRef.current = audio;

      audio.onended = () => {
        setPlayingVoiceId(null);
      };

      audio.onerror = () => {
        setPlayingVoiceId(null);
        setError("Nepodařilo se přehrát ukázku");
      };

      await audio.play();
      setPlayingVoiceId(voiceId);
    } catch {
      setError("Nepodařilo se načíst ukázku hlasu");
    } finally {
      setLoadingVoiceId(null);
    }
  };

  const handleConfirm = async () => {
    if (!selectedVoiceId) return;

    setIsSaving(true);
    stopAudio();
    try {
      await api.updateTenant({ voice_id: selectedVoiceId });
      onSelect(selectedVoiceId);
      onClose();
    } catch {
      setError("Nepodařilo se uložit hlas");
    } finally {
      setIsSaving(false);
    }
  };

  const handleClose = () => {
    stopAudio();
    onClose();
  };

  return (
    <Modal
      opened={opened}
      onClose={handleClose}
      title="Vyberte hlas pro Karen"
      centered
      size="md"
    >
      <Stack gap="md">
        {error && (
          <Alert icon={<IconAlertCircle size={16} />} color="red" variant="light">
            {error}
          </Alert>
        )}

        {isLoading ? (
          <Group justify="center" py="xl">
            <Loader size="md" />
          </Group>
        ) : (
          <Radio.Group value={selectedVoiceId} onChange={setSelectedVoiceId}>
            <Stack gap="sm">
              {voices.map((voice) => (
                <Paper
                  key={voice.id}
                  p="sm"
                  radius="md"
                  withBorder
                  style={{
                    borderColor:
                      selectedVoiceId === voice.id
                        ? "var(--mantine-color-teal-5)"
                        : undefined,
                    cursor: "pointer",
                  }}
                  onClick={() => setSelectedVoiceId(voice.id)}
                >
                  <Group justify="space-between" wrap="nowrap">
                    <Group gap="sm" wrap="nowrap">
                      <Radio value={voice.id} />
                      <ThemeIcon
                        size="md"
                        variant="light"
                        color={voice.gender === "female" ? "pink" : "blue"}
                      >
                        <IconUser size={14} />
                      </ThemeIcon>
                      <div>
                        <Text size="sm" fw={500}>
                          {voice.name}
                        </Text>
                        <Text size="xs" c="dimmed">
                          {voice.description}
                        </Text>
                      </div>
                    </Group>
                    <ActionIcon
                      variant="subtle"
                      color={playingVoiceId === voice.id ? "teal" : "gray"}
                      onClick={(e) => {
                        e.stopPropagation();
                        handlePreview(voice.id);
                      }}
                      loading={loadingVoiceId === voice.id}
                      disabled={loadingVoiceId !== null && loadingVoiceId !== voice.id}
                    >
                      {playingVoiceId === voice.id ? (
                        <IconPlayerStop size={16} />
                      ) : (
                        <IconPlayerPlay size={16} />
                      )}
                    </ActionIcon>
                  </Group>
                </Paper>
              ))}
            </Stack>
          </Radio.Group>
        )}

        <Text size="xs" c="dimmed" ta="center">
          Vybraný hlas se použije při příštím hovoru.
        </Text>

        <Group justify="flex-end" gap="sm">
          <Button variant="subtle" onClick={handleClose}>
            Zrušit
          </Button>
          <Button
            onClick={handleConfirm}
            loading={isSaving}
            disabled={!selectedVoiceId || isLoading}
          >
            Potvrdit
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
}

// Helper to get voice name from ID (for display in settings)
export function useVoiceName(voiceId?: string): string {
  const [voices, setVoices] = useState<Voice[]>([]);

  useEffect(() => {
    api.getVoices().then(({ voices }) => setVoices(voices)).catch(() => {});
  }, []);

  if (!voiceId) return "Výchozí";
  const voice = voices.find((v) => v.id === voiceId);
  return voice?.name || "Výchozí";
}
