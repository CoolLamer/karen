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
  Accordion,
  List,
  Spoiler,
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
} from "@tabler/icons-react";
import { api, Tenant, TenantPhoneNumber } from "../api";
import { useAuth } from "../AuthContext";
import {
  REDIRECT_CODES,
  REDIRECT_ORDER,
  PHONE_SETTINGS_INSTRUCTIONS,
  getDialCode as getRedirectDialCode,
  getDeactivationCode,
  type RedirectType,
} from "../constants/redirectCodes";

export function SettingsPage() {
  const navigate = useNavigate();
  const { user, tenant, logout, setTenant, isAdmin } = useAuth();

  const [phoneNumbers, setPhoneNumbers] = useState<TenantPhoneNumber[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [logoutModalOpen, setLogoutModalOpen] = useState(false);
  const [forwardingModalOpen, setForwardingModalOpen] = useState(false);
  const [showGreetingWarning, setShowGreetingWarning] = useState(false);
  const [warningOldName, setWarningOldName] = useState("");

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
      const data = await api.getTenant();
      setPhoneNumbers(data.phone_numbers || []);

      // Update form state with fresh data
      if (data.tenant) {
        setName(data.tenant.name || "");
        setOriginalName(data.tenant.name || "");
        setGreetingText(data.tenant.greeting_text || "");
        setVipNames(data.tenant.vip_names || []);
        setMarketingEmail(data.tenant.marketing_email || "");
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

  const karenNumber = phoneNumbers.find((p) => p.is_primary)?.twilio_number || "";

  const planLabel = {
    trial: "Trial",
    basic: "Basic",
    pro: "Pro",
  }[tenant?.plan || "trial"];

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

              <Group justify="space-between">
                <Group gap="xs">
                  <Text size="sm">Plán:</Text>
                  <Badge variant="light">{planLabel}</Badge>
                </Group>
                <Button variant="light" size="xs">
                  Upgradovat
                </Button>
              </Group>
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
      <Modal
        opened={forwardingModalOpen}
        onClose={() => setForwardingModalOpen(false)}
        title="Jak nastavit přesměrování"
        centered
        size="lg"
      >
        <Stack gap="md">
          <Alert color="blue" variant="light">
            <Text size="sm">
              Přesměrování se nastavuje vytočením speciálního kódu na telefonu.
              Otevři tuto stránku na mobilu a klikni na tlačítko – automaticky se vytočí
              aktivační kód a na obrazovce uvidíš potvrzení od operátora.
            </Text>
            <Text size="sm" mt="xs" c="dimmed">
              Na počítači tlačítko nefunguje – musíš kód vytočit ručně nebo otevřít stránku na telefonu.
            </Text>
          </Alert>
          <Text size="sm" c="dimmed">
            Pro kompletní pokrytí doporučujeme nastavit všechny tři typy přesměrování.
          </Text>

          <Accordion variant="separated" defaultValue="noAnswer">
            {REDIRECT_ORDER.map((type) => {
              const code = REDIRECT_CODES[type];
              const dialCode = karenNumber ? getRedirectDialCode(type, karenNumber) : "";
              const deactivateCode = getDeactivationCode(type);
              return (
                <Accordion.Item key={type} value={type}>
                  <Accordion.Control>
                    <Group>
                      <Text fw={500}>{code.label}</Text>
                    </Group>
                  </Accordion.Control>
                  <Accordion.Panel>
                    <Stack gap="sm">
                      <Text size="sm" c="dimmed">{code.description}</Text>

                      <Button
                        variant="light"
                        leftSection={<IconPhone size={14} />}
                        disabled={!karenNumber}
                        onClick={() => { window.location.href = `tel:${dialCode}`; }}
                      >
                        Aktivovat přesměrování
                      </Button>
                      <Spoiler maxHeight={0} showLabel="Zobrazit kód" hideLabel="Skrýt kód">
                        <Group gap="xs">
                          <Text size="sm" ff="monospace">{dialCode}</Text>
                          <CopyButton value={dialCode}>
                            {({ copied, copy }) => (
                              <Tooltip label={copied ? "Zkopírováno" : "Kopírovat"}>
                                <ActionIcon size="sm" variant="subtle" onClick={copy} color={copied ? "green" : "gray"}>
                                  {copied ? <IconCheck size={14} /> : <IconCopy size={14} />}
                                </ActionIcon>
                              </Tooltip>
                            )}
                          </CopyButton>
                        </Group>
                      </Spoiler>

                      <Button
                        variant="subtle"
                        size="xs"
                        color="red"
                        leftSection={<IconPhone size={14} />}
                        onClick={() => { window.location.href = `tel:${deactivateCode}`; }}
                        mt="xs"
                      >
                        Zrušit přesměrování
                      </Button>
                      <Spoiler maxHeight={0} showLabel="Zobrazit kód" hideLabel="Skrýt kód">
                        <Group gap="xs">
                          <Text size="sm" ff="monospace">{deactivateCode}</Text>
                          <CopyButton value={deactivateCode}>
                            {({ copied, copy }) => (
                              <Tooltip label={copied ? "Zkopírováno" : "Kopírovat"}>
                                <ActionIcon size="sm" variant="subtle" onClick={copy} color={copied ? "green" : "gray"}>
                                  {copied ? <IconCheck size={14} /> : <IconCopy size={14} />}
                                </ActionIcon>
                              </Tooltip>
                            )}
                          </CopyButton>
                        </Group>
                      </Spoiler>
                    </Stack>
                  </Accordion.Panel>
                </Accordion.Item>
              );
            })}

            {/* Phone settings alternative */}
            <Accordion.Item value="phone-settings">
              <Accordion.Control>
                <Text fw={500}>Nastavení v telefonu (alternativa)</Text>
              </Accordion.Control>
              <Accordion.Panel>
                <Stack gap="md">
                  <Text size="sm" c="dimmed">
                    Místo vytáčení kódů můžeš přesměrování nastavit přímo v nastavení telefonu.
                  </Text>

                  <Box>
                    <Text size="sm" fw={500}>{PHONE_SETTINGS_INSTRUCTIONS.iphone.title}</Text>
                    <List size="sm" c="dimmed" mt="xs">
                      {PHONE_SETTINGS_INSTRUCTIONS.iphone.steps.map((step, i) => (
                        <List.Item key={i}>{step}</List.Item>
                      ))}
                    </List>
                  </Box>

                  <Box>
                    <Text size="sm" fw={500}>{PHONE_SETTINGS_INSTRUCTIONS.android.title}</Text>
                    <List size="sm" c="dimmed" mt="xs">
                      {PHONE_SETTINGS_INSTRUCTIONS.android.steps.map((step, i) => (
                        <List.Item key={i}>{step}</List.Item>
                      ))}
                    </List>
                  </Box>
                </Stack>
              </Accordion.Panel>
            </Accordion.Item>
          </Accordion>
        </Stack>
      </Modal>

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
    </Box>
  );
}
