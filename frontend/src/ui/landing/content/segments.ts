import { SegmentContent, SegmentKey } from "./index";

export const SEGMENTS: Record<SegmentKey, SegmentContent> = {
  technicians: {
    key: "technicians",
    urlPath: "/pro-techniky",
    selectorLabel: "Technik / Remeslnik",
    selectorIcon: "IconTool",
    hero: {
      title: "AI asistentka pro techniky v terenu",
      tagline:
        "Jsi na zakazce a nekdo vola. Zvednu zvedne, zjisti o co jde, a ty vis, jestli je to nova prace nebo spam.",
      ctaText: "Vyzkouset zdarma",
    },
    painPoints: {
      title: "Znate to?",
      items: [
        {
          icon: "IconPhoneOff",
          title: "Zmeskane hovory = ztracene zakazky",
          description: "Kazdy nezvednuty hovor muze byt nova zakazka za desitky tisic.",
        },
        {
          icon: "IconHandStop",
          title: "Nejde zvednout behem prace",
          description: "Spinave ruce, hlucne prostredi, jste u klienta - proste to nejde.",
        },
        {
          icon: "IconMailOff",
          title: "Spam rusi praci",
          description: "Polovina hovoru jsou marketingove nabidky energie a pojisteni.",
        },
        {
          icon: "IconClipboardList",
          title: "Dotazy na stav zakazky",
          description:
            "Zakaznik vola ohledne stavu objednavky - vy to vidite a zavolate az budete u pocitace.",
        },
      ],
    },
    exampleCall: {
      scenario: "Technik Belix je na sanaci, vola zakaznik ohledne stavu zakazky",
      dialogue: [
        {
          speaker: "karen",
          text: "Dobry den, tady Karen, asistentka pana Petra. Petr je momentalne na zakazce. Jak vam mohu pomoct?",
        },
        {
          speaker: "caller",
          text: "Dobry den, chtela jsem se zeptat, jak to vypada s nasi zakazkou - meli jste prijit vysuset sklep.",
        },
        {
          speaker: "karen",
          text: "Rozumim, dotaz na stav zakazky ohledne vysuseni sklepa. Petr se vam ozve jakmile bude mit pristup k systemu. Mohu potvrdit vase cislo?",
        },
      ],
      result: {
        label: "Dotaz na zakazku",
        color: "blue",
        summary: "Stav zakazky vysuseni sklepa - zavolat zpet s info",
      },
    },
    featurePriority: ["context", "spam_filter", "never_miss", "transcript"],
  },

  professionals: {
    key: "professionals",
    urlPath: "/pro-lekare",
    selectorLabel: "Lekar / Pravnik",
    selectorIcon: "IconStethoscope",
    hero: {
      title: "AI asistentka pro profesionaly",
      tagline:
        "Jste v ordinaci a pacient vola. Zvednu zjisti, jestli je to urgentni, a dulezite hovory vam prepoji.",
      ctaText: "Vyzkouset zdarma",
    },
    painPoints: {
      title: "Znate to?",
      items: [
        {
          icon: "IconAlertTriangle",
          title: "Urgentni vs. bezne",
          description: "Potrebujete vedet, co je akutni a co muze pockat na dalsi den.",
        },
        {
          icon: "IconCurrencyDollar",
          title: "Sekretarka je draha",
          description: "Plny uvazek sekretarky stoji 25-40 tisic mesicne.",
        },
        {
          icon: "IconPhoneIncoming",
          title: "Hovory behem ordinace",
          description: "Pacienti a klienti volaji, kdyz mate jine pacienty.",
        },
      ],
    },
    exampleCall: {
      scenario: "Dr. Martina ma ordinaci, vola pacient s akutnim problemem",
      dialogue: [
        {
          speaker: "karen",
          text: "Dobry den, ordinace doktorky Martiny. Doktorka ma prave pacienta. Jak vam mohu pomoct?",
        },
        {
          speaker: "caller",
          text: "Mam silne bolesti zubu uz od vcera, nemuzhu spat.",
        },
        {
          speaker: "karen",
          text: "Rozumim, silne bolesti zubu. Dovolte, abych vas prepojila na doktorku - jde o akutni pripad.",
        },
      ],
      result: {
        label: "Urgentni - prepojeno",
        color: "red",
        summary: "Akutni bolest - pacient prepojen",
      },
    },
    featurePriority: ["forward", "context", "professional", "transcript"],
  },

  sales: {
    key: "sales",
    urlPath: "/pro-maklere",
    selectorLabel: "Obchodnik / Makler",
    selectorIcon: "IconBuildingSkyscraper",
    hero: {
      title: "AI asistentka pro obchodniky",
      tagline:
        "Kazdy zmeskany hovor je potencialni ztraceny deal. Zvednu zachyti kazdy lead a vy zavolate zpet s plnym kontextem.",
      ctaText: "Vyzkouset zdarma",
    },
    painPoints: {
      title: "Znate to?",
      items: [
        {
          icon: "IconUserOff",
          title: "Leady utikaji ke konkurenci",
          description: "Kdyz nezvednes do hodiny, klient uz ma jineho maklere.",
        },
        {
          icon: "IconHome",
          title: "Na prohlidce nelze zvedat",
          description: "Behem prohlidek prichazite o nove potencialni klienty.",
        },
        {
          icon: "IconQuestionMark",
          title: "Chybi kontext pred volanim zpet",
          description: "Nevite, jak moc je lead horky, nez zavolate.",
        },
      ],
    },
    exampleCall: {
      scenario: "Tomas (realitni makler) je na prohlidce, vola novy zajemce o byt",
      dialogue: [
        {
          speaker: "karen",
          text: "Dobry den, kancelar pana Tomase. Tomas je momentalne na prohlidce. S cim vam mohu pomoct?",
        },
        {
          speaker: "caller",
          text: "Hledam 2+kk v Praze, do 6 milionu, nejlepe se zahradou.",
        },
        {
          speaker: "karen",
          text: "Skvele, 2+kk v Praze do 6 milionu se zahradou. Mohu vas pozadat o cislo a Tomas se vam ozve dnes vecer?",
        },
      ],
      result: {
        label: "Hot lead",
        color: "orange",
        summary: "2+kk Praha, 6M, se zahradou - zavolat dnes!",
      },
    },
    featurePriority: ["never_miss", "context", "transcript", "spam_filter"],
  },

  managers: {
    key: "managers",
    urlPath: "/pro-manazery",
    selectorLabel: "Manazer / Vedouci",
    selectorIcon: "IconBriefcase",
    hero: {
      title: "AI asistentka pro manazery",
      tagline:
        "Delegujte filtrovani hovoru. Vidite kontext pred tim, nez zavolate zpet.",
      ctaText: "Vyzkouset zdarma",
    },
    painPoints: {
      title: "Znate to?",
      items: [
        {
          icon: "IconPhone",
          title: "8-10 hovoru denne",
          description: "Vetsina neni urgentni, ale nevite to dopredu.",
        },
        {
          icon: "IconUsers",
          title: "Sdilena nebo zadna asistentka",
          description: "Nemuzete se spolehnout, ze nekdo zvedne.",
        },
        {
          icon: "IconArrowBack",
          title: "Volani zpet bez kontextu",
          description: "Nevite, o cem bude hovor, dokud nezavolate.",
        },
      ],
    },
    exampleCall: {
      scenario: "Jan (vedouci oddeleni) ma meeting, vola kolega z jine pobocky",
      dialogue: [
        {
          speaker: "karen",
          text: "Dobry den, kancelar pana Jana. Jan je momentalne na jednani. Jak vam mohu pomoct?",
        },
        {
          speaker: "caller",
          text: "Potrebuju s nim probrat rozpocet na Q2, neni to urgentni ale chtel bych to stihnout tento tyden.",
        },
        {
          speaker: "karen",
          text: "Rozumim, rozpocet na Q2, neni urgentni, idealne tento tyden. Predam Janovi a ozve se vam.",
        },
      ],
      result: {
        label: "Interni",
        color: "blue",
        summary: "Rozpocet Q2 - zavolat tento tyden",
      },
    },
    featurePriority: ["context", "spam_filter", "rules", "transcript"],
  },
};

export const SEGMENT_LIST = Object.values(SEGMENTS);
