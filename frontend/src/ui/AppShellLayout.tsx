import React from "react";
import { Outlet, Link, useLocation } from "react-router-dom";
import {
  AppShell,
  Group,
  Title,
  Anchor,
  Container,
  ActionIcon,
  Tooltip,
  Badge,
  Button,
  Modal,
  Text,
  Stack,
  CopyButton,
  Accordion,
  Alert,
  Box,
  List,
} from "@mantine/core";
import { useDisclosure, useMediaQuery } from "@mantine/hooks";
import { IconSettings, IconPhone, IconCopy, IconCheck } from "@tabler/icons-react";
import { api, TenantPhoneNumber } from "../api";
import {
  REDIRECT_CODES,
  REDIRECT_ORDER,
  PHONE_SETTINGS_INSTRUCTIONS,
  getDialCode,
  getDeactivationCode,
} from "../constants/redirectCodes";

export function AppShellLayout() {
  const [unresolvedCount, setUnresolvedCount] = React.useState(0);
  const [phoneNumbers, setPhoneNumbers] = React.useState<TenantPhoneNumber[]>([]);
  const [modalOpened, { open: openModal, close: closeModal }] = useDisclosure(false);
  const isMobile = useMediaQuery("(max-width: 768px)");
  const location = useLocation();

  const karenNumber = phoneNumbers.find((p) => p.is_primary)?.twilio_number;

  // Fetch unresolved count and phone numbers
  const fetchData = React.useCallback(() => {
    api.getUnresolvedCount().then((data) => setUnresolvedCount(data.count)).catch(() => {});
    api.getTenant().then((data) => setPhoneNumbers(data.phone_numbers || [])).catch(() => {});
  }, []);

  // Initial fetch + polling every 30 seconds for real-time updates
  React.useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 30000);
    return () => clearInterval(interval);
  }, [fetchData]);

  // Refetch when navigating (e.g., after viewing/resolving calls)
  React.useEffect(() => {
    fetchData();
  }, [location.pathname, fetchData]);

  return (
    <AppShell header={{ height: 56 }} padding="md">
      <AppShell.Header>
        <Group h="100%" px="md" justify="space-between">
          <Group gap="md">
            <Title order={4} c="teal">
              Zvednu
            </Title>
            <Group gap={6}>
              <Anchor component={Link} to="/inbox" underline="never" size="sm">
                Hovory
              </Anchor>
              {unresolvedCount > 0 && (
                <Badge size="sm" color="teal" variant="filled" radius="xl">
                  {unresolvedCount}
                </Badge>
              )}
            </Group>
          </Group>
          <Group gap="sm">
            {karenNumber && (
              isMobile ? (
                <Tooltip label="Nastavit přesměrování">
                  <ActionIcon variant="subtle" size="lg" onClick={openModal}>
                    <IconPhone size={20} />
                  </ActionIcon>
                </Tooltip>
              ) : (
                <Button variant="subtle" size="sm" onClick={openModal}>
                  Nastavit přesměrování
                </Button>
              )
            )}
            <Tooltip label="Nastavení">
              <ActionIcon component={Link} to="/settings" variant="subtle" size="lg">
                <IconSettings size={20} />
              </ActionIcon>
            </Tooltip>
          </Group>
        </Group>
      </AppShell.Header>
      <AppShell.Main>
        <Container size="lg">
          <Outlet />
        </Container>
      </AppShell.Main>

      {/* Forwarding Setup Modal */}
      <Modal
        opened={modalOpened}
        onClose={closeModal}
        title="Jak nastavit přesměrování"
        size="lg"
        centered
      >
        <Stack gap="md">
          <Alert color="blue" variant="light">
            Pro kompletní pokrytí doporučujeme nastavit všechny tři typy přesměrování.
          </Alert>

          <Accordion variant="separated" defaultValue="noAnswer">
            {REDIRECT_ORDER.map((type) => {
              const code = REDIRECT_CODES[type];
              const dialCode = karenNumber ? getDialCode(type, karenNumber) : "";
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

                      <Text size="xs" fw={500} c="teal">Aktivovat:</Text>
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
                      <Button
                        variant="light"
                        size="xs"
                        leftSection={<IconPhone size={14} />}
                        disabled={!karenNumber}
                        onClick={() => { window.location.href = `tel:${dialCode}`; }}
                      >
                        Vytočit
                      </Button>

                      <Text size="xs" fw={500} c="red" mt="xs">Deaktivovat:</Text>
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
                      <Button
                        variant="subtle"
                        size="xs"
                        color="red"
                        leftSection={<IconPhone size={14} />}
                        onClick={() => { window.location.href = `tel:${deactivateCode}`; }}
                      >
                        Vytočit
                      </Button>
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

          <Button variant="subtle" onClick={closeModal} fullWidth>
            Zavřít
          </Button>
        </Stack>
      </Modal>
    </AppShell>
  );
}
