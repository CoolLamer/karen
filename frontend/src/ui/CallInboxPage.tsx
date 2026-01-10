import React from "react";
import { Link } from "react-router-dom";
import { Alert, Badge, Group, Paper, Stack, Table, Text, Title } from "@mantine/core";
import { IconAlertCircle } from "@tabler/icons-react";
import { api, CallListItem, TenantPhoneNumber } from "../api";

function labelColor(label: string | undefined) {
  switch (label) {
    case "legitimate":
    case "legitimní":
      return "green";
    case "marketing":
      return "yellow";
    case "spam":
      return "red";
    case "unknown":
      return "gray";
    default:
      return "gray";
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

export function CallInboxPage() {
  const [calls, setCalls] = React.useState<CallListItem[] | null>(null);
  const [phoneNumbers, setPhoneNumbers] = React.useState<TenantPhoneNumber[]>([]);
  const [error, setError] = React.useState<string | null>(null);

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
        <Paper p="xl" withBorder ta="center">
          <Text c="dimmed" size="lg">
            Zatím žádné hovory
          </Text>
          <Text c="dimmed" size="sm" mt="xs">
            Jakmile někdo zavolá na tvoje Karen číslo, uvidíš hovor zde.
          </Text>
        </Paper>
      )}

      {calls && calls.length > 0 && (
        <Paper withBorder>
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
                const label = c.screening?.legitimacy_label ?? "unknown";
                const intent = c.screening?.intent_text ?? "";
                return (
                  <Table.Tr key={c.provider_call_id}>
                    <Table.Td>
                      <Text size="sm">{new Date(c.started_at).toLocaleString("cs-CZ")}</Text>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm" fw={600}>
                        <Link to={`/calls/${encodeURIComponent(c.provider_call_id)}`}>{c.from_number}</Link>
                      </Text>
                      <Text size="xs" c="dimmed">
                        na {c.to_number}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Badge color={labelColor(label)} variant="light">
                        {formatLabel(label)}
                      </Badge>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm" lineClamp={2}>
                        {intent || "—"}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm">{formatStatus(c.status)}</Text>
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
