import { SegmentContent, SegmentKey } from "./index";

export const SEGMENTS: Record<SegmentKey, SegmentContent> = {
  technicians: {
    key: "technicians",
    urlPath: "/pro-techniky",
    selectorLabel: "Technik / Řemeslník",
    selectorIcon: "IconTool",
    hero: {
      title: "AI asistentka pro techniky v terénu",
      tagline:
        "Jsi na zakázce a někdo volá. Zvednu zvedne, zjistí o co jde, a ty víš, jestli je to nová práce nebo spam.",
      ctaText: "Vyzkoušet zdarma",
    },
    painPoints: {
      title: "Znáte to?",
      items: [
        {
          icon: "IconPhoneOff",
          title: "Zmeškané hovory = ztracené zakázky",
          description: "Každý nezvednutý hovor může být nová zakázka za desítky tisíc.",
        },
        {
          icon: "IconHandStop",
          title: "Nejde zvednout během práce",
          description: "Špinavé ruce, hlučné prostředí, jste u klienta - prostě to nejde.",
        },
        {
          icon: "IconMailOff",
          title: "Spam ruší práci",
          description: "Polovina hovorů jsou marketingové nabídky energie a pojištění.",
        },
        {
          icon: "IconClipboardList",
          title: "Dotazy na stav zakázky",
          description:
            "Zákazník volá ohledně stavu objednávky - vy to vidíte a zavoláte až budete u počítače.",
        },
      ],
    },
    exampleCall: {
      scenario: "Technik Belix je na sanaci, volá zákazník ohledně stavu zakázky",
      dialogue: [
        {
          speaker: "karen",
          text: "Dobrý den, tady Karen, asistentka pana Petra. Petr je momentálně na zakázce. Jak vám mohu pomoct?",
        },
        {
          speaker: "caller",
          text: "Dobrý den, chtěla jsem se zeptat, jak to vypadá s naší zakázkou - měli jste přijít vysušit sklep.",
        },
        {
          speaker: "karen",
          text: "Rozumím, dotaz na stav zakázky ohledně vysušení sklepa. Petr se vám ozve jakmile bude mít přístup k systému. Mohu potvrdit vaše číslo?",
        },
      ],
      result: {
        label: "Dotaz na zakázku",
        color: "blue",
        summary: "Stav zakázky vysušení sklepa - zavolat zpět s info",
      },
    },
    featurePriority: ["context", "spam_filter", "never_miss", "transcript"],
  },

  professionals: {
    key: "professionals",
    urlPath: "/pro-lekare",
    selectorLabel: "Lékař / Právník",
    selectorIcon: "IconStethoscope",
    hero: {
      title: "AI asistentka pro profesionály",
      tagline:
        "Jste v ordinaci a pacient volá. Zvednu zjistí, jestli je to urgentní, a důležité hovory vám přepojí.",
      ctaText: "Vyzkoušet zdarma",
    },
    painPoints: {
      title: "Znáte to?",
      items: [
        {
          icon: "IconAlertTriangle",
          title: "Urgentní vs. běžné",
          description: "Potřebujete vědět, co je akutní a co může počkat na další den.",
        },
        {
          icon: "IconCurrencyDollar",
          title: "Sekretářka je drahá",
          description: "Plný úvazek sekretářky stojí 25-40 tisíc měsíčně.",
        },
        {
          icon: "IconPhoneIncoming",
          title: "Hovory během ordinace",
          description: "Pacienti a klienti volají, když máte jiné pacienty.",
        },
      ],
    },
    exampleCall: {
      scenario: "Dr. Martina má ordinaci, volá pacient s akutním problémem",
      dialogue: [
        {
          speaker: "karen",
          text: "Dobrý den, ordinace doktorky Martiny. Doktorka má právě pacienta. Jak vám mohu pomoct?",
        },
        {
          speaker: "caller",
          text: "Mám silné bolesti zubů už od včera, nemůžu spát.",
        },
        {
          speaker: "karen",
          text: "Rozumím, silné bolesti zubů. Dovolte, abych vás přepojila na doktorku - jde o akutní případ.",
        },
      ],
      result: {
        label: "Urgentní - přepojeno",
        color: "red",
        summary: "Akutní bolest - pacient přepojen",
      },
    },
    featurePriority: ["forward", "context", "professional", "transcript"],
  },

  sales: {
    key: "sales",
    urlPath: "/pro-maklere",
    selectorLabel: "Obchodník / Makléř",
    selectorIcon: "IconBuildingSkyscraper",
    hero: {
      title: "AI asistentka pro obchodníky",
      tagline:
        "Každý zmeškaný hovor je potenciální ztracený deal. Zvednu zachytí každý lead a vy zavoláte zpět s plným kontextem.",
      ctaText: "Vyzkoušet zdarma",
    },
    painPoints: {
      title: "Znáte to?",
      items: [
        {
          icon: "IconUserOff",
          title: "Leady utíkají ke konkurenci",
          description: "Když nezvedneš do hodiny, klient už má jiného makléře.",
        },
        {
          icon: "IconHome",
          title: "Na prohlídce nelze zvedat",
          description: "Během prohlídek přicházíte o nové potenciální klienty.",
        },
        {
          icon: "IconQuestionMark",
          title: "Chybí kontext před voláním zpět",
          description: "Nevíte, jak moc je lead horký, než zavoláte.",
        },
      ],
    },
    exampleCall: {
      scenario: "Tomáš (realitní makléř) je na prohlídce, volá nový zájemce o byt",
      dialogue: [
        {
          speaker: "karen",
          text: "Dobrý den, kancelář pana Tomáše. Tomáš je momentálně na prohlídce. S čím vám mohu pomoct?",
        },
        {
          speaker: "caller",
          text: "Hledám 2+kk v Praze, do 6 milionů, nejlépe se zahradou.",
        },
        {
          speaker: "karen",
          text: "Skvělé, 2+kk v Praze do 6 milionů se zahradou. Mohu vás požádat o číslo a Tomáš se vám ozve dnes večer?",
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
    selectorLabel: "Manažer / Vedoucí",
    selectorIcon: "IconBriefcase",
    hero: {
      title: "AI asistentka pro manažery",
      tagline:
        "Delegujte filtrování hovorů. Vidíte kontext před tím, než zavoláte zpět.",
      ctaText: "Vyzkoušet zdarma",
    },
    painPoints: {
      title: "Znáte to?",
      items: [
        {
          icon: "IconPhone",
          title: "8-10 hovorů denně",
          description: "Většina není urgentní, ale nevíte to dopředu.",
        },
        {
          icon: "IconUsers",
          title: "Sdílená nebo žádná asistentka",
          description: "Nemůžete se spolehnout, že někdo zvedne.",
        },
        {
          icon: "IconArrowBack",
          title: "Volání zpět bez kontextu",
          description: "Nevíte, o čem bude hovor, dokud nezavoláte.",
        },
      ],
    },
    exampleCall: {
      scenario: "Jan (vedoucí oddělení) má meeting, volá kolega z jiné pobočky",
      dialogue: [
        {
          speaker: "karen",
          text: "Dobrý den, kancelář pana Jana. Jan je momentálně na jednání. Jak vám mohu pomoct?",
        },
        {
          speaker: "caller",
          text: "Potřebuju s ním probrat rozpočet na Q2, není to urgentní ale chtěl bych to stihnout tento týden.",
        },
        {
          speaker: "karen",
          text: "Rozumím, rozpočet na Q2, není urgentní, ideálně tento týden. Předám Janovi a ozve se vám.",
        },
      ],
      result: {
        label: "Interní",
        color: "blue",
        summary: "Rozpočet Q2 - zavolat tento týden",
      },
    },
    featurePriority: ["context", "spam_filter", "rules", "transcript"],
  },
};

export const SEGMENT_LIST = Object.values(SEGMENTS);
