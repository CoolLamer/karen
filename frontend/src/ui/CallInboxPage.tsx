import React from "react";
import { Link } from "react-router-dom";
import { Badge, Group, Paper, Stack, Table, Text, Title } from "@mantine/core";
import { api, CallListItem } from "../api";

function labelColor(label: string | undefined) {
  switch (label) {
    case "legitimate":
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

export function CallInboxPage() {
  const [calls, setCalls] = React.useState<CallListItem[] | null>(null);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    api
      .listCalls()
      .then(setCalls)
      .catch((e) => setError(String(e)));
  }, []);

  return (
    <Stack gap="md" py="md">
      <Group justify="space-between">
        <Title order={2}>Call Inbox</Title>
      </Group>

      {error && (
        <Paper p="md" withBorder>
          <Text c="red">Error: {error}</Text>
        </Paper>
      )}

      {!calls && !error && <Text c="dimmed">Loading…</Text>}

      {calls && (
        <Paper withBorder>
          <Table striped highlightOnHover>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Time</Table.Th>
                <Table.Th>From</Table.Th>
                <Table.Th>Label</Table.Th>
                <Table.Th>Intent</Table.Th>
                <Table.Th>Status</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {calls.map((c) => {
                const label = c.screening?.legitimacy_label ?? "unknown";
                const intent = c.screening?.intent_text ?? "";
                return (
                  <Table.Tr key={c.provider_call_id}>
                    <Table.Td>
                      <Text size="sm">{new Date(c.started_at).toLocaleString()}</Text>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm" fw={600}>
                        <Link to={`/calls/${encodeURIComponent(c.provider_call_id)}`}>{c.from_number}</Link>
                      </Text>
                      <Text size="xs" c="dimmed">
                        to {c.to_number}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Badge color={labelColor(label)} variant="light">
                        {label}
                      </Badge>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm" lineClamp={2}>
                        {intent || "—"}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm">{c.status}</Text>
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


