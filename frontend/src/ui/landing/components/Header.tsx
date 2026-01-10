import { useNavigate } from "react-router-dom";
import { Box, Button, Container, Group, Text } from "@mantine/core";
import { SHARED_CONTENT } from "../content/shared";

interface HeaderProps {
  showBackToMain?: boolean;
}

export function Header({ showBackToMain }: HeaderProps) {
  const navigate = useNavigate();

  return (
    <Box component="header" py="md" px="lg">
      <Container size="lg">
        <Group justify="space-between">
          <Text
            size="xl"
            fw={700}
            c="blue"
            style={{ cursor: "pointer" }}
            onClick={() => navigate("/")}
          >
            {SHARED_CONTENT.brand.name}
          </Text>
          <Group gap="sm">
            {showBackToMain && (
              <Button variant="subtle" onClick={() => navigate("/")}>
                Zpet
              </Button>
            )}
            <Button variant="subtle" onClick={() => navigate("/login")}>
              Prihlasit se
            </Button>
          </Group>
        </Group>
      </Container>
    </Box>
  );
}
