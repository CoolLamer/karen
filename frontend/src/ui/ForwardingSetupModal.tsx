import {
  Modal,
  Stack,
  Alert,
  Text,
  Button,
} from "@mantine/core";
import { RedirectSetupAccordion } from "./components/RedirectSetupAccordion";

interface ForwardingSetupModalProps {
  opened: boolean;
  onClose: () => void;
  karenNumber: string | undefined;
}

export function ForwardingSetupModal({ opened, onClose, karenNumber }: ForwardingSetupModalProps) {
  return (
    <Modal
      opened={opened}
      onClose={onClose}
      title="Jak nastavit přesměrování"
      size="lg"
      centered
    >
      <Stack gap="md">
        <Alert color="blue" variant="light">
          <Text size="sm">
            Přesměrování se nastavuje vytočením speciálního kódu na telefonu.
            Otevři tuto stránku na mobilu a klikni na tlačítko – automaticky se vytočí
            aktivační kód a na obrazovce uvidíš potvrzení od operátora.
          </Text>
          <Text size="sm" mt="xs" c="dimmed">
            Na počítači tlačítko nefunguje – musíš kód vytočit ručně nebo otevřít stránku na telefonu.
          </Text>
        </Alert>
        <Text size="sm" c="dimmed">
          Pro kompletní pokrytí doporučujeme nastavit všechny tři typy přesměrování.
        </Text>

        <Alert color="yellow" variant="light">
          <Text size="sm">
            <Text span fw={500}>Zobrazuje se chyba?</Text> Pokud máš již nastavené přesměrování na jiné číslo,
            musíš ho nejdřív zrušit. Použij tlačítko „Zrušit přesměrování" u příslušného typu.
          </Text>
        </Alert>

        <RedirectSetupAccordion karenNumber={karenNumber || ""} defaultValue="noAnswer" />

        <Button variant="subtle" onClick={onClose} fullWidth>
          Zavřít
        </Button>
      </Stack>
    </Modal>
  );
}
