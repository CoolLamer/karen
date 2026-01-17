import React from "react";
import { useNavigate } from "react-router-dom";
import {
  Accordion,
  Badge,
  Box,
  Collapse,
  Container,
  Divider,
  Group,
  Modal,
  NumberInput,
  Paper,
  Select,
  Stack,
  Table,
  Text,
  Textarea,
  TextInput,
  Title,
  Spoiler,
  ThemeIcon,
  UnstyledButton,
  Alert,
  Button,
} from "@mantine/core";
import { useMediaQuery, useDisclosure } from "@mantine/hooks";
import {
  IconChevronDown,
  IconChevronRight,
  IconEdit,
  IconUsers,
  IconSettings,
  IconPhoneCall,
  IconRobot,
  IconUser,
  IconRefresh,
  IconTrash,
  IconCreditCard,
  IconNote,
} from "@tabler/icons-react";
import { api, AdminTenantDetail, AdminUser, CallDetail, TenantCostSummary } from "../api";
import { IconCoin } from "@tabler/icons-react";

const PLAN_OPTIONS = [
  { value: "trial", label: "Trial" },
  { value: "basic", label: "Basic" },
  { value: "pro", label: "Pro" },
];

const STATUS_OPTIONS = [
  { value: "active", label: "Active" },
  { value: "suspended", label: "Suspended" },
  { value: "cancelled", label: "Cancelled" },
];

const PLAN_COLORS: Record<string, string> = {
  trial: "gray",
  basic: "blue",
  pro: "violet",
};

const STATUS_COLORS: Record<string, string> = {
  active: "green",
  suspended: "orange",
  cancelled: "red",
};

function formatRelativeTime(date: Date): string {
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return "Just now";
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  return date.toLocaleDateString();
}

function formatSpeaker(speaker: string) {
  return speaker === "agent" ? "Karen" : "Caller";
}

