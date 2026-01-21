import { Stack, Text, Title, Table, Anchor } from "@mantine/core";
import { LegalPageLayout } from "./LegalPageLayout";

const COMPANY_INFO = {
  name: "BlockSei s.r.o.",
  address: "Korunní 2569/108, Praha 101 00",
  ico: "22345159",
  dic: "CZ22345159",
  spisovaZnacka: "C 415023/MSPH",
  soud: "Městský soud v Praze",
  email: "info@zvednu.cz",
  web: "zvednu.cz",
};

export function CompanyInfoPage() {
  return (
    <LegalPageLayout
      title="Informace o provozovateli"
      lastUpdated="13. ledna 2026"
    >
      <Stack gap="lg">
        <Text>
          Provozovatelem služby Zvednu je společnost {COMPANY_INFO.name},
          zapsaná v obchodním rejstříku vedeném {COMPANY_INFO.soud}.
        </Text>

        <Title order={2} size="h3">
          Identifikační údaje
        </Title>

        <Table>
          <Table.Tbody>
            <Table.Tr>
              <Table.Td fw={500}>Obchodní firma</Table.Td>
              <Table.Td>{COMPANY_INFO.name}</Table.Td>
            </Table.Tr>
            <Table.Tr>
              <Table.Td fw={500}>Sídlo</Table.Td>
              <Table.Td>{COMPANY_INFO.address}</Table.Td>
            </Table.Tr>
            <Table.Tr>
              <Table.Td fw={500}>IČO</Table.Td>
              <Table.Td>{COMPANY_INFO.ico}</Table.Td>
            </Table.Tr>
            <Table.Tr>
              <Table.Td fw={500}>DIČ</Table.Td>
              <Table.Td>{COMPANY_INFO.dic}</Table.Td>
            </Table.Tr>
            <Table.Tr>
              <Table.Td fw={500}>Spisová značka</Table.Td>
              <Table.Td>
                {COMPANY_INFO.spisovaZnacka}, vedená u {COMPANY_INFO.soud}
              </Table.Td>
            </Table.Tr>
          </Table.Tbody>
        </Table>

        <Title order={2} size="h3">
          Kontaktní údaje
        </Title>

        <Table>
          <Table.Tbody>
            <Table.Tr>
              <Table.Td fw={500}>E-mail</Table.Td>
              <Table.Td>
                <Anchor href={`mailto:${COMPANY_INFO.email}`}>
                  {COMPANY_INFO.email}
                </Anchor>
              </Table.Td>
            </Table.Tr>
            <Table.Tr>
              <Table.Td fw={500}>Web</Table.Td>
              <Table.Td>
                <Anchor href={`https://${COMPANY_INFO.web}`} target="_blank">
                  {COMPANY_INFO.web}
                </Anchor>
              </Table.Td>
            </Table.Tr>
          </Table.Tbody>
        </Table>

        <Title order={2} size="h3">
          Předmět podnikání
        </Title>

        <Text>
          Společnost BlockSei s.r.o. provozuje službu Zvednu – AI telefonní
          asistentku, která přijímá a zpracovává příchozí hovory jménem svých
          uživatelů. Služba využívá umělou inteligenci pro rozpoznání řeči,
          vedení konverzace a přepis hovorů.
        </Text>
      </Stack>
    </LegalPageLayout>
  );
}
