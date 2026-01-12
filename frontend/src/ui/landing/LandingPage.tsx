import { useNavigate } from "react-router-dom";
import { Box } from "@mantine/core";
import {
  Header,
  Hero,
  SegmentSelector,
  HowItWorks,
  Features,
  ExampleCall,
  FAQ,
  CTASection,
  Footer,
} from "./components";
import { SHARED_CONTENT } from "./content/shared";

export function LandingPage() {
  const navigate = useNavigate();

  const allFeatures = Object.values(SHARED_CONTENT.features);

  return (
    <Box>
      <Header />
      <Hero
        title="Vaše AI asistentka na telefonu"
        tagline="Když nezvednete, Karen to zvedne za vás. Zjistí, kdo volá a proč. Spam odfiltruje."
        ctaText="Vyzkoušet zdarma"
        onCtaClick={() => navigate("/login")}
      />
      <SegmentSelector />
      <HowItWorks steps={SHARED_CONTENT.howItWorks} />
      <Features features={allFeatures} />
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