export function AdminUsersPage() {
  const navigate = useNavigate();
  const isMobile = useMediaQuery("(max-width: 768px)");
  const [tenants, setTenants] = React.useState<AdminTenantDetail[] | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [success, setSuccess] = React.useState<string | null>(null);

  // Lazy-loaded data per tenant
  const [tenantUsers, setTenantUsers] = React.useState<Record<string, AdminUser[]>>({});
  const [tenantCalls, setTenantCalls] = React.useState<Record<string, CallDetail[]>>({});
  const [tenantCosts, setTenantCosts] = React.useState<Record<string, TenantCostSummary>>({});
  const [loadingUsers, setLoadingUsers] = React.useState<Record<string, boolean>>({});
  const [loadingCalls, setLoadingCalls] = React.useState<Record<string, boolean>>({});
  const [loadingCosts, setLoadingCosts] = React.useState<Record<string, boolean>>({});

  // Mobile expanded sections
  const [expandedTenant, setExpandedTenant] = React.useState<string | null>(null);
  const [expandedSection, setExpandedSection] = React.useState<Record<string, { config?: boolean; billing?: boolean; users?: boolean; calls?: boolean }>>({});
  const [expandedCall, setExpandedCall] = React.useState<string | null>(null);

  // Edit modal
  const [editModalOpened, { open: openEditModal, close: closeEditModal }] = useDisclosure(false);
  const [editingTenant, setEditingTenant] = React.useState<AdminTenantDetail | null>(null);
  const [editPlan, setEditPlan] = React.useState("");
  const [editStatus, setEditStatus] = React.useState("");
  const [editMaxTurnTimeout, setEditMaxTurnTimeout] = React.useState<number | "">("");
  const [editTrialEndsAt, setEditTrialEndsAt] = React.useState("");
  const [editCurrentPeriodCalls, setEditCurrentPeriodCalls] = React.useState<number | "">("");
  const [editAdminNotes, setEditAdminNotes] = React.useState("");
  const [saving, setSaving] = React.useState(false);

  // Reset onboarding modal
  const [resetModalOpened, { open: openResetModal, close: closeResetModal }] = useDisclosure(false);
  const [resettingUser, setResettingUser] = React.useState<AdminUser | null>(null);
  const [resettingTenantId, setResettingTenantId] = React.useState<string | null>(null);
  const [resetting, setResetting] = React.useState(false);

  // Delete tenant modal
  const [deleteModalOpened, { open: openDeleteModal, close: closeDeleteModal }] = useDisclosure(false);
  const [deletingTenant, setDeletingTenant] = React.useState<AdminTenantDetail | null>(null);
  const [deleting, setDeleting] = React.useState(false);

  React.useEffect(() => {
    api
      .adminListTenantsWithDetails()
      .then((data) => setTenants(data.tenants || []))
      .catch((e: unknown) => {
        const err = e as { status?: number };
        if (err.status === 403) {
          navigate("/");
          return;
        }
        setError("Failed to load tenants");
      });
  }, [navigate]);

  const loadUsers = async (tenantId: string) => {
    if (tenantUsers[tenantId] || loadingUsers[tenantId]) return;
    setLoadingUsers((prev) => ({ ...prev, [tenantId]: true }));
    try {
      const data = await api.adminGetTenantUsers(tenantId);
      setTenantUsers((prev) => ({ ...prev, [tenantId]: data.users || [] }));
    } catch {
      setError("Failed to load users");
    } finally {
      setLoadingUsers((prev) => ({ ...prev, [tenantId]: false }));
    }
  };

  const loadCalls = async (tenantId: string) => {
    if (tenantCalls[tenantId] || loadingCalls[tenantId]) return;
    setLoadingCalls((prev) => ({ ...prev, [tenantId]: true }));
    try {
      const data = await api.adminGetTenantCalls(tenantId, 10);
      setTenantCalls((prev) => ({ ...prev, [tenantId]: data.calls || [] }));
    } catch {
      setError("Failed to load calls");
    } finally {
      setLoadingCalls((prev) => ({ ...prev, [tenantId]: false }));
    }
  };

  const loadCosts = async (tenantId: string) => {
    if (tenantCosts[tenantId] || loadingCosts[tenantId]) return;
    setLoadingCosts((prev) => ({ ...prev, [tenantId]: true }));
    try {
      const data = await api.adminGetTenantCosts(tenantId);
      setTenantCosts((prev) => ({ ...prev, [tenantId]: data }));
    } catch {
      setError("Failed to load costs");
    } finally {
      setLoadingCosts((prev) => ({ ...prev, [tenantId]: false }));
    }
  };

  const handleEdit = (tenant: AdminTenantDetail) => {
    setEditingTenant(tenant);
    setEditPlan(tenant.plan);
    setEditStatus(tenant.status);
    setEditMaxTurnTimeout(tenant.max_turn_timeout_ms || "");
    // Initialize billing fields - convert ISO date to date input format (YYYY-MM-DD)
    setEditTrialEndsAt(tenant.trial_ends_at ? tenant.trial_ends_at.slice(0, 10) : "");
    setEditCurrentPeriodCalls(tenant.current_period_calls ?? "");
    setEditAdminNotes(tenant.admin_notes || "");
    openEditModal();
  };

  const handleSaveEdit = async () => {
    if (!editingTenant) return;
    setSaving(true);
    try {
      const maxTimeout = editMaxTurnTimeout ? Number(editMaxTurnTimeout) : undefined;
      // Convert date input (YYYY-MM-DD) to ISO string if provided
      const trialEndsAt = editTrialEndsAt ? new Date(editTrialEndsAt + "T23:59:59Z").toISOString() : undefined;
      const currentPeriodCalls = editCurrentPeriodCalls !== "" ? Number(editCurrentPeriodCalls) : undefined;
      // Send empty string to clear notes, undefined to leave unchanged
      const adminNotes = editAdminNotes;

      await api.adminUpdateTenant(editingTenant.id, {
        plan: editPlan,
        status: editStatus,
        max_turn_timeout_ms: maxTimeout,
        trial_ends_at: trialEndsAt,
        current_period_calls: currentPeriodCalls,
        admin_notes: adminNotes,
      });
      setTenants((prev) =>
        prev?.map((t) =>
          t.id === editingTenant.id
            ? {
                ...t,
                plan: editPlan,
                status: editStatus,
                max_turn_timeout_ms: maxTimeout,
                trial_ends_at: trialEndsAt,
                current_period_calls: currentPeriodCalls ?? t.current_period_calls,
                admin_notes: adminNotes,
              }
            : t
        ) ?? null
      );
      setSuccess(`Updated ${editingTenant.name}`);
      closeEditModal();
      setTimeout(() => setSuccess(null), 3000);
    } catch {
      setError("Failed to update tenant");
    } finally {
      setSaving(false);
    }
  };

  const handleResetOnboarding = (user: AdminUser, tenantId: string) => {
    setResettingUser(user);
    setResettingTenantId(tenantId);
    openResetModal();
  };

  const handleConfirmReset = async () => {
    if (!resettingUser || !resettingTenantId) return;
    setResetting(true);
    try {
      await api.adminResetUserOnboarding(resettingUser.id);
      // Remove user from local state for that tenant
      setTenantUsers((prev) => ({
        ...prev,
        [resettingTenantId]: prev[resettingTenantId]?.filter((u) => u.id !== resettingUser.id) || [],
      }));
      // Update tenant user count
      setTenants((prev) =>
        prev?.map((t) =>
          t.id === resettingTenantId ? { ...t, user_count: t.user_count - 1 } : t
        ) ?? null
      );
      setSuccess(`Reset onboarding for ${resettingUser.phone}`);
      closeResetModal();
      setTimeout(() => setSuccess(null), 3000);
    } catch {
      setError("Failed to reset onboarding");
    } finally {
      setResetting(false);
    }
  };

  const handleDeleteTenant = (tenant: AdminTenantDetail) => {
    setDeletingTenant(tenant);
    openDeleteModal();
  };

  const handleConfirmDelete = async () => {
    if (!deletingTenant) return;
    setDeleting(true);
    try {
      await api.adminDeleteTenant(deletingTenant.id);
      // Remove tenant from local state
      setTenants((prev) => prev?.filter((t) => t.id !== deletingTenant.id) ?? null);
      // Clean up any loaded data for this tenant
      setTenantUsers((prev) => {
        const { [deletingTenant.id]: _, ...rest } = prev;
        return rest;
      });
      setTenantCalls((prev) => {
        const { [deletingTenant.id]: _, ...rest } = prev;
        return rest;
      });
      setSuccess(`Deleted tenant "${deletingTenant.name}" and all associated data`);
      closeDeleteModal();
      setTimeout(() => setSuccess(null), 3000);
    } catch {
      setError("Failed to delete tenant");
    } finally {
      setDeleting(false);
    }
  };

  const toggleMobileSection = (tenantId: string, section: "config" | "billing" | "users" | "calls") => {
    setExpandedSection((prev) => ({
      ...prev,
      [tenantId]: {
        ...prev[tenantId],
        [section]: !prev[tenantId]?.[section],
      },
    }));
    if (section === "users") loadUsers(tenantId);
    if (section === "calls") loadCalls(tenantId);
  };

  // Stats
  const stats = React.useMemo(() => {
    if (!tenants) return null;
    return {
      active: tenants.filter((t) => t.status === "active").length,
      trial: tenants.filter((t) => t.plan === "trial").length,
      total: tenants.length,
    };
  }, [tenants]);

  return (
    <Container size="lg" py="xl">
      <Stack gap="md">
        {/* Header */}
        <Title order={2}>Users</Title>

        {/* Stats badges */}
        {stats && (
          <Group gap="xs">
            <Badge color="green" variant="light">
              Active: {stats.active}
            </Badge>
            <Badge color="gray" variant="light">
              Trial: {stats.trial}
            </Badge>
            <Badge variant="outline">Total: {stats.total}</Badge>
          </Group>
        )}

        {/* Alerts */}
        {error && (
          <Alert color="red" onClose={() => setError(null)} withCloseButton>
            {error}
          </Alert>
        )}
        {success && (
          <Alert color="green" onClose={() => setSuccess(null)} withCloseButton>
            {success}
          </Alert>
        )}

        {/* Loading state */}
        {!tenants && !error && <Text c="dimmed">Loading...</Text>}

        {/* Empty state */}
        {tenants && tenants.length === 0 && (
          <Paper p="xl" withBorder ta="center" radius="md">
            <ThemeIcon size={60} radius="xl" variant="light" color="gray" mb="md" mx="auto">
              <IconUsers size={30} />
            </ThemeIcon>
            <Text c="dimmed" size="lg">
              No tenants yet
            </Text>
          </Paper>
        )}

        {/* Mobile layout */}
        {tenants && tenants.length > 0 && isMobile && (
          <Stack gap="sm">
            {tenants.map((tenant) => (
              <Paper
              key={tenant.id}
              p="md"
              radius="md"
              withBorder
              style={{
                borderLeft: `4px solid var(--mantine-color-${STATUS_COLORS[tenant.status] || "gray"}-5)`,
              }}
            >
              {/* Tenant header */}
              <UnstyledButton
                onClick={() => setExpandedTenant(expandedTenant === tenant.id ? null : tenant.id)}
                style={{ width: "100%" }}
              >
                <Group justify="space-between" wrap="nowrap">
                  <Box style={{ flex: 1, minWidth: 0 }}>
                    <Text fw={600} truncate>
                      {tenant.name}
                    </Text>
                    <Group gap="xs" mt={4}>
                      <Badge size="xs" color={PLAN_COLORS[tenant.plan] || "gray"}>
                        {tenant.plan}
                      </Badge>
                      <Badge size="xs" color={STATUS_COLORS[tenant.status] || "gray"}>
                        {tenant.status}
                      </Badge>
                    </Group>
                  </Box>
                  {expandedTenant === tenant.id ? (
                    <IconChevronDown size={20} color="gray" />
                  ) : (
                    <IconChevronRight size={20} color="gray" />
                  )}
                </Group>
              </UnstyledButton>

              <Collapse in={expandedTenant === tenant.id}>
                <Stack gap="xs" mt="md">
                  {/* Config section */}
                  <UnstyledButton
                    onClick={() => toggleMobileSection(tenant.id, "config")}
                    style={{ width: "100%" }}
                  >
                    <Group gap="xs">
                      <IconSettings size={16} />
                      <Text size="sm" fw={500}>
                        Configuration
                      </Text>
                      {expandedSection[tenant.id]?.config ? (
                        <IconChevronDown size={14} />
                      ) : (
                        <IconChevronRight size={14} />
                      )}
                    </Group>
                  </UnstyledButton>
                  <Collapse in={expandedSection[tenant.id]?.config || false}>
                    <Paper p="sm" bg="gray.0" radius="sm">
                      <Stack gap="xs">
                        <Text size="xs">
                          <Text span fw={500}>Language:</Text> {tenant.language}
                        </Text>
                        {tenant.voice_id && (
                          <Text size="xs">
                            <Text span fw={500}>Voice:</Text> {tenant.voice_id}
                          </Text>
                        )}
                        {tenant.forward_number && (
                          <Text size="xs">
                            <Text span fw={500}>Forward:</Text> {tenant.forward_number}
                          </Text>
                        )}
                        {tenant.vip_names && tenant.vip_names.length > 0 && (
                          <Text size="xs">
                            <Text span fw={500}>VIP:</Text> {tenant.vip_names.join(", ")}
                          </Text>
                        )}
                        <Text size="xs">
                          <Text span fw={500}>Max Turn Timeout:</Text>{" "}
                          {tenant.max_turn_timeout_ms ? `${tenant.max_turn_timeout_ms}ms` : "default (4000ms)"}
                        </Text>
                        <Box>
                          <Text size="xs" fw={500} mb={2}>
                            System Prompt:
                          </Text>
                          <Spoiler maxHeight={60} showLabel="Show more" hideLabel="Hide">
                            <Text size="xs" style={{ whiteSpace: "pre-wrap" }}>
                              {tenant.system_prompt}
                            </Text>
                          </Spoiler>
                        </Box>
                      </Stack>
                    </Paper>
                  </Collapse>

                  {/* Billing section */}
                  <UnstyledButton
                    onClick={() => toggleMobileSection(tenant.id, "billing")}
                    style={{ width: "100%" }}
                  >
                    <Group gap="xs">
                      <IconCreditCard size={16} />
                      <Text size="sm" fw={500}>
                        Billing
                      </Text>
                      {expandedSection[tenant.id]?.billing ? (
                        <IconChevronDown size={14} />
                      ) : (
                        <IconChevronRight size={14} />
                      )}
                    </Group>
                  </UnstyledButton>
                  <Collapse in={expandedSection[tenant.id]?.billing || false}>
                    <Paper p="sm" bg="gray.0" radius="sm">
                      <Stack gap="xs">
                        {tenant.stripe_customer_id && (
                          <Text size="xs">
                            <Text span fw={500}>Stripe Customer:</Text>{" "}
                            <Text span ff="monospace">{tenant.stripe_customer_id}</Text>
                          </Text>
                        )}
                        {tenant.stripe_subscription_id && (
                          <Text size="xs">
                            <Text span fw={500}>Subscription:</Text>{" "}
                            <Text span ff="monospace">{tenant.stripe_subscription_id}</Text>
                          </Text>
                        )}
                        {tenant.trial_ends_at && (
                          <Group gap="xs">
                            <Text size="xs">
                              <Text span fw={500}>Trial Ends:</Text>{" "}
                              {new Date(tenant.trial_ends_at).toLocaleDateString()}
                            </Text>
                            {new Date(tenant.trial_ends_at) < new Date() && (
                              <Badge color="red" size="xs">Expired</Badge>
                            )}
                          </Group>
                        )}
                        {tenant.current_period_start && (
                          <Text size="xs">
                            <Text span fw={500}>Period Start:</Text>{" "}
                            {new Date(tenant.current_period_start).toLocaleDateString()}
                          </Text>
                        )}
                        <Text size="xs">
                          <Text span fw={500}>Period Calls:</Text>{" "}
                          <Badge variant="outline" size="xs">{tenant.current_period_calls}</Badge>
                        </Text>
                        <Text size="xs">
                          <Text span fw={500}>Time Saved:</Text>{" "}
                          {Math.floor(tenant.time_saved_seconds / 60)} minutes
                        </Text>
                        <Text size="xs">
                          <Text span fw={500}>Spam Blocked:</Text> {tenant.spam_calls_blocked}
                        </Text>
                        {tenant.admin_notes && (
                          <Box>
                            <Text size="xs" fw={500} mb={2}>
                              <IconNote size={12} style={{ verticalAlign: "middle", marginRight: 4 }} />
                              Admin Notes:
                            </Text>
                            <Paper p="xs" bg="yellow.0" radius="sm">
                              <Text size="xs" style={{ whiteSpace: "pre-wrap" }}>
                                {tenant.admin_notes}
                              </Text>
                            </Paper>
                          </Box>
                        )}
                      </Stack>
                    </Paper>
                  </Collapse>

                  {/* Users section */}
                  <UnstyledButton
                    onClick={() => toggleMobileSection(tenant.id, "users")}
                    style={{ width: "100%" }}
                  >
                    <Group gap="xs">
                      <IconUsers size={16} />
                      <Text size="sm" fw={500}>
                        Users ({tenant.user_count})
                      </Text>
                      {expandedSection[tenant.id]?.users ? (
                        <IconChevronDown size={14} />
                      ) : (
                        <IconChevronRight size={14} />
                      )}
                    </Group>
                  </UnstyledButton>
                  <Collapse in={expandedSection[tenant.id]?.users || false}>
                    {loadingUsers[tenant.id] && <Text size="xs" c="dimmed">Loading...</Text>}
                    {tenantUsers[tenant.id]?.length === 0 && (
                      <Text size="xs" c="dimmed">No users</Text>
                    )}
                    {tenantUsers[tenant.id]?.map((user) => (
                      <Paper key={user.id} p="xs" bg="gray.0" radius="sm" mb="xs">
                        <Group justify="space-between" wrap="nowrap">
                          <Box style={{ minWidth: 0 }}>
                            <Text size="sm" fw={500} truncate>
                              {user.phone}
                            </Text>
                            <Text size="xs" c="dimmed" truncate>
                              {user.name || "No name"} | {user.role}
                            </Text>
                          </Box>
                          <Group gap="xs" wrap="nowrap">
                            {user.last_login_at && (
                              <Text size="xs" c="dimmed">
                                {formatRelativeTime(new Date(user.last_login_at))}
                              </Text>
                            )}
                            <Button
                              variant="subtle"
                              color="orange"
                              size="xs"
                              p={4}
                              onClick={() => handleResetOnboarding(user, tenant.id)}
                            >
                              <IconRefresh size={14} />
                            </Button>
                          </Group>
                        </Group>
                      </Paper>
                    ))}
                  </Collapse>

                  {/* Calls section */}
                  <UnstyledButton
                    onClick={() => toggleMobileSection(tenant.id, "calls")}
                    style={{ width: "100%" }}
                  >
                    <Group gap="xs">
                      <IconPhoneCall size={16} />
                      <Text size="sm" fw={500}>
                        Calls ({tenant.call_count})
                      </Text>
                      {expandedSection[tenant.id]?.calls ? (
                        <IconChevronDown size={14} />
                      ) : (
                        <IconChevronRight size={14} />
                      )}
                    </Group>
                  </UnstyledButton>
                  <Collapse in={expandedSection[tenant.id]?.calls || false}>
                    {loadingCalls[tenant.id] && <Text size="xs" c="dimmed">Loading...</Text>}
                    {tenantCalls[tenant.id]?.length === 0 && (
                      <Text size="xs" c="dimmed">No calls</Text>
                    )}
                    {tenantCalls[tenant.id]?.map((call) => (
                      <Paper key={call.provider_call_id} p="xs" bg="gray.0" radius="sm" mb="xs">
                        <UnstyledButton
                          onClick={() =>
                            setExpandedCall(
                              expandedCall === call.provider_call_id ? null : call.provider_call_id
                            )
                          }
                          style={{ width: "100%" }}
                        >
                          <Group justify="space-between">
                            <Box>
                              <Text size="sm" fw={500}>
                                {call.from_number}
                              </Text>
                              <Text size="xs" c="dimmed">
                                {formatRelativeTime(new Date(call.started_at))} | {call.status}
                              </Text>
                            </Box>
                            {expandedCall === call.provider_call_id ? (
                              <IconChevronDown size={14} />
                            ) : (
                              <IconChevronRight size={14} />
                            )}
                          </Group>
                        </UnstyledButton>
                        <Collapse in={expandedCall === call.provider_call_id}>
                          {call.utterances?.length > 0 ? (
                            <Stack gap="xs" mt="sm">
                              {call.utterances.map((u) => (
                                <Box
                                  key={u.sequence}
                                  style={{
                                    display: "flex",
                                    justifyContent: u.speaker === "agent" ? "flex-end" : "flex-start",
                                  }}
                                >
                                  <Paper
                                    p="xs"
                                    radius="md"
                                    style={{
                                      maxWidth: "85%",
                                      backgroundColor:
                                        u.speaker === "agent"
                                          ? "var(--mantine-color-teal-0)"
                                          : "var(--mantine-color-gray-1)",
                                    }}
                                  >
                                    <Group gap="xs" mb={2}>
                                      <ThemeIcon
                                        size="xs"
                                        color={u.speaker === "agent" ? "teal" : "gray"}
                                        variant="transparent"
                                      >
                                        {u.speaker === "agent" ? (
                                          <IconRobot size={10} />
                                        ) : (
                                          <IconUser size={10} />
                                        )}
                                      </ThemeIcon>
                                      <Text size="xs" fw={500}>
                                        {formatSpeaker(u.speaker)}
                                      </Text>
                                    </Group>
                                    <Text size="xs">{u.text}</Text>
                                  </Paper>
                                </Box>
                              ))}
                            </Stack>
                          ) : (
                            <Text size="xs" c="dimmed" mt="sm">
                              No transcript
                            </Text>
                          )}
                        </Collapse>
                      </Paper>
                    ))}
                  </Collapse>

                  {/* Action buttons */}
                  <Group gap="xs" mt="xs">
                    <Button
                      variant="light"
                      size="xs"
                      leftSection={<IconEdit size={14} />}
                      onClick={() => handleEdit(tenant)}
                    >
                      Edit Plan/Status
                    </Button>
                    <Button
                      variant="light"
                      color="red"
                      size="xs"
                      leftSection={<IconTrash size={14} />}
                      onClick={() => handleDeleteTenant(tenant)}
                    >
                      Delete
                    </Button>
                  </Group>
                </Stack>
              </Collapse>
            </Paper>
          ))}
        </Stack>
      )}

      {/* Desktop layout */}
      {tenants && tenants.length > 0 && !isMobile && (
        <Accordion variant="separated">
          {tenants.map((tenant) => (
            <Accordion.Item key={tenant.id} value={tenant.id}>
              <Accordion.Control
                onClick={() => {
                  loadUsers(tenant.id);
                  loadCalls(tenant.id);
                  loadCosts(tenant.id);
                }}
              >
                <Group justify="space-between" wrap="nowrap" style={{ flex: 1 }}>
                  <Group gap="sm">
                    <Text fw={600}>{tenant.name}</Text>
                    <Badge size="sm" color={PLAN_COLORS[tenant.plan] || "gray"}>
                      {tenant.plan}
                    </Badge>
                    <Badge size="sm" color={STATUS_COLORS[tenant.status] || "gray"}>
                      {tenant.status}
                    </Badge>
                  </Group>
                  <Group gap="xs">
                    <Badge variant="outline" size="sm">
                      {tenant.user_count} users
                    </Badge>
                    <Badge variant="outline" size="sm">
                      {tenant.call_count} calls
                    </Badge>
                    <Button
                      variant="subtle"
                      size="xs"
                      leftSection={<IconEdit size={14} />}
                      onClick={(e) => {
                        e.stopPropagation();
                        handleEdit(tenant);
                      }}
                    >
                      Edit
                    </Button>
                    <Button
                      variant="subtle"
                      color="red"
                      size="xs"
                      leftSection={<IconTrash size={14} />}
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDeleteTenant(tenant);
                      }}
                    >
                      Delete
                    </Button>
                  </Group>
                </Group>
              </Accordion.Control>
              <Accordion.Panel>
                <Stack gap="md">
                  {/* Configuration */}
                  <Box>
                    <Text fw={600} size="sm" mb="xs">
                      Configuration
                    </Text>
                    <Table>
                      <Table.Tbody>
                        <Table.Tr>
                          <Table.Td w={150}>Language</Table.Td>
                          <Table.Td>{tenant.language}</Table.Td>
                        </Table.Tr>
                        {tenant.voice_id && (
                          <Table.Tr>
                            <Table.Td>Voice ID</Table.Td>
                            <Table.Td>{tenant.voice_id}</Table.Td>
                          </Table.Tr>
                        )}
                        {tenant.forward_number && (
                          <Table.Tr>
                            <Table.Td>Forward Number</Table.Td>
                            <Table.Td>{tenant.forward_number}</Table.Td>
                          </Table.Tr>
                        )}
                        {tenant.vip_names && tenant.vip_names.length > 0 && (
                          <Table.Tr>
                            <Table.Td>VIP Names</Table.Td>
                            <Table.Td>{tenant.vip_names.join(", ")}</Table.Td>
                          </Table.Tr>
                        )}
                        <Table.Tr>
                          <Table.Td>Max Turn Timeout</Table.Td>
                          <Table.Td>
                            {tenant.max_turn_timeout_ms
                              ? `${tenant.max_turn_timeout_ms}ms`
                              : "default (4000ms)"}
                          </Table.Td>
                        </Table.Tr>
                        <Table.Tr>
                          <Table.Td style={{ verticalAlign: "top" }}>System Prompt</Table.Td>
                          <Table.Td>
                            <Spoiler maxHeight={100} showLabel="Show full" hideLabel="Hide">
                              <Text size="sm" style={{ whiteSpace: "pre-wrap" }}>
                                {tenant.system_prompt}
                              </Text>
                            </Spoiler>
                          </Table.Td>
                        </Table.Tr>
                      </Table.Tbody>
                    </Table>
                  </Box>

                  {/* Billing */}
                  <Box>
                    <Text fw={600} size="sm" mb="xs">
                      Billing
                    </Text>
                    <Table>
                      <Table.Tbody>
                        {tenant.stripe_customer_id && (
                          <Table.Tr>
                            <Table.Td w={180}>Stripe Customer</Table.Td>
                            <Table.Td>
                              <Text size="sm" ff="monospace">{tenant.stripe_customer_id}</Text>
                            </Table.Td>
                          </Table.Tr>
                        )}
                        {tenant.stripe_subscription_id && (
                          <Table.Tr>
                            <Table.Td>Stripe Subscription</Table.Td>
                            <Table.Td>
                              <Text size="sm" ff="monospace">{tenant.stripe_subscription_id}</Text>
                            </Table.Td>
                          </Table.Tr>
                        )}
                        {tenant.trial_ends_at && (
                          <Table.Tr>
                            <Table.Td>Trial Ends</Table.Td>
                            <Table.Td>
                              <Group gap="xs">
                                <Text size="sm">
                                  {new Date(tenant.trial_ends_at).toLocaleDateString()}
                                </Text>
                                {new Date(tenant.trial_ends_at) < new Date() && (
                                  <Badge color="red" size="xs">Expired</Badge>
                                )}
                              </Group>
                            </Table.Td>
                          </Table.Tr>
                        )}
                        {tenant.current_period_start && (
                          <Table.Tr>
                            <Table.Td>Period Start</Table.Td>
                            <Table.Td>
                              {new Date(tenant.current_period_start).toLocaleDateString()}
                            </Table.Td>
                          </Table.Tr>
                        )}
                        <Table.Tr>
                          <Table.Td>Period Calls</Table.Td>
                          <Table.Td>
                            <Badge variant="outline">{tenant.current_period_calls}</Badge>
                          </Table.Td>
                        </Table.Tr>
                        <Table.Tr>
                          <Table.Td>Time Saved</Table.Td>
                          <Table.Td>
                            {Math.floor(tenant.time_saved_seconds / 60)} minutes
                          </Table.Td>
                        </Table.Tr>
                        <Table.Tr>
                          <Table.Td>Spam Blocked</Table.Td>
                          <Table.Td>{tenant.spam_calls_blocked}</Table.Td>
                        </Table.Tr>
                      </Table.Tbody>
                    </Table>
                  </Box>

                  {/* Cost Breakdown */}
                  <Box>
                    <Text fw={600} size="sm" mb="xs">
                      <IconCoin size={14} style={{ verticalAlign: "middle", marginRight: 4 }} />
                      Cost Breakdown (Current Month)
                    </Text>
                    {loadingCosts[tenant.id] && <Text size="sm" c="dimmed">Loading...</Text>}
                    {tenantCosts[tenant.id] && (
                      <Table>
                        <Table.Tbody>
                          <Table.Tr>
                            <Table.Td w={180}>Twilio (Voice)</Table.Td>
                            <Table.Td>${(tenantCosts[tenant.id].twilio_cost_cents / 100).toFixed(2)}</Table.Td>
                          </Table.Tr>
                          <Table.Tr>
                            <Table.Td>Deepgram (STT)</Table.Td>
                            <Table.Td>${(tenantCosts[tenant.id].stt_cost_cents / 100).toFixed(2)}</Table.Td>
                          </Table.Tr>
                          <Table.Tr>
                            <Table.Td>OpenAI (LLM)</Table.Td>
                            <Table.Td>${(tenantCosts[tenant.id].llm_cost_cents / 100).toFixed(2)}</Table.Td>
                          </Table.Tr>
                          <Table.Tr>
                            <Table.Td>ElevenLabs (TTS)</Table.Td>
                            <Table.Td>${(tenantCosts[tenant.id].tts_cost_cents / 100).toFixed(2)}</Table.Td>
                          </Table.Tr>
                          <Table.Tr style={{ borderTop: "1px solid var(--mantine-color-gray-3)" }}>
                            <Table.Td fw={500}>API Subtotal</Table.Td>
                            <Table.Td fw={500}>${(tenantCosts[tenant.id].total_api_cost_cents / 100).toFixed(2)}</Table.Td>
                          </Table.Tr>
                          <Table.Tr>
                            <Table.Td>Phone Rental ({tenantCosts[tenant.id].phone_number_count} numbers)</Table.Td>
                            <Table.Td>${(tenantCosts[tenant.id].phone_rental_cents / 100).toFixed(2)}</Table.Td>
                          </Table.Tr>
                          <Table.Tr style={{ backgroundColor: "var(--mantine-color-gray-0)" }}>
                            <Table.Td fw={600}>Total Cost</Table.Td>
                            <Table.Td fw={600}>${(tenantCosts[tenant.id].total_cost_cents / 100).toFixed(2)}</Table.Td>
                          </Table.Tr>
                        </Table.Tbody>
                      </Table>
                    )}
                    {!loadingCosts[tenant.id] && !tenantCosts[tenant.id] && (
                      <Text size="sm" c="dimmed">No cost data available</Text>
                    )}
                  </Box>

                  {/* Admin Notes */}
                  {tenant.admin_notes && (
                    <Box>
                      <Text fw={600} size="sm" mb="xs">
                        Admin Notes
                      </Text>
                      <Paper p="sm" bg="yellow.0" radius="sm">
                        <Text size="sm" style={{ whiteSpace: "pre-wrap" }}>
                          {tenant.admin_notes}
                        </Text>
                      </Paper>
                    </Box>
                  )}

                  {/* Users */}
                  <Box>
                    <Text fw={600} size="sm" mb="xs">
                      Users ({tenant.user_count})
                    </Text>
                    {loadingUsers[tenant.id] && <Text size="sm" c="dimmed">Loading...</Text>}
                    {tenantUsers[tenant.id]?.length === 0 && (
                      <Text size="sm" c="dimmed">No users</Text>
                    )}
                    {tenantUsers[tenant.id] && tenantUsers[tenant.id].length > 0 && (
                      <Table striped>
                        <Table.Thead>
                          <Table.Tr>
                            <Table.Th>Phone</Table.Th>
                            <Table.Th>Name</Table.Th>
                            <Table.Th>Role</Table.Th>
                            <Table.Th>Last Login</Table.Th>
                            <Table.Th w={100}>Actions</Table.Th>
                          </Table.Tr>
                        </Table.Thead>
                        <Table.Tbody>
                          {tenantUsers[tenant.id].map((user) => (
                            <Table.Tr key={user.id}>
                              <Table.Td>{user.phone}</Table.Td>
                              <Table.Td>{user.name || "—"}</Table.Td>
                              <Table.Td>
                                <Badge size="xs" variant="light">
                                  {user.role}
                                </Badge>
                              </Table.Td>
                              <Table.Td>
                                {user.last_login_at
                                  ? formatRelativeTime(new Date(user.last_login_at))
                                  : "Never"}
                              </Table.Td>
                              <Table.Td>
                                <Button
                                  variant="subtle"
                                  color="orange"
                                  size="xs"
                                  leftSection={<IconRefresh size={14} />}
                                  onClick={() => handleResetOnboarding(user, tenant.id)}
                                >
                                  Reset
                                </Button>
                              </Table.Td>
                            </Table.Tr>
                          ))}
                        </Table.Tbody>
                      </Table>
                    )}
                  </Box>

                  {/* Calls */}
                  <Box>
                    <Text fw={600} size="sm" mb="xs">
                      Recent Calls ({tenant.call_count} total)
                    </Text>
                    {loadingCalls[tenant.id] && <Text size="sm" c="dimmed">Loading...</Text>}
                    {tenantCalls[tenant.id]?.length === 0 && (
                      <Text size="sm" c="dimmed">No calls</Text>
                    )}
                    {tenantCalls[tenant.id] && tenantCalls[tenant.id].length > 0 && (
                      <Accordion variant="filled" chevronPosition="left">
                        {tenantCalls[tenant.id].map((call) => (
                          <Accordion.Item key={call.provider_call_id} value={call.provider_call_id}>
                            <Accordion.Control>
                              <Group justify="space-between">
                                <Group gap="sm">
                                  <Text size="sm" fw={500}>
                                    {call.from_number}
                                  </Text>
                                  <Text size="xs" c="dimmed">
                                    {formatRelativeTime(new Date(call.started_at))}
                                  </Text>
                                </Group>
                                <Badge size="xs" variant="light">
                                  {call.status}
                                </Badge>
                              </Group>
                            </Accordion.Control>
                            <Accordion.Panel>
                              {call.screening?.intent_text && (
                                <Text size="sm" c="dimmed" mb="md">
                                  Intent: {call.screening.intent_text}
                                </Text>
                              )}
                              {call.utterances?.length > 0 ? (
                                <Stack gap="sm">
                                  {call.utterances.map((u) => (
                                    <Box
                                      key={u.sequence}
                                      style={{
                                        display: "flex",
                                        justifyContent:
                                          u.speaker === "agent" ? "flex-end" : "flex-start",
                                      }}
                                    >
                                      <Paper
                                        p="sm"
                                        radius="md"
                                        style={{
                                          maxWidth: "80%",
                                          backgroundColor:
                                            u.speaker === "agent"
                                              ? "var(--mantine-color-teal-0)"
                                              : "var(--mantine-color-gray-1)",
                                        }}
                                      >
                                        <Group gap="xs" mb={4}>
                                          <ThemeIcon
                                            size="xs"
                                            color={u.speaker === "agent" ? "teal" : "gray"}
                                            variant="transparent"
                                          >
                                            {u.speaker === "agent" ? (
                                              <IconRobot size={12} />
                                            ) : (
                                              <IconUser size={12} />
                                            )}
                                          </ThemeIcon>
                                          <Text size="xs" fw={500}>
                                            {formatSpeaker(u.speaker)}
                                          </Text>
                                        </Group>
                                        <Text size="sm">{u.text}</Text>
                                      </Paper>
                                    </Box>
                                  ))}
                                </Stack>
                              ) : (
                                <Text size="sm" c="dimmed">
                                  No transcript available
                                </Text>
                              )}
                            </Accordion.Panel>
                          </Accordion.Item>
                        ))}
                      </Accordion>
                    )}
                  </Box>
                </Stack>
              </Accordion.Panel>
            </Accordion.Item>
          ))}
        </Accordion>
      )}

      {/* Edit Modal */}
      <Modal
        opened={editModalOpened}
        onClose={closeEditModal}
        title={`Edit: ${editingTenant?.name}`}
        size="md"
      >
        <Stack gap="md">
          <Group grow>
            <Select
              label="Plan"
              value={editPlan}
              onChange={(v) => setEditPlan(v || "")}
              data={PLAN_OPTIONS}
            />
            <Select
              label="Status"
              value={editStatus}
              onChange={(v) => setEditStatus(v || "")}
              data={STATUS_OPTIONS}
            />
          </Group>
          <NumberInput
            label="Max Turn Timeout (ms)"
            description="Hard timeout for speech_final (default: 4000)"
            placeholder="4000"
            min={1000}
            max={15000}
            step={500}
            value={editMaxTurnTimeout}
            onChange={(val) => setEditMaxTurnTimeout(val === "" ? "" : Number(val))}
          />

          <Divider label="Billing Controls" labelPosition="center" mt="sm" />

          <TextInput
            label="Trial Ends At"
            description="Extend or set trial expiration date"
            type="date"
            value={editTrialEndsAt}
            onChange={(e) => setEditTrialEndsAt(e.currentTarget.value)}
          />

          <NumberInput
            label="Current Period Calls"
            description="Set to 0 to reset call counter for this billing period"
            placeholder="0"
            min={0}
            value={editCurrentPeriodCalls}
            onChange={(val) => setEditCurrentPeriodCalls(val === "" ? "" : Number(val))}
          />

          <Group gap="xs">
            <Button
              variant="light"
              size="xs"
              onClick={() => setEditCurrentPeriodCalls(0)}
            >
              Reset Calls to 0
            </Button>
            <Button
              variant="light"
              size="xs"
              onClick={() => {
                const newDate = new Date();
                newDate.setDate(newDate.getDate() + 14);
                setEditTrialEndsAt(newDate.toISOString().slice(0, 10));
              }}
            >
              +14 Days Trial
            </Button>
          </Group>

          <Textarea
            label="Admin Notes"
            description="Internal notes (only visible to admins)"
            placeholder="Add internal notes about this tenant..."
            minRows={3}
            maxRows={6}
            autosize
            value={editAdminNotes}
            onChange={(e) => setEditAdminNotes(e.currentTarget.value)}
          />

          <Group justify="flex-end" mt="md">
            <Button variant="subtle" onClick={closeEditModal}>
              Cancel
            </Button>
            <Button onClick={handleSaveEdit} loading={saving}>
              Save
            </Button>
          </Group>
        </Stack>
      </Modal>

      {/* Reset Onboarding Modal */}
      <Modal
        opened={resetModalOpened}
        onClose={closeResetModal}
        title="Reset Onboarding"
        size="sm"
      >
        <Stack gap="md">
          <Alert color="orange" variant="light">
            This will reset onboarding for <strong>{resettingUser?.phone}</strong>.
            <Text size="sm" mt="sm">
              <strong>What happens:</strong>
            </Text>
            <Text size="sm" component="ul" style={{ paddingLeft: 20, marginTop: 4 }}>
              <li>User&apos;s phone number will be released back to the pool</li>
              <li>User will need to complete onboarding again</li>
              <li>Tenant and call history are preserved</li>
            </Text>
          </Alert>
          <Group justify="flex-end" mt="md">
            <Button variant="subtle" onClick={closeResetModal}>
              Cancel
            </Button>
            <Button color="orange" onClick={handleConfirmReset} loading={resetting}>
              Reset Onboarding
            </Button>
          </Group>
        </Stack>
      </Modal>

      {/* Delete Tenant Modal */}
      <Modal
        opened={deleteModalOpened}
        onClose={closeDeleteModal}
        title="Delete Tenant"
        size="sm"
      >
        <Stack gap="md">
          <Alert color="red" variant="light">
            This will permanently delete <strong>{deletingTenant?.name}</strong>.
            <Text size="sm" mt="sm">
              <strong>This action cannot be undone. The following will be deleted:</strong>
            </Text>
            <Text size="sm" component="ul" style={{ paddingLeft: 20, marginTop: 4 }}>
              <li>All users ({deletingTenant?.user_count || 0}) associated with this tenant</li>
              <li>All calls ({deletingTenant?.call_count || 0}) and transcripts</li>
              <li>Phone numbers will be unassigned (returned to pool)</li>
              <li>All configuration and settings</li>
            </Text>
          </Alert>
          <Group justify="flex-end" mt="md">
            <Button variant="subtle" onClick={closeDeleteModal}>
              Cancel
            </Button>
            <Button color="red" onClick={handleConfirmDelete} loading={deleting}>
              Delete Tenant
            </Button>
          </Group>
        </Stack>
      </Modal>
      </Stack>
    </Container>
  );
}
