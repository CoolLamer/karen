import { Box, Container, Text } from "@mantine/core";

interface FooterProps {
  tagline: string;
}

export function Footer({ tagline }: FooterProps) {
  return (
    <Box py="xl" bg="gray.9">
      <Container size="lg">
        <Text c="gray.5" ta="center" size="sm">
          {tagline}
        </Text>
      </Container>
    </Box>
  );
}
