import React from "react";
import { useParams } from "react-router-dom";
import { Badge, Group, Paper, Stack, Text, Title } from "@mantine/core";
import { api, CallDetail } from "../api";

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
      <Title order={2}>Call Detail</Title>

      {error && (
        <Paper p="md" withBorder>
          <Text c="red">Error: {error}</Text>
        </Paper>
      )}

      {!call && !error && <Text c="dimmed">Loading…</Text>}

      {call && (
        <>
          <Paper p="md" withBorder>
            <Group justify="space-between" align="flex-start">
              <Stack gap={4}>
                <Text fw={700}>{call.from_number}</Text>
                <Text size="sm" c="dimmed">
                  to {call.to_number}
                </Text>
                <Text size="sm">Started: {new Date(call.started_at).toLocaleString()}</Text>
                <Text size="sm">Status: {call.status}</Text>
              </Stack>
              <Stack gap={6} align="flex-end">
                <Badge variant="light">{call.screening?.legitimacy_label ?? "unknown"}</Badge>
                {typeof call.screening?.legitimacy_confidence === "number" && (
                  <Text size="xs" c="dimmed">
                    confidence: {(call.screening.legitimacy_confidence * 100).toFixed(0)}%
                  </Text>
                )}
              </Stack>
            </Group>
            <Text mt="sm" size="sm">
              <strong>Intent:</strong> {call.screening?.intent_text || "—"}
            </Text>
          </Paper>

          <Paper p="md" withBorder>
            <Title order={4} mb="sm">
              Transcript
            </Title>
            {call.utterances?.length ? (
              <Stack gap="xs">
                {call.utterances.map((u) => (
                  <Paper key={u.sequence} p="sm" withBorder>
                    <Text size="xs" c="dimmed">
                      {u.speaker} • #{u.sequence}
                    </Text>
                    <Text size="sm">{u.text}</Text>
                  </Paper>
                ))}
              </Stack>
            ) : (
              <Text size="sm" c="dimmed">
                No transcript yet.
              </Text>
            )}
          </Paper>
        </>
      )}
    </Stack>
  );
}


