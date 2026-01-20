import { Box, Container, Group, Text, ThemeIcon } from "@mantine/core";
import { IconShieldCheck, IconMapPin, IconPhone, IconBuilding } from "@tabler/icons-react";

interface TrustBadge {
  icon: React.ReactNode;
  text: string;
}

const badges: TrustBadge[] = [
  {
    icon: <IconBuilding size={18} />,
    text: "Česká firma",
  },
  {
    icon: <IconShieldCheck size={18} />,
    text: "V souladu s GDPR",
  },
  {
    icon: <IconMapPin size={18} />,
    text: "Data v EU",
  },
  {
    icon: <IconPhone size={18} />,
    text: "Všichni operátoři",
  },
];

export function TrustBadges() {
  return (
    <Box py={30} bg="gray.1">
      <Container size="lg">
        <Group justify="center" gap="xl" wrap="wrap">
          {badges.map((badge) => (
            <Group key={badge.text} gap="xs">
              <ThemeIcon variant="light" color="teal" size="md" radius="xl">
                {badge.icon}
              </ThemeIcon>
              <Text size="sm" c="dimmed">
                {badge.text}
              </Text>
            </Group>
          ))}
        </Group>
      </Container>
    </Box>
  );
}
