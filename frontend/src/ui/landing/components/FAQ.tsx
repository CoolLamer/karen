import { Accordion, Container, Title, Text } from "@mantine/core";
import { FAQItem } from "../content/index";

interface FAQProps {
  items: FAQItem[];
  compact?: boolean;
}

export function FAQ({ items, compact = false }: FAQProps) {
  const accordion = (
    <Accordion variant="separated">
      {items.map((item, index) => (
        <Accordion.Item key={index} value={`faq-${index}`}>
          <Accordion.Control>
            <Text fw={500}>{item.question}</Text>
          </Accordion.Control>
          <Accordion.Panel>
            <Text c="dimmed">{item.answer}</Text>
          </Accordion.Panel>
        </Accordion.Item>
      ))}
    </Accordion>
  );

  if (compact) {
    return accordion;
  }

  return (
    <Container size="md" py={60} id="faq">
      <Title order={2} ta="center" mb={40}>
        Časté dotazy
      </Title>
      {accordion}
    </Container>
  );
}
