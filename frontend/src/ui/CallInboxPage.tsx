import React from "react";
import { useNavigate } from "react-router-dom";
import {
  Alert,
  Badge,
  Box,
  Group,
  Paper,
  Stack,
  Table,
  Text,
  Title,
  ThemeIcon,
  UnstyledButton,
  ActionIcon,
  Tooltip,
} from "@mantine/core";
import { useMediaQuery } from "@mantine/hooks";
import {
  IconAlertCircle,
  IconPhone,
  IconChevronRight,
  IconCircleFilled,
  IconCircleCheck,
  IconCircle,
} from "@tabler/icons-react";
import { api, CallListItem, TenantPhoneNumber } from "../api";
import { getLegitimacyConfig, getLeadLabelConfig } from "./callLabels";

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

function formatRelativeTime(date: Date): string {
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return "Právě teď";
  if (diffMins < 60) return `Před ${diffMins} min`;
  if (diffHours < 24) return `Před ${diffHours} h`;
  if (diffDays < 7) return `Před ${diffDays} dny`;
  return date.toLocaleDateString("cs-CZ");
}

// Call resolution status
type ResolutionStatus = "new" | "viewed" | "resolved";

function getResolutionStatus(call: CallListItem): ResolutionStatus {
  if (call.resolved_at) return "resolved";
  if (call.first_viewed_at) return "viewed";
  return "new";
}

