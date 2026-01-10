import { useNavigate } from "react-router-dom";
import { Box } from "@mantine/core";
import {
  Header,
  Hero,
  SegmentSelector,
  HowItWorks,
  Features,
  ExampleCall,
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
        title="Tvoje AI asistentka na telefonu"
        tagline="Kdyz nezvednes, Karen zvedne za tebe. Zjisti kdo vola a proc. Spam odfiltruje."
        ctaText="Vyzkouset zdarma"
        onCtaClick={() => navigate("/login")}
      />
      <SegmentSelector />
      <HowItWorks steps={SHARED_CONTENT.howItWorks} />
      <Features features={allFeatures} />
      <ExampleCall />
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
