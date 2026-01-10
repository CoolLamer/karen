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
  Code,
  List,
  CopyButton,
  Paper,
} from "@mantine/core";
import { useDisclosure, useMediaQuery } from "@mantine/hooks";
import { IconSettings, IconPhone, IconCopy, IconCheck } from "@tabler/icons-react";
import { api, TenantPhoneNumber } from "../api";

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

  React.useEffect(() => {
    fetchData();
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
        size="md"
      >
        <Stack gap="md">
          <Text size="sm">
            Přesměrujte hovory z vašeho telefonu na Karen číslo, když nemůžete zvednout.
          </Text>

          <Paper p="md" withBorder radius="md" bg="gray.0">
            <Text size="sm" fw={500} mb="xs">
              Vaše Karen číslo:
            </Text>
            <Group gap="xs">
              <Code style={{ fontSize: "1.1rem" }}>{karenNumber}</Code>
              <CopyButton value={karenNumber || ""}>
                {({ copied, copy }) => (
                  <Tooltip label={copied ? "Zkopírováno!" : "Kopírovat"}>
                    <ActionIcon
                      color={copied ? "teal" : "gray"}
                      variant="subtle"
                      onClick={copy}
                    >
                      {copied ? <IconCheck size={16} /> : <IconCopy size={16} />}
                    </ActionIcon>
                  </Tooltip>
                )}
              </CopyButton>
            </Group>
          </Paper>

          <Text size="sm" fw={500}>
            Postup pro nastavení:
          </Text>
          <List size="sm" spacing="xs">
            <List.Item>
              Otevřete <strong>Nastavení</strong> &gt; <strong>Telefon</strong> &gt;{" "}
              <strong>Přesměrování hovorů</strong>
            </List.Item>
            <List.Item>
              Vyberte <strong>"Při obsazení"</strong> nebo <strong>"Při nedostupnosti"</strong>
            </List.Item>
            <List.Item>
              Zadejte Karen číslo: <Code>{karenNumber}</Code>
            </List.Item>
            <List.Item>Uložte nastavení</List.Item>
          </List>

          <Text size="xs" c="dimmed">
            Tip: Můžete také nastavit podmíněné přesměrování přes kód operátora. Kontaktujte svého
            operátora pro více informací.
          </Text>

          <Button onClick={closeModal} fullWidth>
            Rozumím
          </Button>
        </Stack>
      </Modal>
    </AppShell>
  );
}
