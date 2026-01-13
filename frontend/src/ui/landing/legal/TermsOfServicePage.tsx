import { Stack, Text, Title, List, Anchor } from "@mantine/core";
import { Link } from "react-router-dom";
import { LegalPageLayout } from "./LegalPageLayout";

export function TermsOfServicePage() {
  return (
    <LegalPageLayout title="Obchodní podmínky" lastUpdated="13. ledna 2026">
      <Stack gap="xl">
        {/* 1. Úvodní ustanovení */}
        <section>
          <Title order={2} size="h3" mb="md">
            1. Úvodní ustanovení
          </Title>
          <Text mb="sm">
            Tyto obchodní podmínky (dále jen „Podmínky") upravují vztahy mezi
            společností BlockSei s.r.o., IČO: 22345159, se sídlem Korunní
            2569/108, Praha 101 00, zapsanou v obchodním rejstříku vedeném
            Městským soudem v Praze, spisová značka C 415023/MSPH (dále jen
            „Provozovatel"), a uživateli služby Zvednu (dále jen „Uživatel").
          </Text>
          <Text>
            Služba Zvednu je AI telefonní asistentka, která přijímá příchozí
            hovory přesměrované z telefonu Uživatele, vede s volajícími
            konverzaci a poskytuje Uživateli přehled o hovorech včetně přepisů.
          </Text>
        </section>

        {/* 2. Definice pojmů */}
        <section>
          <Title order={2} size="h3" mb="md">
            2. Definice pojmů
          </Title>
          <List spacing="xs">
            <List.Item>
              <Text component="span" fw={500}>
                Služba
              </Text>{" "}
              – služba Zvednu poskytovaná Provozovatelem prostřednictvím webové
              aplikace a telefonního systému.
            </List.Item>
            <List.Item>
              <Text component="span" fw={500}>
                Uživatel
              </Text>{" "}
              – fyzická nebo právnická osoba, která se zaregistrovala ke Službě
              a využívá ji.
            </List.Item>
            <List.Item>
              <Text component="span" fw={500}>
                Účet
              </Text>{" "}
              – uživatelský účet vytvořený při registraci, přístupný po ověření
              telefonního čísla.
            </List.Item>
            <List.Item>
              <Text component="span" fw={500}>
                Hovor
              </Text>{" "}
              – telefonní hovor přesměrovaný na telefonní číslo Služby a
              zpracovaný AI asistentkou.
            </List.Item>
            <List.Item>
              <Text component="span" fw={500}>
                Přepis
              </Text>{" "}
              – textový záznam obsahu Hovoru vytvořený pomocí technologie
              rozpoznávání řeči.
            </List.Item>
          </List>
        </section>

        {/* 3. Registrace a účet */}
        <section>
          <Title order={2} size="h3" mb="md">
            3. Registrace a účet
          </Title>
          <Text mb="sm">
            Pro využívání Služby je nutná registrace prostřednictvím ověření
            telefonního čísla. Uživatel je povinen uvádět pravdivé údaje.
          </Text>
          <Text mb="sm">
            Uživatel je odpovědný za zabezpečení přístupu ke svému Účtu a nesmí
            sdílet přístupové údaje s třetími osobami.
          </Text>
          <Text>
            Provozovatel si vyhrazuje právo zrušit Účet Uživatele v případě
            porušení těchto Podmínek.
          </Text>
        </section>

        {/* 4. Popis služby */}
        <section>
          <Title order={2} size="h3" mb="md">
            4. Popis služby
          </Title>
          <Text mb="sm">Služba Zvednu poskytuje následující funkce:</Text>
          <List spacing="xs" mb="sm">
            <List.Item>
              Přijímání příchozích hovorů přesměrovaných z telefonu Uživatele
            </List.Item>
            <List.Item>
              Vedení konverzace s volajícími pomocí AI asistentky
            </List.Item>
            <List.Item>Nahrávání a přepis hovorů</List.Item>
            <List.Item>
              Zobrazení přehledu hovorů včetně shrnutí a kontextu
            </List.Item>
            <List.Item>Filtrování spamu a marketingových hovorů</List.Item>
            <List.Item>Přepojení urgentních hovorů na Uživatele</List.Item>
          </List>
          <Text>
            Služba vyžaduje nastavení přesměrování hovorů na straně Uživatele.
            Přesměrování se aktivuje pomocí standardních GSM kódů operátora.
          </Text>
        </section>

        {/* 5. Podmínky užívání */}
        <section>
          <Title order={2} size="h3" mb="md">
            5. Podmínky užívání
          </Title>
          <Text mb="sm">Uživatel se zavazuje:</Text>
          <List spacing="xs" mb="sm">
            <List.Item>
              Využívat Službu pouze k zákonným účelům
            </List.Item>
            <List.Item>
              Nezneužívat Službu k obtěžování třetích osob
            </List.Item>
            <List.Item>
              Nezasahovat do technického provozu Služby
            </List.Item>
            <List.Item>
              Informovat volající o tom, že hovor může být nahráván (toto
              zajišťuje AI asistentka automaticky při představení)
            </List.Item>
          </List>
          <Text>
            Provozovatel si vyhrazuje právo omezit nebo ukončit poskytování
            Služby Uživateli, který porušuje tyto Podmínky.
          </Text>
        </section>

        {/* 6. Platební podmínky */}
        <section>
          <Title order={2} size="h3" mb="md">
            6. Platební podmínky
          </Title>
          <Text mb="sm">
            Služba je poskytována v různých cenových plánech. Aktuální ceník je
            k dispozici na webových stránkách Služby.
          </Text>
          <Text mb="sm">
            Platby jsou účtovány předem na měsíční bázi. V případě nezaplacení
            může být Služba pozastavena.
          </Text>
          <Text>
            Uživatel má právo na vrácení nevyužité části předplatného v případě
            zrušení Účtu, pokud o to požádá do 14 dnů od platby (pro
            spotřebitele v souladu s právem na odstoupení od smlouvy).
          </Text>
        </section>

        {/* 7. Omezení odpovědnosti */}
        <section>
          <Title order={2} size="h3" mb="md">
            7. Omezení odpovědnosti
          </Title>
          <Text mb="sm">
            Provozovatel nezaručuje nepřetržitou dostupnost Služby. Služba může
            být dočasně nedostupná z důvodu údržby nebo technických problémů.
          </Text>
          <Text mb="sm">
            AI asistentka využívá technologie rozpoznávání řeči a umělé
            inteligence, které nemusí být vždy 100% přesné. Provozovatel
            neodpovídá za případné nepřesnosti v přepisech nebo shrnutích
            hovorů.
          </Text>
          <Text>
            Odpovědnost Provozovatele za škodu je omezena na výši zaplaceného
            předplatného za poslední 3 měsíce.
          </Text>
        </section>

        {/* 8. Duševní vlastnictví */}
        <section>
          <Title order={2} size="h3" mb="md">
            8. Duševní vlastnictví
          </Title>
          <Text mb="sm">
            Veškerá práva k Službě, včetně softwaru, grafiky a obsahu, náleží
            Provozovateli nebo jeho poskytovatelům licence.
          </Text>
          <Text>
            Uživatel si ponechává veškerá práva k obsahu svých hovorů.
            Provozovatel získává licenci k užití tohoto obsahu pouze za účelem
            poskytování Služby.
          </Text>
        </section>

        {/* 9. Změna podmínek */}
        <section>
          <Title order={2} size="h3" mb="md">
            9. Změna podmínek
          </Title>
          <Text mb="sm">
            Provozovatel si vyhrazuje právo tyto Podmínky kdykoliv změnit. O
            změnách bude Uživatel informován e-mailem nebo prostřednictvím
            aplikace nejméně 14 dní před jejich účinností.
          </Text>
          <Text>
            Pokud Uživatel se změnami nesouhlasí, má právo Účet zrušit před
            nabytím účinnosti změn.
          </Text>
        </section>

        {/* 10. Řešení sporů */}
        <section>
          <Title order={2} size="h3" mb="md">
            10. Řešení sporů
          </Title>
          <Text mb="sm">
            Tyto Podmínky se řídí právním řádem České republiky. Případné spory
            budou řešeny příslušnými soudy České republiky.
          </Text>
          <Text mb="sm">
            Spotřebitelé mají právo na mimosoudní řešení spotřebitelských sporů.
            Příslušným subjektem je Česká obchodní inspekce (
            <Anchor href="https://www.coi.cz" target="_blank">
              www.coi.cz
            </Anchor>
            ).
          </Text>
          <Text>
            Pro online řešení sporů mohou spotřebitelé využít platformu ODR
            Evropské komise na adrese{" "}
            <Anchor
              href="https://ec.europa.eu/consumers/odr"
              target="_blank"
            >
              ec.europa.eu/consumers/odr
            </Anchor>
            .
          </Text>
        </section>

        {/* 11. Kontaktní údaje */}
        <section>
          <Title order={2} size="h3" mb="md">
            11. Kontaktní údaje
          </Title>
          <Text mb="sm">
            V případě dotazů nebo připomínek nás kontaktujte:
          </Text>
          <Text mb="sm">
            E-mail:{" "}
            <Anchor href="mailto:info@zvednu.cz">info@zvednu.cz</Anchor>
          </Text>
          <Text>
            Kompletní údaje o Provozovateli naleznete na stránce{" "}
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
