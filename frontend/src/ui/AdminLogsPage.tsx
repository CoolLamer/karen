import React, { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
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
  Code,
  Loader,
  Alert,
} from "@mantine/core";
import { useMediaQuery } from "@mantine/hooks";
import { IconRefresh, IconAlertCircle } from "@tabler/icons-react";
import { api, CallListItem, CallEvent } from "../api";

const eventTypeColors: Record<string, string> = {
  call_started: "blue",
  stt_result: "gray",
  turn_finalized: "green",
  barge_in: "orange",
  filler_spoken: "cyan",
  filler_skipped: "gray",
  llm_started: "violet",
  llm_completed: "violet",
  llm_error: "red",
  tts_started: "pink",
  tts_completed: "pink",
  tts_error: "red",
  audio_mark: "gray",
  goodbye_detected: "yellow",
  forward_detected: "yellow",
  call_forwarded: "teal",
  call_hangup: "red",
  call_ended: "red",
  marketing_call: "orange",
};

function formatTime(ts: string): string {
  const d = new Date(ts);
  return d.toLocaleTimeString();
}

function formatDateTime(ts: string): string {
  const d = new Date(ts);
  return d.toLocaleString();
}

export function AdminLogsPage() {
  const navigate = useNavigate();
  const { providerCallId } = useParams<{ providerCallId?: string }>();
  const isMobile = useMediaQuery("(max-width: 768px)");

  const [calls, setCalls] = useState<CallListItem[]>([]);
  const [selectedCall, setSelectedCall] = useState<string | null>(providerCallId || null);
  const [events, setEvents] = useState<CallEvent[]>([]);
  const [isLoadingCalls, setIsLoadingCalls] = useState(true);
  const [isLoadingEvents, setIsLoadingEvents] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadCalls();
  }, []);

  useEffect(() => {
    if (selectedCall) {
      loadEvents(selectedCall);
    }
  }, [selectedCall]);

  const loadCalls = async () => {
    setIsLoadingCalls(true);
    try {
      const data = await api.adminListCalls(100);
      setCalls(data.calls || []);
      setError(null);
    } catch (e: unknown) {
      const err = e as { status?: number };
      if (err.status === 403) {
        navigate("/");
        return;
      }
      setError("Failed to load calls");
    } finally {
      setIsLoadingCalls(false);
    }
  };

  const loadEvents = async (callId: string) => {
    setIsLoadingEvents(true);
    try {
      const data = await api.adminGetCallEvents(callId);
      setEvents(data.events || []);
    } catch {
      setEvents([]);
    } finally {
      setIsLoadingEvents(false);
    }
  };

  const handleSelectCall = (callId: string) => {
    setSelectedCall(callId);
    navigate(`/admin/logs/${encodeURIComponent(callId)}`, { replace: true });
  };

  // Call list component
  const CallList = (
    <Paper withBorder p="md" style={isMobile ? undefined : { width: 380, flexShrink: 0 }}>
      <Title order={4} mb="md">
        Recent Calls
      </Title>
      {isLoadingCalls ? (
        <Loader />
      ) : (
        <ScrollArea h={isMobile ? 300 : 600}>
          <Stack gap="xs">
            {calls.length === 0 ? (
              <Text c="dimmed" ta="center" py="md">
                No calls found
              </Text>
            ) : (
              calls.map((call) => (
                <Paper
                  key={call.provider_call_id}
                  p="sm"
                  withBorder
                  style={{
                    cursor: "pointer",
                    background:
                      selectedCall === call.provider_call_id ? "#e7f5ff" : undefined,
                  }}
                  onClick={() => handleSelectCall(call.provider_call_id)}
                >
                  <Group justify="space-between" mb={4}>
                    <Text size="sm" fw={500}>
                      {call.from_number}
                    </Text>
                    <Badge
                      size="xs"
                      color={call.status === "completed" ? "green" : "gray"}
                    >
                      {call.status}
                    </Badge>
                  </Group>
                  <Text size="xs" c="dimmed">
                    {formatDateTime(call.started_at)}
                  </Text>
                  {call.ended_by && (
                    <Text size="xs" c="dimmed">
                      Ended by: {call.ended_by}
                    </Text>
                  )}
                </Paper>
              ))
            )}
          </Stack>
        </ScrollArea>
      )}
    </Paper>
  );

  // Events panel component
  const EventsPanel = (
    <Paper withBorder p="md" style={{ flex: 1, minWidth: 0 }}>
      <Title order={4} mb="md">
        Events{" "}
        {selectedCall && (
          <Text span size="sm" c="dimmed">
            ({events.length})
          </Text>
        )}
      </Title>
      {!selectedCall ? (
        <Text c="dimmed">Select a call to view events</Text>
      ) : isLoadingEvents ? (
        <Loader />
      ) : events.length === 0 ? (
        <Text c="dimmed">No events found for this call</Text>
      ) : (
        <ScrollArea h={isMobile ? 400 : 600}>
          <Table striped highlightOnHover>
            <Table.Thead>
              <Table.Tr>
                <Table.Th style={{ width: 90 }}>Time</Table.Th>
                <Table.Th style={{ width: 140 }}>Event</Table.Th>
                <Table.Th>Details</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {events.map((event) => (
                <Table.Tr key={event.id}>
                  <Table.Td>
                    <Text size="xs" ff="monospace">
                      {formatTime(event.created_at)}
                    </Text>
                  </Table.Td>
                  <Table.Td>
                    <Badge
                      color={eventTypeColors[event.event_type] || "gray"}
                      size="sm"
                    >
                      {event.event_type}
                    </Badge>
                  </Table.Td>
                  <Table.Td>
                    <Code
                      block
                      style={{
                        fontSize: "11px",
                        maxWidth: "100%",
                        overflow: "auto",
                        whiteSpace: "pre-wrap",
                        wordBreak: "break-word",
                      }}
                    >
                      {JSON.stringify(event.event_data, null, 2)}
                    </Code>
                  </Table.Td>
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        </ScrollArea>
      )}
    </Paper>
  );

  return (
    <Container size="xl" py="xl">
      <Stack gap="xl">
        {/* Header */}
        <Group justify="space-between">
          <Title order={2}>Call Logs</Title>
          <Button
            leftSection={<IconRefresh size={16} />}
            variant="light"
            onClick={() => {
              loadCalls();
              if (selectedCall) loadEvents(selectedCall);
            }}
          >
            Refresh
          </Button>
        </Group>

        {error && (
          <Alert icon={<IconAlertCircle size={16} />} color="red" variant="light">
            {error}
          </Alert>
        )}

        {/* Responsive layout */}
        {isMobile ? (
          <Stack gap="lg">
            {CallList}
            {selectedCall && EventsPanel}
          </Stack>
        ) : (
          <Group align="flex-start" gap="lg" wrap="nowrap">
            {CallList}
            {EventsPanel}
          </Group>
        )}
      </Stack>
    </Container>
  );
}
