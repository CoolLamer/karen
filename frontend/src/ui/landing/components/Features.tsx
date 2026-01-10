import { Box, Container, SimpleGrid, Title } from "@mantine/core";
import {
  IconShieldCheck,
  IconMessage,
  IconRobot,
  IconPhone,
  IconListCheck,
  IconSettings,
  IconPhoneCall,
} from "@tabler/icons-react";
import { FeatureDefinition } from "../content/index";
import { FeatureCard } from "./FeatureCard";

const iconMap: Record<string, React.ReactNode> = {
  IconShieldCheck: <IconShieldCheck size={24} />,
  IconMessage: <IconMessage size={24} />,
  IconRobot: <IconRobot size={24} />,
  IconPhone: <IconPhone size={24} />,
  IconListCheck: <IconListCheck size={24} />,
  IconSettings: <IconSettings size={24} />,
  IconPhoneCall: <IconPhoneCall size={24} />,
};

interface FeaturesProps {
  features: FeatureDefinition[];
  title?: string;
}

export function Features({ features, title = "Proč Zvednu?" }: FeaturesProps) {
  return (
    <Box py={60} bg="gray.0">
      <Container size="lg">
        <Title order={2} ta="center" mb={40}>
          {title}
        </Title>
        <SimpleGrid cols={{ base: 1, sm: 2, md: 3 }} spacing="xl">
          {features.map((feature) => (
            <FeatureCard
              key={feature.title}
              icon={iconMap[feature.icon] || <IconPhone size={24} />}
              title={feature.title}
              description={feature.description}
            />
          ))}
        </SimpleGrid>
      </Container>
    </Box>
  );
}
