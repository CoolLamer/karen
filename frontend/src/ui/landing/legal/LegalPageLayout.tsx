import { Box, Container, Stack, Title, Text } from "@mantine/core";
import { Header } from "../components/Header";
import { Footer } from "../components/Footer";
import { SHARED_CONTENT } from "../content/shared";

interface LegalPageLayoutProps {
  title: string;
  lastUpdated: string;
  children: React.ReactNode;
}

export function LegalPageLayout({
  title,
  lastUpdated,
  children,
}: LegalPageLayoutProps) {
  return (
    <Box>
      <Header showBackToMain />
      <Container size="md" py={60}>
        <Stack gap="xl">
          <Box>
            <Title order={1} mb="xs">
              {title}
            </Title>
            <Text c="dimmed" size="sm">
              Poslední aktualizace: {lastUpdated}
            </Text>
          </Box>
          {children}
        </Stack>
      </Container>
      <Footer tagline={SHARED_CONTENT.footer.tagline} />
    </Box>
  );
}
