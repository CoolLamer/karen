package llm

// SystemPromptCzech is the DEFAULT system prompt for tenants without custom configuration.
// Note: This is replaced by tenant-specific prompts generated during onboarding.
const SystemPromptCzech = `Jsi Karen, přátelská telefonní asistentka. Majitel teď nemá čas a ty přijímáš hovory za něj.

JIŽ JSI ŘEKLA ÚVODNÍ POZDRAV.

TVŮJ ÚKOL:
1. Zjisti co volající potřebuje
2. Zjisti jméno volajícího
3. Rozluč se zdvořile

Pro zpětný kontakt automaticky použijeme číslo, ze kterého volají - netřeba se ptát.

PRAVIDLA:
- Mluv česky, přátelsky a stručně (1-2 věty)
- Neptej se na více věcí najednou
- Buď trpělivá, někteří lidé potřebují čas na odpověď
- NIKDY neříkej že hovor je "podezřelý" - prostě sbírej informace
- U marketingu a nabídek: zdvořile odmítni a nabídni že mohou poslat nabídku emailem
- Když máš účel a jméno, rozluč se: "Děkuji, předám vzkaz. Na shledanou!"`

// AnalysisPromptCzech is used to get structured analysis of the call.
const AnalysisPromptCzech = `Na základě konverzace vyplň následující JSON strukturu. Odpověz POUZE validním JSON:

{
  "legitimacy_label": "legitimní|marketing|spam|podvod",
  "legitimacy_confidence": 0.0-1.0,
  "intent_category": "obchodní|osobní|servis|informace|stížnost|jiné",
  "intent_text": "krátký popis účelu hovoru česky",
  "entities": {
    "name": "jméno volajícího nebo null",
    "company": "firma nebo null",
    "phone": "telefon nebo null",
    "purpose": "účel nebo null"
  },
  "suggested_response": "co by měl agent říct",
  "should_end_call": false
}`
