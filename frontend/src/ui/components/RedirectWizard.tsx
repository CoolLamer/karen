import { useState } from "react";
import {
  ActionIcon,
  Box,
  Button,
  CopyButton,
  Group,
  List,
  Paper,
  Progress,
  SegmentedControl,
  Stack,
  Text,
  ThemeIcon,
  Tooltip,
  UnstyledButton,
} from "@mantine/core";
import {
  IconCheck,
  IconCopy,
  IconPhone,
  IconArrowRight,
  IconRefresh,
  IconPlayerSkipForward,
  IconCircleCheck,
  IconCircle,
} from "@tabler/icons-react";
import {
  CLEAR_ALL_REDIRECTS_CODE,
  NO_ANSWER_TIME_OPTIONS,
  DEFAULT_NO_ANSWER_TIME,
  getDialCode,
  type NoAnswerTime,
} from "../../constants/redirectCodes";

export type WizardStep = "intro" | "clear" | "noAnswer" | "busy" | "unreachable" | "complete";

export type StepStatus = "pending" | "completed" | "skipped";

export interface StepStatuses {
  clear: StepStatus;
  noAnswer: StepStatus;
  busy: StepStatus;
  unreachable: StepStatus;
}

interface RedirectWizardProps {
  karenNumber: string;
  onComplete: () => void;
}

const WIZARD_STEPS: WizardStep[] = ["intro", "clear", "noAnswer", "busy", "unreachable", "complete"];

const REDIRECT_STEPS: (keyof StepStatuses)[] = ["clear", "noAnswer", "busy", "unreachable"];

function getStepIndex(step: WizardStep): number {
  return WIZARD_STEPS.indexOf(step);
}

function getProgressPercent(step: WizardStep): number {
  const index = getStepIndex(step);
  // intro = 0%, clear = 20%, noAnswer = 40%, busy = 60%, unreachable = 80%, complete = 100%
  return (index / (WIZARD_STEPS.length - 1)) * 100;
}

interface WizardProgressProps {
  currentStep: WizardStep;
  stepStatuses: StepStatuses;
}

function WizardProgress({ currentStep, stepStatuses }: WizardProgressProps) {
  const labels = {
    clear: "Vymazat",
    noAnswer: "Nezvedám",
    busy: "Obsazeno",
    unreachable: "Nedostupný",
  };

  return (
    <Stack gap="xs">
      <Progress value={getProgressPercent(currentStep)} size="sm" radius="xl" />
      <Group justify="space-between" gap="xs">
        {REDIRECT_STEPS.map((step) => {
          const status = stepStatuses[step];
          const isCurrent = currentStep === step;
          const isPast = getStepIndex(currentStep) > getStepIndex(step);

          return (
            <Group key={step} gap={4}>
              {status === "completed" ? (
                <ThemeIcon size="xs" color="green" variant="filled" radius="xl">
                  <IconCircleCheck size={12} />
                </ThemeIcon>
              ) : status === "skipped" ? (
                <ThemeIcon size="xs" color="gray" variant="light" radius="xl">
                  <IconPlayerSkipForward size={10} />
                </ThemeIcon>
              ) : isCurrent ? (
                <ThemeIcon size="xs" color="blue" variant="filled" radius="xl">
                  <IconCircle size={12} />
                </ThemeIcon>
              ) : isPast ? (
                <ThemeIcon size="xs" color="gray" variant="light" radius="xl">
                  <IconCircle size={12} />
                </ThemeIcon>
              ) : (
                <ThemeIcon size="xs" color="gray" variant="light" radius="xl">
                  <IconCircle size={12} />
                </ThemeIcon>
              )}
              <Text
                size="xs"
                c={isCurrent ? "blue" : status === "completed" ? "green" : "dimmed"}
                fw={isCurrent ? 500 : 400}
              >
                {labels[step]}
              </Text>
            </Group>
          );
        })}
      </Group>
    </Stack>
  );
}

interface IntroStepProps {
  onStart: () => void;
}

function IntroStep({ onStart }: IntroStepProps) {
  return (
    <Stack gap="lg">
      <Stack gap="xs" ta="center">
        <Text size="lg" fw={500}>
          Nastavíme přesměrování hovorů
        </Text>
        <Text size="sm" c="dimmed">
          Provedeme tě 4 krátkými kroky. Na mobilu stačí klikat na tlačítka.
        </Text>
      </Stack>

      <Paper p="md" radius="md" bg="gray.0">
        <List size="sm" spacing="xs">
          <List.Item>Vymažeme stávající přesměrování</List.Item>
          <List.Item>Nastavíme přesměrování, když nezvedáš</List.Item>
          <List.Item>Nastavíme přesměrování, když máš obsazeno</List.Item>
          <List.Item>Nastavíme přesměrování, když jsi nedostupný</List.Item>
        </List>
      </Paper>

      <Text size="xs" c="dimmed" ta="center">
        Každý krok můžeš přeskočit, pokud nechceš daný typ přesměrování nastavit.
      </Text>

      <Button size="lg" fullWidth rightSection={<IconArrowRight size={18} />} onClick={onStart}>
        Začít
      </Button>
    </Stack>
  );
}

