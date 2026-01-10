import { Container, SimpleGrid, Paper, Text, ThemeIcon, Title, Box } from "@mantine/core";
import {
  IconPhoneOff,
  IconHandStop,
  IconMailOff,
  IconClipboardList,
  IconAlertTriangle,
  IconCurrencyDollar,
  IconPhoneIncoming,
  IconUserOff,
  IconHome,
  IconQuestionMark,
  IconPhone,
  IconUsers,
  IconArrowBack,
} from "@tabler/icons-react";
import { PainPointItem } from "../content/index";

const iconMap: Record<string, React.ReactNode> = {
  IconPhoneOff: <IconPhoneOff size={24} />,
  IconHandStop: <IconHandStop size={24} />,
  IconMailOff: <IconMailOff size={24} />,
  IconClipboardList: <IconClipboardList size={24} />,
  IconAlertTriangle: <IconAlertTriangle size={24} />,
  IconCurrencyDollar: <IconCurrencyDollar size={24} />,
  IconPhoneIncoming: <IconPhoneIncoming size={24} />,
  IconUserOff: <IconUserOff size={24} />,
  IconHome: <IconHome size={24} />,
  IconQuestionMark: <IconQuestionMark size={24} />,
  IconPhone: <IconPhone size={24} />,
  IconUsers: <IconUsers size={24} />,
  IconArrowBack: <IconArrowBack size={24} />,
};

interface PainPointsProps {
  title: string;
  items: PainPointItem[];
}

export function PainPoints({ title, items }: PainPointsProps) {
  return (
    <Container size="lg" py={60}>
      <Title order={2} ta="center" mb={40}>
        {title}
      </Title>
      <SimpleGrid cols={{ base: 1, sm: 2 }} spacing="xl">
        {items.map((item) => (
          <Paper
            key={item.title}
            p="xl"
            radius="md"
            withBorder
            style={{
              borderLeft: "4px solid var(--mantine-color-orange-5)",
            }}
          >
            <Box style={{ display: "flex", gap: 16 }}>
              <ThemeIcon size={48} radius="md" variant="light" color="orange">
                {iconMap[item.icon] || <IconPhone size={24} />}
              </ThemeIcon>
              <Box>
                <Text size="lg" fw={600} mb="xs">
                  {item.title}
                </Text>
                <Text c="dimmed" size="sm">
                  {item.description}
                </Text>
              </Box>
            </Box>
          </Paper>
        ))}
      </SimpleGrid>
    </Container>
  );
}
