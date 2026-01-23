import React from "react";
import { Outlet, Link, useLocation } from "react-router-dom";
import {
  AppShell,
  Group,
  Title,
  Anchor,
  ActionIcon,
  Tooltip,
  Badge,
  Burger,
  Drawer,
  Stack,
  NavLink,
  Box,
} from "@mantine/core";
import { useDisclosure, useMediaQuery } from "@mantine/hooks";
import {
  IconArrowLeft,
  IconPhone,
  IconUsers,
  IconFileText,
  IconSettings,
} from "@tabler/icons-react";

const ADMIN_NAV_ITEMS = [
  { path: "/admin", label: "Phone Numbers", icon: IconPhone, exact: true },
  { path: "/admin/users", label: "Users", icon: IconUsers, exact: false },
  { path: "/admin/logs", label: "Logs", icon: IconFileText, exact: false },
  { path: "/admin/config", label: "Config", icon: IconSettings, exact: false },
];

export function AdminShellLayout() {
  const location = useLocation();
  const isMobile = useMediaQuery("(max-width: 768px)");
  const [drawerOpened, { toggle: toggleDrawer, close: closeDrawer }] =
    useDisclosure(false);

  const isActive = (path: string, exact: boolean) => {
    if (exact) {
      return location.pathname === path;
    }
    return location.pathname.startsWith(path);
  };

  return (
    <AppShell header={{ height: 56 }} padding="md">
      <AppShell.Header bg="dark.6">
        <Group h="100%" px="md" justify="space-between">
          <Group gap="md">
            {isMobile && (
              <Burger
                opened={drawerOpened}
                onClick={toggleDrawer}
                size="sm"
                color="white"
              />
            )}
            <Group gap="xs">
              <Title order={4} c="white">
                Zvednu
              </Title>
              <Badge color="violet" variant="filled" size="sm">
                Admin
              </Badge>
            </Group>

            {/* Desktop nav links */}
            {!isMobile && (
              <Group gap="xs" ml="lg">
                {ADMIN_NAV_ITEMS.map((item) => (
                  <Anchor
                    key={item.path}
                    component={Link}
                    to={item.path}
                    underline="never"
                    size="sm"
                    c={isActive(item.path, item.exact) ? "white" : "gray.4"}
                    fw={isActive(item.path, item.exact) ? 600 : 400}
                  >
                    {item.label}
                  </Anchor>
                ))}
              </Group>
            )}
          </Group>

          <Tooltip label="Back to App">
            <ActionIcon
              component={Link}
              to="/inbox"
              variant="subtle"
              size="lg"
              c="white"
            >
              <IconArrowLeft size={20} />
            </ActionIcon>
          </Tooltip>
        </Group>
      </AppShell.Header>

      {/* Mobile navigation drawer */}
      <Drawer
        opened={drawerOpened}
        onClose={closeDrawer}
        size="xs"
        title="Admin Navigation"
      >
        <Stack gap="xs">
          {ADMIN_NAV_ITEMS.map((item) => (
            <NavLink
              key={item.path}
              component={Link}
              to={item.path}
              label={item.label}
              leftSection={<item.icon size={18} />}
              active={isActive(item.path, item.exact)}
              onClick={closeDrawer}
            />
          ))}
          <NavLink
            component={Link}
            to="/inbox"
            label="Back to App"
            leftSection={<IconArrowLeft size={18} />}
            onClick={closeDrawer}
            mt="md"
          />
        </Stack>
      </Drawer>

      <AppShell.Main>
        <Box bg="gray.0" mih="calc(100vh - 56px)">
          <Outlet />
        </Box>
      </AppShell.Main>
    </AppShell>
  );
}
