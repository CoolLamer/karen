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
  Select,
} from "@mantine/core";
import {
  IconArrowLeft,
  IconTrash,
  IconPlus,
  IconAlertCircle,
  IconCheck,
  IconUserOff,
  IconUserPlus,
} from "@tabler/icons-react";
import { api, AdminPhoneNumber, AdminTenant } from "../api";

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

  // Assign modal state
  const [assignModalOpen, setAssignModalOpen] = useState(false);
  const [assigningPhoneId, setAssigningPhoneId] = useState<string | null>(null);
  const [assigningPhoneNumber, setAssigningPhoneNumber] = useState<string>("");
  const [selectedTenantId, setSelectedTenantId] = useState<string | null>(null);
  const [tenants, setTenants] = useState<AdminTenant[]>([]);
  const [isAssigning, setIsAssigning] = useState(false);

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

  const openAssignModal = async (phoneId: string, phoneNumber: string) => {
    setAssigningPhoneId(phoneId);
    setAssigningPhoneNumber(phoneNumber);
    setSelectedTenantId(null);

    try {
      const data = await api.adminListTenants();
      setTenants(data.tenants || []);
      setAssignModalOpen(true);
    } catch {
      setError("Failed to load tenants");
    }
  };

  const handleAssign = async () => {
    if (!assigningPhoneId || !selectedTenantId) return;

    setIsAssigning(true);
    setError(null);
    try {
      await api.adminUpdatePhoneNumber(assigningPhoneId, selectedTenantId);
      setSuccess("Phone number assigned");
      setAssignModalOpen(false);
      setAssigningPhoneId(null);
      setSelectedTenantId(null);
      loadPhoneNumbers();
      setTimeout(() => setSuccess(null), 3000);
    } catch {
      setError("Failed to assign phone number");
    } finally {
      setIsAssigning(false);
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
            <Group>
              <Button variant="light" onClick={() => navigate("/admin/users")}>
                Users
              </Button>
              <Button variant="light" onClick={() => navigate("/admin/logs")}>
                View Logs
              </Button>
              <Button leftSection={<IconPlus size={16} />} onClick={() => setAddModalOpen(true)}>
                Add Number
              </Button>
            </Group>
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
                          {!pn.tenant_id && (
                            <Tooltip label="Assign to tenant">
                              <ActionIcon
                                color="blue"
                                variant="light"
                                onClick={() => openAssignModal(pn.id, pn.twilio_number)}
                              >
                                <IconUserPlus size={16} />
                              </ActionIcon>
                            </Tooltip>
                          )}
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

      {/* Assign Modal */}
      <Modal
        opened={assignModalOpen}
        onClose={() => setAssignModalOpen(false)}
        title="Assign Phone Number"
      >
        <Stack>
          <Text size="sm" c="dimmed">
            Assign <Text span fw={500}>{assigningPhoneNumber}</Text> to a tenant
          </Text>
          <Select
            label="Select Tenant"
            placeholder="Choose a tenant"
            data={tenants.map((t) => ({ value: t.id, label: t.name }))}
            value={selectedTenantId}
            onChange={setSelectedTenantId}
            searchable
            required
          />
          <Group justify="flex-end">
            <Button variant="subtle" onClick={() => setAssignModalOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleAssign} loading={isAssigning} disabled={!selectedTenantId}>
              Assign
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Box>
  );
}
