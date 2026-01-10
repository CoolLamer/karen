export type SegmentKey = "technicians" | "professionals" | "sales" | "managers";

export interface DialogueLine {
  speaker: "karen" | "caller";
  text: string;
}

export interface ExampleCallResult {
  label: string;
  color: string;
  summary: string;
}

export interface PainPointItem {
  icon: string;
  title: string;
  description: string;
}

export interface SegmentContent {
  key: SegmentKey;
  urlPath: string;
  selectorLabel: string;
  selectorIcon: string;
  hero: {
    title: string;
    tagline: string;
    ctaText: string;
  };
  painPoints: {
    title: string;
    items: PainPointItem[];
  };
  exampleCall: {
    scenario: string;
    dialogue: DialogueLine[];
    result: ExampleCallResult;
  };
  featurePriority: string[];
}

export interface HowItWorksStep {
  step: number;
  title: string;
  description: string;
}

export interface FeatureDefinition {
  icon: string;
  title: string;
  description: string;
}

export interface SharedContent {
  brand: {
    name: string;
    assistantName: string;
    tagline: string;
  };
  howItWorks: HowItWorksStep[];
  features: Record<string, FeatureDefinition>;
  cta: {
    title: string;
    subtitle: string;
    buttonText: string;
  };
  footer: {
    tagline: string;
  };
}
