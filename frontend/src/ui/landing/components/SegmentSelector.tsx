import { useNavigate } from "react-router-dom";
import {
  Container,
  Title,
  SimpleGrid,
  Paper,
  Text,
  ThemeIcon,
  Group,
  UnstyledButton,
} from "@mantine/core";
import {
  IconTool,
  IconStethoscope,
  IconBuildingSkyscraper,
  IconBriefcase,
  IconChevronRight,
} from "@tabler/icons-react";
import { SEGMENT_LIST } from "../content/segments";

const iconMap: Record<string, React.ReactNode> = {
  IconTool: <IconTool size={24} />,
  IconStethoscope: <IconStethoscope size={24} />,
  IconBuildingSkyscraper: <IconBuildingSkyscraper size={24} />,
  IconBriefcase: <IconBriefcase size={24} />,
};

export function SegmentSelector() {
  const navigate = useNavigate();

  return (
    <Container size="lg" py={60}>
      <Title order={2} ta="center" mb="md">
        Jsem:
      </Title>
      <Text c="dimmed" ta="center" mb={40}>
        Vyber svuj obor a zjisti, jak ti Zvednu pomuze
      </Text>

      <SimpleGrid cols={{ base: 1, sm: 2, md: 4 }} spacing="lg">
        {SEGMENT_LIST.map((segment) => (
          <UnstyledButton
            key={segment.key}
            onClick={() => navigate(segment.urlPath)}
            style={{ width: "100%" }}
          >
            <Paper
              p="xl"
              radius="md"
              withBorder
              style={{
                cursor: "pointer",
                transition: "all 0.2s",
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.borderColor = "#228be6";
                e.currentTarget.style.transform = "translateY(-2px)";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.borderColor = "";
                e.currentTarget.style.transform = "";
              }}
            >
              <Group justify="space-between" mb="md">
                <ThemeIcon size={48} radius="md" variant="light" color="blue">
                  {iconMap[segment.selectorIcon]}
                </ThemeIcon>
                <IconChevronRight size={20} color="gray" />
              </Group>
              <Text fw={600}>{segment.selectorLabel}</Text>
            </Paper>
          </UnstyledButton>
        ))}
      </SimpleGrid>
    </Container>
  );
}
