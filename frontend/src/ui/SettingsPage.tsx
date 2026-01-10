import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import {
  Box,
  Button,
  Container,
  Divider,
  Group,
  Paper,
  Stack,
  Text,
  TextInput,
  Textarea,
  Title,
  Badge,
  CopyButton,
  ActionIcon,
  Tooltip,
  TagsInput,
  Alert,
  Modal,
} from "@mantine/core";
import {
  IconArrowLeft,
  IconCopy,
  IconCheck,
  IconLogout,
  IconAlertCircle,
  IconSettings,
} from "@tabler/icons-react";
import { api, Tenant, TenantPhoneNumber } from "../api";
import { useAuth } from "../AuthContext";

export function SettingsPage() {
  const navigate = useNavigate();
  const { user, tenant, logout, setTenant } = useAuth();

  const [phoneNumbers, setPhoneNumbers] = useState<TenantPhoneNumber[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [logoutModalOpen, setLogoutModalOpen] = useState(false);

  // Form state
  const [name, setName] = useState(tenant?.name || "");
  const [greetingText, setGreetingText] = useState(tenant?.greeting_text || "");
  const [vipNames, setVipNames] = useState<string[]>(tenant?.vip_names || []);
  const [marketingEmail, setMarketingEmail] = useState(tenant?.marketing_email || "");
  const [forwardNumber, setForwardNumber] = useState(tenant?.forward_number || "");

  useEffect(() => {
    loadTenantData();
  }, []);

  const loadTenantData = async () => {
    try {
      const data = await api.getTenant();
      setPhoneNumbers(data.phone_numbers || []);

      // Update form state with fresh data
      if (data.tenant) {
        setName(data.tenant.name || "");
        setGreetingText(data.tenant.greeting_text || "");
        setVipNames(data.tenant.vip_names || []);
        setMarketingEmail(data.tenant.marketing_email || "");
        setForwardNumber(data.tenant.forward_number || "");
      }
    } catch {
      setError("Nepodarilo se nacist data");
    } finally {
      setIsLoading(false);
    }
  };

  const handleSave = async () => {
    setIsSaving(true);
    setError(null);
    setSuccess(null);

    try {
      const response = await api.updateTenant({
        name,
        greeting_text: greetingText || undefined,
        vip_names: vipNames.length > 0 ? vipNames : undefined,
        marketing_email: marketingEmail || undefined,
        forward_number: forwardNumber || undefined,
      });

      setTenant(response.tenant);
      setSuccess("Zmeny byly ulozeny");

      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(null), 3000);
    } catch {
      setError("Nepodarilo se ulozit zmeny");
    } finally {
      setIsSaving(false);
    }
  };

  const handleLogout = async () => {
    await logout();
    navigate("/");
  };

  const karenNumber = phoneNumbers.find((p) => p.is_primary)?.twilio_number || "";

  const planLabel = {
    trial: "Trial",
    basic: "Basic",
    pro: "Pro",
  }[tenant?.plan || "trial"];

  if (isLoading) {
    return (
      <Container size="sm" py="xl">
        <Text c="dimmed">Nacitam...</Text>
      </Container>
    );
  }

  return (
    <Box mih="100vh" bg="gray.0">
      <Container size="sm" py="xl">
        <Stack gap="xl">
          {/* Header */}
          <Group justify="space-between">
            <Button
              variant="subtle"
              leftSection={<IconArrowLeft size={16} />}
              onClick={() => navigate("/")}
              px={0}
            >
              Zpet
            </Button>
            <Title order={2}>Nastaveni</Title>
            <Box w={80} /> {/* Spacer for centering */}
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

          {/* Profile section */}
          <Paper p="lg" radius="md" withBorder>
            <Stack gap="md">
              <Text size="sm" fw={600} tt="uppercase" c="dimmed">
                Profil
              </Text>

              <Group justify="space-between">
                <Text size="sm">Telefon</Text>
                <Text size="sm" fw={500}>
                  {user?.phone}
                </Text>
              </Group>

              <TextInput
                label="Jmeno"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </Stack>
          </Paper>

          {/* Karen number section */}
          {karenNumber && (
            <Paper p="lg" radius="md" withBorder>
              <Stack gap="md">
                <Text size="sm" fw={600} tt="uppercase" c="dimmed">
                  Karen cislo
                </Text>

                <Group justify="space-between">
                  <Text size="lg" fw={600}>
                    {karenNumber}
                  </Text>
                  <CopyButton value={karenNumber.replace(/\s/g, "")}>
                    {({ copied, copy }) => (
                      <Tooltip label={copied ? "Skopirovano" : "Kopirovat"}>
                        <ActionIcon
                          variant="subtle"
                          onClick={copy}
                          color={copied ? "green" : "gray"}
                        >
                          {copied ? <IconCheck size={16} /> : <IconCopy size={16} />}
                        </ActionIcon>
                      </Tooltip>
                    )}
                  </CopyButton>
                </Group>

                <Button variant="light" size="xs">
                  Jak nastavit presmerovani
                </Button>
              </Stack>
            </Paper>
          )}

          {/* Assistant settings */}
          <Paper p="lg" radius="md" withBorder>
            <Stack gap="md">
              <Text size="sm" fw={600} tt="uppercase" c="dimmed">
                Asistentka
              </Text>

              <Textarea
                label="Pozdrav"
                description="Text, kterym Karen zacina hovor"
                placeholder="Dobry den, tady Karen, asistentka..."
                value={greetingText}
                onChange={(e) => setGreetingText(e.target.value)}
                minRows={2}
              />

              <TagsInput
                label="VIP kontakty"
                description="Jmena osob, ktere Karen vzdy prepoji (napr. rodina)"
                placeholder="Pridej jmeno a stiskni Enter"
                value={vipNames}
                onChange={setVipNames}
              />

              <TextInput
                label="Marketing email"
                description="Email, kam Karen odkaze marketingove volajici"
                placeholder="nabidky@email.cz"
                type="email"
                value={marketingEmail}
                onChange={(e) => setMarketingEmail(e.target.value)}
              />

              <TextInput
                label="Cislo pro presmerovani"
                description="Cislo, kam Karen prepoji dulezite hovory"
                placeholder="+420 777 123 456"
                type="tel"
                value={forwardNumber}
                onChange={(e) => setForwardNumber(e.target.value)}
              />
            </Stack>
          </Paper>

          {/* Save button */}
          <Button size="lg" onClick={handleSave} loading={isSaving}>
            Ulozit zmeny
          </Button>

          {/* Subscription */}
          <Paper p="lg" radius="md" withBorder>
            <Stack gap="md">
              <Text size="sm" fw={600} tt="uppercase" c="dimmed">
                Predplatne
              </Text>

              <Group justify="space-between">
                <Group gap="xs">
                  <Text size="sm">Plan:</Text>
                  <Badge variant="light">{planLabel}</Badge>
                </Group>
                <Button variant="light" size="xs">
                  Upgradovat
                </Button>
              </Group>
            </Stack>
          </Paper>

          <Divider />

          {/* Admin link (only visible to admins - will show 403 if not admin) */}
          <Button
            variant="subtle"
            c="dimmed"
            leftSection={<IconSettings size={16} />}
            onClick={() => navigate("/admin")}
          >
            Admin Panel
          </Button>

          {/* Logout */}
          <Button
            variant="subtle"
            color="red"
            leftSection={<IconLogout size={16} />}
            onClick={() => setLogoutModalOpen(true)}
          >
            Odhlasit se
          </Button>
        </Stack>
      </Container>

      {/* Logout confirmation modal */}
      <Modal
        opened={logoutModalOpen}
        onClose={() => setLogoutModalOpen(false)}
        title="Odhlasit se?"
        centered
      >
        <Stack gap="md">
          <Text size="sm">Opravdu se chces odhlasit?</Text>
          <Group justify="flex-end">
            <Button variant="subtle" onClick={() => setLogoutModalOpen(false)}>
              Zrusit
            </Button>
            <Button color="red" onClick={handleLogout}>
              Odhlasit
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Box>
  );
}
