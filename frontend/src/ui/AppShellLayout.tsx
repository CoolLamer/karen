import React from "react";
import { Outlet, Link } from "react-router-dom";
import { AppShell, Group, Title, Anchor, Container, ActionIcon, Tooltip } from "@mantine/core";
import { IconSettings } from "@tabler/icons-react";

export function AppShellLayout() {
  return (
    <AppShell header={{ height: 56 }} padding="md">
      <AppShell.Header>
        <Group h="100%" px="md" justify="space-between">
          <Group gap="md">
            <Title order={4} c="blue">
              Zvednu
            </Title>
            <Anchor component={Link} to="/inbox" underline="never" size="sm">
              Hovory
            </Anchor>
          </Group>
          <Tooltip label="Nastaveni">
            <ActionIcon component={Link} to="/settings" variant="subtle" size="lg">
              <IconSettings size={20} />
            </ActionIcon>
          </Tooltip>
        </Group>
      </AppShell.Header>
      <AppShell.Main>
        <Container size="lg">
          <Outlet />
        </Container>
      </AppShell.Main>
    </AppShell>
  );
}
