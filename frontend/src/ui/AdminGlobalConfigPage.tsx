import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import {
  Button,
  Container,
  Group,
  Paper,
  Stack,
  Table,
  Text,
  TextInput,
  Title,
  Badge,
  Alert,
  NumberInput,
  Switch,
  Loader,
  Tooltip,
} from "@mantine/core";
import {
  IconAlertCircle,
  IconCheck,
  IconRefresh,
  IconInfoCircle,
} from "@tabler/icons-react";
import { api, GlobalConfigEntry } from "../api";

// Group configs by feature
const CONFIG_GROUPS: Record<string, { title: string; keys: string[] }> = {
  turn: {
    title: "Turn Timeout Settings",
    keys: [
      "max_turn_timeout_ms",
      "adaptive_turn_enabled",
      "adaptive_min_timeout_ms",
      "adaptive_text_decay_rate_ms",
      "adaptive_sentence_end_bonus_ms",
    ],
  },
  robocall: {
    title: "Robocall Detection Settings",
    keys: [
      "robocall_detection_enabled",
      "robocall_max_call_duration_ms",
      "robocall_silence_threshold_ms",
      "robocall_barge_in_threshold",
      "robocall_barge_in_window_ms",
      "robocall_repetition_threshold",
      "robocall_hold_keywords",
    ],
  },
};

// Config type for input rendering
const NUMERIC_KEYS = new Set([
  "max_turn_timeout_ms",
  "adaptive_min_timeout_ms",
  "adaptive_text_decay_rate_ms",
  "adaptive_sentence_end_bonus_ms",
  "robocall_max_call_duration_ms",
  "robocall_silence_threshold_ms",
  "robocall_barge_in_threshold",
  "robocall_barge_in_window_ms",
  "robocall_repetition_threshold",
]);

const BOOLEAN_KEYS = new Set([
  "adaptive_turn_enabled",
  "robocall_detection_enabled",
]);

