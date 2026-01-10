import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Box,
  Button,
  Container,
  Group,
  Paper,
  Stack,
  Text,
  TextInput,
  Title,
  ThemeIcon,
  SegmentedControl,
  CopyButton,
  ActionIcon,
  Tooltip,
  Alert,
  Anchor,
  Stepper,
} from "@mantine/core";
import {
  IconRobot,
  IconArrowRight,
  IconCheck,
  IconCopy,
  IconPhone,
  IconConfetti,
  IconAlertCircle,
  IconUser,
  IconDeviceMobile,
} from "@tabler/icons-react";
import { api, Tenant, TenantPhoneNumber, setAuthToken } from "../api";
import { useAuth } from "../AuthContext";

type OnboardingStep = 0 | 1 | 2 | 3;

const CARRIER_CODES: Record<string, { noAnswer: string; description: string }> = {
  o2: {
    noAnswer: "**61*{number}#",
    description: "Přesměrování když nezvedneš (po 20s)",
  },
  tmobile: {
    noAnswer: "**61*{number}#",
    description: "Přesměrování když nezvedneš (po 20s)",
  },
  vodafone: {
    noAnswer: "**61*{number}#",
    description: "Přesměrování když nezvedneš (po 20s)",
  },
  other: {
    noAnswer: "**61*{number}#",
    description: "Přesměrování když nezvedneš (standardní kód)",
  },
};

