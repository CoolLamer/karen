import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Box,
  Button,
  Container,
  Group,
  Paper,
  PinInput,
  Stack,
  Text,
  TextInput,
  Title,
  Anchor,
  Alert,
  ThemeIcon,
} from "@mantine/core";
import { IconArrowLeft, IconAlertCircle, IconPhone } from "@tabler/icons-react";
import { useMediaQuery } from "@mantine/hooks";
import { api } from "../api";
import { useAuth } from "../AuthContext";

type Step = "phone" | "otp";

export function LoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const isMobile = useMediaQuery("(max-width: 480px)");

  const [step, setStep] = useState<Step>("phone");
  const [phone, setPhone] = useState("");
  const [code, setCode] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [canResend, setCanResend] = useState(false);
  const [resendCountdown, setResendCountdown] = useState(0);

  const formatPhone = (value: string) => {
    // Remove all non-digits except +
    let cleaned = value.replace(/[^\d+]/g, "");

    // Ensure it starts with +
    if (cleaned && !cleaned.startsWith("+")) {
      cleaned = "+" + cleaned;
    }

    return cleaned;
  };

  const handlePhoneChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPhone(formatPhone(e.target.value));
    setError(null);
  };

  const handleSendCode = async () => {
    if (!phone || phone.length < 10) {
      setError("Zadej platné telefonní číslo");
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      await api.sendCode(phone);
      setStep("otp");
      startResendTimer();
    } catch {
      setError("Nepodařilo se odeslat kód. Zkontroluj číslo a zkus to znovu.");
    } finally {
      setIsLoading(false);
    }
  };

  const startResendTimer = () => {
    setCanResend(false);
    setResendCountdown(30);

    const interval = setInterval(() => {
      setResendCountdown((prev) => {
        if (prev <= 1) {
          clearInterval(interval);
          setCanResend(true);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
  };

  const handleResendCode = async () => {
    if (!canResend) return;

    setIsLoading(true);
    setError(null);

    try {
      await api.sendCode(phone);
      startResendTimer();
    } catch {
      setError("Nepodařilo se odeslat kód.");
    } finally {
      setIsLoading(false);
    }
  };

  const handleVerifyCode = async (value: string) => {
    if (value.length !== 6) return;

    setIsLoading(true);
    setError(null);

    try {
      const response = await api.verifyCode(phone, value);
      const needsOnboarding = login(response.token, response.user);

      // Navigate based on whether user needs onboarding
      // Use the returned value from login() to avoid React state race condition
      if (needsOnboarding) {
        navigate("/onboarding");
      } else {
        navigate("/");
      }
    } catch {
      setError("Neplatný kód. Zkus to znovu.");
      setCode("");
    } finally {
      setIsLoading(false);
    }
  };

  const handleBack = () => {
    if (step === "otp") {
      setStep("phone");
      setCode("");
      setError(null);
    } else {
      navigate("/");
    }
  };

  return (
    <Box mih="100vh" bg="gray.0">
      <Container size="xs" py={{ base: 20, sm: 60 }}>
        <Paper p="xl" radius="md" withBorder>
          <Stack gap="lg">
            {/* Back button */}
            <Button
              variant="subtle"
              leftSection={<IconArrowLeft size={16} />}
              onClick={handleBack}
              style={{ alignSelf: "flex-start" }}
              px={0}
            >
              Zpět
            </Button>

            {step === "phone" && (
              <>
                <Stack gap="md" ta="center" align="center">
                  <ThemeIcon size={60} radius="xl" variant="light" color="teal">
                    <IconPhone size={30} />
                  </ThemeIcon>
                  <Title order={2}>Přihlášení</Title>
                  <Text c="dimmed">Zadej své telefonní číslo</Text>
                </Stack>

                {error && (
                  <Alert icon={<IconAlertCircle size={16} />} color="red" variant="light">
                    {error}
                  </Alert>
                )}

                <TextInput
                  size="lg"
                  placeholder="+420 777 123 456"
                  value={phone}
                  onChange={handlePhoneChange}
                  disabled={isLoading}
                  type="tel"
                  autoComplete="tel"
                  autoFocus
                />

                <Button
                  size="lg"
                  fullWidth
                  onClick={handleSendCode}
                  loading={isLoading}
                >
                  Poslat ověřovací kód
                </Button>

                <Text size="xs" c="dimmed" ta="center">
                  Pokračováním souhlasíte s podmínkami služby
                </Text>
              </>
            )}

            {step === "otp" && (
              <>
                <Stack gap="xs" ta="center">
                  <Title order={2}>Ověřovací kód</Title>
                  <Text c="dimmed">
                    Poslali jsme SMS na{" "}
                    <Text span fw={500}>
                      {phone}
                    </Text>
                  </Text>
                </Stack>

                {error && (
                  <Alert icon={<IconAlertCircle size={16} />} color="red" variant="light">
                    {error}
                  </Alert>
                )}

                <Group justify="center">
                  <PinInput
                    length={6}
                    size={isMobile ? "md" : "xl"}
                    value={code}
                    onChange={setCode}
                    onComplete={handleVerifyCode}
                    disabled={isLoading}
                    type="number"
                    autoFocus
                  />
                </Group>

                <Text size="sm" c="dimmed" ta="center">
                  Nepřišel kód?{" "}
                  {canResend ? (
                    <Anchor component="button" onClick={handleResendCode} disabled={isLoading}>
                      Poslat znovu
                    </Anchor>
                  ) : (
                    <Text span c="dimmed">
                      Poslat znovu za {resendCountdown}s
                    </Text>
                  )}
                </Text>
              </>
            )}
          </Stack>
        </Paper>
      </Container>
    </Box>
  );
}