export function CallInboxPage() {
  const [calls, setCalls] = React.useState<CallListItem[] | null>(null);
  const [phoneNumbers, setPhoneNumbers] = React.useState<TenantPhoneNumber[]>([]);
  const [error, setError] = React.useState<string | null>(null);
  const isMobile = useMediaQuery("(max-width: 768px)");
  const navigate = useNavigate();

  const handleCallClick = (providerCallId: string) => {
    navigate(`/calls/${encodeURIComponent(providerCallId)}`);
  };

  const handleToggleResolved = async (
    e: React.MouseEvent,
    providerCallId: string,
    currentlyResolved: boolean
  ) => {
    e.stopPropagation();

    // Optimistic update
    setCalls((prev) =>
      prev
        ? prev.map((c) =>
            c.provider_call_id === providerCallId
              ? {
                  ...c,
                  resolved_at: currentlyResolved ? null : new Date().toISOString(),
                }
              : c
          )
        : prev
    );

    try {
      if (currentlyResolved) {
        await api.markCallUnresolved(providerCallId);
      } else {
        await api.markCallResolved(providerCallId);
      }
    } catch {
      // Revert on error
      setCalls((prev) =>
        prev
          ? prev.map((c) =>
              c.provider_call_id === providerCallId
                ? {
                    ...c,
                    resolved_at: currentlyResolved ? new Date().toISOString() : null,
                  }
                : c
            )
          : prev
      );
    }
  };

  React.useEffect(() => {
    api
      .listCalls()
      .then(setCalls)
      .catch((e) => setError(String(e)));

    api
      .getTenant()
      .then((data) => setPhoneNumbers(data.phone_numbers || []))
      .catch(() => {});
  }, []);

  const hasPhoneNumber = phoneNumbers.some((p) => p.is_primary);
  const karenNumber = phoneNumbers.find((p) => p.is_primary)?.twilio_number;

  return (
    <Stack gap="md" py="md">
      <Group justify="space-between">
        <Title order={2}>Příchozí hovory</Title>
        {hasPhoneNumber && (
          <Text size="sm" c="dimmed">
            Karen číslo:{" "}
            <Text span fw={600}>
              {karenNumber}
            </Text>
          </Text>
        )}
      </Group>

      {!hasPhoneNumber && calls !== null && (
        <Alert icon={<IconAlertCircle size={16} />} color="yellow" variant="light">
          Zatím ti nebylo přiřazeno telefonní číslo. Jakmile bude dostupné, budeme tě informovat.
        </Alert>
      )}

      {error && (
        <Paper p="md" withBorder>
          <Text c="red">Chyba: {error}</Text>
        </Paper>
      )}

      {!calls && !error && <Text c="dimmed">Načítání…</Text>}

      {calls && calls.length === 0 && (
        <Paper p="xl" withBorder ta="center" radius="md">
          <ThemeIcon size={60} radius="xl" variant="light" color="gray" mb="md" mx="auto">
            <IconPhone size={30} />
          </ThemeIcon>
          <Text c="dimmed" size="lg">
            Zatím žádné hovory
          </Text>
          <Text c="dimmed" size="sm" mt="xs">
            Jakmile někdo zavolá na tvoje Karen číslo, uvidíš hovor zde.
          </Text>
        </Paper>
      )}

      {/* Mobile card layout */}
      {calls && calls.length > 0 && isMobile && (
        <Stack gap="sm">
          {calls.map((c) => {
            const legitimacy = getLegitimacyConfig(c.screening?.legitimacy_label, 12);
            const lead = getLeadLabelConfig(c.screening?.lead_label, 12);
            const intent = c.screening?.intent_text ?? "";
            const resolutionStatus = getResolutionStatus(c);
            const isResolved = resolutionStatus === "resolved";
            const isNew = resolutionStatus === "new";

            return (
              <UnstyledButton
                key={c.provider_call_id}
                onClick={() => handleCallClick(c.provider_call_id)}
                aria-label={`Zobrazit hovor od ${c.from_number}`}
                style={{ width: "100%" }}
              >
                <Paper
                  p="md"
                  radius="md"
                  withBorder
                  style={{
                    borderLeft: `4px solid var(--mantine-color-${legitimacy.color}-5)`,
                    cursor: "pointer",
                    transition: "background-color 0.15s",
                    opacity: isResolved ? 0.7 : 1,
                  }}
                >
                  {/* Row 1: Resolution icon + Phone number + Lead badge + Chevron */}
                  <Group justify="space-between" wrap="nowrap" mb={4}>
                    <Group gap="sm" wrap="nowrap" style={{ flex: 1, minWidth: 0 }}>
                      {/* Resolution status indicator */}
                      <Tooltip
                        label={
                          isResolved
                            ? "Vyřešeno - kliknutím označíte jako nevyřešené"
                            : "Označit jako vyřešené"
                        }
                      >
                        <ActionIcon
                          variant="subtle"
                          size="md"
                          color={isResolved ? "teal" : isNew ? "blue" : "gray"}
                          onClick={(e) => handleToggleResolved(e, c.provider_call_id, isResolved)}
                        >
                          {isResolved ? (
                            <IconCircleCheck size={20} />
                          ) : isNew ? (
                            <IconCircleFilled size={12} />
                          ) : (
                            <IconCircle size={20} />
                          )}
                        </ActionIcon>
                      </Tooltip>
                      <Text size="sm" fw={isNew ? 700 : 600} truncate style={{ flex: 1, minWidth: 0 }}>
                        {c.from_number}
                      </Text>
                      <Badge
                        color={lead.color}
                        variant="light"
                        leftSection={lead.icon}
                        size="sm"
                        style={{ flexShrink: 0 }}
                      >
                        {lead.label}
                      </Badge>
                    </Group>
                    <IconChevronRight size={20} color="gray" style={{ flexShrink: 0 }} />
                  </Group>

                  {/* Row 2: to_number + relative time + legitimacy label */}
                  <Group gap="xs" mb="xs" pl={30}>
                    <Text size="xs" c="dimmed">
                      na {c.to_number} | {formatRelativeTime(new Date(c.started_at))}
                    </Text>
                    <Text size="xs" c={legitimacy.color} fw={500}>
                      • {legitimacy.label}
                    </Text>
                  </Group>

                  {/* Row 3: Intent text */}
                  {intent && (
                    <Text size="sm" lineClamp={2} mb="xs" pl={30}>
                      "{intent}"
                    </Text>
                  )}

                  {/* Row 4: Status badge */}
                  <Box ta="right">
                    <Badge variant="light" color="gray" size="sm">
                      {formatStatus(c.status)}
                    </Badge>
                  </Box>
                </Paper>
              </UnstyledButton>
            );
          })}
        </Stack>
      )}

      {/* Desktop table layout */}
      {calls && calls.length > 0 && !isMobile && (
        <Paper withBorder radius="md" style={{ overflow: "hidden" }}>
          <Table striped highlightOnHover>
            <Table.Thead>
              <Table.Tr>
                <Table.Th style={{ width: 40 }}></Table.Th>
                <Table.Th>Čas</Table.Th>
                <Table.Th>Od</Table.Th>
                <Table.Th>Hodnocení</Table.Th>
                <Table.Th>Lead</Table.Th>
                <Table.Th>Účel</Table.Th>
                <Table.Th>Stav</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {calls.map((c) => {
                const legitimacy = getLegitimacyConfig(c.screening?.legitimacy_label);
                const lead = getLeadLabelConfig(c.screening?.lead_label);
                const intent = c.screening?.intent_text ?? "";
                const resolutionStatus = getResolutionStatus(c);
                const isResolved = resolutionStatus === "resolved";
                const isNew = resolutionStatus === "new";

                return (
                  <Table.Tr
                    key={c.provider_call_id}
                    onClick={() => handleCallClick(c.provider_call_id)}
                    style={{
                      borderLeft: `4px solid var(--mantine-color-${legitimacy.color}-5)`,
                      cursor: "pointer",
                      opacity: isResolved ? 0.7 : 1,
                    }}
                  >
                    <Table.Td>
                      <Tooltip
                        label={
                          isResolved
                            ? "Vyřešeno - kliknutím označíte jako nevyřešené"
                            : "Označit jako vyřešené"
                        }
                      >
                        <ActionIcon
                          variant="subtle"
                          size="md"
                          color={isResolved ? "teal" : isNew ? "blue" : "gray"}
                          onClick={(e) => handleToggleResolved(e, c.provider_call_id, isResolved)}
                        >
                          {isResolved ? (
                            <IconCircleCheck size={20} />
                          ) : isNew ? (
                            <IconCircleFilled size={12} />
                          ) : (
                            <IconCircle size={20} />
                          )}
                        </ActionIcon>
                      </Tooltip>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm" fw={isNew ? 600 : 400}>
                        {formatRelativeTime(new Date(c.started_at))}
                      </Text>
                      <Text size="xs" c="dimmed">
                        {new Date(c.started_at).toLocaleTimeString("cs-CZ", {
                          hour: "2-digit",
                          minute: "2-digit",
                        })}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm" fw={isNew ? 700 : 600}>
                        {c.from_number}
                      </Text>
                      <Text size="xs" c="dimmed">
                        na {c.to_number}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Badge color={legitimacy.color} variant="light" leftSection={legitimacy.icon}>
                        {legitimacy.label}
                      </Badge>
                    </Table.Td>
                    <Table.Td>
                      <Badge
                        color={lead.color}
                        variant="light"
                        leftSection={lead.icon}
                      >
                        {lead.label}
                      </Badge>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm" lineClamp={2} maw={200}>
                        {intent || "—"}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Badge variant="light" color="gray" size="sm">
                        {formatStatus(c.status)}
                      </Badge>
                    </Table.Td>
                  </Table.Tr>
                );
              })}
            </Table.Tbody>
          </Table>
        </Paper>
      )}
    </Stack>
  );
}