interface DialStepProps {
  title: string;
  description: string;
  dialCode: string;
  karenNumber?: string;
  onConfirm: () => void;
  onRetry: () => void;
  onSkip: () => void;
  showTimingControl?: boolean;
  noAnswerTime?: NoAnswerTime;
  onTimeChange?: (time: NoAnswerTime) => void;
}

function DialStep({
  title,
  description,
  dialCode,
  karenNumber,
  onConfirm,
  onRetry,
  onSkip,
  showTimingControl,
  noAnswerTime,
  onTimeChange,
}: DialStepProps) {
  const [dialed, setDialed] = useState(false);
  const [showTiming, setShowTiming] = useState(false);

  const handleDial = () => {
    window.location.href = `tel:${dialCode}`;
    setDialed(true);
  };

  const handleRetry = () => {
    setDialed(false);
    onRetry();
  };

  return (
    <Stack gap="lg">
      <Stack gap="xs" ta="center">
        <Text size="lg" fw={500}>
          {title}
        </Text>
        <Text size="sm" c="dimmed">
          {description}
        </Text>
      </Stack>

      {karenNumber && (
        <Paper p="md" radius="md" bg="teal.0" ta="center">
          <Text size="xs" c="teal.7" mb={4}>
            Karen číslo
          </Text>
          <Text size="lg" fw={700} c="teal.8">
            {karenNumber}
          </Text>
        </Paper>
      )}

      <Paper p="md" radius="md" withBorder ta="center">
        <Text size="xs" c="dimmed" mb={4}>
          Kód k vytočení
        </Text>
        <Group justify="center" gap="xs">
          <Text size="lg" ff="monospace" fw={500}>
            {dialCode}
          </Text>
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
      </Paper>

      {showTimingControl && (
        <Box>
          {!showTiming ? (
            <UnstyledButton onClick={() => setShowTiming(true)}>
              <Text size="xs" c="blue" td="underline">
                Změnit časování ({noAnswerTime}s)
              </Text>
            </UnstyledButton>
          ) : (
            <Box>
              <Text size="xs" fw={500} mb={4}>
                Po kolika sekundách přesměrovat?
              </Text>
              <SegmentedControl
                size="xs"
                fullWidth
                value={String(noAnswerTime)}
                onChange={(value) => onTimeChange?.(Number(value) as NoAnswerTime)}
                data={NO_ANSWER_TIME_OPTIONS.map((t) => ({
                  value: String(t),
                  label: `${t}s`,
                }))}
              />
            </Box>
          )}
        </Box>
      )}

      {!dialed ? (
        <Button size="lg" fullWidth leftSection={<IconPhone size={18} />} onClick={handleDial}>
          Vytočit kód
        </Button>
      ) : (
        <Stack gap="md">
          <Text size="sm" ta="center" c="dimmed">
            Viděl/a jsi potvrzení od operátora?
          </Text>
          <Group grow>
            <Button variant="filled" color="green" leftSection={<IconCheck size={16} />} onClick={onConfirm}>
              Ano, aktivováno
            </Button>
          </Group>
          <Group grow>
            <Button variant="light" leftSection={<IconRefresh size={16} />} onClick={handleRetry}>
              Zkusit znovu
            </Button>
            <Button variant="subtle" leftSection={<IconPlayerSkipForward size={16} />} onClick={onSkip}>
              Přeskočit
            </Button>
          </Group>
        </Stack>
      )}

      {!dialed && (
        <Button variant="subtle" size="sm" onClick={onSkip}>
          Přeskočit tento krok
        </Button>
      )}
    </Stack>
  );
}

interface CompleteStepProps {
  stepStatuses: StepStatuses;
  onFinish: () => void;
}

