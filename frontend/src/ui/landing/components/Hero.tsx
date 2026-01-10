import { Box, Button, Container, Text, Title } from "@mantine/core";

interface HeroProps {
  title: string;
  tagline: string;
  ctaText: string;
  onCtaClick: () => void;
}

export function Hero({ title, tagline, ctaText, onCtaClick }: HeroProps) {
  return (
    <Box py={80} style={{ background: "linear-gradient(180deg, #f8f9fa 0%, #fff 100%)" }}>
      <Container size="md" ta="center">
        <Title order={1} size={48} mb="md">
          {title}
        </Title>
        <Text size="xl" c="dimmed" mb="xl" maw={600} mx="auto">
          {tagline}
        </Text>
        <Button size="xl" radius="md" onClick={onCtaClick}>
          {ctaText}
        </Button>
      </Container>
    </Box>
  );
}
