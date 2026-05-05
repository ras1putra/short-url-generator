"use client";

import Navbar from "@/components/landing/Navbar";
import Hero from "@/components/landing/Hero";
import Marquee from "@/components/landing/Marquee";
import Stats from "@/components/landing/Stats";
import Features from "@/components/landing/Features";
import HowItWorks from "@/components/landing/HowItWorks";
import CTA from "@/components/landing/CTA";
import Footer from "@/components/landing/Footer";

export default function LandingPage() {
  return (
    <div className="min-h-screen bg-[#0A0A0A] text-white overflow-x-hidden grain-overlay font-syne">
      <Navbar />

      <main>
        <Hero />
        <Marquee />
        <Stats />

        <div className="divider-line mx-6 md:mx-12" />
        <Features />

        <div className="divider-line mx-6 md:mx-12" />
        <HowItWorks />

        <div className="divider-line mx-6 md:mx-12" />

        <section className="py-24 px-6 md:px-12 max-w-7xl mx-auto text-center">
          <p className="text-lg text-white/40">
            Built with care. Designed for you.
          </p>
        </section>

        <div className="divider-line mx-6 md:mx-12" />

        <CTA />
      </main>

      <Footer />
    </div>
  );
}