export function AdminGlobalConfigPage() {
  const navigate = useNavigate();
  const [configs, setConfigs] = useState<GlobalConfigEntry[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [editValue, setEditValue] = useState<string>("");
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    loadConfigs();
  }, []);

  const loadConfigs = async () => {
    try {
      setIsLoading(true);
      const data = await api.adminListGlobalConfig();
      setConfigs(data.config || []);
      setError(null);
    } catch (e: unknown) {
      const err = e as { status?: number };
      if (err.status === 403) {
        navigate("/");
        return;
      }
      setError("Failed to load configuration");
    } finally {
      setIsLoading(false);
    }
  };

  const handleEdit = (config: GlobalConfigEntry) => {
    setEditingKey(config.key);
    setEditValue(config.value);
  };

  const handleCancel = () => {
    setEditingKey(null);
    setEditValue("");
  };

  const handleSave = async () => {
    if (!editingKey) return;

    setIsSaving(true);
    setError(null);
    try {
      await api.adminUpdateGlobalConfig(editingKey, editValue);
      setSuccess(`Updated ${editingKey}`);
      setEditingKey(null);
      setEditValue("");
      loadConfigs();
      setTimeout(() => setSuccess(null), 3000);
    } catch (e: unknown) {
      const err = e as { message?: string };
      setError(err.message || "Failed to update configuration");
    } finally {
      setIsSaving(false);
    }
  };

  const getConfigByKey = (key: string): GlobalConfigEntry | undefined => {
    return configs.find((c) => c.key === key);
  };

  const renderConfigRow = (config: GlobalConfigEntry) => {
    const isEditing = editingKey === config.key;
    const isNumeric = NUMERIC_KEYS.has(config.key);
    const isBoolean = BOOLEAN_KEYS.has(config.key);

    return (
      <Table.Tr key={config.key}>
        <Table.Td>
          <Group gap="xs">
            <Text size="sm" fw={500}>
              {config.key}
            </Text>
            {config.description && (
              <Tooltip label={config.description} multiline w={300}>
                <IconInfoCircle size={14} style={{ opacity: 0.5 }} />
              </Tooltip>
            )}
          </Group>
        </Table.Td>
        <Table.Td>
          {isEditing ? (
            <Group gap="xs">
              {isBoolean ? (
                <Switch
                  checked={editValue === "true"}
                  onChange={(e) =>
                    setEditValue(e.currentTarget.checked ? "true" : "false")
                  }
                />
              ) : isNumeric ? (
                <NumberInput
                  value={parseInt(editValue) || 0}
                  onChange={(val) => setEditValue(String(val))}
                  min={0}
                  style={{ width: 150 }}
                />
              ) : (
                <TextInput
                  value={editValue}
                  onChange={(e) => setEditValue(e.currentTarget.value)}
                  style={{ flex: 1, minWidth: 200 }}
                />
              )}
              <Button
                size="xs"
                color="green"
                onClick={handleSave}
                loading={isSaving}
              >
                Save
              </Button>
              <Button size="xs" variant="subtle" onClick={handleCancel}>
                Cancel
              </Button>
            </Group>
          ) : (
            <Group gap="xs">
              {isBoolean ? (
                <Badge color={config.value === "true" ? "green" : "gray"}>
                  {config.value === "true" ? "Enabled" : "Disabled"}
                </Badge>
              ) : (
                <Text size="sm" style={{ fontFamily: "monospace" }}>
                  {config.value}
                </Text>
              )}
              <Button
                size="xs"
                variant="subtle"
                onClick={() => handleEdit(config)}
              >
                Edit
              </Button>
            </Group>
          )}
        </Table.Td>
      </Table.Tr>
    );
  };

  const renderGroup = (groupKey: string, group: { title: string; keys: string[] }) => {
    const groupConfigs = group.keys
      .map((key) => getConfigByKey(key))
      .filter(Boolean) as GlobalConfigEntry[];

    if (groupConfigs.length === 0) return null;

    return (
      <Paper key={groupKey} p="md" withBorder>
        <Stack gap="md">
          <Title order={4}>{group.title}</Title>
          <Table>
            <Table.Thead>
              <Table.Tr>
                <Table.Th style={{ width: "40%" }}>Setting</Table.Th>
                <Table.Th>Value</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {groupConfigs.map((config) => renderConfigRow(config))}
            </Table.Tbody>
          </Table>
        </Stack>
      </Paper>
    );
  };

  // Get configs that don't belong to any group
  const ungroupedConfigs = configs.filter((config) => {
    return !Object.values(CONFIG_GROUPS).some((group) =>
      group.keys.includes(config.key)
    );
  });

  if (isLoading) {
    return (
      <Container size="lg" py="xl">
        <Group justify="center">
          <Loader size="lg" />
        </Group>
      </Container>
    );
  }

  return (
    <Container size="lg" py="xl">
      <Stack gap="lg">
        <Group justify="space-between">
          <Title order={2}>Global Configuration</Title>
          <Button
            leftSection={<IconRefresh size={16} />}
            variant="subtle"
            onClick={loadConfigs}
          >
            Refresh
          </Button>
        </Group>

        {error && (
          <Alert
            color="red"
            icon={<IconAlertCircle size={16} />}
            onClose={() => setError(null)}
            withCloseButton
          >
            {error}
          </Alert>
        )}

        {success && (
          <Alert
            color="green"
            icon={<IconCheck size={16} />}
            onClose={() => setSuccess(null)}
            withCloseButton
          >
            {success}
          </Alert>
        )}

        <Text c="dimmed" size="sm">
          These settings control system-wide behavior. Changes take effect
          immediately for new calls.
        </Text>

        {Object.entries(CONFIG_GROUPS).map(([key, group]) =>
          renderGroup(key, group)
        )}

        {ungroupedConfigs.length > 0 && (
          <Paper p="md" withBorder>
            <Stack gap="md">
              <Title order={4}>Other Settings</Title>
              <Table>
                <Table.Thead>
                  <Table.Tr>
                    <Table.Th style={{ width: "40%" }}>Setting</Table.Th>
                    <Table.Th>Value</Table.Th>
                  </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                  {ungroupedConfigs.map((config) => renderConfigRow(config))}
                </Table.Tbody>
              </Table>
            </Stack>
          </Paper>
        )}
      </Stack>
    </Container>
  );
}
