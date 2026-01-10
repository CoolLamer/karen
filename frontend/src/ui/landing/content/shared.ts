import { SharedContent } from "./index";

export const SHARED_CONTENT: SharedContent = {
  brand: {
    name: "Zvednu",
    assistantName: "Karen",
    tagline: "Zvednu to za vás",
  },
  howItWorks: [
    {
      step: 1,
      title: "Přesměrujte hovory",
      description: "Nastavte přesměrování na své Zvednu číslo, když nezvedáte.",
    },
    {
      step: 2,
      title: "Karen zvedne",
      description: "Karen přijme hovor a zeptá se volajícího, o co jde.",
    },
    {
      step: 3,
      title: "Uvidíte přehled",
      description: "V aplikaci uvidíte, kdo volal, proč a jestli má smysl zavolat zpět.",
    },
  ],
  features: {
    spam_filter: {
      icon: "IconShieldCheck",
      title: "Filtruje spam",
      description:
        "Karen rozpozná marketingové hovory a spam. Vy uvidíte jen ty, které jsou pro vás důležité.",
    },
    context: {
      icon: "IconMessage",
      title: "Zjistí kontext",
      description: "Karen se zeptá, proč volají, a shrne to pro vás. Víte, o čem hovor bude.",
    },
    professional: {
      icon: "IconRobot",
      title: "Přirozený hlas",
      description:
        "Karen mluví česky s přirozeně znějícím hlasem. Volající nemají pocit, že mluví s robotem.",
    },
    never_miss: {
      icon: "IconPhone",
      title: "Nic důležitého nezmeškáte",
      description: "Důležité hovory vám Karen přepošle nebo zanechá podrobnou zprávu.",
    },
    transcript: {
      icon: "IconListCheck",
      title: "Přepis hovoru",
      description: "Každý hovor má kompletní přepis. Přečtete si ho místo poslechu.",
    },
    rules: {
      icon: "IconSettings",
      title: "Vaše pravidla",
      description: "Nastavte si VIP kontakty, způsoby oslovení a vlastní instrukce.",
    },
    forward: {
      icon: "IconPhoneCall",
      title: "VIP přepojení",
      description: "Urgentní hovory vám Karen přepojí okamžitě na váš telefon.",
    },
  },
  cta: {
    title: "Začněte používat Zvednu ještě dnes",
    subtitle: "Registrace je zdarma. Za 2 minuty budete mít svou asistentku.",
    buttonText: "Vyzkoušet zdarma",
  },
  footer: {
    tagline: "Zvednu - vaše AI asistentka na telefonu",
  },
};
