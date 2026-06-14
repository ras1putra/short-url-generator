"use client";

import Navbar from "@/components/landing/Navbar";
import Hero from "@/components/landing/Hero";
import Marquee from "@/components/landing/Marquee";
import Stats from "@/components/landing/Stats";
import Features from "@/components/landing/Features";
import NFTSection from "@/components/landing/NFTSection";
import HowItWorks from "@/components/landing/HowItWorks";
import Earn from "@/components/landing/Earn";
import CTA from "@/components/landing/CTA";
import Footer from "@/components/landing/Footer";

export default function LandingPage() {
  return (
    <div className="min-h-screen bg-[#0A0A0A] text-white overflow-x-hidden grain-overlay font-syne">
      <Navbar />

      <main>
        <Hero />

        <div className="divider-line mx-6 md:mx-12" />
        <NFTSection />

        <div className="divider-line mx-6 md:mx-12" />
        <Marquee />
        <Stats />

        <div className="divider-line mx-6 md:mx-12" />
        <Features />

        <div className="divider-line mx-6 md:mx-12" />
        <HowItWorks />

        <div className="divider-line mx-6 md:mx-12" />
        <Earn />

        <div className="divider-line mx-6 md:mx-12" />

        <CTA />
      </main>

      <Footer />
    </div>
  );
}
