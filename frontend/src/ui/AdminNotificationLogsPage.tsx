import { useState, useEffect, useCallback } from "react";
import {
  Button,
  Container,
  Group,
  Paper,
  Stack,
  Table,
  Text,
  Title,
  Badge,
  ScrollArea,
  Loader,
  Alert,
  Select,
} from "@mantine/core";
import { useMediaQuery } from "@mantine/hooks";
import { IconRefresh, IconAlertCircle } from "@tabler/icons-react";
import { api, NotificationLogEntry, AdminTenant } from "../api";

const channelColors: Record<string, string> = {
  sms: "blue",
  apns: "grape",
};

const typeColors: Record<string, string> = {
  trial_day10: "cyan",
  trial_day12: "teal",
  trial_expired: "orange",
  grace_warning: "yellow",
  phone_released: "red",
  call_completed: "indigo",
  usage_warning: "pink",
};

const typeLabels: Record<string, string> = {
  trial_day10: "Day 10",
  trial_day12: "Day 12",
  trial_expired: "Expired",
  grace_warning: "Grace",
  phone_released: "Released",
  call_completed: "Call",
  usage_warning: "Usage",
};

function formatTime(iso: string) {
  const d = new Date(iso);
  return d.toLocaleString("cs-CZ", {
    day: "2-digit",
    month: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

export function AdminNotificationLogsPage() {
  const isMobile = useMediaQuery("(max-width: 768px)");
  const [logs, setLogs] = useState<NotificationLogEntry[]>([]);
  const [tenants, setTenants] = useState<AdminTenant[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filters
  const [channelFilter, setChannelFilter] = useState<string | null>(null);
  const [tenantFilter, setTenantFilter] = useState<string | null>(null);

  const loadLogs = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await api.adminListNotificationLogs({
        channel: channelFilter || undefined,
        tenant_id: tenantFilter || undefined,
        limit: 200,
      });
      setLogs(data.logs || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load logs");
    } finally {
      setIsLoading(false);
    }
  }, [channelFilter, tenantFilter]);

  useEffect(() => {
    loadLogs();
  }, [loadLogs]);

  useEffect(() => {
    api.adminListTenants().then((data) => setTenants(data.tenants || []));
  }, []);

  const sentCount = logs.filter((l) => l.status === "sent").length;
  const failedCount = logs.filter((l) => l.status === "failed").length;

  return (
    <Container size="xl" py="md">
      <Stack gap="md">
        <Group justify="space-between" align="center">
          <Title order={3}>Notification Logs</Title>
          <Group gap="xs">
            <Badge color="green" variant="light" size="lg">
              {sentCount} sent
            </Badge>
            <Badge color="red" variant="light" size="lg">
              {failedCount} failed
            </Badge>
          </Group>
        </Group>

        <Paper p="sm" withBorder>
          <Group gap="sm" wrap="wrap">
            <Select
              placeholder="All channels"
              clearable
              size="sm"
              w={isMobile ? "100%" : 150}
              value={channelFilter}
              onChange={setChannelFilter}
              data={[
                { value: "sms", label: "SMS" },
                { value: "apns", label: "Push (APNs)" },
              ]}
            />
            <Select
              placeholder="All tenants"
              clearable
              searchable
              size="sm"
              w={isMobile ? "100%" : 220}
              value={tenantFilter}
              onChange={setTenantFilter}
              data={tenants.map((t) => ({ value: t.id, label: t.name }))}
            />
            <Button
              size="sm"
              variant="light"
              leftSection={<IconRefresh size={16} />}
              onClick={loadLogs}
              loading={isLoading}
            >
              Refresh
            </Button>
          </Group>
        </Paper>

        {error && (
          <Alert
            color="red"
            icon={<IconAlertCircle size={16} />}
            withCloseButton
            onClose={() => setError(null)}
          >
            {error}
          </Alert>
        )}

        {isLoading ? (
          <Group justify="center" py="xl">
            <Loader />
          </Group>
        ) : logs.length === 0 ? (
          <Paper p="xl" withBorder>
            <Text c="dimmed" ta="center">
              No notification logs found.
            </Text>
          </Paper>
        ) : (
          <Paper withBorder>
            <ScrollArea>
              <Table striped highlightOnHover>
                <Table.Thead>
                  <Table.Tr>
                    <Table.Th>Time</Table.Th>
                    <Table.Th>Channel</Table.Th>
                    <Table.Th>Type</Table.Th>
                    <Table.Th>Recipient</Table.Th>
                    {!isMobile && <Table.Th>Body</Table.Th>}
                    <Table.Th>Status</Table.Th>
                  </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                  {logs.map((log) => (
                    <Table.Tr key={log.id}>
                      <Table.Td>
                        <Text size="xs" c="dimmed" style={{ whiteSpace: "nowrap" }}>
                          {formatTime(log.created_at)}
                        </Text>
                      </Table.Td>
                      <Table.Td>
                        <Badge
                          color={channelColors[log.channel] || "gray"}
                          variant="light"
                          size="sm"
                        >
                          {log.channel.toUpperCase()}
                        </Badge>
                      </Table.Td>
                      <Table.Td>
                        <Badge
                          color={typeColors[log.notification_type] || "gray"}
                          variant="dot"
                          size="sm"
                        >
                          {typeLabels[log.notification_type] || log.notification_type}
                        </Badge>
                      </Table.Td>
                      <Table.Td>
                        <Text size="xs" ff="monospace">
                          {log.recipient}
                        </Text>
                      </Table.Td>
                      {!isMobile && (
                        <Table.Td style={{ maxWidth: 300 }}>
                          <Text size="xs" truncate="end" c="dimmed">
                            {log.body || "—"}
                          </Text>
                        </Table.Td>
                      )}
                      <Table.Td>
                        {log.status === "sent" ? (
                          <Badge color="green" variant="light" size="sm">
                            sent
                          </Badge>
                        ) : (
                          <Badge color="red" variant="light" size="sm" title={log.error_message}>
                            failed
                          </Badge>
                        )}
                      </Table.Td>
                    </Table.Tr>
                  ))}
                </Table.Tbody>
              </Table>
            </ScrollArea>
          </Paper>
        )}
      </Stack>
    </Container>
  );
}
