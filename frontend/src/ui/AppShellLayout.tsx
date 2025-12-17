import React from "react";
import { Outlet, Link } from "react-router-dom";
import { AppShell, Group, Title, Anchor, Container } from "@mantine/core";

export function AppShellLayout() {
  return (
    <AppShell header={{ height: 56 }} padding="md">
      <AppShell.Header>
        <Group h="100%" px="md" justify="space-between">
          <Group gap="sm">
            <Title order={4}>karen</Title>
            <Anchor component={Link} to="/" underline="never">
              Inbox
            </Anchor>
          </Group>
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


