import { Container, SimpleGrid, Title } from "@mantine/core";
import { HowItWorksStep } from "../content/index";
import { StepCard } from "./StepCard";

interface HowItWorksProps {
  steps: HowItWorksStep[];
}

export function HowItWorks({ steps }: HowItWorksProps) {
  return (
    <Container size="lg" py={60} id="how-it-works">
      <Title order={2} ta="center" mb={40}>
        Jak to funguje
      </Title>
      <SimpleGrid cols={{ base: 1, sm: 3 }} spacing="xl">
        {steps.map((step) => (
          <StepCard
            key={step.step}
            step={step.step}
            title={step.title}
            description={step.description}
          />
        ))}
      </SimpleGrid>
    </Container>
  );
}
