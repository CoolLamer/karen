export type RedirectType = "noAnswer" | "busy" | "unreachable";

/** Code to clear all existing call forwards */
export const CLEAR_ALL_REDIRECTS_CODE = "##002#";

export interface RedirectCode {
  code: string;
  deactivateCode: string;
  label: string;
  description: string;
}

/** Available timeout options for "no answer" forwarding (in seconds) */
export const NO_ANSWER_TIME_OPTIONS = [5, 10, 15, 20, 25, 30] as const;
export type NoAnswerTime = (typeof NO_ANSWER_TIME_OPTIONS)[number];
export const DEFAULT_NO_ANSWER_TIME: NoAnswerTime = 10;

export const REDIRECT_CODES: Record<RedirectType, RedirectCode> = {
  noAnswer: {
    code: "**61*{number}**{time}#",
    deactivateCode: "##61#",
    label: "Když nezvednu",
    description: "Když nezvedneš do {time} sekund, hovor se přesměruje na Karen",
  },
  busy: {
    code: "**67*{number}#",
    deactivateCode: "##67#",
    label: "Když mám obsazeno",
    description: "Když máš obsazeno nebo odmítneš hovor, přesměruje se na Karen",
  },
  unreachable: {
    code: "**62*{number}#",
    deactivateCode: "##62#",
    label: "Když jsem nedostupný",
    description: "Když nemáš signál nebo máš vypnutý telefon, hovor jde na Karen",
  },
};

export const REDIRECT_ORDER: RedirectType[] = ["noAnswer", "busy", "unreachable"];

export const PHONE_SETTINGS_INSTRUCTIONS = {
  iphone: {
    title: "iPhone",
    steps: [
      "Otevři Nastavení",
      "Klikni na Telefon",
      "Vyber Přesměrování hovorů",
      "Zapni přesměrování a zadej číslo Karen",
    ],
  },
  android: {
    title: "Android",
    steps: [
      "Otevři aplikaci Telefon",
      "Klikni na ⋮ (tři tečky) → Nastavení",
      "Vyber Přesměrování hovorů",
      "Nastav jednotlivé typy přesměrování na číslo Karen",
    ],
  },
};

export function getDialCode(type: RedirectType, phoneNumber: string, time?: NoAnswerTime): string {
  let code = REDIRECT_CODES[type].code.replace("{number}", phoneNumber.replace(/\s/g, ""));
  if (type === "noAnswer") {
    code = code.replace("{time}", String(time ?? DEFAULT_NO_ANSWER_TIME));
  }
  return code;
}

export function getDescription(type: RedirectType, time?: NoAnswerTime): string {
  let description = REDIRECT_CODES[type].description;
  if (type === "noAnswer") {
    description = description.replace("{time}", String(time ?? DEFAULT_NO_ANSWER_TIME));
  }
  return description;
}

export function getDeactivationCode(type: RedirectType): string {
  return REDIRECT_CODES[type].deactivateCode;
}
