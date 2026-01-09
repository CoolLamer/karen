import React from "react";
import { useNavigate } from "react-router-dom";
import {
  Box,
  Button,
  Container,
  Group,
  Paper,
  SimpleGrid,
  Stack,
  Text,
  ThemeIcon,
  Title,
} from "@mantine/core";
import {
  IconPhone,
  IconRobot,
  IconListCheck,
  IconShieldCheck,
  IconMessage,
} from "@tabler/icons-react";

function FeatureCard({
  icon,
  title,
  description,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
}) {
  return (
    <Paper p="xl" radius="md" withBorder>
      <ThemeIcon size={48} radius="md" variant="light" color="blue" mb="md">
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

function StepCard({ step, title, description }: { step: number; title: string; description: string }) {
  return (
    <Stack gap="xs" align="center" ta="center">
      <ThemeIcon size={48} radius="xl" variant="filled" color="blue">
        <Text fw={700} size="lg">
          {step}
        </Text>
      </ThemeIcon>
      <Text fw={600}>{title}</Text>
      <Text c="dimmed" size="sm">
        {description}
      </Text>
    </Stack>
  );
}

export function LandingPage() {
  const navigate = useNavigate();

  return (
    <Box>
      {/* Header */}
      <Box component="header" py="md" px="lg">
        <Container size="lg">
          <Group justify="space-between">
            <Text size="xl" fw={700} c="blue">
              Karen
            </Text>
            <Button variant="subtle" onClick={() => navigate("/login")}>
              Prihlasit se
            </Button>
          </Group>
        </Container>
      </Box>

      {/* Hero Section */}
      <Box py={80} style={{ background: "linear-gradient(180deg, #f8f9fa 0%, #fff 100%)" }}>
        <Container size="md" ta="center">
          <Title order={1} size={48} mb="md">
            Tvoje AI asistentka na telefonu
          </Title>
          <Text size="xl" c="dimmed" mb="xl" maw={600} mx="auto">
            Kdyz nezvednes, Karen zvedne za tebe. Zjisti kdo vola a proc. Spam odfiltruje.
          </Text>
          <Button size="xl" radius="md" onClick={() => navigate("/login")}>
            Vyzkouset zdarma
          </Button>
        </Container>
      </Box>

      {/* How it works */}
      <Container size="lg" py={60}>
        <Title order={2} ta="center" mb={40}>
          Jak to funguje
        </Title>
        <SimpleGrid cols={{ base: 1, sm: 3 }} spacing="xl">
          <StepCard
            step={1}
            title="Presmeruj hovory"
            description="Nastav presmerovani na sve Karen cislo, kdyz nezvedas."
          />
          <StepCard
            step={2}
            title="Karen zvedne"
            description="Karen prijme hovor a zepta se volajiciho, o co jde."
          />
          <StepCard
            step={3}
            title="Vidis prehled"
            description="V aplikaci vidis kdo volal, proc a zda stoji za to zavolat zpet."
          />
        </SimpleGrid>
      </Container>

      {/* Features */}
      <Box py={60} bg="gray.0">
        <Container size="lg">
          <Title order={2} ta="center" mb={40}>
            Proc Karen?
          </Title>
          <SimpleGrid cols={{ base: 1, sm: 2, md: 3 }} spacing="xl">
            <FeatureCard
              icon={<IconShieldCheck size={24} />}
              title="Filtruje spam"
              description="Karen rozpozna marketingove hovory a spam. Ty vidis jen dulezite hovory."
            />
            <FeatureCard
              icon={<IconMessage size={24} />}
              title="Zjisti kontext"
              description="Karen se zepta proc volaji a shrne to pro tebe. Vis o cem hovor bude."
            />
            <FeatureCard
              icon={<IconRobot size={24} />}
              title="Prirozeny hlas"
              description="Karen mluvi cesky s prirozene znejicim hlasem. Volajici nemaji pocit ze mluvi s robotem."
            />
            <FeatureCard
              icon={<IconPhone size={24} />}
              title="Nikdy nepropasnes"
              description="Dulezite hovory ti Karen preposle nebo zanecha podrobnou zpravu."
            />
            <FeatureCard
              icon={<IconListCheck size={24} />}
              title="Prepis hovoru"
              description="Kazdy hovor ma kompletni prepis. Muzes si ho precist misto poslechu."
            />
            <FeatureCard
              icon={<IconShieldCheck size={24} />}
              title="Tvoje pravidla"
              description="Nastav si VIP kontakty, zpusoby osloveni a vlastni instrukce."
            />
          </SimpleGrid>
        </Container>
      </Box>

      {/* Example Call */}
      <Container size="md" py={60}>
        <Title order={2} ta="center" mb={40}>
          Ukazka hovoru
        </Title>
        <Paper p="xl" radius="md" withBorder>
          <Stack gap="md">
            <Group gap="xs">
              <ThemeIcon size="sm" color="blue" variant="light">
                <IconRobot size={14} />
              </ThemeIcon>
              <Text size="sm" fw={500}>
                Karen:
              </Text>
            </Group>
            <Text c="dimmed" size="sm" pl={28}>
              "Dobry den, tady Karen, asistentka pana Lukase. Lukas ted nemuze prijmout hovor. Jak vam
              mohu pomoct?"
            </Text>

            <Group gap="xs">
              <ThemeIcon size="sm" color="gray" variant="light">
                <IconPhone size={14} />
              </ThemeIcon>
              <Text size="sm" fw={500}>
                Volajici:
              </Text>
            </Group>
            <Text c="dimmed" size="sm" pl={28}>
              "Dobry den, volam ohledne nabidky elektricke energie..."
            </Text>

            <Group gap="xs">
              <ThemeIcon size="sm" color="blue" variant="light">
                <IconRobot size={14} />
              </ThemeIcon>
              <Text size="sm" fw={500}>
                Karen:
              </Text>
            </Group>
            <Text c="dimmed" size="sm" pl={28}>
              "Dekuji za zavolani. Lukas nema zajem o marketingove nabidky. Pokud chcete, poslete
              nabidku na email nabidky@bauerlukas.cz. Na shledanou."
            </Text>

            <Paper p="md" radius="sm" bg="gray.0" mt="md">
              <Group justify="space-between">
                <Text size="sm" fw={500}>
                  Vysledek:
                </Text>
                <Text size="sm" c="yellow.7" fw={600}>
                  Marketing
                </Text>
              </Group>
              <Text size="sm" c="dimmed">
                Nabidka energie - ignorovat
              </Text>
            </Paper>
          </Stack>
        </Paper>
      </Container>

      {/* CTA */}
      <Box py={60} bg="blue.6">
        <Container size="sm" ta="center">
          <Title order={2} c="white" mb="md">
            Zacni pouzivat Karen jeste dnes
          </Title>
          <Text c="blue.1" mb="xl">
            Registrace je zdarma. Za 2 minuty budes mit svoji asistentku.
          </Text>
          <Button size="xl" radius="md" color="white" variant="white" onClick={() => navigate("/login")}>
            Vyzkouset zdarma
          </Button>
        </Container>
      </Box>

      {/* Footer */}
      <Box py="xl" bg="gray.9">
        <Container size="lg">
          <Text c="gray.5" ta="center" size="sm">
            Karen - Tvoje AI asistentka na telefonu
          </Text>
        </Container>
      </Box>
    </Box>
  );
}
