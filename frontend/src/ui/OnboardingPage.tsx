import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import {
  Box,
  Button,
  Container,
  Group,
  Paper,
  Progress,
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
} from "@mantine/core";
import {
  IconRobot,
  IconArrowRight,
  IconCheck,
  IconCopy,
  IconPhone,
  IconConfetti,
  IconAlertCircle,
} from "@tabler/icons-react";
import { api, Tenant, TenantPhoneNumber, setAuthToken } from "../api";
import { useAuth } from "../AuthContext";

type OnboardingStep = 1 | 2 | 3 | 4;

const CARRIER_CODES: Record<string, { noAnswer: string; description: string }> = {
  o2: {
    noAnswer: "**61*{number}#",
    description: "Presmerovani kdyz nezvednes (po 20s)",
  },
  tmobile: {
    noAnswer: "**61*{number}#",
    description: "Presmerovani kdyz nezvednes (po 20s)",
  },
  vodafone: {
    noAnswer: "**61*{number}#",
    description: "Presmerovani kdyz nezvednes (po 20s)",
  },
  other: {
    noAnswer: "**61*{number}#",
    description: "Presmerovani kdyz nezvednes (standardni kod)",
  },
};

export function OnboardingPage() {
  const navigate = useNavigate();
  const { setTenant, refreshUser, user } = useAuth();

  const [step, setStep] = useState<OnboardingStep>(1);
  const [name, setName] = useState("");
  const [, setTenantState] = useState<Tenant | null>(null);
  const [phoneNumbers, setPhoneNumbers] = useState<TenantPhoneNumber[]>([]);
  const [selectedCarrier, setSelectedCarrier] = useState("o2");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const progress = (step / 4) * 100;
  const primaryPhone = phoneNumbers.find((p) => p.is_primary)?.twilio_number;
  const hasPhoneNumber = !!primaryPhone;

  const handleCompleteOnboarding = async () => {
    if (!name.trim()) {
      setError("Zadej sve jmeno");
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await api.completeOnboarding(name.trim());
      // Save the new token that includes the tenant_id
      setAuthToken(response.token);
      setTenantState(response.tenant);
      setTenant(response.tenant);

      // Use phone number from onboarding response if available
      if (response.phone_number) {
        setPhoneNumbers([response.phone_number]);
      } else {
        // Fallback: Load phone numbers from API
        try {
          const tenantData = await api.getTenant();
          setPhoneNumbers(tenantData.phone_numbers || []);
        } catch {
          // Phone numbers might not be assigned yet
        }
      }

      setStep(3);
    } catch (err) {
      setError("Nepodarilo se dokoncit registraci. Zkus to znovu.");
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
        {/* Progress bar */}
        <Progress value={progress} size="sm" mb="xl" radius="xl" />

        <Paper p="xl" radius="md" withBorder>
          {/* Step 1: Welcome */}
          {step === 1 && (
            <Stack gap="xl" align="center" ta="center">
              <ThemeIcon size={80} radius="xl" variant="light" color="blue">
                <IconRobot size={40} />
              </ThemeIcon>

              <Stack gap="xs">
                <Title order={2}>Vitej!</Title>
                <Text c="dimmed" maw={400}>
                  Jsem Karen, tvoje nova telefonni asistentka. Za chvili te provedu nastavenim.
                  Bude to trvat asi 2 minuty.
                </Text>
              </Stack>

              <Button
                size="lg"
                rightSection={<IconArrowRight size={18} />}
                onClick={() => setStep(2)}
              >
                Pojdme na to
              </Button>

              {/* Step indicator */}
              <Group gap="xs">
                {[1, 2, 3, 4].map((s) => (
                  <Box
                    key={s}
                    w={8}
                    h={8}
                    style={{ borderRadius: "50%" }}
                    bg={s === step ? "blue" : "gray.3"}
                  />
                ))}
              </Group>
            </Stack>
          )}

          {/* Step 2: Name */}
          {step === 2 && (
            <Stack gap="xl">
              <Stack gap="xs" ta="center">
                <Title order={2}>Jak se jmenujes?</Title>
                <Text c="dimmed">Karen bude oslovovat volajici tvym jmenem</Text>
              </Stack>

              {error && (
                <Alert icon={<IconAlertCircle size={16} />} color="red" variant="light">
                  {error}
                </Alert>
              )}

              <TextInput
                size="lg"
                placeholder="Lukas"
                value={name}
                onChange={(e) => {
                  setName(e.target.value);
                  setError(null);
                }}
                autoFocus
              />

              <Paper p="md" radius="sm" bg="gray.0">
                <Text size="sm" c="dimmed">
                  Karen bude rikat: "{name || "Lukas"} ted nemuze prijmout hovor, mohu vam pomoct?"
                </Text>
              </Paper>

              <Button
                size="lg"
                fullWidth
                rightSection={<IconArrowRight size={18} />}
                onClick={handleCompleteOnboarding}
                loading={isLoading}
              >
                Pokracovat
              </Button>

              {/* Step indicator */}
              <Group gap="xs" justify="center">
                {[1, 2, 3, 4].map((s) => (
                  <Box
                    key={s}
                    w={8}
                    h={8}
                    style={{ borderRadius: "50%" }}
                    bg={s <= step ? "blue" : "gray.3"}
                  />
                ))}
              </Group>
            </Stack>
          )}

          {/* Step 3: Phone number & forwarding */}
          {step === 3 && (
            <Stack gap="xl">
              <Stack gap="xs" ta="center">
                <Title order={2}>{hasPhoneNumber ? "Tvoje Karen cislo" : "Skoro hotovo!"}</Title>
                <Text c="dimmed">
                  {hasPhoneNumber
                    ? "Na toto cislo presmerujes hovory kdyz budes nedostupny"
                    : "Cislo ti pridelime co nejdrive"}
                </Text>
              </Stack>

              {hasPhoneNumber ? (
                <>
                  {/* Karen number */}
                  <Paper p="lg" radius="md" bg="blue.0" ta="center">
                    <Text size="xl" fw={700} c="blue.8">
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
                          {copied ? "Skopirovano" : "Kopirovat"}
                        </Button>
                      )}
                    </CopyButton>
                  </Paper>

                  {/* Carrier selection */}
                  <Stack gap="xs">
                    <Text size="sm" fw={500}>
                      Vyber sveho operatora:
                    </Text>
                    <SegmentedControl
                      fullWidth
                      value={selectedCarrier}
                      onChange={setSelectedCarrier}
                      data={[
                        { label: "O2", value: "o2" },
                        { label: "T-Mobile", value: "tmobile" },
                        { label: "Vodafone", value: "vodafone" },
                        { label: "Jiny", value: "other" },
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
                        1. Otevri aplikaci Telefon
                      </Text>
                      <Group>
                        <Text size="sm" c="dimmed">
                          2. Vytoc:
                        </Text>
                        <Text size="sm" fw={600} ff="monospace">
                          {getDialCode()}
                        </Text>
                        <CopyButton value={getDialCode()}>
                          {({ copied, copy }) => (
                            <Tooltip label={copied ? "Skopirovano" : "Kopirovat"}>
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
                        3. Uslysite potvrzeni "Sluzba aktivovana"
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
                      Vytocit automaticky
                    </Button>
                  </Anchor>

                  <Button
                    size="lg"
                    fullWidth
                    rightSection={<IconArrowRight size={18} />}
                    onClick={() => setStep(4)}
                  >
                    Hotovo, presmerovani funguje
                  </Button>
                </>
              ) : (
                <>
                  {/* No phone number available */}
                  <Alert icon={<IconAlertCircle size={16} />} color="yellow" variant="light">
                    Momentalne nemame volne cislo. Jakmile bude dostupne, priradime ti ho a oznamime ti to.
                    Presmerovani nastavis v nastaveni.
                  </Alert>

                  <Button
                    size="lg"
                    fullWidth
                    rightSection={<IconArrowRight size={18} />}
                    onClick={() => setStep(4)}
                  >
                    Pokracovat
                  </Button>
                </>
              )}

              {/* Step indicator */}
              <Group gap="xs" justify="center">
                {[1, 2, 3, 4].map((s) => (
                  <Box
                    key={s}
                    w={8}
                    h={8}
                    style={{ borderRadius: "50%" }}
                    bg={s <= step ? "blue" : "gray.3"}
                  />
                ))}
              </Group>
            </Stack>
          )}

          {/* Step 4: Test & Complete */}
          {step === 4 && (
            <Stack gap="xl" align="center" ta="center">
              <ThemeIcon size={80} radius="xl" variant="light" color="green">
                <IconConfetti size={40} />
              </ThemeIcon>

              <Stack gap="xs">
                <Title order={2}>Hotovo!</Title>
                <Text c="dimmed" maw={400}>
                  Karen je pripravena prijimat hovory. Az nekdo zavola a ty nezvednes, Karen to
                  vyridi za tebe.
                </Text>
              </Stack>

              <Paper p="md" radius="md" bg="gray.0" w="100%">
                <Stack gap="sm">
                  <Text size="sm" fw={500}>
                    Tip: Otestuj to!
                  </Text>
                  <Text size="sm" c="dimmed">
                    Zavolej na sve cislo z jineho telefonu a nech to vyzvanet. Karen by mela
                    zvednout po 20 sekundach.
                  </Text>
                </Stack>
              </Paper>

              <Button
                size="lg"
                fullWidth
                rightSection={<IconArrowRight size={18} />}
                onClick={handleFinish}
              >
                Jit do prehledu hovoru
              </Button>

              {/* Step indicator */}
              <Group gap="xs">
                {[1, 2, 3, 4].map((s) => (
                  <Box
                    key={s}
                    w={8}
                    h={8}
                    style={{ borderRadius: "50%" }}
                    bg={s <= step ? "blue" : "gray.3"}
                  />
                ))}
              </Group>
            </Stack>
          )}
        </Paper>
      </Container>
    </Box>
  );
}
