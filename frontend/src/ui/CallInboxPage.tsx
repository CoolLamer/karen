import React from "react";
import { useNavigate } from "react-router-dom";
import {
  Alert,
  Badge,
  Box,
  Group,
  Paper,
  Stack,
  Switch,
  Table,
  Text,
  Title,
  ThemeIcon,
  UnstyledButton,
  ActionIcon,
  Tooltip,
  Progress,
  SimpleGrid,
} from "@mantine/core";
import { useMediaQuery } from "@mantine/hooks";
import {
  IconAlertCircle,
  IconPhone,
  IconChevronRight,
  IconCircleFilled,
  IconCircleCheck,
  IconCircle,
  IconClock,
  IconPhoneOff,
} from "@tabler/icons-react";
import { api, CallListItem, TenantPhoneNumber, BillingInfo } from "../api";
import { getLegitimacyConfig, getLeadLabelConfig } from "./callLabels";

// Format time saved in a human-readable format (Czech)
function formatTimeSaved(seconds: number): string {
  if (seconds < 60) return `${seconds} sekund`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes} minut`;
  const hours = Math.floor(minutes / 60);
  const remainingMins = minutes % 60;
  if (remainingMins === 0) return `${hours}h`;
  return `${hours}h ${remainingMins}min`;
}

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
  const [billing, setBilling] = React.useState<BillingInfo | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [hideResolved, setHideResolved] = React.useState(() => {
    return localStorage.getItem("inbox.hideResolved") === "true";
  });
  const isMobile = useMediaQuery("(max-width: 768px)");
  const navigate = useNavigate();

  const handleHideResolvedChange = (checked: boolean) => {
    setHideResolved(checked);
    localStorage.setItem("inbox.hideResolved", String(checked));
  };

  const filteredCalls = React.useMemo(() => {
    if (!calls) return null;
    return hideResolved ? calls.filter((c) => !c.resolved_at) : calls;
  }, [calls, hideResolved]);

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

    api
      .getBilling()
      .then(setBilling)
      .catch(() => {});
  }, []);

  const hasPhoneNumber = phoneNumbers.some((p) => p.is_primary);
  const karenNumber = phoneNumbers.find((p) => p.is_primary)?.twilio_number;

  return (
    <Stack gap="md" py="md">
      <Group justify="space-between" wrap="wrap">
        <Title order={2}>Příchozí hovory</Title>
        <Group gap="md">
          <Switch
            label="Skrýt vyřešené"
            checked={hideResolved}
            onChange={(e) => handleHideResolvedChange(e.currentTarget.checked)}
            size="sm"
          />
          {hasPhoneNumber && (
            <Text size="sm" c="dimmed">
              Karen číslo:{" "}
              <Text span fw={600}>
                {karenNumber}
              </Text>
            </Text>
          )}
        </Group>
      </Group>

      {/* Time Saved Widget + Trial Status */}
      {billing && (
        <SimpleGrid cols={{ base: 1, sm: 2 }} spacing="md">
          {/* Time Saved Card */}
          <Paper p="md" radius="md" withBorder>
            <Group gap="md" wrap="nowrap">
              <ThemeIcon size={48} radius="xl" variant="light" color="teal">
                <IconClock size={24} />
              </ThemeIcon>
              <Box>
                <Text size="sm" c="dimmed">
                  Karen ti ušetřila
                </Text>
                <Text size="xl" fw={700} c="teal">
                  {formatTimeSaved(billing.current_usage?.time_saved_seconds ?? 0)}
                </Text>
                <Text size="xs" c="dimmed">
                  tento měsíc ({billing.current_usage?.calls_count ?? 0} hovorů
                  {billing.current_usage?.spam_calls_blocked
                    ? `, ${billing.current_usage.spam_calls_blocked} spam blokováno`
                    : ""}
                  )
                </Text>
              </Box>
            </Group>
            {billing.total_time_saved > 0 && billing.total_time_saved !== billing.current_usage?.time_saved_seconds && (
              <Text size="xs" c="dimmed" mt="xs" ta="right">
                Celkem od začátku: {formatTimeSaved(billing.total_time_saved)}
              </Text>
            )}
          </Paper>

          {/* Trial Status Card (only for trial plan) */}
          {billing.plan === "trial" && (
            <Paper p="md" radius="md" withBorder>
              <Group gap="md" wrap="nowrap">
                <ThemeIcon
                  size={48}
                  radius="xl"
                  variant="light"
                  color={billing.call_status.can_receive ? "blue" : "red"}
                >
                  {billing.call_status.can_receive ? (
                    <IconPhone size={24} />
                  ) : (
                    <IconPhoneOff size={24} />
                  )}
                </ThemeIcon>
                <Box style={{ flex: 1 }}>
                  <Text size="sm" c="dimmed">
                    Trial status
                  </Text>
                  {billing.call_status.can_receive ? (
                    <>
                      <Text size="lg" fw={600}>
                        {billing.call_status.trial_calls_left} hovorů zbývá
                      </Text>
                      <Text size="xs" c="dimmed">
                        {billing.call_status.trial_days_left} dní do konce trialu
                      </Text>
                    </>
                  ) : (
                    <>
                      <Text size="lg" fw={600} c="red">
                        Trial vypršel
                      </Text>
                      <Text size="xs" c="dimmed">
                        {billing.call_status.reason === "limit_exceeded"
                          ? "Dosáhli jste limitu hovorů"
                          : "Trial skončil"}
                      </Text>
                    </>
                  )}
                  {/* Progress bar showing calls used */}
                  {billing.call_status.calls_limit > 0 && (
                    <Progress
                      value={
                        (billing.call_status.calls_used / billing.call_status.calls_limit) * 100
                      }
                      size="sm"
                      color={
                        billing.call_status.calls_used >= billing.call_status.calls_limit
                          ? "red"
                          : billing.call_status.calls_used >= billing.call_status.calls_limit * 0.8
                          ? "yellow"
                          : "blue"
                      }
                      mt="xs"
                    />
                  )}
                </Box>
              </Group>
            </Paper>
          )}

          {/* Basic/Pro plan - just show usage if there's a limit */}
          {billing.plan === "basic" && (
            <Paper p="md" radius="md" withBorder>
              <Group gap="md" wrap="nowrap">
                <ThemeIcon size={48} radius="xl" variant="light" color="blue">
                  <IconPhone size={24} />
                </ThemeIcon>
                <Box style={{ flex: 1 }}>
                  <Text size="sm" c="dimmed">
                    Využití tohoto měsíce
                  </Text>
                  <Text size="lg" fw={600}>
                    {billing.call_status.calls_used} / {billing.call_status.calls_limit} hovorů
                  </Text>
                  <Progress
                    value={
                      (billing.call_status.calls_used / billing.call_status.calls_limit) * 100
                    }
                    size="sm"
                    color={
                      billing.call_status.calls_used >= billing.call_status.calls_limit
                        ? "red"
                        : billing.call_status.calls_used >= billing.call_status.calls_limit * 0.8
                        ? "yellow"
                        : "blue"
                    }
                    mt="xs"
                  />
                </Box>
              </Group>
            </Paper>
          )}
        </SimpleGrid>
      )}

      {/* Trial expired alert */}
      {billing && !billing.call_status.can_receive && (
        <Alert icon={<IconAlertCircle size={16} />} color="red" variant="light">
          {billing.call_status.reason === "trial_expired"
            ? "Tvůj trial vypršel. Karen nebude přijímat hovory. Upgraduj na zvednu.cz"
            : "Dosáhli jste limitu hovorů. Karen nebude přijímat další hovory. Upgraduj na zvednu.cz"}
        </Alert>
      )}

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

      {filteredCalls && filteredCalls.length === 0 && calls && calls.length > 0 && (
        <Paper p="xl" withBorder ta="center" radius="md">
          <ThemeIcon size={60} radius="xl" variant="light" color="teal" mb="md" mx="auto">
            <IconPhone size={30} />
          </ThemeIcon>
          <Text c="dimmed" size="lg">
            Všechny hovory jsou vyřešené
          </Text>
          <Text c="dimmed" size="sm" mt="xs">
            Vypni filtr „Skrýt vyřešené" pro zobrazení všech hovorů.
          </Text>
        </Paper>
      )}

      {/* Mobile card layout */}
      {filteredCalls && filteredCalls.length > 0 && isMobile && (
        <Stack gap="sm">
          {filteredCalls.map((c) => {
            const isRejected = c.status === "rejected_limit";
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
                    borderLeft: isRejected
                      ? "4px solid var(--mantine-color-orange-5)"
                      : `4px solid var(--mantine-color-${legitimacy.color}-5)`,
                    cursor: "pointer",
                    transition: "background-color 0.15s",
                    opacity: isResolved ? 0.7 : isRejected ? 0.85 : 1,
                    backgroundColor: isRejected ? "var(--mantine-color-gray-0)" : undefined,
                  }}
                >
                  {/* Row 1: Resolution icon + Phone number + Lead/Rejected badge + Chevron */}
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
                          color={isRejected ? "orange" : isResolved ? "teal" : isNew ? "blue" : "gray"}
                          onClick={(e) => handleToggleResolved(e, c.provider_call_id, isResolved)}
                        >
                          {isRejected ? (
                            <IconPhoneOff size={20} />
                          ) : isResolved ? (
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
                      {isRejected ? (
                        <Badge
                          color="orange"
                          variant="light"
                          leftSection={<IconPhoneOff size={12} />}
                          size="sm"
                          style={{ flexShrink: 0 }}
                        >
                          {formatStatus(c.status, c.rejection_reason)}
                        </Badge>
                      ) : (
                        <Badge
                          color={lead.color}
                          variant="light"
                          leftSection={lead.icon}
                          size="sm"
                          style={{ flexShrink: 0 }}
                        >
                          {lead.label}
                        </Badge>
                      )}
                    </Group>
                    <IconChevronRight size={20} color="gray" style={{ flexShrink: 0 }} />
                  </Group>

                  {/* Row 2: to_number + relative time + legitimacy label (or nothing for rejected) */}
                  <Group gap="xs" mb="xs" pl={30}>
                    <Text size="xs" c="dimmed">
                      na {c.to_number} | {formatRelativeTime(new Date(c.started_at))}
                    </Text>
                    {!isRejected && (
                      <Text size="xs" c={legitimacy.color} fw={500}>
                        • {legitimacy.label}
                      </Text>
                    )}
                  </Group>

                  {/* Row 3: Intent text or rejection explanation */}
                  {isRejected ? (
                    <Text size="sm" c="dimmed" mb="xs" pl={30}>
                      {getRejectionExplanation(c.rejection_reason)}
                    </Text>
                  ) : intent ? (
                    <Text size="sm" lineClamp={2} mb="xs" pl={30}>
                      "{intent}"
                    </Text>
                  ) : null}

                  {/* Row 4: Status badge (hide for rejected calls since badge is shown at top) */}
                  {!isRejected && (
                    <Box ta="right">
                      <Badge variant="light" color="gray" size="sm">
                        {formatStatus(c.status, c.rejection_reason)}
                      </Badge>
                    </Box>
                  )}
                </Paper>
              </UnstyledButton>
            );
          })}
        </Stack>
      )}

      {/* Desktop table layout */}
      {filteredCalls && filteredCalls.length > 0 && !isMobile && (
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
              {filteredCalls.map((c) => {
                const isRejected = c.status === "rejected_limit";
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
                      borderLeft: isRejected
                        ? "4px solid var(--mantine-color-orange-5)"
                        : `4px solid var(--mantine-color-${legitimacy.color}-5)`,
                      cursor: "pointer",
                      opacity: isResolved ? 0.7 : isRejected ? 0.85 : 1,
                      backgroundColor: isRejected ? "var(--mantine-color-gray-0)" : undefined,
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
                          color={isRejected ? "orange" : isResolved ? "teal" : isNew ? "blue" : "gray"}
                          onClick={(e) => handleToggleResolved(e, c.provider_call_id, isResolved)}
                        >
                          {isRejected ? (
                            <IconPhoneOff size={20} />
                          ) : isResolved ? (
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
                      {isRejected ? (
                        <Text size="sm" c="dimmed">—</Text>
                      ) : (
                        <Badge color={legitimacy.color} variant="light" leftSection={legitimacy.icon}>
                          {legitimacy.label}
                        </Badge>
                      )}
                    </Table.Td>
                    <Table.Td>
                      {isRejected ? (
                        <Text size="sm" c="dimmed">—</Text>
                      ) : (
                        <Badge
                          color={lead.color}
                          variant="light"
                          leftSection={lead.icon}
                        >
                          {lead.label}
                        </Badge>
                      )}
                    </Table.Td>
                    <Table.Td>
                      {isRejected ? (
                        <Text size="sm" c="dimmed" lineClamp={2} maw={200}>
                          {getRejectionExplanation(c.rejection_reason)}
                        </Text>
                      ) : (
                        <Text size="sm" lineClamp={2} maw={200}>
                          {intent || "—"}
                        </Text>
                      )}
                    </Table.Td>
                    <Table.Td>
                      <Badge variant="light" color={isRejected ? "orange" : "gray"} size="sm">
                        {formatStatus(c.status, c.rejection_reason)}
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
