import {
  Box,
  Container,
  Title,
  Text,
  SimpleGrid,
  Card,
  Badge,
  Button,
  List,
  ThemeIcon,
  Group,
} from "@mantine/core";
import { IconCheck, IconStar } from "@tabler/icons-react";

interface PricingTier {
  name: string;
  price: string;
  period?: string;
  callLimit: string;
  features: string[];
  buttonText: string;
  buttonVariant: "filled" | "outline" | "light";
  highlighted?: boolean;
  onButtonClick: () => void;
}

interface PricingSectionProps {
  onTrialClick: () => void;
  onBuyClick: () => void;
  onContactClick: () => void;
}

export function PricingSection({ onTrialClick, onBuyClick, onContactClick }: PricingSectionProps) {
  const tiers: PricingTier[] = [
    {
      name: "Zkušební",
      price: "Zdarma",
      callLimit: "20 hovorů nebo 14 dní",
      features: ["Základní přepis", "Push notifikace", "Klasifikace hovorů"],
      buttonText: "Vyzkoušet zdarma",
      buttonVariant: "outline",
      onButtonClick: onTrialClick,
    },
    {
      name: "Základ",
      price: "199 Kč",
      period: "/měsíc",
      callLimit: "50 hovorů",
      features: ["Plná klasifikace", "Push notifikace", "Přepis hovoru", "E-mailové shrnutí"],
      buttonText: "Koupit",
      buttonVariant: "outline",
      onButtonClick: onBuyClick,
    },
    {
      name: "Pro",
      price: "499 Kč",
      period: "/měsíc",
      callLimit: "200 hovorů",
      features: [
        "Vše ze Základu",
        "VIP přepojení",
        "Vlastní hlas",
        "Prioritní podpora",
      ],
      buttonText: "Koupit",
      buttonVariant: "filled",
      highlighted: true,
      onButtonClick: onBuyClick,
    },
    {
      name: "Firma",
      price: "Na míru",
      callLimit: "Neomezeno",
      features: [
        "Vše z Pro",
        "Více telefonních čísel",
        "Týmový přístup",
        "API přístup",
        "Garance SLA",
      ],
      buttonText: "Kontaktujte nás",
      buttonVariant: "outline",
      onButtonClick: onContactClick,
    },
  ];

  return (
    <Box py={60} id="pricing">
      <Container size="lg">
        <Title order={2} ta="center" mb="sm">
          Ceník
        </Title>
        <Text ta="center" c="dimmed" mb={40} maw={600} mx="auto">
          Vyberte si tarif podle počtu hovorů. Tarif můžete kdykoliv změnit.
        </Text>
        <SimpleGrid cols={{ base: 1, sm: 2, lg: 4 }} spacing="lg">
          {tiers.map((tier) => (
            <Card
              key={tier.name}
              shadow={tier.highlighted ? "md" : "sm"}
              padding="lg"
              radius="md"
              withBorder
              style={{
                borderColor: tier.highlighted ? "var(--mantine-color-teal-5)" : undefined,
                borderWidth: tier.highlighted ? 2 : 1,
              }}
            >
              <Group justify="space-between" mb="xs">
                <Text fw={500} size="lg">
                  {tier.name}
                </Text>
                {tier.highlighted && (
                  <Badge color="teal" variant="light" leftSection={<IconStar size={12} />}>
                    Oblíbený
                  </Badge>
                )}
              </Group>

              <Group align="baseline" gap={4} mb="xs">
                <Text fz={32} fw={700}>
                  {tier.price}
                </Text>
                {tier.period && (
                  <Text size="sm" c="dimmed">
                    {tier.period}
                  </Text>
                )}
              </Group>

              <Text size="sm" c="dimmed" mb="md">
                {tier.callLimit}
              </Text>

              <List
                spacing="xs"
                size="sm"
                mb="xl"
                icon={
                  <ThemeIcon color="teal" size={20} radius="xl">
                    <IconCheck size={12} />
                  </ThemeIcon>
                }
              >
                {tier.features.map((feature) => (
                  <List.Item key={feature}>{feature}</List.Item>
                ))}
              </List>

              <Button
                fullWidth
                variant={tier.buttonVariant}
                color={tier.highlighted ? "teal" : "gray"}
                onClick={tier.onButtonClick}
              >
                {tier.buttonText}
              </Button>
            </Card>
          ))}
        </SimpleGrid>

        <Text ta="center" c="dimmed" mt="xl" size="sm">
          Roční předplatné: sleva 20 % (2 měsíce zdarma)
        </Text>
      </Container>
    </Box>
  );
}
