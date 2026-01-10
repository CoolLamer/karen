import { Box, Container, Text, Group, Anchor, Stack } from "@mantine/core";

interface FooterProps {
  tagline: string;
}

export function Footer({ tagline }: FooterProps) {
  return (
    <Box py="xl" bg="gray.9">
      <Container size="lg">
        <Stack gap="md">
          <Group justify="center" gap="xl">
            <Anchor href="mailto:info@zvednu.cz" c="gray.5" size="sm" underline="hover">
              Kontakt
            </Anchor>
          </Group>
          <Text c="gray.6" ta="center" size="sm">
            {tagline}
          </Text>
          <Text c="gray.7" ta="center" size="xs">
            {new Date().getFullYear()} Zvednu. Všechna práva vyhrazena.
          </Text>
        </Stack>
      </Container>
    </Box>
  );
}
