import { Box, Button, Container, Text, Title, Anchor } from "@mantine/core";

interface CTASectionProps {
  title: string;
  subtitle: string;
  buttonText: string;
  onCtaClick: () => void;
}

export function CTASection({ title, subtitle, buttonText, onCtaClick }: CTASectionProps) {
  return (
    <Box
      py={80}
      style={{
        background: "linear-gradient(135deg, var(--mantine-color-teal-6) 0%, var(--mantine-color-teal-7) 100%)",
      }}
    >
      <Container size="sm" ta="center">
        <Title order={2} c="white" mb="md">
          {title}
        </Title>
        <Text c="teal.1" mb="xl" size="lg">
          {subtitle}
        </Text>
        <Button size="xl" radius="lg" color="white" variant="white" onClick={onCtaClick}>
          {buttonText}
        </Button>
        <Text c="teal.2" mt="lg" size="sm">
          Máte otázky?{" "}
          <Anchor c="white" href="mailto:info@zvednu.cz" underline="always">
            Napište nám
          </Anchor>
        </Text>
      </Container>
    </Box>
  );
}
