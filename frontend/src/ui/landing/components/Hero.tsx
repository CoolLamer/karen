import { Box, Button, Container, Text, Title, Group, Badge } from "@mantine/core";

interface HeroProps {
  title: string;
  tagline: string;
  ctaText: string;
  onCtaClick: () => void;
}

export function Hero({ title, tagline, ctaText, onCtaClick }: HeroProps) {
  return (
    <Box
      py={{ base: 60, sm: 100 }}
      style={{
        background: "linear-gradient(135deg, #E6FFFA 0%, #B2F5EA 50%, #E9D8FD 100%)",
        position: "relative",
        overflow: "hidden",
      }}
    >
      <Container size="md" ta="center" style={{ position: "relative", zIndex: 1 }}>
        <Badge size="lg" variant="light" color="purple" mb="lg">
          AI asistentka
        </Badge>
        <Title order={1} fz={{ base: 36, sm: 48, md: 56 }} mb="md" style={{ lineHeight: 1.2 }}>
          {title}
        </Title>
        <Text size="xl" c="dimmed" mb="xl" maw={600} mx="auto">
          {tagline}
        </Text>
        <Group justify="center" gap="md">
          <Button size="xl" radius="lg" onClick={onCtaClick}>
            {ctaText}
          </Button>
          <Button size="xl" radius="lg" variant="outline" color="gray" component="a" href="#how-it-works">
            Jak to funguje
          </Button>
        </Group>
        <Group justify="center" mt="xl" gap="xl">
          <Text size="sm" c="dimmed">Bez kreditky</Text>
          <Text size="sm" c="dimmed">Aktivace za 2 minuty</Text>
          <Text size="sm" c="dimmed">Zrušení kdykoliv</Text>
        </Group>
      </Container>
    </Box>
  );
}
