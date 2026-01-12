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
  Tabs,
  Box,
  ThemeIcon,
} from "@mantine/core";
import { useMediaQuery } from "@mantine/hooks";
import { IconRefresh, IconAlertCircle, IconCopy, IconCheck, IconRobot, IconUser, IconMessage, IconCode } from "@tabler/icons-react";
import { api, CallListItem, CallEvent, CallDetail } from "../api";

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
  vad_speech_started: "lime",
  vad_utterance_end: "lime",
  max_turn_timeout: "red",
};

function formatTime(ts: string): string {
  const d = new Date(ts);
  return d.toLocaleTimeString();
}

function formatDateTime(ts: string): string {
  const d = new Date(ts);
  return d.toLocaleString();
}

function formatEventsForAI(events: CallEvent[]): string {
  const lines = events.map((e) => {
    const time = new Date(e.created_at).toLocaleTimeString();
    const data = e.event_data ? JSON.stringify(e.event_data) : "";
    return `${time}\n\n${e.event_type}\n${data}`;
  });
  return lines.join("\n");
}

export function AdminLogsPage() {
  const navigate = useNavigate();
  const { providerCallId } = useParams<{ providerCallId?: string }>();
  const isMobile = useMediaQuery("(max-width: 768px)");

  const [calls, setCalls] = useState<CallListItem[]>([]);
  const [selectedCall, setSelectedCall] = useState<string | null>(providerCallId || null);
  const [callDetail, setCallDetail] = useState<CallDetail | null>(null);
  const [events, setEvents] = useState<CallEvent[]>([]);
  const [isLoadingCalls, setIsLoadingCalls] = useState(true);
  const [isLoadingDetail, setIsLoadingDetail] = useState(false);
  const [isLoadingEvents, setIsLoadingEvents] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [activeTab, setActiveTab] = useState<string | null>("transcript");

  const handleCopyEvents = async () => {
    const text = formatEventsForAI(events);
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  useEffect(() => {
    loadCalls();
  }, []);

  useEffect(() => {
    if (selectedCall) {
      loadCallDetail(selectedCall);
      loadEvents(selectedCall);
    }
  }, [selectedCall]);

  const loadCallDetail = async (callId: string) => {
    setIsLoadingDetail(true);
    try {
      const data = await api.adminGetCallDetail(callId);
      setCallDetail(data);
    } catch {
      setCallDetail(null);
    } finally {
      setIsLoadingDetail(false);
    }
  };

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
    setCallDetail(null);
    setEvents([]);
    navigate(`/admin/logs/${encodeURIComponent(callId)}`, { replace: true });
  };

  const formatSpeaker = (speaker: string) => {
    switch (speaker) {
      case "agent":
        return "Karen";
      case "caller":
        return "Caller";
      default:
        return speaker;
    }
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

  // Detail panel component with tabs for Transcript and Events
  const DetailPanel = (
    <Paper withBorder p="md" style={{ flex: 1, minWidth: 0 }}>
      {!selectedCall ? (
        <Text c="dimmed">Select a call to view details</Text>
      ) : (
        <Tabs value={activeTab} onChange={setActiveTab}>
          <Tabs.List mb="md">
            <Tabs.Tab value="transcript" leftSection={<IconMessage size={14} />}>
              Transcript
              {callDetail?.utterances && (
                <Text span size="xs" c="dimmed" ml={4}>
                  ({callDetail.utterances.length})
                </Text>
              )}
            </Tabs.Tab>
            <Tabs.Tab value="events" leftSection={<IconCode size={14} />}>
              Events
              {events.length > 0 && (
                <Text span size="xs" c="dimmed" ml={4}>
                  ({events.length})
                </Text>
              )}
            </Tabs.Tab>
          </Tabs.List>

          <Tabs.Panel value="transcript">
            {isLoadingDetail ? (
              <Loader />
            ) : !callDetail?.utterances?.length ? (
              <Text c="dimmed">No transcript available for this call</Text>
            ) : (
              <ScrollArea h={isMobile ? 400 : 550}>
                <Stack gap="md">
                  {callDetail.utterances.map((u) => (
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
                          maxWidth: "85%",
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
                          {u.interrupted && (
                            <Badge size="xs" variant="light" color="orange">
                              interrupted
                            </Badge>
                          )}
                        </Group>
                        <Text size="sm">{u.text}</Text>
                      </Paper>
                    </Box>
                  ))}
                  {callDetail.ended_by && (
                    <Text size="sm" c="dimmed" ta="center" fs="italic">
                      — Ended by {callDetail.ended_by} —
                    </Text>
                  )}
                </Stack>
              </ScrollArea>
            )}
          </Tabs.Panel>

          <Tabs.Panel value="events">
            <Group justify="flex-end" mb="sm">
              {events.length > 0 && (
                <Button
                  size="xs"
                  variant="light"
                  color={copied ? "green" : "gray"}
                  leftSection={copied ? <IconCheck size={14} /> : <IconCopy size={14} />}
                  onClick={handleCopyEvents}
                >
                  {copied ? "Copied!" : "Copy for AI"}
                </Button>
              )}
            </Group>
            {isLoadingEvents ? (
              <Loader />
            ) : events.length === 0 ? (
              <Text c="dimmed">No events found for this call</Text>
            ) : (
              <ScrollArea h={isMobile ? 400 : 520}>
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
          </Tabs.Panel>
        </Tabs>
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
              if (selectedCall) {
                loadCallDetail(selectedCall);
                loadEvents(selectedCall);
              }
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
            {DetailPanel}
          </Stack>
        ) : (
          <Group align="flex-start" gap="lg" wrap="nowrap">
            {CallList}
            {DetailPanel}
          </Group>
        )}
      </Stack>
    </Container>
  );
}
