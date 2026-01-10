import React from "react";
import { Paper, Text, ThemeIcon } from "@mantine/core";

interface FeatureCardProps {
  icon: React.ReactNode;
  title: string;
  description: string;
}

export function FeatureCard({ icon, title, description }: FeatureCardProps) {
  return (
    <Paper
      p="xl"
      radius="md"
      withBorder
      style={{
        transition: "all 0.2s ease",
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.transform = "translateY(-4px)";
        e.currentTarget.style.boxShadow = "0 10px 20px rgba(0, 0, 0, 0.08)";
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.transform = "";
        e.currentTarget.style.boxShadow = "";
      }}
    >
      <ThemeIcon size={48} radius="md" variant="light" color="teal" mb="md">
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
