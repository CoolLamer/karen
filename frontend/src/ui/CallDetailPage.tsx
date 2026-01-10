import React from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Badge, Group, Paper, Stack, Text, Title, Box, ThemeIcon, Button, Progress } from "@mantine/core";
import { IconArrowLeft, IconRobot, IconUser, IconCheck, IconX, IconQuestionMark, IconMail } from "@tabler/icons-react";
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

function getLegitimacyConfig(label: string) {
  switch (label) {
    case "legitimate":
    case "legitimní":
      return { color: "green", label: "Legitimní", icon: <IconCheck size={14} /> };
    case "marketing":
      return { color: "yellow", label: "Marketing", icon: <IconMail size={14} /> };
    case "spam":
      return { color: "red", label: "Spam", icon: <IconX size={14} /> };
    default:
      return { color: "gray", label: "Neznámé", icon: <IconQuestionMark size={14} /> };
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
  const navigate = useNavigate();
  const [call, setCall] = React.useState<CallDetail | null>(null);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    if (!providerCallId) return;
    api
      .getCall(providerCallId)
      .then(setCall)
      .catch((e) => setError(String(e)));
  }, [providerCallId]);

  const legitimacyConfig = getLegitimacyConfig(call?.screening?.legitimacy_label ?? "unknown");

  return (
    <Stack gap="md" py="md">
      <Group>
        <Button
          variant="subtle"
          leftSection={<IconArrowLeft size={16} />}
          onClick={() => navigate("/inbox")}
          px={0}
        >
          Zpět na hovory
        </Button>
      </Group>

      <Title order={2}>Detail hovoru</Title>

      {error && (
        <Paper p="md" withBorder>
          <Text c="red">Chyba: {error}</Text>
        </Paper>
      )}

      {!call && !error && <Text c="dimmed">Načítání…</Text>}

      {call && (
        <>
          {/* Call info header */}
          <Paper p="lg" withBorder radius="md">
            <Group justify="space-between" align="flex-start">
              <Stack gap={4}>
                <Text size="xl" fw={700}>{call.from_number}</Text>
                <Text size="sm" c="dimmed">
                  na {call.to_number}
                </Text>
                <Text size="sm" mt="xs">
                  {new Date(call.started_at).toLocaleString("cs-CZ")}
                </Text>
                <Badge variant="light" color="gray" size="sm" mt="xs">
                  {formatStatus(call.status)}
                </Badge>
              </Stack>
              <Stack gap="xs" align="flex-end">
                <Badge
                  variant="light"
                  color={legitimacyConfig.color}
                  size="lg"
                  leftSection={legitimacyConfig.icon}
                >
                  {legitimacyConfig.label}
                </Badge>
                {typeof call.screening?.legitimacy_confidence === "number" && (
                  <Box w={100}>
                    <Text size="xs" c="dimmed" ta="right" mb={4}>
                      Spolehlivost: {(call.screening.legitimacy_confidence * 100).toFixed(0)}%
                    </Text>
                    <Progress
                      value={call.screening.legitimacy_confidence * 100}
                      size="xs"
                      color={legitimacyConfig.color}
                    />
                  </Box>
                )}
              </Stack>
            </Group>
            {call.screening?.intent_text && (
              <Paper p="sm" radius="sm" bg="gray.0" mt="md">
                <Text size="sm" fw={500}>Účel hovoru:</Text>
                <Text size="sm" c="dimmed">{call.screening.intent_text}</Text>
              </Paper>
            )}
          </Paper>

          {/* Transcript with chat bubbles */}
          <Paper p="lg" withBorder radius="md">
            <Title order={4} mb="md">
              Přepis hovoru
            </Title>
            {call.utterances?.length ? (
              <Stack gap="md">
                {call.utterances.map((u) => (
                  <Box
                    key={u.sequence}
                    style={{
                      display: "flex",
                      justifyContent: u.speaker === "agent" ? "flex-end" : "flex-start",
                    }}
                  >
                    <Paper
                      p="sm"
                      radius="md"
                      style={{
                        maxWidth: "80%",
                        backgroundColor:
                          u.speaker === "agent"
                            ? "var(--mantine-color-teal-0)"
                            : "var(--mantine-color-gray-1)",
                        borderBottomRightRadius: u.speaker === "agent" ? 4 : undefined,
                        borderBottomLeftRadius: u.speaker === "caller" ? 4 : undefined,
                      }}
                    >
                      <Group gap="xs" mb={4}>
                        <ThemeIcon
                          size="xs"
                          color={u.speaker === "agent" ? "teal" : "gray"}
                          variant="transparent"
                        >
                          {u.speaker === "agent" ? <IconRobot size={12} /> : <IconUser size={12} />}
                        </ThemeIcon>
                        <Text size="xs" fw={500} c={u.speaker === "agent" ? "teal.7" : "gray.7"}>
                          {formatSpeaker(u.speaker)}
                        </Text>
                      </Group>
                      <Text size="sm">{u.text}</Text>
                    </Paper>
                  </Box>
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
