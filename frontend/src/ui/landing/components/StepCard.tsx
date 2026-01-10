import { Stack, Text, ThemeIcon } from "@mantine/core";

interface StepCardProps {
  step: number;
  title: string;
  description: string;
}

export function StepCard({ step, title, description }: StepCardProps) {
  return (
    <Stack gap="xs" align="center" ta="center">
      <ThemeIcon size={48} radius="xl" variant="filled" color="blue">
        <Text fw={700} size="lg">
          {step}
        </Text>
      </ThemeIcon>
      <Text fw={600}>{title}</Text>
      <Text c="dimmed" size="sm">
        {description}
      </Text>
    </Stack>
  );
}
