import { useNavigate } from "react-router-dom";
import { Box } from "@mantine/core";
import {
  Header,
  Hero,
  TrustBadges,
  PainPoints,
  HowItWorks,
  Features,
  PricingSection,
  ComparisonTable,
  ExampleCall,
  CTASection,
  Footer,
} from "./components";
import { SegmentKey } from "./content/index";
import { SEGMENTS } from "./content/segments";
import { SHARED_CONTENT } from "./content/shared";

interface SegmentLandingPageProps {
  segmentKey: SegmentKey;
}

export function SegmentLandingPage({ segmentKey }: SegmentLandingPageProps) {
  const navigate = useNavigate();
  const segment = SEGMENTS[segmentKey];

  const prioritizedFeatures = segment.featurePriority
    .map((key) => SHARED_CONTENT.features[key])
    .filter(Boolean);

  const handleContactClick = () => {
    window.location.href = "mailto:info@zvednu.cz?subject=Zájem o tarif Firma";
  };

  return (
    <Box>
      <Header showBackToMain />
      <Hero
        title={segment.hero.title}
        tagline={segment.hero.tagline}
        ctaText={segment.hero.ctaText}
        onCtaClick={() => navigate("/login")}
      />
      <TrustBadges />
      <PainPoints title={segment.painPoints.title} items={segment.painPoints.items} />
      <HowItWorks steps={SHARED_CONTENT.howItWorks} />
      <Features features={prioritizedFeatures} />
      <PricingSection
        onTrialClick={() => navigate("/login")}
        onBuyClick={() => navigate("/login")}
        onContactClick={handleContactClick}
      />
      <ComparisonTable />
      <ExampleCall
        scenario={segment.exampleCall.scenario}
        dialogue={segment.exampleCall.dialogue}
        result={segment.exampleCall.result}
      />
      <CTASection
        title={SHARED_CONTENT.cta.title}
        subtitle={SHARED_CONTENT.cta.subtitle}
        buttonText={SHARED_CONTENT.cta.buttonText}
        onCtaClick={() => navigate("/login")}
      />
      <Footer tagline={SHARED_CONTENT.footer.tagline} />
    </Box>
  );
}
