import { useNavigate } from "react-router-dom";
import { Box } from "@mantine/core";
import {
  Header,
  Hero,
  TrustBadges,
  SegmentSelector,
  HowItWorks,
  Features,
  PricingSection,
  ComparisonTable,
  ExampleCall,
  FAQ,
  CTASection,
  Footer,
} from "./components";
import { SHARED_CONTENT, CONTACT_MAILTO_FIRMA } from "./content/shared";

export function LandingPage() {
  const navigate = useNavigate();

  const allFeatures = Object.values(SHARED_CONTENT.features);

  const handleContactClick = () => {
    window.location.href = CONTACT_MAILTO_FIRMA;
  };

  return (
    <Box>
      <Header />
      <Hero
        title="Vaše AI asistentka na telefonu"
        tagline="Když nezvednete, Karen to zvedne za vás. Zjistí, kdo volá a proč. Spam odfiltruje."
        ctaText="Aktivovat Karen zdarma"
        onCtaClick={() => navigate("/login")}
      />
      <TrustBadges />
      <SegmentSelector />
      <HowItWorks steps={SHARED_CONTENT.howItWorks} />
      <Features features={allFeatures} />
      <PricingSection
        onTrialClick={() => navigate("/login")}
        onBuyClick={() => navigate("/login")}
        onContactClick={handleContactClick}
      />
      <ComparisonTable />
      <ExampleCall />
      <FAQ items={SHARED_CONTENT.faq} />
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
