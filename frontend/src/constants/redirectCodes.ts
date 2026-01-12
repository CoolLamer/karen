export type RedirectType = "noAnswer" | "busy" | "unreachable";

export interface RedirectCode {
  code: string;
  deactivateCode: string;
  label: string;
  description: string;
}

export const REDIRECT_CODES: Record<RedirectType, RedirectCode> = {
  noAnswer: {
    code: "**61*{number}#",
    deactivateCode: "##61#",
    label: "Když nezvednu",
    description: "Přesměrování po 5 zazvoněních - když telefon nezvedneš",
  },
  busy: {
    code: "**67*{number}#",
    deactivateCode: "##67#",
    label: "Když mám obsazeno",
    description: "Přesměrování když probíhá jiný hovor nebo odmítneš volání",
  },
  unreachable: {
    code: "**62*{number}#",
    deactivateCode: "##62#",
    label: "Když jsem nedostupný",
    description: "Přesměrování když nemáš signál nebo je telefon vypnutý",
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

export function getDialCode(type: RedirectType, phoneNumber: string): string {
  return REDIRECT_CODES[type].code.replace("{number}", phoneNumber.replace(/\s/g, ""));
}

export function getDeactivationCode(type: RedirectType): string {
  return REDIRECT_CODES[type].deactivateCode;
}
