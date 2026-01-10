import { Container, Group, Paper, Stack, Text, ThemeIcon, Title, Badge } from "@mantine/core";
import { IconRobot, IconPhone } from "@tabler/icons-react";
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
    text: "Dobry den, tady Karen, asistentka pana Lukase. Lukas ted nemuze prijmout hovor. Jak vam mohu pomoct?",
  },
  {
    speaker: "caller",
    text: "Dobry den, volam ohledne nabidky elektricke energie...",
  },
  {
    speaker: "karen",
    text: "Dekuji za zavolani. Lukas nema zajem o marketingove nabidky. Pokud chcete, poslete nabidku na email. Na shledanou.",
  },
];

const defaultResult: ExampleCallResult = {
  label: "Marketing",
  color: "yellow",
  summary: "Nabidka energie - ignorovat",
};

export function ExampleCall({
  scenario,
  dialogue = defaultDialogue,
  result = defaultResult,
}: ExampleCallProps) {
  return (
    <Container size="md" py={60}>
      <Title order={2} ta="center" mb="md">
        Ukazka hovoru
      </Title>
      {scenario && (
        <Text c="dimmed" ta="center" mb={40}>
          {scenario}
        </Text>
      )}
      <Paper p="xl" radius="md" withBorder>
        <Stack gap="md">
          {dialogue.map((line, index) => (
            <div key={index}>
              <Group gap="xs">
                <ThemeIcon
                  size="sm"
                  color={line.speaker === "karen" ? "blue" : "gray"}
                  variant="light"
                >
                  {line.speaker === "karen" ? <IconRobot size={14} /> : <IconPhone size={14} />}
                </ThemeIcon>
                <Text size="sm" fw={500}>
                  {line.speaker === "karen" ? SHARED_CONTENT.brand.assistantName : "Volajici"}:
                </Text>
              </Group>
              <Text c="dimmed" size="sm" pl={28}>
                "{line.text}"
              </Text>
            </div>
          ))}

          <Paper p="md" radius="sm" bg="gray.0" mt="md">
            <Group justify="space-between">
              <Text size="sm" fw={500}>
                Vysledek:
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
