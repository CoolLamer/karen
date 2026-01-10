import React from "react";
import { useParams } from "react-router-dom";
import { Badge, Group, Paper, Stack, Text, Title } from "@mantine/core";
import { api, CallDetail } from "../api";

function formatStatus(status: string) {
  switch (status) {
    case "in_progress":
      return "Probíhá";
    case "completed":
      return "Dokončeno";
    case "queued":
      return "Čeká";
    case "ringing":
      return "Vyzvání";
    default:
      return status;
  }
}

function formatLabel(label: string) {
  switch (label) {
    case "legitimate":
      return "legitimní";
    case "unknown":
      return "neznámé";
    default:
      return label;
  }
}

function formatSpeaker(speaker: string) {
  switch (speaker) {
    case "agent":
      return "Karen";
    case "caller":
      return "Volající";
    default:
      return speaker;
  }
}

export function CallDetailPage() {
  const { providerCallId } = useParams();
  const [call, setCall] = React.useState<CallDetail | null>(null);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    if (!providerCallId) return;
    api
      .getCall(providerCallId)
      .then(setCall)
      .catch((e) => setError(String(e)));
  }, [providerCallId]);

  return (
    <Stack gap="md" py="md">
      <Title order={2}>Detail hovoru</Title>

      {error && (
        <Paper p="md" withBorder>
          <Text c="red">Chyba: {error}</Text>
        </Paper>
      )}

      {!call && !error && <Text c="dimmed">Načítání…</Text>}

      {call && (
        <>
          <Paper p="md" withBorder>
            <Group justify="space-between" align="flex-start">
              <Stack gap={4}>
                <Text fw={700}>{call.from_number}</Text>
                <Text size="sm" c="dimmed">
                  na {call.to_number}
                </Text>
                <Text size="sm">Začátek: {new Date(call.started_at).toLocaleString("cs-CZ")}</Text>
                <Text size="sm">Stav: {formatStatus(call.status)}</Text>
              </Stack>
              <Stack gap={6} align="flex-end">
                <Badge variant="light">{formatLabel(call.screening?.legitimacy_label ?? "unknown")}</Badge>
                {typeof call.screening?.legitimacy_confidence === "number" && (
                  <Text size="xs" c="dimmed">
                    spolehlivost: {(call.screening.legitimacy_confidence * 100).toFixed(0)}%
                  </Text>
                )}
              </Stack>
            </Group>
            <Text mt="sm" size="sm">
              <strong>Účel:</strong> {call.screening?.intent_text || "—"}
            </Text>
          </Paper>

          <Paper p="md" withBorder>
            <Title order={4} mb="sm">
              Přepis
            </Title>
            {call.utterances?.length ? (
              <Stack gap="xs">
                {call.utterances.map((u) => (
                  <Paper key={u.sequence} p="sm" withBorder>
                    <Text size="xs" c="dimmed">
                      {formatSpeaker(u.speaker)} • #{u.sequence}
                    </Text>
                    <Text size="sm">{u.text}</Text>
                  </Paper>
                ))}
              </Stack>
            ) : (
              <Text size="sm" c="dimmed">
                Přepis zatím není k dispozici.
              </Text>
            )}
          </Paper>
        </>
      )}
    </Stack>
  );
}
