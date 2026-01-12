import {
  Accordion,
  ActionIcon,
  Box,
  Button,
  CopyButton,
  Group,
  List,
  Spoiler,
  Stack,
  Text,
  Tooltip,
} from "@mantine/core";
import { IconCheck, IconCopy, IconPhone } from "@tabler/icons-react";
import {
  REDIRECT_CODES,
  REDIRECT_ORDER,
  PHONE_SETTINGS_INSTRUCTIONS,
  getDialCode as getRedirectDialCode,
  getDeactivationCode,
  type RedirectType,
} from "../../constants/redirectCodes";

interface RedirectSetupAccordionProps {
  karenNumber: string;
  /** Which redirect types to show. Defaults to all types. */
  redirectTypes?: RedirectType[];
  /** Default expanded accordion item */
  defaultValue?: string;
  /** Whether to include the phone settings alternative section */
  showPhoneSettings?: boolean;
}

export function RedirectSetupAccordion({
  karenNumber,
  redirectTypes,
  defaultValue,
  showPhoneSettings = true,
}: RedirectSetupAccordionProps) {
  const typesToShow = redirectTypes || REDIRECT_ORDER;

  return (
    <Accordion variant="separated" defaultValue={defaultValue || typesToShow[0]}>
      {typesToShow.map((type) => {
        const code = REDIRECT_CODES[type];
        const dialCode = karenNumber ? getRedirectDialCode(type, karenNumber) : "";
        const deactivateCode = getDeactivationCode(type);
        return (
          <Accordion.Item key={type} value={type}>
            <Accordion.Control>
              <Text size="sm" fw={500}>{code.label}</Text>
            </Accordion.Control>
            <Accordion.Panel>
              <Stack gap="sm">
                <Text size="xs" c="dimmed">{code.description}</Text>

                <Button
                  variant="light"
                  leftSection={<IconPhone size={14} />}
                  disabled={!karenNumber}
                  onClick={() => { window.location.href = `tel:${dialCode}`; }}
                >
                  Aktivovat přesměrování
                </Button>
                <Spoiler maxHeight={0} showLabel="Zobrazit kód" hideLabel="Skrýt kód">
                  <Group gap="xs">
                    <Text size="sm" ff="monospace">{dialCode}</Text>
                    <CopyButton value={dialCode}>
                      {({ copied, copy }) => (
                        <Tooltip label={copied ? "Zkopírováno" : "Kopírovat"}>
                          <ActionIcon size="sm" variant="subtle" onClick={copy} color={copied ? "green" : "gray"}>
                            {copied ? <IconCheck size={14} /> : <IconCopy size={14} />}
                          </ActionIcon>
                        </Tooltip>
                      )}
                    </CopyButton>
                  </Group>
                </Spoiler>

                <Button
                  variant="subtle"
                  size="xs"
                  color="red"
                  leftSection={<IconPhone size={14} />}
                  onClick={() => { window.location.href = `tel:${deactivateCode}`; }}
                  mt="xs"
                >
                  Zrušit přesměrování
                </Button>
                <Spoiler maxHeight={0} showLabel="Zobrazit kód" hideLabel="Skrýt kód">
                  <Group gap="xs">
                    <Text size="sm" ff="monospace">{deactivateCode}</Text>
                    <CopyButton value={deactivateCode}>
                      {({ copied, copy }) => (
                        <Tooltip label={copied ? "Zkopírováno" : "Kopírovat"}>
                          <ActionIcon size="sm" variant="subtle" onClick={copy} color={copied ? "green" : "gray"}>
                            {copied ? <IconCheck size={14} /> : <IconCopy size={14} />}
                          </ActionIcon>
                        </Tooltip>
                      )}
                    </CopyButton>
                  </Group>
                </Spoiler>
              </Stack>
            </Accordion.Panel>
          </Accordion.Item>
        );
      })}

      {/* Phone settings alternative */}
      {showPhoneSettings && (
        <Accordion.Item value="phone-settings">
          <Accordion.Control>
            <Text size="sm" fw={500}>Nastavení v telefonu (alternativa)</Text>
          </Accordion.Control>
          <Accordion.Panel>
            <Stack gap="md">
              <Text size="xs" c="dimmed">
                Místo vytáčení kódů můžeš přesměrování nastavit přímo v nastavení telefonu.
              </Text>

              <Box>
                <Text size="xs" fw={500}>{PHONE_SETTINGS_INSTRUCTIONS.iphone.title}</Text>
                <List size="xs" c="dimmed" mt="xs">
                  {PHONE_SETTINGS_INSTRUCTIONS.iphone.steps.map((step, i) => (
                    <List.Item key={i}>{step}</List.Item>
                  ))}
                </List>
              </Box>

              <Box>
                <Text size="xs" fw={500}>{PHONE_SETTINGS_INSTRUCTIONS.android.title}</Text>
                <List size="xs" c="dimmed" mt="xs">
                  {PHONE_SETTINGS_INSTRUCTIONS.android.steps.map((step, i) => (
                    <List.Item key={i}>{step}</List.Item>
                  ))}
                </List>
              </Box>
            </Stack>
          </Accordion.Panel>
        </Accordion.Item>
      )}
    </Accordion>
  );
}
