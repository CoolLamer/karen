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
  faq: [
    {
      question: "Budou mi lidi volat na jiné číslo?",
      answer:
        "Ne, lidé vám volají stále na vaše původní telefonní číslo. Karen se zapojí pouze když hovor nezvednete nebo odmítnete – pak se hovor automaticky přesměruje na asistentku. Volající nepozná rozdíl, prostě někdo zvedne telefon.",
    },
    {
      question: "Jak nastavím přesměrování hovorů?",
      answer:
        'Přesměrování se nastavuje vytočením speciálního kódu na telefonu. Otevřete aplikaci Zvednu na mobilu, přejděte do nastavení a klikněte na tlačítko "Aktivovat přesměrování" – automaticky se vytočí aktivační kód a na obrazovce uvidíte potvrzení od operátora. Na počítači tlačítko nefunguje, musíte to udělat na telefonu.',
    },
    {
      question: "Můžu si změnit jméno asistentky?",
      answer:
        'Ano! V nastavení si můžete upravit "Představení" – tam si nastavíte jak se asistentka představí volajícímu. Například: "Dobrý den, tady Alfréd, pan Novák právě nemá čas, můžete mi říct co potřebujete."',
    },
    {
      question: "Funguje to s mým operátorem?",
      answer:
        "Ano, Zvednu funguje se všemi českými operátory (O2, T-Mobile, Vodafone) i virtuálními operátory. Přesměrování hovorů je standardní funkce GSM sítě.",
    },
    {
      question: "Co když nechci přijímat spam?",
      answer:
        "Karen automaticky rozpozná marketingové hovory a spam. V přehledu hovorů uvidíte jen ty důležité – spam je označený a můžete ho ignorovat.",
    },
  ],
  cta: {
    title: "Začněte používat Zvednu ještě dnes",
    subtitle: "Registrace je zdarma. Za 2 minuty budete mít svou asistentku.",
    buttonText: "Vyzkoušet zdarma",
  },
  footer: {
    tagline: "Zvednu - vaše AI asistentka na telefonu",
  },
};
