import { Box, Container, Title, Text, Table, ThemeIcon } from "@mantine/core";
import { IconCheck, IconX } from "@tabler/icons-react";

interface ComparisonRow {
  feature: string;
  zvednu: boolean | string;
  secretary: boolean | string;
  voicemail: boolean | string;
  truecaller: boolean | string;
}

const comparisonData: ComparisonRow[] = [
  {
    feature: "Zvedne a mluví",
    zvednu: true,
    secretary: true,
    voicemail: false,
    truecaller: false,
  },
  {
    feature: "Zjistí důvod volání",
    zvednu: true,
    secretary: true,
    voicemail: false,
    truecaller: false,
  },
  {
    feature: "Filtruje spam",
    zvednu: true,
    secretary: false,
    voicemail: false,
    truecaller: true,
  },
  {
    feature: "Přepis hovoru",
    zvednu: true,
    secretary: "Někdy",
    voicemail: false,
    truecaller: false,
  },
  {
    feature: "Dostupnost 24/7",
    zvednu: true,
    secretary: false,
    voicemail: true,
    truecaller: true,
  },
  {
    feature: "Plně česky",
    zvednu: true,
    secretary: true,
    voicemail: false,
    truecaller: false,
  },
];

function CellValue({ value }: { value: boolean | string }) {
  if (typeof value === "string") {
    return <Text size="sm">{value}</Text>;
  }
  if (value) {
    return (
      <ThemeIcon color="teal" variant="light" size="sm" radius="xl">
        <IconCheck size={14} />
      </ThemeIcon>
    );
  }
  return (
    <ThemeIcon color="gray" variant="light" size="sm" radius="xl">
      <IconX size={14} />
    </ThemeIcon>
  );
}

export function ComparisonTable() {
  return (
    <Box py={60}>
      <Container size="lg">
        <Title order={2} ta="center" mb="sm">
          Proč Zvednu?
        </Title>
        <Text ta="center" c="dimmed" mb={40} maw={600} mx="auto">
          Porovnání s alternativami
        </Text>
        <Box style={{ overflowX: "auto" }}>
          <Table highlightOnHover withTableBorder withColumnBorders>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Funkce</Table.Th>
                <Table.Th ta="center" style={{ backgroundColor: "var(--mantine-color-teal-0)" }}>
                  <Text fw={700} c="teal">Zvednu</Text>
                  <Text size="xs" c="dimmed">od 199 Kč/měs</Text>
                </Table.Th>
                <Table.Th ta="center">
                  <Text fw={500}>Sekretářka</Text>
                  <Text size="xs" c="dimmed">25–40 tis. Kč/měs</Text>
                </Table.Th>
                <Table.Th ta="center">
                  <Text fw={500}>Voicemail</Text>
                  <Text size="xs" c="dimmed">Zdarma</Text>
                </Table.Th>
                <Table.Th ta="center">
                  <Text fw={500}>Truecaller</Text>
                  <Text size="xs" c="dimmed">~50 Kč/měs</Text>
                </Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {comparisonData.map((row) => (
                <Table.Tr key={row.feature}>
                  <Table.Td>{row.feature}</Table.Td>
                  <Table.Td ta="center" style={{ backgroundColor: "var(--mantine-color-teal-0)" }}>
                    <CellValue value={row.zvednu} />
                  </Table.Td>
                  <Table.Td ta="center">
                    <CellValue value={row.secretary} />
                  </Table.Td>
                  <Table.Td ta="center">
                    <CellValue value={row.voicemail} />
                  </Table.Td>
                  <Table.Td ta="center">
                    <CellValue value={row.truecaller} />
                  </Table.Td>
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        </Box>
      </Container>
    </Box>
  );
}
