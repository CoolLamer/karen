import { Link } from "react-router-dom";
import { Box, Container, Text, Group, Anchor, Stack } from "@mantine/core";

interface FooterProps {
  tagline: string;
}

export function Footer({ tagline }: FooterProps) {
  return (
    <Box py="xl" bg="gray.9">
      <Container size="lg">
        <Stack gap="md">
          <Group justify="center" gap="xl" wrap="wrap">
            <Anchor
              component={Link}
              to="/obchodni-podminky"
              c="gray.5"
              size="sm"
              underline="hover"
            >
              Obchodní podmínky
            </Anchor>
            <Anchor
              component={Link}
              to="/ochrana-osobnich-udaju"
              c="gray.5"
              size="sm"
              underline="hover"
            >
              Ochrana osobních údajů
            </Anchor>
            <Anchor
              component={Link}
              to="/informace-o-provozovateli"
              c="gray.5"
              size="sm"
              underline="hover"
            >
              Informace o provozovateli
            </Anchor>
            <Anchor
              href="mailto:info@zvednu.cz"
              c="gray.5"
              size="sm"
              underline="hover"
            >
              Kontakt
            </Anchor>
          </Group>
          <Text c="gray.6" ta="center" size="xs">
            BlockSei s.r.o. | IČO: 22345159 | Korunní 2569/108, Praha 101 00
          </Text>
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
