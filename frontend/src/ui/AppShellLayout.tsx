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
} from "@mantine/core";
import { useDisclosure, useMediaQuery } from "@mantine/hooks";
import { IconSettings, IconPhone, IconQuestionMark } from "@tabler/icons-react";
import { FAQ } from "./landing/components";
import { SHARED_CONTENT } from "./landing/content/shared";
import { ForwardingSetupModal } from "./ForwardingSetupModal";
import { api, TenantPhoneNumber } from "../api";

export function AppShellLayout() {
  const [unresolvedCount, setUnresolvedCount] = React.useState(0);
  const [phoneNumbers, setPhoneNumbers] = React.useState<TenantPhoneNumber[]>([]);
  const [modalOpened, { open: openModal, close: closeModal }] = useDisclosure(false);
  const [faqOpened, { open: openFaq, close: closeFaq }] = useDisclosure(false);
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
            <Tooltip label="Nápověda">
              <ActionIcon variant="subtle" size="lg" onClick={openFaq}>
                <IconQuestionMark size={20} />
              </ActionIcon>
            </Tooltip>
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

      <ForwardingSetupModal
        opened={modalOpened}
        onClose={closeModal}
        karenNumber={karenNumber}
      />

      {/* FAQ Modal */}
      <Modal
        opened={faqOpened}
        onClose={closeFaq}
        title="Časté dotazy"
        size="lg"
        centered
      >
        <FAQ items={SHARED_CONTENT.faq} compact />
      </Modal>
    </AppShell>
  );
}