export function OnboardingPage() {
  const navigate = useNavigate();
  const { setTenant, refreshUser } = useAuth();

  const [step, setStep] = useState<OnboardingStep>(0);
  const [name, setName] = useState("");
  const [, setTenantState] = useState<Tenant | null>(null);
  const [phoneNumbers, setPhoneNumbers] = useState<TenantPhoneNumber[]>([]);
  const [selectedCarrier, setSelectedCarrier] = useState("o2");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const primaryPhone = phoneNumbers.find((p) => p.is_primary)?.twilio_number;
  const hasPhoneNumber = !!primaryPhone;

  const handleCompleteOnboarding = async () => {
    if (!name.trim()) {
      setError("Zadej své jméno");
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await api.completeOnboarding(name.trim());
      setAuthToken(response.token);
      setTenantState(response.tenant);
      setTenant(response.tenant);

      if (response.phone_number) {
        setPhoneNumbers([response.phone_number]);
      } else {
        try {
          const tenantData = await api.getTenant();
          setPhoneNumbers(tenantData.phone_numbers || []);
        } catch {
          // Phone numbers might not be assigned yet
        }
      }

      setStep(2);
    } catch {
      setError("Nepodařilo se dokončit registraci. Zkus to znovu.");
    } finally {
      setIsLoading(false);
    }
  };

  const getDialCode = () => {
    if (!primaryPhone) return "";
    const carrier = CARRIER_CODES[selectedCarrier];
    return carrier.noAnswer.replace("{number}", primaryPhone.replace(/\s/g, ""));
  };

  const handleFinish = async () => {
    await refreshUser();
    navigate("/");
  };

  return (
    <Box mih="100vh" bg="gray.0">
      <Container size="sm" py={40}>
        {/* Stepper */}
        <Stepper active={step} mb="xl" size="sm">
          <Stepper.Step label="Vítej" icon={<IconRobot size={18} />} />
          <Stepper.Step label="Jméno" icon={<IconUser size={18} />} />
          <Stepper.Step label="Číslo" icon={<IconDeviceMobile size={18} />} />
          <Stepper.Step label="Hotovo" icon={<IconCheck size={18} />} />
        </Stepper>

        <Paper p="xl" radius="md" withBorder>
          {/* Step 0: Welcome */}
          {step === 0 && (
            <Stack gap="xl" align="center" ta="center">
              <ThemeIcon size={80} radius="xl" variant="light" color="teal">
                <IconRobot size={40} />
              </ThemeIcon>

              <Stack gap="xs">
                <Title order={2}>Vítej!</Title>
                <Text c="dimmed" maw={400}>
                  Jsem Karen, tvoje nová telefonní asistentka. Za chvíli tě provedu nastavením.
                  Bude to trvat asi 2 minuty.
                </Text>
              </Stack>

              <Button
                size="lg"
                rightSection={<IconArrowRight size={18} />}
                onClick={() => setStep(1)}
              >
                Pojďme na to
              </Button>
            </Stack>
          )}

          {/* Step 1: Name */}
          {step === 1 && (
            <Stack gap="xl">
              <Stack gap="xs" ta="center">
                <Title order={2}>Jak se jmenuješ?</Title>
                <Text c="dimmed">Karen bude oslovovat volající tvým jménem</Text>
              </Stack>

              {error && (
                <Alert icon={<IconAlertCircle size={16} />} color="red" variant="light">
                  {error}
                </Alert>
              )}

              <TextInput
                size="lg"
                placeholder="Lukáš"
                value={name}
                onChange={(e) => {
                  setName(e.target.value);
                  setError(null);
                }}
                autoFocus
              />

              <Paper p="md" radius="md" bg="gray.0">
                <Text size="sm" c="dimmed">
                  Karen bude říkat: „{name || "Lukáš"} teď nemůže přijmout hovor, mohu vám pomoct?"
                </Text>
              </Paper>

              <Button
                size="lg"
                fullWidth
                rightSection={<IconArrowRight size={18} />}
                onClick={handleCompleteOnboarding}
                loading={isLoading}
              >
                Pokračovat
              </Button>
            </Stack>
          )}

          {/* Step 2: Phone number & forwarding */}
          {step === 2 && (
            <Stack gap="xl">
              <Stack gap="xs" ta="center">
                <Title order={2}>{hasPhoneNumber ? "Tvoje Zvednu číslo" : "Skoro hotovo!"}</Title>
                <Text c="dimmed">
                  {hasPhoneNumber
                    ? "Na toto číslo přesměruješ hovory když budeš nedostupný"
                    : "Číslo ti přidělíme co nejdříve"}
                </Text>
              </Stack>

              {hasPhoneNumber ? (
                <>
                  {/* Karen number */}
                  <Paper p="lg" radius="md" style={{ backgroundColor: "var(--mantine-color-teal-0)" }} ta="center">
                    <Text size="xl" fw={700} c="teal.8">
                      {primaryPhone}
                    </Text>
                    <CopyButton value={primaryPhone.replace(/\s/g, "")}>
                      {({ copied, copy }) => (
                        <Button
                          variant="subtle"
                          size="xs"
                          leftSection={copied ? <IconCheck size={14} /> : <IconCopy size={14} />}
                          onClick={copy}
                          mt="xs"
                        >
                          {copied ? "Zkopírováno" : "Kopírovat"}
                        </Button>
                      )}
                    </CopyButton>
                  </Paper>

                  {/* Carrier selection */}
                  <Stack gap="xs">
                    <Text size="sm" fw={500}>
                      Vyber svého operátora:
                    </Text>
                    <SegmentedControl
                      fullWidth
                      value={selectedCarrier}
                      onChange={setSelectedCarrier}
                      data={[
                        { label: "O2", value: "o2" },
                        { label: "T-Mobile", value: "tmobile" },
                        { label: "Vodafone", value: "vodafone" },
                        { label: "Jiný", value: "other" },
                      ]}
                    />
                  </Stack>

                  {/* Forwarding instructions */}
                  <Paper p="md" radius="md" withBorder>
                    <Stack gap="md">
                      <Text size="sm" fw={500}>
                        {CARRIER_CODES[selectedCarrier].description}
                      </Text>
                      <Text size="sm" c="dimmed">
                        1. Otevři aplikaci Telefon
                      </Text>
                      <Group>
                        <Text size="sm" c="dimmed">
                          2. Vytoč:
                        </Text>
                        <Text size="sm" fw={600} ff="monospace">
                          {getDialCode()}
                        </Text>
                        <CopyButton value={getDialCode()}>
                          {({ copied, copy }) => (
                            <Tooltip label={copied ? "Zkopírováno" : "Kopírovat"}>
                              <ActionIcon
                                size="sm"
                                variant="subtle"
                                onClick={copy}
                                color={copied ? "green" : "gray"}
                              >
                                {copied ? <IconCheck size={14} /> : <IconCopy size={14} />}
                              </ActionIcon>
                            </Tooltip>
                          )}
                        </CopyButton>
                      </Group>
                      <Text size="sm" c="dimmed">
                        3. Uslyšíte potvrzení „Služba aktivována"
                      </Text>
                    </Stack>
                  </Paper>

                  <Anchor
                    href={`tel:${getDialCode()}`}
                    style={{ textDecoration: "none" }}
                  >
                    <Button
                      variant="light"
                      fullWidth
                      leftSection={<IconPhone size={18} />}
                    >
                      Vytočit automaticky
                    </Button>
                  </Anchor>

                  <Button
                    size="lg"
                    fullWidth
                    rightSection={<IconArrowRight size={18} />}
                    onClick={() => setStep(3)}
                  >
                    Hotovo, přesměrování funguje
                  </Button>
                </>
              ) : (
                <>
                  {/* No phone number available */}
                  <Alert icon={<IconAlertCircle size={16} />} color="yellow" variant="light">
                    Momentálně nemáme volné číslo. Jakmile bude dostupné, přiřadíme ti ho a oznámíme ti to.
                    Přesměrování nastavíš v nastavení.
                  </Alert>

                  <Button
                    size="lg"
                    fullWidth
                    rightSection={<IconArrowRight size={18} />}
                    onClick={() => setStep(3)}
                  >
                    Pokračovat
                  </Button>
                </>
              )}
            </Stack>
          )}

          {/* Step 3: Test & Complete */}
          {step === 3 && (
            <Stack gap="xl" align="center" ta="center">
              <ThemeIcon size={80} radius="xl" variant="light" color="green">
                <IconConfetti size={40} />
              </ThemeIcon>

              <Stack gap="xs">
                <Title order={2}>Hotovo!</Title>
                <Text c="dimmed" maw={400}>
                  Karen je připravená přijímat hovory. Až někdo zavolá a ty nezvedneš, Karen to
                  vyřídí za tebe.
                </Text>
              </Stack>

              <Paper p="md" radius="md" bg="gray.0" w="100%">
                <Stack gap="sm">
                  <Text size="sm" fw={500}>
                    Tip: Otestuj to!
                  </Text>
                  <Text size="sm" c="dimmed">
                    Zavolej na své číslo z jiného telefonu a nech to vyzvánět. Karen by měla
                    zvednout po 20 sekundách.
                  </Text>
                </Stack>
              </Paper>

              <Button
                size="lg"
                fullWidth
                rightSection={<IconArrowRight size={18} />}
                onClick={handleFinish}
              >
                Jít do přehledu hovorů
              </Button>
            </Stack>
          )}
        </Paper>
      </Container>
    </Box>
  );
}
