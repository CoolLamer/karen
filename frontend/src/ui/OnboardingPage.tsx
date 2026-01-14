import React, { useState, useEffect } from "react";
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
  CopyButton,
  Alert,
  Progress,
  TagsInput,
  Radio,
  List,
  Checkbox,
} from "@mantine/core";
import {
  IconRobot,
  IconArrowRight,
  IconCheck,
  IconCopy,
  IconConfetti,
  IconAlertCircle,
} from "@tabler/icons-react";
import { RedirectSetupAccordion } from "./components/RedirectSetupAccordion";
import { api, Tenant, TenantPhoneNumber, setAuthToken } from "../api";
import { useAuth } from "../AuthContext";
import {
  REDIRECT_CODES,
  REDIRECT_ORDER,
  type RedirectType,
} from "../constants/redirectCodes";

type OnboardingStep = 0 | 1 | 2 | 3 | 4 | 5;

export function OnboardingPage() {
  const navigate = useNavigate();
  const { startOnboarding, finishOnboarding } = useAuth();

  const [step, setStep] = useState<OnboardingStep>(0);
  const [name, setName] = useState("");
  const [, setTenantState] = useState<Tenant | null>(null);
  const [phoneNumbers, setPhoneNumbers] = useState<TenantPhoneNumber[]>([]);
  const [selectedRedirects, setSelectedRedirects] = useState<RedirectType[]>(["noAnswer"]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // New state for VIP contacts and marketing
  const [vipNames, setVipNames] = useState<string[]>([]);
  const [marketingOption, setMarketingOption] = useState<"reject" | "email">("reject");
  const [marketingEmail, setMarketingEmail] = useState("");
  const [_testOption, _setTestOption] = useState<"forwarding" | "direct" | null>(null);

  const primaryPhone = phoneNumbers.find((p) => p.is_primary)?.twilio_number;
  const hasPhoneNumber = !!primaryPhone;

  // Call when component mounts to prevent redirects during onboarding
  useEffect(() => {
    startOnboarding();
  }, [startOnboarding]);

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
      // Don't call setTenant here - it would set needsOnboarding=false and
      // trigger a redirect before onboarding steps are complete

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

  const toggleRedirect = (type: RedirectType) => {
    setSelectedRedirects((prev) =>
      prev.includes(type) ? prev.filter((t) => t !== type) : [...prev, type]
    );
  };

  // Save VIP names and marketing email after configuration steps
  const handleSaveConfiguration = async () => {
    // Only save if user configured something
    if (vipNames.length > 0 || (marketingOption === "email" && marketingEmail)) {
      try {
        const updateData: Partial<Tenant> = {};
        if (vipNames.length > 0) {
          updateData.vip_names = vipNames;
        }
        if (marketingOption === "email" && marketingEmail) {
          updateData.marketing_email = marketingEmail;
        }
        await api.updateTenant(updateData);
        // Don't update auth context here - refreshUser() in handleFinish() will do it
      } catch {
        // Non-critical, continue anyway - user can configure later
      }
    }
    setStep(4);
  };

  const handleFinish = async () => {
    await finishOnboarding(); // Clear flag and refresh
    navigate("/");
  };

  return (
    <Box mih="100vh" bg="gray.0">
      <Container size="sm" py={40}>
        {/* Progress indicator */}
        {step > 0 && step < 5 && (
          <Stack gap="xs" mb="xl">
            <Group justify="space-between">
              <Text size="sm" c="dimmed">Krok {step} z 5</Text>
              <Text size="sm" c="dimmed">{Math.round((step / 5) * 100)}%</Text>
            </Group>
            <Progress value={(step / 5) * 100} size="sm" radius="xl" />
          </Stack>
        )}

        <Paper p="xl" radius="md" withBorder>
          {/* Step 0: Welcome + Value Proposition */}
          {step === 0 && (
            <Stack gap="xl" align="center">
              <ThemeIcon size={80} radius="xl" variant="light" color="teal">
                <IconRobot size={40} />
              </ThemeIcon>

              <Stack gap="xs" ta="center">
                <Title order={2}>Vítej! Jsem Karen</Title>
                <Text c="dimmed" maw={400}>
                  Jsem tvoje AI telefonní asistentka. Když nezvedneš telefon, přebírám hovory za tebe.
                </Text>
              </Stack>

              <Paper p="md" radius="md" bg="gray.0" w="100%">
                <Stack gap="sm">
                  <Text size="sm" fw={500}>Co pro tebe udělám:</Text>
                  <List size="sm" spacing="xs" c="dimmed">
                    <List.Item>Zvednu hovory, když budeš zaneprázdněný</List.Item>
                    <List.Item>Zjistím, kdo volá a co potřebuje</List.Item>
                    <List.Item>Odmítnu marketing a spam</List.Item>
                    <List.Item>Okamžitě přepojím důležité kontakty (rodina, kolegové)</List.Item>
                    <List.Item>Pošlu ti přehled hovorů v aplikaci</List.Item>
                  </List>
                </Stack>
              </Paper>

              <Text size="sm" c="dimmed">Nastavení zabere asi 3 minuty.</Text>

              <Button
                size="lg"
                fullWidth
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

          {/* Step 2: VIP Contacts */}
          {step === 2 && (
            <Stack gap="xl">
              <Stack gap="xs" ta="center">
                <Title order={2}>Koho má Karen vždy přepojit?</Title>
                <Text c="dimmed">
                  Některé hovory jsou důležité a nechceš, aby je Karen vyřizovala.
                </Text>
              </Stack>

              <Paper p="md" radius="md" bg="gray.0">
                <Text size="sm" c="dimmed">
                  Když se volající představí jedním z těchto jmen, Karen ho okamžitě přepojí na tebe.
                  Například: „Tady máma" → Karen řekne „Přepojuji" a zavolá ti.
                </Text>
              </Paper>

              <TagsInput
                size="md"
                placeholder="Máma, Táta, Jana, Šéf..."
                description="Přidej jméno a stiskni Enter"
                value={vipNames}
                onChange={setVipNames}
              />

              <Group grow>
                <Button
                  variant="subtle"
                  onClick={() => setStep(3)}
                >
                  Nastavím později
                </Button>
                <Button
                  size="lg"
                  rightSection={<IconArrowRight size={18} />}
                  onClick={() => setStep(3)}
                >
                  Pokračovat
                </Button>
              </Group>
            </Stack>
          )}

          {/* Step 3: Marketing Handling */}
          {step === 3 && (
            <Stack gap="xl">
              <Stack gap="xs" ta="center">
                <Title order={2}>Jak nakládat s marketingem?</Title>
                <Text c="dimmed">
                  Karen automaticky rozpozná marketingové a obchodní hovory.
                </Text>
              </Stack>

              <Radio.Group
                value={marketingOption}
                onChange={(value) => setMarketingOption(value as "reject" | "email")}
              >
                <Stack gap="md">
                  <Radio
                    value="reject"
                    label="Zdvořile odmítne a ukončí hovor"
                    description="Standardní nastavení pro většinu uživatelů"
                  />
                  <Radio
                    value="email"
                    label="Odmítne, ale nabídne můj email pro písemné nabídky"
                    description="Užitečné, pokud občas chceš vidět nabídky"
                  />
                </Stack>
              </Radio.Group>

              {marketingOption === "email" && (
                <TextInput
                  size="md"
                  placeholder="nabidky@email.cz"
                  label="Email pro marketingové nabídky"
                  value={marketingEmail}
                  onChange={(e) => setMarketingEmail(e.target.value)}
                />
              )}

              <Group grow>
                <Button
                  variant="subtle"
                  onClick={handleSaveConfiguration}
                >
                  Přeskočit
                </Button>
                <Button
                  size="lg"
                  rightSection={<IconArrowRight size={18} />}
                  onClick={handleSaveConfiguration}
                >
                  Pokračovat
                </Button>
              </Group>
            </Stack>
          )}

          {/* Step 4: Phone number & forwarding */}
          {step === 4 && (
            <Stack gap="xl">
              <Stack gap="xs" ta="center">
                <Title order={2}>{hasPhoneNumber ? "Tvoje Karen číslo" : "Skoro hotovo!"}</Title>
                <Text c="dimmed">
                  {hasPhoneNumber
                    ? "Toto je číslo, na které přesměruješ hovory"
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

                  {/* Karen is ready notice */}
                  <Alert icon={<IconCheck size={16} />} color="green" variant="light">
                    Karen je připravená! Můžeš ji hned vyzkoušet zavoláním na číslo výše z jiného telefonu.
                  </Alert>

                  {/* Two options explanation */}
                  <Paper p="md" radius="md" bg="blue.0">
                    <Stack gap="sm">
                      <Text size="sm" fw={500} c="blue.8">Jak Karen používat dlouhodobě?</Text>
                      <Text size="sm" c="blue.7">
                        <strong>Varianta A:</strong> Nastav přesměrování hovorů (doporučeno) – když nezvedneš, hovor se automaticky přepojí na Karen.
                      </Text>
                      <Text size="sm" c="blue.7">
                        <strong>Varianta B:</strong> Zavolej přímo na Karen číslo – ideální pro rychlé vyzkoušení.
                      </Text>
                    </Stack>
                  </Paper>

                  {/* Redirect type selection */}
                  <Stack gap="md">
                    <Text size="sm" fw={500}>
                      Varianta A: Nastav přesměrování
                    </Text>
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
                      Vyber, které typy přesměrování chceš nastavit. Pro kompletní pokrytí doporučujeme všechny tři.
                    </Text>
                    <Stack gap="xs">
                      {REDIRECT_ORDER.map((type) => (
                        <Checkbox
                          key={type}
                          label={REDIRECT_CODES[type].label}
                          description={REDIRECT_CODES[type].description}
                          checked={selectedRedirects.includes(type)}
                          onChange={() => toggleRedirect(type)}
                        />
                      ))}
                    </Stack>
                  </Stack>

                  {/* Forwarding instructions for selected types */}
                  {selectedRedirects.length > 0 && (
                    <RedirectSetupAccordion
                      karenNumber={primaryPhone || ""}
                      redirectTypes={selectedRedirects}
                    />
                  )}

                  <Button
                    size="lg"
                    fullWidth
                    rightSection={<IconArrowRight size={18} />}
                    onClick={() => setStep(5)}
                  >
                    Pokračovat
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
                    onClick={() => setStep(5)}
                  >
                    Pokračovat
                  </Button>
                </>
              )}
            </Stack>
          )}

          {/* Step 5: Done + Next Steps */}
          {step === 5 && (
            <Stack gap="xl" align="center">
              <ThemeIcon size={80} radius="xl" variant="light" color="green">
                <IconConfetti size={40} />
              </ThemeIcon>

              <Stack gap="xs" ta="center">
                <Title order={2}>Hotovo! Karen je připravená.</Title>
                <Text c="dimmed" maw={400}>
                  Když někdo zavolá a ty nezvedneš, Karen to vyřídí za tebe.
                </Text>
              </Stack>

              {hasPhoneNumber && (
                <Paper p="md" radius="md" bg="teal.0" w="100%">
                  <Stack gap="sm">
                    <Text size="sm" fw={500} c="teal.8">
                      Vyzkoušej Karen!
                    </Text>
                    <Text size="sm" c="teal.7">
                      <strong>S přesměrováním:</strong> Zavolej na své číslo z jiného telefonu a nech vyzvánět 20 sekund. Karen zvedne.
                    </Text>
                    <Text size="sm" c="teal.7">
                      <strong>Přímo:</strong> Zavolej na {primaryPhone} – Karen zvedne okamžitě.
                    </Text>
                  </Stack>
                </Paper>
              )}

              <Paper p="md" radius="md" bg="gray.0" w="100%">
                <Stack gap="sm">
                  <Text size="sm" fw={500}>Co dál:</Text>
                  <List size="sm" spacing="xs" c="dimmed">
                    <List.Item>Přehled všech hovorů najdeš v aplikaci</List.Item>
                    <List.Item>Nastavení můžeš kdykoliv změnit v sekci Nastavení</List.Item>
                    {vipNames.length === 0 && (
                      <List.Item>Přidej VIP kontakty, které má Karen vždy přepojit</List.Item>
                    )}
                  </List>
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
