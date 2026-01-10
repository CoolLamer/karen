import { Container, Paper, Stack, Text, ThemeIcon, Title, Badge, Box, Group } from "@mantine/core";
import { IconRobot, IconUser } from "@tabler/icons-react";
import { DialogueLine, ExampleCallResult } from "../content/index";
import { SHARED_CONTENT } from "../content/shared";

interface ExampleCallProps {
  scenario?: string;
  dialogue?: DialogueLine[];
  result?: ExampleCallResult;
}

const defaultDialogue: DialogueLine[] = [
  {
    speaker: "karen",
    text: "Dobrý den, tady Karen, asistentka pana Lukáše. Lukáš teď nemůže přijmout hovor. Jak vám mohu pomoct?",
  },
  {
    speaker: "caller",
    text: "Dobrý den, volám ohledně nabídky elektrické energie...",
  },
  {
    speaker: "karen",
    text: "Děkuji za zavolání. Lukáš nemá zájem o marketingové nabídky. Pokud chcete, pošlete nabídku na email. Na shledanou.",
  },
];

const defaultResult: ExampleCallResult = {
  label: "Marketing",
  color: "yellow",
  summary: "Nabídka energie - ignorovat",
};

export function ExampleCall({
  scenario,
  dialogue = defaultDialogue,
  result = defaultResult,
}: ExampleCallProps) {
  return (
    <Container size="md" py={60}>
      <Title order={2} ta="center" mb="md">
        Ukázka hovoru
      </Title>
      {scenario && (
        <Text c="dimmed" ta="center" mb={40}>
          {scenario}
        </Text>
      )}

      {/* Phone frame */}
      <Paper
        p="xl"
        radius="lg"
        withBorder
        style={{
          background: "linear-gradient(180deg, var(--mantine-color-gray-0) 0%, #fff 100%)",
          maxWidth: 500,
          margin: "0 auto",
        }}
      >
        <Stack gap="md">
          {dialogue.map((line, index) => (
            <Box
              key={index}
              style={{
                display: "flex",
                justifyContent: line.speaker === "karen" ? "flex-end" : "flex-start",
              }}
            >
              <Paper
                p="sm"
                radius="md"
                style={{
                  maxWidth: "85%",
                  backgroundColor:
                    line.speaker === "karen"
                      ? "var(--mantine-color-teal-0)"
                      : "var(--mantine-color-gray-1)",
                  borderBottomRightRadius: line.speaker === "karen" ? 4 : undefined,
                  borderBottomLeftRadius: line.speaker === "caller" ? 4 : undefined,
                }}
              >
                <Group gap="xs" mb={4}>
                  <ThemeIcon
                    size="xs"
                    color={line.speaker === "karen" ? "teal" : "gray"}
                    variant="transparent"
                  >
                    {line.speaker === "karen" ? <IconRobot size={12} /> : <IconUser size={12} />}
                  </ThemeIcon>
                  <Text size="xs" fw={500} c={line.speaker === "karen" ? "teal.7" : "gray.7"}>
                    {line.speaker === "karen" ? SHARED_CONTENT.brand.assistantName : "Volající"}
                  </Text>
                </Group>
                <Text size="sm" c="dark">
                  {line.text}
                </Text>
              </Paper>
            </Box>
          ))}

          <Paper p="md" radius="md" bg="gray.0" mt="md" withBorder>
            <Group justify="space-between" mb="xs">
              <Text size="sm" fw={500}>
                Výsledek:
              </Text>
              <Badge color={result.color} variant="light">
                {result.label}
              </Badge>
            </Group>
            <Text size="sm" c="dimmed">
              {result.summary}
            </Text>
          </Paper>
        </Stack>
      </Paper>
    </Container>
  );
}
