import { SharedContent } from "./index";

export const SHARED_CONTENT: SharedContent = {
  brand: {
    name: "Zvednu",
    assistantName: "Karen",
    tagline: "Zvednu to za tebe",
  },
  howItWorks: [
    {
      step: 1,
      title: "Přesměruj hovory",
      description: "Nastav přesměrování na své Zvednu číslo, když nezvedáš.",
    },
    {
      step: 2,
      title: "Karen zvedne",
      description: "Karen přijme hovor a zeptá se volajícího, o co jde.",
    },
    {
      step: 3,
      title: "Vidíš přehled",
      description: "V aplikaci vidíš kdo volal, proč a zda stojí za to zavolat zpět.",
    },
  ],
  features: {
    spam_filter: {
      icon: "IconShieldCheck",
      title: "Filtruje spam",
      description: "Karen rozpozná marketingové hovory a spam. Ty vidíš jen důležité hovory.",
    },
    context: {
      icon: "IconMessage",
      title: "Zjistí kontext",
      description: "Karen se zeptá proč volají a shrne to pro tebe. Víš o čem hovor bude.",
    },
    professional: {
      icon: "IconRobot",
      title: "Přirozený hlas",
      description:
        "Karen mluví česky s přirozeně znějícím hlasem. Volající nemají pocit že mluví s robotem.",
    },
    never_miss: {
      icon: "IconPhone",
      title: "Nikdy nepromeškáš",
      description: "Důležité hovory ti Karen přepošle nebo zanechá podrobnou zprávu.",
    },
    transcript: {
      icon: "IconListCheck",
      title: "Přepis hovoru",
      description: "Každý hovor má kompletní přepis. Můžeš si ho přečíst místo poslechu.",
    },
    rules: {
      icon: "IconSettings",
      title: "Tvoje pravidla",
      description: "Nastav si VIP kontakty, způsoby oslovení a vlastní instrukce.",
    },
    forward: {
      icon: "IconPhoneCall",
      title: "VIP přepojení",
      description: "Urgentní hovory ti Karen přepojí okamžitě na tvůj telefon.",
    },
  },
  cta: {
    title: "Začni používat Zvednu ještě dnes",
    subtitle: "Registrace je zdarma. Za 2 minuty budeš mít svoji asistentku.",
    buttonText: "Vyzkoušet zdarma",
  },
  footer: {
    tagline: "Zvednu - tvoje AI asistentka na telefonu",
  },
};
