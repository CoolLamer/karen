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
  ThemeIcon,
  Progress,
  SimpleGrid,
  Card,
} from "@mantine/core";
import {
  IconArrowLeft,
  IconCopy,
  IconCheck,
  IconLogout,
  IconAlertCircle,
  IconSettings,
  IconPhone,
  IconUser,
  IconRobot,
  IconCreditCard,
  IconClock,
  IconExternalLink,
  IconVolume,
} from "@tabler/icons-react";
import { ForwardingSetupModal } from "./ForwardingSetupModal";
import { VoicePickerModal, useVoiceName } from "./VoicePickerModal";
import { api, TenantPhoneNumber, BillingInfo } from "../api";
import { useAuth } from "../AuthContext";

export function SettingsPage() {
  const navigate = useNavigate();
  const { user, tenant, logout, setTenant, isAdmin } = useAuth();

  const [phoneNumbers, setPhoneNumbers] = useState<TenantPhoneNumber[]>([]);
  const [billing, setBilling] = useState<BillingInfo | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [isUpgrading, setIsUpgrading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [logoutModalOpen, setLogoutModalOpen] = useState(false);
  const [forwardingModalOpen, setForwardingModalOpen] = useState(false);
  const [upgradeModalOpen, setUpgradeModalOpen] = useState(false);
  const [showGreetingWarning, setShowGreetingWarning] = useState(false);
  const [warningOldName, setWarningOldName] = useState("");
  const [voiceModalOpen, setVoiceModalOpen] = useState(false);
  const voiceName = useVoiceName(tenant?.voice_id);

  // Form state
  const [name, setName] = useState(tenant?.name || "");
  const [originalName, setOriginalName] = useState(tenant?.name || "");
  const [greetingText, setGreetingText] = useState(tenant?.greeting_text || "");
  const [vipNames, setVipNames] = useState<string[]>(tenant?.vip_names || []);
  const [marketingEmail, setMarketingEmail] = useState(tenant?.marketing_email || "");

  useEffect(() => {
    loadTenantData();
  }, []);

  const loadTenantData = async () => {
    try {
      const [tenantData, billingData] = await Promise.all([
        api.getTenant(),
        api.getBilling().catch(() => null),
      ]);
      setPhoneNumbers(tenantData.phone_numbers || []);
      setBilling(billingData);

      // Update form state with fresh data
      if (tenantData.tenant) {
        setName(tenantData.tenant.name || "");
        setOriginalName(tenantData.tenant.name || "");
        setGreetingText(tenantData.tenant.greeting_text || "");
        setVipNames(tenantData.tenant.vip_names || []);
        setMarketingEmail(tenantData.tenant.marketing_email || "");
      }
    } catch {
      setError("Nepodařilo se načíst data");
    } finally {
      setIsLoading(false);
    }
  };

  const handleSave = async () => {
    setIsSaving(true);
    setError(null);
    setSuccess(null);
    setShowGreetingWarning(false);

    try {
      const response = await api.updateTenant({
        name,
        greeting_text: greetingText || undefined,
        vip_names: vipNames.length > 0 ? vipNames : undefined,
        marketing_email: marketingEmail || undefined,
      });

      setTenant(response.tenant);

      // Check if name changed and user has a custom greeting
      const nameChanged = name !== originalName && name.trim() !== "" && originalName.trim() !== "";
      const hasCustomGreeting = greetingText && greetingText.trim() !== "";

      if (nameChanged && hasCustomGreeting) {
        setWarningOldName(originalName);
        setShowGreetingWarning(true);
        setSuccess("Změny byly uloženy");
        // Longer timeout for warning
        setTimeout(() => {
          setSuccess(null);
          setShowGreetingWarning(false);
        }, 10000);
      } else {
        setSuccess("Změny byly uloženy");
        setTimeout(() => setSuccess(null), 3000);
      }

      // Update original name to new value
      setOriginalName(name);
    } catch {
      setError("Nepodařilo se uložit změny");
    } finally {
      setIsSaving(false);
    }
  };

  const handleLogout = async () => {
    await logout();
    navigate("/");
  };

  const handleUpgrade = async (plan: "basic" | "pro", interval: "monthly" | "annual") => {
    setIsUpgrading(true);
    setError(null);
    try {
      const { checkout_url } = await api.createCheckout(plan, interval);
      window.location.href = checkout_url;
    } catch {
      setError("Nepodařilo se vytvořit platební session");
      setIsUpgrading(false);
    }
  };

  const handleManageSubscription = async () => {
    setIsUpgrading(true);
    setError(null);
    try {
      const { portal_url } = await api.createPortal();
      window.location.href = portal_url;
    } catch {
      setError("Nepodařilo se otevřít správu předplatného");
      setIsUpgrading(false);
    }
  };

  const karenNumber = phoneNumbers.find((p) => p.is_primary)?.twilio_number || "";

  const planLabel = {
    trial: "Trial",
    basic: "Základ",
    pro: "Pro",
  }[billing?.plan || tenant?.plan || "trial"];

  const formatTimeSaved = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    if (hours > 0) {
      return `${hours}h ${minutes}min`;
    }
    return `${minutes}min`;
  };

  const usagePercentage = billing?.call_status
    ? billing.call_status.calls_limit > 0
      ? (billing.call_status.calls_used / billing.call_status.calls_limit) * 100
      : 0
    : 0;

  if (isLoading) {
    return (
      <Container size="sm" py="xl">
        <Text c="dimmed">Načítám...</Text>
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
              Zpět
            </Button>
            <Title order={2}>Nastavení</Title>
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

          {showGreetingWarning && (
            <Alert
              icon={<IconAlertCircle size={16} />}
              color="yellow"
              variant="light"
              title="Zkontroluj pozdrav"
              withCloseButton
              onClose={() => setShowGreetingWarning(false)}
            >
              <Stack gap="xs">
                <Text size="sm">
                  Změnil/a jsi jméno z &quot;{warningOldName}&quot; na &quot;{name}&quot;. Tvůj vlastní pozdrav může stále obsahovat staré jméno.
                </Text>
                <Button
                  variant="light"
                  size="xs"
                  onClick={() => {
                    document.getElementById("greeting-textarea")?.focus();
                    setShowGreetingWarning(false);
                  }}
                >
                  Upravit pozdrav
                </Button>
              </Stack>
            </Alert>
          )}

          {/* Profile section */}
          <Paper p="lg" radius="md" withBorder>
            <Stack gap="md">
              <Group gap="xs">
                <ThemeIcon size="sm" variant="light" color="teal">
                  <IconUser size={14} />
                </ThemeIcon>
                <Text size="sm" fw={600} tt="uppercase" c="dimmed">
                  Profil
                </Text>
              </Group>

              <Group justify="space-between">
                <Text size="sm">Telefon</Text>
                <Text size="sm" fw={500}>
                  {user?.phone}
                </Text>
              </Group>

              <TextInput
                label="Jméno"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </Stack>
          </Paper>

          {/* Karen number section */}
          <Paper p="lg" radius="md" withBorder>
            <Stack gap="md">
              <Group gap="xs">
                <ThemeIcon size="sm" variant="light" color="teal">
                  <IconPhone size={14} />
                </ThemeIcon>
                <Text size="sm" fw={600} tt="uppercase" c="dimmed">
                  Zvednu číslo
                </Text>
              </Group>

              {karenNumber ? (
                <>
                  <Group justify="space-between">
                    <Text size="lg" fw={600}>
                      {karenNumber}
                    </Text>
                    <Group gap="xs">
                      <CopyButton value={karenNumber.replace(/\s/g, "")}>
                        {({ copied, copy }) => (
                          <Tooltip label={copied ? "Zkopírováno" : "Kopírovat"}>
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
                      <Tooltip label="Zavolat">
                        <ActionIcon
                          variant="subtle"
                          component="a"
                          href={`tel:${karenNumber.replace(/\s/g, "")}`}
                        >
                          <IconPhone size={16} />
                        </ActionIcon>
                      </Tooltip>
                    </Group>
                  </Group>

                  <Button variant="light" size="xs" onClick={() => setForwardingModalOpen(true)}>
                    Jak nastavit přesměrování
                  </Button>
                </>
              ) : (
                <Alert icon={<IconAlertCircle size={16} />} color="yellow" variant="light">
                  Zatím ti nebylo přiřazeno telefonní číslo. Jakmile bude dostupné, oznámíme ti to.
                </Alert>
              )}
            </Stack>
          </Paper>

          {/* Assistant settings */}
          <Paper p="lg" radius="md" withBorder>
            <Stack gap="md">
              <Group gap="xs">
                <ThemeIcon size="sm" variant="light" color="teal">
                  <IconRobot size={14} />
                </ThemeIcon>
                <Text size="sm" fw={600} tt="uppercase" c="dimmed">
                  Asistentka
                </Text>
              </Group>

              <Textarea
                id="greeting-textarea"
                label="Pozdrav"
                description="Text, kterým Karen začíná hovor"
                placeholder="Dobrý den, tady Karen, asistentka..."
                value={greetingText}
                onChange={(e) => setGreetingText(e.target.value)}
                minRows={2}
              />

              <TagsInput
                label="VIP kontakty"
                description="Jména osob, které Karen vždy přepojí (např. rodina)"
                placeholder="Přidej jméno a stiskni Enter"
                value={vipNames}
                onChange={setVipNames}
              />

              <TextInput
                label="Marketing email"
                description="Email, kam Karen odkáže marketingové volající"
                placeholder="nabidky@email.cz"
                type="email"
                value={marketingEmail}
                onChange={(e) => setMarketingEmail(e.target.value)}
              />

              <TextInput
                label="Číslo pro přesměrování"
                description="Hovory budou přesměrovány na vaše registrované číslo"
                type="tel"
                value={user?.phone || ""}
                disabled
                styles={{
                  input: {
                    backgroundColor: "var(--mantine-color-gray-1)",
                    cursor: "not-allowed",
                  },
                }}
              />
            </Stack>
          </Paper>

          {/* Voice selection */}
          <Paper p="lg" radius="md" withBorder>
            <Stack gap="md">
              <Group gap="xs">
                <ThemeIcon size="sm" variant="light" color="teal">
                  <IconVolume size={14} />
                </ThemeIcon>
                <Text size="sm" fw={600} tt="uppercase" c="dimmed">
                  Hlas asistentky
                </Text>
              </Group>

              <Group justify="space-between" align="center">
                <div>
                  <Text size="sm" fw={500}>
                    {voiceName}
                  </Text>
                  <Text size="xs" c="dimmed">
                    Hlas, kterým Karen mluví
                  </Text>
                </div>
                <Button variant="light" size="xs" onClick={() => setVoiceModalOpen(true)}>
                  Změnit
                </Button>
              </Group>
            </Stack>
          </Paper>

          {/* Save button */}
          <Button size="lg" onClick={handleSave} loading={isSaving}>
            Uložit změny
          </Button>

          {/* Subscription */}
          <Paper p="lg" radius="md" withBorder>
            <Stack gap="md">
              <Group gap="xs">
                <ThemeIcon size="sm" variant="light" color="teal">
                  <IconCreditCard size={14} />
                </ThemeIcon>
                <Text size="sm" fw={600} tt="uppercase" c="dimmed">
                  Předplatné
                </Text>
              </Group>

              {/* Current Plan Info */}
              <Group justify="space-between" align="flex-start">
                <Stack gap="xs">
                  <Group gap="xs">
                    <Text size="sm">Plán:</Text>
                    <Badge
                      variant="light"
                      color={billing?.plan === "pro" ? "violet" : billing?.plan === "basic" ? "blue" : "gray"}
                    >
                      {planLabel}
                    </Badge>
                    {billing?.status === "past_due" && (
                      <Badge color="red" variant="filled" size="xs">
                        Nezaplaceno
                      </Badge>
                    )}
                  </Group>

                  {/* Trial Info */}
                  {billing?.plan === "trial" && billing?.call_status && (
                    <Stack gap={4}>
                      <Text size="xs" c="dimmed">
                        {billing.call_status.trial_calls_left !== undefined && (
                          <>Zbývá {billing.call_status.trial_calls_left} hovorů</>
                        )}
                        {billing.call_status.trial_days_left !== undefined && (
                          <> • {billing.call_status.trial_days_left} dní</>
                        )}
                      </Text>
                      <Progress
                        value={usagePercentage}
                        color={usagePercentage >= 100 ? "red" : usagePercentage >= 80 ? "yellow" : "blue"}
                        size="xs"
                      />
                    </Stack>
                  )}
                </Stack>

                {billing?.plan === "trial" || billing?.plan === undefined ? (
                  <Button
                    variant="light"
                    size="xs"
                    onClick={() => setUpgradeModalOpen(true)}
                    loading={isUpgrading}
                  >
                    Upgradovat
                  </Button>
                ) : (
                  <Button
                    variant="subtle"
                    size="xs"
                    rightSection={<IconExternalLink size={14} />}
                    onClick={handleManageSubscription}
                    loading={isUpgrading}
                  >
                    Spravovat
                  </Button>
                )}
              </Group>

              {/* Time Saved */}
              {billing && billing.total_time_saved > 0 && (
                <Paper p="sm" bg="teal.0" radius="md">
                  <Group gap="xs">
                    <ThemeIcon size="md" variant="light" color="teal">
                      <IconClock size={16} />
                    </ThemeIcon>
                    <div>
                      <Text size="sm" c="dimmed">
                        Karen ti ušetřila
                      </Text>
                      <Text size="lg" fw={600} c="teal">
                        {formatTimeSaved(billing.total_time_saved)}
                      </Text>
                    </div>
                  </Group>
                </Paper>
              )}

              {/* Trial Expired Warning */}
              {billing?.call_status && !billing.call_status.can_receive && (
                <Alert
                  icon={<IconAlertCircle size={16} />}
                  color="red"
                  variant="light"
                  title="Trial vypršel"
                >
                  {billing.call_status.reason === "limit_exceeded"
                    ? "Dosáhli jste limitu hovorů. Karen nebude přijímat nové hovory."
                    : "Váš trial skončil. Karen nebude přijímat nové hovory."}
                  <Button
                    variant="filled"
                    size="xs"
                    mt="xs"
                    onClick={() => setUpgradeModalOpen(true)}
                  >
                    Upgradovat nyní
                  </Button>
                </Alert>
              )}
            </Stack>
          </Paper>

          <Divider />

          {/* Admin link (only visible to admins) */}
          {isAdmin && (
            <Button
              variant="subtle"
              c="dimmed"
              leftSection={<IconSettings size={16} />}
              onClick={() => navigate("/admin")}
            >
              Admin Panel
            </Button>
          )}

          {/* Logout */}
          <Button
            variant="subtle"
            color="red"
            leftSection={<IconLogout size={16} />}
            onClick={() => setLogoutModalOpen(true)}
          >
            Odhlásit se
          </Button>
        </Stack>
      </Container>

      {/* Forwarding instructions modal */}
      <ForwardingSetupModal
        opened={forwardingModalOpen}
        onClose={() => setForwardingModalOpen(false)}
        karenNumber={karenNumber}
      />

      {/* Voice picker modal */}
      <VoicePickerModal
        opened={voiceModalOpen}
        onClose={() => setVoiceModalOpen(false)}
        currentVoiceId={tenant?.voice_id}
        onSelect={(voiceId) => {
          if (tenant) {
            setTenant({ ...tenant, voice_id: voiceId });
          }
        }}
      />

      {/* Logout confirmation modal */}
      <Modal
        opened={logoutModalOpen}
        onClose={() => setLogoutModalOpen(false)}
        title="Odhlásit se?"
        centered
      >
        <Stack gap="md">
          <Text size="sm">Opravdu se chceš odhlásit?</Text>
          <Group justify="flex-end">
            <Button variant="subtle" onClick={() => setLogoutModalOpen(false)}>
              Zrušit
            </Button>
            <Button color="red" onClick={handleLogout}>
              Odhlásit
            </Button>
          </Group>
        </Stack>
      </Modal>

      {/* Upgrade modal */}
      <Modal
        opened={upgradeModalOpen}
        onClose={() => setUpgradeModalOpen(false)}
        title="Vyberte plán"
        centered
        size="lg"
      >
        <Stack gap="lg">
          <SimpleGrid cols={2}>
            {/* Basic Plan */}
            <Card withBorder padding="lg" radius="md">
              <Stack gap="md">
                <div>
                  <Text size="lg" fw={600}>
                    Základ
                  </Text>
                  <Text size="sm" c="dimmed">
                    Pro OSVČ a malé firmy
                  </Text>
                </div>
                <div>
                  <Text size="xl" fw={700}>
                    199 Kč
                    <Text span size="sm" fw={400} c="dimmed">
                      /měsíc
                    </Text>
                  </Text>
                  <Text size="xs" c="dimmed">
                    nebo 159 Kč/měsíc ročně
                  </Text>
                </div>
                <Stack gap={4}>
                  <Text size="sm">✓ 50 hovorů měsíčně</Text>
                  <Text size="sm">✓ Kompletní přepisy</Text>
                  <Text size="sm">✓ SMS notifikace</Text>
                </Stack>
                <Stack gap="xs">
                  <Button
                    variant="filled"
                    onClick={() => handleUpgrade("basic", "monthly")}
                    loading={isUpgrading}
                  >
                    Měsíční platba
                  </Button>
                  <Button
                    variant="light"
                    onClick={() => handleUpgrade("basic", "annual")}
                    loading={isUpgrading}
                  >
                    Roční platba (ušetři 20%)
                  </Button>
                </Stack>
              </Stack>
            </Card>

            {/* Pro Plan */}
            <Card withBorder padding="lg" radius="md" style={{ borderColor: "var(--mantine-color-violet-5)" }}>
              <Stack gap="md">
                <Group justify="space-between">
                  <div>
                    <Text size="lg" fw={600}>
                      Pro
                    </Text>
                    <Text size="sm" c="dimmed">
                      Pro profesionály
                    </Text>
                  </div>
                  <Badge color="violet">Populární</Badge>
                </Group>
                <div>
                  <Text size="xl" fw={700}>
                    499 Kč
                    <Text span size="sm" fw={400} c="dimmed">
                      /měsíc
                    </Text>
                  </Text>
                  <Text size="xs" c="dimmed">
                    nebo 399 Kč/měsíc ročně
                  </Text>
                </div>
                <Stack gap={4}>
                  <Text size="sm">✓ Neomezené hovory</Text>
                  <Text size="sm">✓ VIP přepojování</Text>
                  <Text size="sm">✓ Vlastní hlas</Text>
                  <Text size="sm">✓ Prioritní podpora</Text>
                </Stack>
                <Stack gap="xs">
                  <Button
                    variant="filled"
                    color="violet"
                    onClick={() => handleUpgrade("pro", "monthly")}
                    loading={isUpgrading}
                  >
                    Měsíční platba
                  </Button>
                  <Button
                    variant="light"
                    color="violet"
                    onClick={() => handleUpgrade("pro", "annual")}
                    loading={isUpgrading}
                  >
                    Roční platba (ušetři 20%)
                  </Button>
                </Stack>
              </Stack>
            </Card>
          </SimpleGrid>

          <Text size="xs" c="dimmed" ta="center">
            Platbu zpracovává Stripe. Předplatné můžete kdykoli zrušit.
          </Text>
        </Stack>
      </Modal>
    </Box>
  );
}
