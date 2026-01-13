import { Stack, Text, Title, List, Table, Anchor } from "@mantine/core";
import { Link } from "react-router-dom";
import { LegalPageLayout } from "./LegalPageLayout";

export function PrivacyPolicyPage() {
  return (
    <LegalPageLayout
      title="Ochrana osobních údajů"
      lastUpdated="13. ledna 2026"
    >
      <Stack gap="xl">
        {/* 1. Správce osobních údajů */}
        <section>
          <Title order={2} size="h3" mb="md">
            1. Správce osobních údajů
          </Title>
          <Text mb="sm">
            Správcem osobních údajů je společnost BlockSei s.r.o., IČO:
            22345159, se sídlem Korunní 2569/108, Praha 101 00 (dále jen
            „Správce").
          </Text>
          <Text>
            Kontakt pro záležitosti ochrany osobních údajů:{" "}
            <Anchor href="mailto:info@zvednu.cz">info@zvednu.cz</Anchor>
          </Text>
        </section>

        {/* 2. Jaké údaje zpracováváme */}
        <section>
          <Title order={2} size="h3" mb="md">
            2. Jaké údaje zpracováváme
          </Title>
          <Text mb="sm">
            V rámci poskytování služby Zvednu zpracováváme následující kategorie
            osobních údajů:
          </Text>

          <Title order={3} size="h4" mb="xs" mt="md">
            Registrační údaje
          </Title>
          <List spacing="xs" mb="sm">
            <List.Item>Telefonní číslo (slouží jako identifikátor účtu)</List.Item>
            <List.Item>Jméno (volitelné, pro personalizaci)</List.Item>
          </List>

          <Title order={3} size="h4" mb="xs" mt="md">
            Údaje o hovorech
          </Title>
          <List spacing="xs" mb="sm">
            <List.Item>
              Telefonní čísla volajících
            </List.Item>
            <List.Item>
              Audio nahrávky hovorů
            </List.Item>
            <List.Item>
              Textové přepisy hovorů
            </List.Item>
            <List.Item>
              Shrnutí a klasifikace hovorů generované AI
            </List.Item>
            <List.Item>
              Metadata hovorů (datum, čas, délka trvání)
            </List.Item>
          </List>

          <Title order={3} size="h4" mb="xs" mt="md">
            Technické údaje
          </Title>
          <List spacing="xs">
            <List.Item>IP adresa při přístupu do aplikace</List.Item>
            <List.Item>Informace o zařízení a prohlížeči</List.Item>
            <List.Item>Logy přístupu a aktivit v aplikaci</List.Item>
          </List>
        </section>

        {/* 3. Účely zpracování */}
        <section>
          <Title order={2} size="h3" mb="md">
            3. Účely zpracování
          </Title>
          <Table mb="sm">
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Účel</Table.Th>
                <Table.Th>Právní základ</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              <Table.Tr>
                <Table.Td>Poskytování služby (příjem hovorů, přepisy)</Table.Td>
                <Table.Td>Plnění smlouvy</Table.Td>
              </Table.Tr>
              <Table.Tr>
                <Table.Td>Správa uživatelského účtu</Table.Td>
                <Table.Td>Plnění smlouvy</Table.Td>
              </Table.Tr>
              <Table.Tr>
                <Table.Td>Zlepšování služby a AI modelů</Table.Td>
                <Table.Td>Oprávněný zájem</Table.Td>
              </Table.Tr>
              <Table.Tr>
                <Table.Td>Fakturace a účetnictví</Table.Td>
                <Table.Td>Plnění právní povinnosti</Table.Td>
              </Table.Tr>
              <Table.Tr>
                <Table.Td>Zasílání obchodních sdělení</Table.Td>
                <Table.Td>Souhlas</Table.Td>
              </Table.Tr>
              <Table.Tr>
                <Table.Td>Zabezpečení a prevence zneužití</Table.Td>
                <Table.Td>Oprávněný zájem</Table.Td>
              </Table.Tr>
            </Table.Tbody>
          </Table>
        </section>

        {/* 4. Nahrávání hovorů */}
        <section>
          <Title order={2} size="h3" mb="md">
            4. Nahrávání hovorů
          </Title>
          <Text mb="sm">
            Služba Zvednu nahrává příchozí hovory za účelem:
          </Text>
          <List spacing="xs" mb="sm">
            <List.Item>Vytvoření textového přepisu pro uživatele</List.Item>
            <List.Item>Generování shrnutí a klasifikace hovoru pomocí AI</List.Item>
            <List.Item>Umožnění zpětného poslechu hovoru uživatelem</List.Item>
          </List>
          <Text mb="sm">
            <Text component="span" fw={500}>
              Informování volajících:
            </Text>{" "}
            AI asistentka se při každém hovoru představí a volající jsou tak
            informováni, že hovoří s automatizovaným systémem. Nahrávání probíhá
            na základě oprávněného zájmu uživatele služby na dokumentaci
            komunikace.
          </Text>
          <Text>
            Nahrávky jsou přístupné pouze uživateli, kterému hovor patří, a
            oprávněným zaměstnancům Správce pro účely technické podpory.
          </Text>
        </section>

        {/* 5. AI zpracování */}
        <section>
          <Title order={2} size="h3" mb="md">
            5. Zpracování umělou inteligencí
          </Title>
          <Text mb="sm">
            Služba využívá technologie umělé inteligence pro:
          </Text>
          <List spacing="xs" mb="sm">
            <List.Item>
              <Text component="span" fw={500}>
                Rozpoznávání řeči (STT):
              </Text>{" "}
              Převod audio nahrávky na text
            </List.Item>
            <List.Item>
              <Text component="span" fw={500}>
                Konverzační AI (LLM):
              </Text>{" "}
              Vedení dialogu s volajícím, generování odpovědí
            </List.Item>
            <List.Item>
              <Text component="span" fw={500}>
                Syntéza řeči (TTS):
              </Text>{" "}
              Převod textových odpovědí na hlas
            </List.Item>
            <List.Item>
              <Text component="span" fw={500}>
                Klasifikace a shrnutí:
              </Text>{" "}
              Automatická kategorizace hovorů (spam, důležité, apod.)
            </List.Item>
          </List>
          <Text mb="sm">
            <Text component="span" fw={500}>
              Automatizované rozhodování:
            </Text>{" "}
            AI může automaticky klasifikovat hovory jako spam nebo marketingové.
            Tato klasifikace slouží pouze jako doporučení pro uživatele a
            neomezuje přístup k údajům o hovoru. Uživatel má vždy přístup ke
            kompletním informacím o všech hovorech.
          </Text>
          <Text>
            Na požádání zajistíme lidský přezkum jakéhokoli automatizovaného
            rozhodnutí.
          </Text>
        </section>

        {/* 6. Doba uchovávání */}
        <section>
          <Title order={2} size="h3" mb="md">
            6. Doba uchovávání údajů
          </Title>
          <Table>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Typ údajů</Table.Th>
                <Table.Th>Doba uchovávání</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              <Table.Tr>
                <Table.Td>Audio nahrávky hovorů</Table.Td>
                <Table.Td>30 dní od hovoru</Table.Td>
              </Table.Tr>
              <Table.Tr>
                <Table.Td>Přepisy hovorů</Table.Td>
                <Table.Td>Po dobu trvání účtu + 30 dní</Table.Td>
              </Table.Tr>
              <Table.Tr>
                <Table.Td>Údaje o účtu</Table.Td>
                <Table.Td>Po dobu trvání účtu + 30 dní</Table.Td>
              </Table.Tr>
              <Table.Tr>
                <Table.Td>Fakturační údaje</Table.Td>
                <Table.Td>10 let (zákonná povinnost)</Table.Td>
              </Table.Tr>
              <Table.Tr>
                <Table.Td>Bezpečnostní logy</Table.Td>
                <Table.Td>6 měsíců</Table.Td>
              </Table.Tr>
            </Table.Tbody>
          </Table>
        </section>

        {/* 7. Sdílení údajů */}
        <section>
          <Title order={2} size="h3" mb="md">
            7. Sdílení údajů s třetími stranami
          </Title>
          <Text mb="sm">
            Pro poskytování služby spolupracujeme s následujícími kategoriemi
            zpracovatelů:
          </Text>
          <List spacing="xs" mb="sm">
            <List.Item>
              <Text component="span" fw={500}>
                Poskytovatel telefonních služeb (Twilio):
              </Text>{" "}
              Zpracování telefonních hovorů
            </List.Item>
            <List.Item>
              <Text component="span" fw={500}>
                Poskytovatelé AI služeb:
              </Text>{" "}
              Rozpoznávání řeči, konverzační AI, syntéza hlasu
            </List.Item>
            <List.Item>
              <Text component="span" fw={500}>
                Poskytovatel hostingu:
              </Text>{" "}
              Ukládání dat a provoz aplikace
            </List.Item>
            <List.Item>
              <Text component="span" fw={500}>
                Platební brána:
              </Text>{" "}
              Zpracování plateb
            </List.Item>
          </List>
          <Text>
            Se všemi zpracovateli máme uzavřeny smlouvy o zpracování osobních
            údajů v souladu s GDPR.
          </Text>
        </section>

        {/* 8. Mezinárodní přenosy */}
        <section>
          <Title order={2} size="h3" mb="md">
            8. Předávání údajů mimo EU
          </Title>
          <Text mb="sm">
            Někteří naši zpracovatelé (zejména poskytovatelé AI služeb a Twilio)
            mohou mít sídlo nebo servery v USA. Pro tyto přenosy využíváme:
          </Text>
          <List spacing="xs">
            <List.Item>
              Standardní smluvní doložky schválené Evropskou komisí
            </List.Item>
            <List.Item>
              Posouzení úrovně ochrany v cílové zemi
            </List.Item>
            <List.Item>
              Dodatečná technická a organizační opatření
            </List.Item>
          </List>
        </section>

        {/* 9. Vaše práva */}
        <section>
          <Title order={2} size="h3" mb="md">
            9. Vaše práva
          </Title>
          <Text mb="sm">
            Jako subjekt údajů máte podle GDPR následující práva:
          </Text>
          <List spacing="xs" mb="sm">
            <List.Item>
              <Text component="span" fw={500}>
                Právo na přístup:
              </Text>{" "}
              Můžete požádat o kopii všech údajů, které o vás zpracováváme
            </List.Item>
            <List.Item>
              <Text component="span" fw={500}>
                Právo na opravu:
              </Text>{" "}
              Můžete požádat o opravu nepřesných údajů
            </List.Item>
            <List.Item>
              <Text component="span" fw={500}>
                Právo na výmaz:
              </Text>{" "}
              Můžete požádat o smazání vašich údajů („právo být zapomenut")
            </List.Item>
            <List.Item>
              <Text component="span" fw={500}>
                Právo na omezení zpracování:
              </Text>{" "}
              Můžete požádat o omezení způsobu zpracování
            </List.Item>
            <List.Item>
              <Text component="span" fw={500}>
                Právo na přenositelnost:
              </Text>{" "}
              Můžete požádat o export vašich údajů ve strojově čitelném formátu
            </List.Item>
            <List.Item>
              <Text component="span" fw={500}>
                Právo vznést námitku:
              </Text>{" "}
              Můžete vznést námitku proti zpracování založenému na oprávněném
              zájmu
            </List.Item>
            <List.Item>
              <Text component="span" fw={500}>
                Právo odvolat souhlas:
              </Text>{" "}
              Pokud zpracování probíhá na základě souhlasu, můžete jej kdykoliv
              odvolat
            </List.Item>
          </List>
          <Text>
            Pro uplatnění těchto práv nás kontaktujte na{" "}
            <Anchor href="mailto:info@zvednu.cz">info@zvednu.cz</Anchor>. Na
            žádost odpovíme do 30 dnů.
          </Text>
        </section>

        {/* 10. Zabezpečení */}
        <section>
          <Title order={2} size="h3" mb="md">
            10. Zabezpečení údajů
          </Title>
          <Text mb="sm">
            Pro ochranu vašich údajů používáme následující opatření:
          </Text>
          <List spacing="xs">
            <List.Item>Šifrování dat při přenosu (TLS/HTTPS)</List.Item>
            <List.Item>Šifrování citlivých dat v klidu</List.Item>
            <List.Item>Přístup k údajům pouze oprávněným osobám</List.Item>
            <List.Item>Pravidelné bezpečnostní audity</List.Item>
            <List.Item>Monitoring a detekce bezpečnostních incidentů</List.Item>
          </List>
        </section>

        {/* 11. Stížnosti */}
        <section>
          <Title order={2} size="h3" mb="md">
            11. Podání stížnosti
          </Title>
          <Text mb="sm">
            Pokud se domníváte, že zpracování vašich osobních údajů porušuje
            GDPR, máte právo podat stížnost u dozorového úřadu:
          </Text>
          <Text mb="sm">
            <Text component="span" fw={500}>
              Úřad pro ochranu osobních údajů
            </Text>
            <br />
            Pplk. Sochora 27
            <br />
            170 00 Praha 7
            <br />
            <Anchor href="https://www.uoou.cz" target="_blank">
              www.uoou.cz
            </Anchor>
          </Text>
          <Text>
            Doporučujeme vám však nejprve kontaktovat nás, abychom mohli vaše
            obavy vyřešit přímo.
          </Text>
        </section>

        {/* 12. Změny */}
        <section>
          <Title order={2} size="h3" mb="md">
            12. Změny těchto zásad
          </Title>
          <Text mb="sm">
            Tyto zásady můžeme příležitostně aktualizovat. O významných změnách
            vás budeme informovat e-mailem nebo prostřednictvím aplikace.
          </Text>
          <Text>
            Aktuální verze je vždy dostupná na této stránce s uvedením data
            poslední aktualizace.
          </Text>
        </section>

        {/* 13. Kontakt */}
        <section>
          <Title order={2} size="h3" mb="md">
            13. Kontakt
          </Title>
          <Text mb="sm">
            S dotazy ohledně ochrany osobních údajů se na nás můžete obrátit:
          </Text>
          <Text mb="sm">
            E-mail:{" "}
            <Anchor href="mailto:info@zvednu.cz">info@zvednu.cz</Anchor>
          </Text>
          <Text>
            Kompletní kontaktní údaje naleznete na stránce{" "}
            <Anchor component={Link} to="/informace-o-provozovateli">
              Informace o provozovateli
            </Anchor>
            .
          </Text>
        </section>
      </Stack>
    </LegalPageLayout>
  );
}
