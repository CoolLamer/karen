import React from "react";
import {
  IconCheck,
  IconX,
  IconQuestionMark,
  IconMail,
  IconFlame,
  IconAlertTriangle,
  IconClockHour4,
  IconInfoCircle,
} from "@tabler/icons-react";

export type LabelConfig = {
  color: string;
  label: string;
  icon: React.ReactNode;
};

export function getLegitimacyConfig(
  label: string | undefined,
  iconSize = 14
): LabelConfig {
  switch (label) {
    case "legitimate":
    case "legitimní":
      return { color: "green", label: "Legitimní", icon: <IconCheck size={iconSize} /> };
    case "marketing":
      return { color: "yellow", label: "Marketing", icon: <IconMail size={iconSize} /> };
    case "spam":
      return { color: "red", label: "Spam", icon: <IconX size={iconSize} /> };
    default:
      return { color: "gray", label: "Neznámé", icon: <IconQuestionMark size={iconSize} /> };
  }
}

export function formatCallStatus(status: string, rejectionReason?: string | null) {
  switch (status) {
    case "in_progress":
      return "Probíhá";
    case "completed":
      return "Dokončeno";
    case "queued":
      return "Čeká";
    case "ringing":
      return "Vyzvání";
    case "rejected_limit":
      switch (rejectionReason) {
        case "trial_expired":
          return "Trial vypršel";
        case "limit_exceeded":
          return "Limit dosažen";
        case "subscription_cancelled":
          return "Předplatné zrušeno";
        case "subscription_suspended":
          return "Předplatné pozastaveno";
        default:
          return "Nepřijato";
      }
    default:
      return status;
  }
}

export function getRejectionExplanation(rejectionReason?: string | null): string {
  switch (rejectionReason) {
    case "trial_expired":
      return "Asistentka nemohla přijmout hovor - váš trial skončil";
    case "limit_exceeded":
      return "Asistentka nemohla přijmout hovor - dosáhli jste měsíčního limitu";
    case "subscription_cancelled":
      return "Asistentka nemohla přijmout hovor - předplatné bylo zrušeno";
    case "subscription_suspended":
      return "Asistentka nemohla přijmout hovor - předplatné bylo pozastaveno";
    default:
      return "Asistentka nemohla přijmout tento hovor";
  }
}

export function getLeadLabelConfig(
  label: string | undefined,
  iconSize = 14
): LabelConfig {
  switch (label) {
    case "hot_lead":
      return { color: "red", label: "Hot lead", icon: <IconFlame size={iconSize} /> };
    case "urgentni":
      return { color: "orange", label: "Urgentní", icon: <IconAlertTriangle size={iconSize} /> };
    case "follow_up":
      return { color: "blue", label: "Follow-up", icon: <IconClockHour4 size={iconSize} /> };
    case "informacni":
      return { color: "gray", label: "Informační", icon: <IconInfoCircle size={iconSize} /> };
    default:
      return { color: "gray", label: "Nezjištěno", icon: <IconQuestionMark size={iconSize} /> };
  }
}
