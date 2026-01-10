import React from "react";
import { Link, useNavigate } from "react-router-dom";
import { Alert, Badge, Box, Group, Paper, Stack, Table, Text, Title, ThemeIcon, UnstyledButton } from "@mantine/core";
import { useMediaQuery } from "@mantine/hooks";
import { IconAlertCircle, IconCheck, IconX, IconQuestionMark, IconMail, IconPhone, IconChevronRight } from "@tabler/icons-react";
import { api, CallListItem, TenantPhoneNumber } from "../api";

function getLegitimacyConfig(label: string | undefined) {
  switch (label) {
    case "legitimate":
    case "legitimní":
      return { color: "green", label: "Legitimní", icon: <IconCheck size={12} /> };
    case "marketing":
      return { color: "yellow", label: "Marketing", icon: <IconMail size={12} /> };
    case "spam":
      return { color: "red", label: "Spam", icon: <IconX size={12} /> };
    default:
      return { color: "gray", label: "Neznámé", icon: <IconQuestionMark size={12} /> };
  }
}

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

export function CallInboxPage() {
  const [calls, setCalls] = React.useState<CallListItem[] | null>(null);
  const [phoneNumbers, setPhoneNumbers] = React.useState<TenantPhoneNumber[]>([]);
  const [error, setError] = React.useState<string | null>(null);
  const isMobile = useMediaQuery("(max-width: 768px)");
  const navigate = useNavigate();

  const handleCallClick = (providerCallId: string) => {
    navigate(`/calls/${encodeURIComponent(providerCallId)}`);
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
            Karen číslo: <Text span fw={600}>{karenNumber}</Text>
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
            const legitimacy = getLegitimacyConfig(c.screening?.legitimacy_label);
            const intent = c.screening?.intent_text ?? "";
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
                  }}
                >
                  {/* Row 1: Phone number + Legitimacy badge + Chevron */}
                  <Group justify="space-between" wrap="nowrap" mb={4}>
                    <Group gap="sm" wrap="nowrap" style={{ flex: 1, minWidth: 0 }}>
                      <Text size="sm" fw={600} truncate>
                        {c.from_number}
                      </Text>
                      <Badge
                        color={legitimacy.color}
                        variant="light"
                        leftSection={legitimacy.icon}
                        size="sm"
                        style={{ flexShrink: 0 }}
                      >
                        {legitimacy.label}
                      </Badge>
                    </Group>
                    <IconChevronRight size={20} color="gray" style={{ flexShrink: 0 }} />
                  </Group>

                  {/* Row 2: to_number + relative time */}
                  <Text size="xs" c="dimmed" mb="xs">
                    na {c.to_number} | {formatRelativeTime(new Date(c.started_at))}
                  </Text>

                  {/* Row 3: Intent text */}
                  {intent && (
                    <Text size="sm" lineClamp={2} mb="xs">
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
                <Table.Th>Čas</Table.Th>
                <Table.Th>Od</Table.Th>
                <Table.Th>Hodnocení</Table.Th>
                <Table.Th>Účel</Table.Th>
                <Table.Th>Stav</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {calls.map((c) => {
                const legitimacy = getLegitimacyConfig(c.screening?.legitimacy_label);
                const intent = c.screening?.intent_text ?? "";
                return (
                  <Table.Tr
                    key={c.provider_call_id}
                    style={{
                      borderLeft: `4px solid var(--mantine-color-${legitimacy.color}-5)`,
                    }}
                  >
                    <Table.Td>
                      <Text size="sm">{formatRelativeTime(new Date(c.started_at))}</Text>
                      <Text size="xs" c="dimmed">
                        {new Date(c.started_at).toLocaleTimeString("cs-CZ", { hour: "2-digit", minute: "2-digit" })}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm" fw={600}>
                        <Link
                          to={`/calls/${encodeURIComponent(c.provider_call_id)}`}
                          style={{ color: "inherit", textDecoration: "none" }}
                        >
                          {c.from_number}
                        </Link>
                      </Text>
                      <Text size="xs" c="dimmed">
                        na {c.to_number}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Badge
                        color={legitimacy.color}
                        variant="light"
                        leftSection={legitimacy.icon}
                      >
                        {legitimacy.label}
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