function CompleteStep({ stepStatuses, onFinish }: CompleteStepProps) {
  const activatedSteps = REDIRECT_STEPS.filter((step) => stepStatuses[step] === "completed");
  const skippedSteps = REDIRECT_STEPS.filter((step) => stepStatuses[step] === "skipped");

  const labels: Record<keyof StepStatuses, string> = {
    clear: "Vymazání stávajících přesměrování",
    noAnswer: "Přesměrování když nezvedáš",
    busy: "Přesměrování při obsazení",
    unreachable: "Přesměrování při nedostupnosti",
  };

  return (
    <Stack gap="lg">
      <Stack gap="xs" ta="center">
        <ThemeIcon size={60} radius="xl" variant="light" color="green">
          <IconCircleCheck size={30} />
        </ThemeIcon>
        <Text size="lg" fw={500}>
          Nastavení dokončeno!
        </Text>
      </Stack>

      {activatedSteps.length > 0 && (
        <Paper p="md" radius="md" bg="green.0">
          <Text size="sm" fw={500} c="green.8" mb="xs">
            Aktivováno ({activatedSteps.length})
          </Text>
          <List size="sm" spacing="xs">
            {activatedSteps.map((step) => (
              <List.Item key={step} c="green.7">
                {labels[step]}
              </List.Item>
            ))}
          </List>
        </Paper>
      )}

      {skippedSteps.length > 0 && (
        <Paper p="md" radius="md" bg="gray.0">
          <Text size="sm" fw={500} c="dimmed" mb="xs">
            Přeskočeno ({skippedSteps.length})
          </Text>
          <List size="sm" spacing="xs">
            {skippedSteps.map((step) => (
              <List.Item key={step} c="dimmed">
                {labels[step]}
              </List.Item>
            ))}
          </List>
        </Paper>
      )}

      <Text size="sm" c="dimmed" ta="center">
        Nastavení můžeš kdykoliv změnit v sekci Nastavení.
      </Text>

      <Button size="lg" fullWidth rightSection={<IconArrowRight size={18} />} onClick={onFinish}>
        Pokračovat
      </Button>
    </Stack>
  );
}

export function RedirectWizard({ karenNumber, onComplete }: RedirectWizardProps) {
  const [currentStep, setCurrentStep] = useState<WizardStep>("intro");
  const [stepStatuses, setStepStatuses] = useState<StepStatuses>({
    clear: "pending",
    noAnswer: "pending",
    busy: "pending",
    unreachable: "pending",
  });
  const [noAnswerTime, setNoAnswerTime] = useState<NoAnswerTime>(DEFAULT_NO_ANSWER_TIME);

  const updateStepStatus = (step: keyof StepStatuses, status: StepStatus) => {
    setStepStatuses((prev) => ({ ...prev, [step]: status }));
  };

  const goToNextStep = () => {
    const currentIndex = WIZARD_STEPS.indexOf(currentStep);
    if (currentIndex < WIZARD_STEPS.length - 1) {
      setCurrentStep(WIZARD_STEPS[currentIndex + 1]);
    }
  };

  const handleStepConfirm = (step: keyof StepStatuses) => {
    updateStepStatus(step, "completed");
    goToNextStep();
  };

  const handleStepSkip = (step: keyof StepStatuses) => {
    updateStepStatus(step, "skipped");
    goToNextStep();
  };

  const cleanKarenNumber = karenNumber.replace(/\s/g, "");

  return (
    <Stack gap="xl">
      {currentStep !== "intro" && currentStep !== "complete" && (
        <WizardProgress currentStep={currentStep} stepStatuses={stepStatuses} />
      )}

      {currentStep === "intro" && <IntroStep onStart={goToNextStep} />}

      {currentStep === "clear" && (
        <DialStep
          title="Krok 1: Vymazat stávající přesměrování"
          description="Nejdřív vymažeme případná existující přesměrování, aby nedošlo ke konfliktu."
          dialCode={CLEAR_ALL_REDIRECTS_CODE}
          onConfirm={() => handleStepConfirm("clear")}
          onRetry={() => {}}
          onSkip={() => handleStepSkip("clear")}
        />
      )}

      {currentStep === "noAnswer" && (
        <DialStep
          title="Krok 2: Když nezvedáš"
          description={`Když nezvedneš do ${noAnswerTime} sekund, hovor se přesměruje na Karen.`}
          dialCode={getDialCode("noAnswer", cleanKarenNumber, noAnswerTime)}
          karenNumber={karenNumber}
          onConfirm={() => handleStepConfirm("noAnswer")}
          onRetry={() => {}}
          onSkip={() => handleStepSkip("noAnswer")}
          showTimingControl
          noAnswerTime={noAnswerTime}
          onTimeChange={setNoAnswerTime}
        />
      )}

      {currentStep === "busy" && (
        <DialStep
          title="Krok 3: Když máš obsazeno"
          description="Když máš obsazeno nebo odmítneš hovor, přesměruje se na Karen."
          dialCode={getDialCode("busy", cleanKarenNumber)}
          karenNumber={karenNumber}
          onConfirm={() => handleStepConfirm("busy")}
          onRetry={() => {}}
          onSkip={() => handleStepSkip("busy")}
        />
      )}

      {currentStep === "unreachable" && (
        <DialStep
          title="Krok 4: Když jsi nedostupný"
          description="Když nemáš signál nebo máš vypnutý telefon, hovor jde na Karen."
          dialCode={getDialCode("unreachable", cleanKarenNumber)}
          karenNumber={karenNumber}
          onConfirm={() => handleStepConfirm("unreachable")}
          onRetry={() => {}}
          onSkip={() => handleStepSkip("unreachable")}
        />
      )}

      {currentStep === "complete" && <CompleteStep stepStatuses={stepStatuses} onFinish={onComplete} />}
    </Stack>
  );
}
