import React from "react";
import { useParams, useNavigate } from "react-router-dom";
import {
  Alert,
  Badge,
  Group,
  Paper,
  Stack,
  Text,
  Title,
  Box,
  ThemeIcon,
  Button,
  Progress,
  Anchor,
} from "@mantine/core";
import {
  IconAlertCircle,
  IconArrowLeft,
  IconRobot,
  IconUser,
  IconCircleCheck,
  IconCircle,
  IconPhoneOff,
} from "@tabler/icons-react";
import { api, CallDetail } from "../api";
import { getLegitimacyConfig, getLeadLabelConfig } from "./callLabels";

function formatStatus(status: string, rejectionReason?: string | null) {
  switch (status) {
    case "in_progress":
      return "Probíhá";
    case "completed":
      return "Dokončeno";
    case "queued":
      return "Čeká";
    case "ringing":
      return "Vyzvání";
    case "rejected_limit":
      // Specific reason messaging
      switch (rejectionReason) {
        case "trial_expired":
          return "Trial vypršel";
        case "limit_exceeded":
          return "Limit dosažen";
        case "subscription_cancelled":
          return "Předplatné zrušeno";
        case "subscription_suspended":
          return "Předplatné pozastaveno";
        default:
          return "Nepřijato";
      }
    default:
      return status;
  }
}

// Get explanation text for rejected calls
function getRejectionExplanation(rejectionReason?: string | null): string {
  switch (rejectionReason) {
    case "trial_expired":
      return "Asistentka nemohla přijmout hovor - váš trial skončil";
    case "limit_exceeded":
      return "Asistentka nemohla přijmout hovor - dosáhli jste měsíčního limitu";
    case "subscription_cancelled":
      return "Asistentka nemohla přijmout hovor - předplatné bylo zrušeno";
    case "subscription_suspended":
      return "Asistentka nemohla přijmout hovor - předplatné bylo pozastaveno";
    default:
      return "Asistentka nemohla přijmout tento hovor";
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

function formatCallTermination(endedBy: string | null | undefined) {
  if (!endedBy) return null;

  switch (endedBy) {
    case "agent":
      return "Zavěsila asistentka";
    case "caller":
      return "Zavěsil volající";
    default:
      return null;
  }
}

export function CallDetailPage() {
  const { providerCallId } = useParams();
  const navigate = useNavigate();
  const [call, setCall] = React.useState<CallDetail | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [isResolved, setIsResolved] = React.useState(false);
  const [isTogglingResolved, setIsTogglingResolved] = React.useState(false);

  React.useEffect(() => {
    if (!providerCallId) return;

    // Fetch call detail
    api
      .getCall(providerCallId)
      .then((data) => {
        setCall(data);
        setIsResolved(!!data.resolved_at);
      })
      .catch((e) => setError(String(e)));

    // Mark as viewed (fire and forget)
    api.markCallViewed(providerCallId).catch(() => {});
  }, [providerCallId]);

  const handleToggleResolved = async () => {
    if (!providerCallId || isTogglingResolved) return;

    setIsTogglingResolved(true);
    const wasResolved = isResolved;

    // Optimistic update
    setIsResolved(!wasResolved);

    try {
      if (wasResolved) {
        await api.markCallUnresolved(providerCallId);
      } else {
        await api.markCallResolved(providerCallId);
      }
    } catch {
      // Revert on error
      setIsResolved(wasResolved);
    } finally {
      setIsTogglingResolved(false);
    }
  };

  const legitimacyConfig = getLegitimacyConfig(call?.screening?.legitimacy_label ?? "unknown");
  const leadConfig = getLeadLabelConfig(call?.screening?.lead_label);

  return (
    <Stack gap="md" py="md">
      <Group justify="space-between">
        <Button
          variant="subtle"
          leftSection={<IconArrowLeft size={16} />}
          onClick={() => navigate("/inbox")}
          px={0}
        >
          Zpět na hovory
        </Button>

        {call && (
          <Button
            variant={isResolved ? "light" : "filled"}
            color={isResolved ? "gray" : "teal"}
            leftSection={isResolved ? <IconCircleCheck size={18} /> : <IconCircle size={18} />}
            onClick={handleToggleResolved}
            loading={isTogglingResolved}
          >
            {isResolved ? "Vyřešeno" : "Označit jako vyřešené"}
          </Button>
        )}
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
                <Text size="xl" fw={700}>
                  {call.from_number}
                </Text>
                <Text size="sm" c="dimmed">
                  na {call.to_number}
                </Text>
                <Text size="sm" mt="xs">
                  {new Date(call.started_at).toLocaleString("cs-CZ")}
                </Text>
                <Group gap="xs" mt="xs">
                  <Badge
                    variant="light"
                    color={call.status === "rejected_limit" ? "orange" : "gray"}
                    size="sm"
                    leftSection={call.status === "rejected_limit" ? <IconPhoneOff size={12} /> : undefined}
                  >
                    {formatStatus(call.status, call.rejection_reason)}
                  </Badge>
                  {isResolved && (
                    <Badge variant="light" color="teal" size="sm" leftSection={<IconCircleCheck size={12} />}>
                      Vyřešeno
                    </Badge>
                  )}
                </Group>
              </Stack>
              {/* Only show legitimacy/lead badges for non-rejected calls */}
              {call.status !== "rejected_limit" && (
                <Stack gap="xs" align="flex-end">
                  <Group gap="xs">
                    <Badge
                      variant="light"
                      color={legitimacyConfig.color}
                      size="lg"
                      leftSection={legitimacyConfig.icon}
                    >
                      {legitimacyConfig.label}
                    </Badge>
                    <Badge
                      variant="light"
                      color={leadConfig.color}
                      size="lg"
                      leftSection={leadConfig.icon}
                    >
                      {leadConfig.label}
                    </Badge>
                  </Group>
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
              )}
            </Group>
            {/* Intent for non-rejected calls */}
            {call.status !== "rejected_limit" && call.screening?.intent_text && (
              <Paper p="sm" radius="sm" bg="gray.0" mt="md">
                <Text size="sm" fw={500}>
                  Účel hovoru:
                </Text>
                <Text size="sm" c="dimmed">
                  {call.screening.intent_text}
                </Text>
              </Paper>
            )}
          </Paper>

          {/* Rejection explanation alert */}
          {call.status === "rejected_limit" && (
            <Alert
              icon={<IconPhoneOff size={16} />}
              title={formatStatus(call.status, call.rejection_reason)}
              color="orange"
              variant="light"
            >
              <Text size="sm">
                {getRejectionExplanation(call.rejection_reason)}
              </Text>
              <Text size="sm" mt="xs">
                Pro obnovení služby <Anchor href="/settings" onClick={(e) => { e.preventDefault(); navigate("/settings"); }}>upgradujte svůj plán</Anchor>.
              </Text>
            </Alert>
          )}

          {/* Transcript with chat bubbles (not shown for rejected calls) */}
          {call.status !== "rejected_limit" && (
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
                  {formatCallTermination(call.ended_by) && (
                    <Text size="sm" c="dimmed" mt="xs" fs="italic">
                      — {formatCallTermination(call.ended_by)} —
                    </Text>
                  )}
                </Stack>
              ) : (
                <Text size="sm" c="dimmed">
                  Přepis zatím není k dispozici.
                </Text>
              )}
            </Paper>
          )}
        </>
      )}
    </Stack>
  );
}
