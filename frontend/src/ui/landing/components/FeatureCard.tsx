import React from "react";
import { Paper, Text, ThemeIcon } from "@mantine/core";

interface FeatureCardProps {
  icon: React.ReactNode;
  title: string;
  description: string;
}

export function FeatureCard({ icon, title, description }: FeatureCardProps) {
  return (
    <Paper p="xl" radius="md" withBorder>
      <ThemeIcon size={48} radius="md" variant="light" color="blue" mb="md">
        {icon}
      </ThemeIcon>
      <Text size="lg" fw={600} mb="xs">
        {title}
      </Text>
      <Text c="dimmed" size="sm">
        {description}
      </Text>
    </Paper>
  );
}
