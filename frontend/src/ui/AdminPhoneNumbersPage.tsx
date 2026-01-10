import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import {
  Box,
  Button,
  Container,
  Group,
  Modal,
  Paper,
  Stack,
  Table,
  Text,
  TextInput,
  Title,
  Badge,
  ActionIcon,
  Tooltip,
  Alert,
} from "@mantine/core";
import {
  IconArrowLeft,
  IconTrash,
  IconPlus,
  IconAlertCircle,
  IconCheck,
  IconUserOff,
} from "@tabler/icons-react";
import { api, AdminPhoneNumber } from "../api";

export function AdminPhoneNumbersPage() {
  const navigate = useNavigate();
  const [phoneNumbers, setPhoneNumbers] = useState<AdminPhoneNumber[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Add modal state
  const [addModalOpen, setAddModalOpen] = useState(false);
  const [newNumber, setNewNumber] = useState("");
  const [newSid, setNewSid] = useState("");
  const [isAdding, setIsAdding] = useState(false);

  useEffect(() => {
    loadPhoneNumbers();
  }, []);

  const loadPhoneNumbers = async () => {
    try {
      const data = await api.adminListPhoneNumbers();
      setPhoneNumbers(data.phone_numbers || []);
      setError(null);
    } catch (e: unknown) {
      const err = e as { status?: number };
      if (err.status === 403) {
        navigate("/");
        return;
      }
      setError("Failed to load phone numbers");
    } finally {
      setIsLoading(false);
    }
  };

  const handleAdd = async () => {
    setIsAdding(true);
    setError(null);
    try {
      await api.adminAddPhoneNumber(newNumber, newSid || undefined);
      setSuccess("Phone number added");
      setAddModalOpen(false);
      setNewNumber("");
      setNewSid("");
      loadPhoneNumbers();
      setTimeout(() => setSuccess(null), 3000);
    } catch (e: unknown) {
      const err = e as { message?: string };
      setError(err.message || "Failed to add phone number");
    } finally {
      setIsAdding(false);
    }
  };

  const handleDelete = async (id: string, number: string) => {
    if (!confirm(`Delete ${number}? This cannot be undone.`)) return;
    try {
      await api.adminDeletePhoneNumber(id);
      setSuccess("Phone number deleted");
      loadPhoneNumbers();
      setTimeout(() => setSuccess(null), 3000);
    } catch {
      setError("Failed to delete phone number");
    }
  };

  const handleUnassign = async (id: string) => {
    try {
      await api.adminUpdatePhoneNumber(id, null);
      setSuccess("Phone number unassigned");
      loadPhoneNumbers();
      setTimeout(() => setSuccess(null), 3000);
    } catch {
      setError("Failed to unassign phone number");
    }
  };

  if (isLoading) {
    return (
      <Container size="lg" py="xl">
        <Text c="dimmed">Loading...</Text>
      </Container>
    );
  }

  const availableCount = phoneNumbers.filter((p) => !p.tenant_id).length;
  const assignedCount = phoneNumbers.filter((p) => p.tenant_id).length;

  return (
    <Box mih="100vh" bg="gray.0">
      <Container size="lg" py="xl">
        <Stack gap="xl">
          {/* Header */}
          <Group justify="space-between">
            <Group>
              <Button
                variant="subtle"
                leftSection={<IconArrowLeft size={16} />}
                onClick={() => navigate("/")}
                px={0}
              >
                Back
              </Button>
              <Title order={2}>Admin: Phone Numbers</Title>
            </Group>
            <Button leftSection={<IconPlus size={16} />} onClick={() => setAddModalOpen(true)}>
              Add Number
            </Button>
          </Group>

          {/* Stats */}
          <Group>
            <Badge color="green" size="lg">
              Available: {availableCount}
            </Badge>
            <Badge color="blue" size="lg">
              Assigned: {assignedCount}
            </Badge>
            <Badge color="gray" size="lg">
              Total: {phoneNumbers.length}
            </Badge>
          </Group>

          {error && (
            <Alert icon={<IconAlertCircle size={16} />} color="red" variant="light">
              {error}
            </Alert>
          )}
          {success && (
            <Alert icon={<IconCheck size={16} />} color="green" variant="light">
              {success}
            </Alert>
          )}

          {/* Table */}
          <Paper withBorder>
            <Table striped highlightOnHover>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Phone Number</Table.Th>
                  <Table.Th>Status</Table.Th>
                  <Table.Th>Tenant</Table.Th>
                  <Table.Th>Created</Table.Th>
                  <Table.Th>Actions</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {phoneNumbers.length === 0 ? (
                  <Table.Tr>
                    <Table.Td colSpan={5}>
                      <Text c="dimmed" ta="center" py="md">
                        No phone numbers in the pool
                      </Text>
                    </Table.Td>
                  </Table.Tr>
                ) : (
                  phoneNumbers.map((pn) => (
                    <Table.Tr key={pn.id}>
                      <Table.Td>
                        <Text fw={500}>{pn.twilio_number}</Text>
                        {pn.twilio_sid && (
                          <Text size="xs" c="dimmed">
                            {pn.twilio_sid}
                          </Text>
                        )}
                      </Table.Td>
                      <Table.Td>
                        <Badge color={pn.tenant_id ? "blue" : "green"}>
                          {pn.tenant_id ? "Assigned" : "Available"}
                        </Badge>
                      </Table.Td>
                      <Table.Td>{pn.tenant_name || (pn.tenant_id ? pn.tenant_id : "—")}</Table.Td>
                      <Table.Td>
                        <Text size="sm">{new Date(pn.created_at).toLocaleDateString()}</Text>
                      </Table.Td>
                      <Table.Td>
                        <Group gap="xs">
                          {pn.tenant_id && (
                            <Tooltip label="Unassign">
                              <ActionIcon
                                color="orange"
                                variant="light"
                                onClick={() => handleUnassign(pn.id)}
                              >
                                <IconUserOff size={16} />
                              </ActionIcon>
                            </Tooltip>
                          )}
                          <Tooltip label="Delete">
                            <ActionIcon
                              color="red"
                              variant="light"
                              onClick={() => handleDelete(pn.id, pn.twilio_number)}
                            >
                              <IconTrash size={16} />
                            </ActionIcon>
                          </Tooltip>
                        </Group>
                      </Table.Td>
                    </Table.Tr>
                  ))
                )}
              </Table.Tbody>
            </Table>
          </Paper>
        </Stack>
      </Container>

      {/* Add Modal */}
      <Modal opened={addModalOpen} onClose={() => setAddModalOpen(false)} title="Add Phone Number">
        <Stack>
          <TextInput
            label="Phone Number"
            placeholder="+420123456789"
            description="E.164 format (e.g. +420123456789)"
            value={newNumber}
            onChange={(e) => setNewNumber(e.target.value)}
            required
          />
          <TextInput
            label="Twilio SID"
            placeholder="PN..."
            description="Optional - Twilio Phone Number SID"
            value={newSid}
            onChange={(e) => setNewSid(e.target.value)}
          />
          <Group justify="flex-end">
            <Button variant="subtle" onClick={() => setAddModalOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleAdd} loading={isAdding} disabled={!newNumber}>
              Add
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Box>
  );
}
