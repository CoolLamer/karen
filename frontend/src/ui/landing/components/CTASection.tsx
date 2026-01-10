import { Box, Button, Container, Text, Title } from "@mantine/core";

interface CTASectionProps {
  title: string;
  subtitle: string;
  buttonText: string;
  onCtaClick: () => void;
}

export function CTASection({ title, subtitle, buttonText, onCtaClick }: CTASectionProps) {
  return (
    <Box py={60} bg="blue.6">
      <Container size="sm" ta="center">
        <Title order={2} c="white" mb="md">
          {title}
        </Title>
        <Text c="blue.1" mb="xl">
          {subtitle}
        </Text>
        <Button size="xl" radius="md" color="white" variant="white" onClick={onCtaClick}>
          {buttonText}
        </Button>
      </Container>
    </Box>
  );
}
