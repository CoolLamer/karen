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
      title: "Presmeruj hovory",
      description: "Nastav presmerovani na sve Zvednu cislo, kdyz nezvedas.",
    },
    {
      step: 2,
      title: "Karen zvedne",
      description: "Karen prijme hovor a zepta se volajiciho, o co jde.",
    },
    {
      step: 3,
      title: "Vidis prehled",
      description: "V aplikaci vidis kdo volal, proc a zda stoji za to zavolat zpet.",
    },
  ],
  features: {
    spam_filter: {
      icon: "IconShieldCheck",
      title: "Filtruje spam",
      description: "Karen rozpozna marketingove hovory a spam. Ty vidis jen dulezite hovory.",
    },
    context: {
      icon: "IconMessage",
      title: "Zjisti kontext",
      description: "Karen se zepta proc volaji a shrne to pro tebe. Vis o cem hovor bude.",
    },
    professional: {
      icon: "IconRobot",
      title: "Prirozeny hlas",
      description:
        "Karen mluvi cesky s prirozene znejicim hlasem. Volajici nemaji pocit ze mluvi s robotem.",
    },
    never_miss: {
      icon: "IconPhone",
      title: "Nikdy nepropasnes",
      description: "Dulezite hovory ti Karen preposle nebo zanecha podrobnou zpravu.",
    },
    transcript: {
      icon: "IconListCheck",
      title: "Prepis hovoru",
      description: "Kazdy hovor ma kompletni prepis. Muzes si ho precist misto poslechu.",
    },
    rules: {
      icon: "IconSettings",
      title: "Tvoje pravidla",
      description: "Nastav si VIP kontakty, zpusoby osloveni a vlastni instrukce.",
    },
    forward: {
      icon: "IconPhoneCall",
      title: "VIP prepojeni",
      description: "Urgentni hovory ti Karen prepoji okamzite na tvuj telefon.",
    },
  },
  cta: {
    title: "Zacni pouzivat Zvednu jeste dnes",
    subtitle: "Registrace je zdarma. Za 2 minuty budes mit svoji asistentku.",
    buttonText: "Vyzkouset zdarma",
  },
  footer: {
    tagline: "Zvednu - tvoje AI asistentka na telefonu",
  },
};
