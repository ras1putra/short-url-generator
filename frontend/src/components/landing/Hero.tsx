"use client";

import { useState } from "react";
import Link from "next/link";
import { ArrowRight, ChevronRight, Zap, Copy, Check, Megaphone } from "lucide-react";

export default function Hero() {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText("https://go-short.dev/porto");
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <section className="relative pt-40 pb-32 px-6 md:px-12 max-w-7xl mx-auto">
      {/* Background glow blobs */}
      <div className="absolute top-20 left-1/4 w-96 h-96 rounded-full pointer-events-none bg-[radial-gradient(circle,rgba(110,231,183,0.06)_0%,transparent_70%)]" />
      <div className="absolute top-40 right-1/4 w-80 h-80 rounded-full pointer-events-none bg-[radial-gradient(circle,rgba(52,211,153,0.04)_0%,transparent_70%)]" />

      {/* Badge */}
      <div className="fade-in inline-flex items-center gap-2 mb-8 px-4 py-2 rounded-full text-xs font-mono-dm border border-[#6EE7B7]/30 text-[#6EE7B7] bg-[#6EE7B7]/5">
        <Zap size={12} />
        <span>Lightning-fast redirects · Earn crypto · Trusted by thousands</span>
      </div>

      {/* Headline */}
      <h1 className="fade-in text-5xl md:text-7xl lg:text-8xl font-black tracking-tight leading-none mb-6 text-glow [animation-delay:0.1s]">
        SHORT LINKS.<br />
        <span className="stat-number">LONG REACH.</span>
      </h1>

      <p className="fade-in text-lg md:text-xl mb-10 max-w-xl leading-relaxed text-white/80 [animation-delay:0.2s]">
        Create short, memorable links. Run targeted ad campaigns. Earn crypto with every click.
      </p>

      <div className="fade-in flex flex-col sm:flex-row items-start sm:items-center gap-4 [animation-delay:0.3s]">
        <Link href="/register" className="btn-primary px-8 py-4 rounded-xl text-base inline-flex items-center gap-2">
          Start shortening <ArrowRight size={18} />
        </Link>
        <Link href="/register/advertiser" className="px-8 py-4 rounded-xl text-base font-bold tracking-tight inline-flex items-center gap-2 bg-[#22D3EE] text-[#0A0A0A] hover:bg-[#67E8F9] hover:-translate-y-0.5 hover:shadow-[0_8px_30px_rgba(34,211,238,0.3)] transition-all duration-200">
          <Megaphone size={18} />
          Advertise with us
        </Link>
        <a href="#how" className="inline-flex items-center gap-2 text-sm text-white/70 hover:text-white transition-colors">
          See how it works <ChevronRight size={16} />
        </a>
      </div>

      {/* Demo card */}
      <div className="fade-in mt-20 rounded-2xl p-6 md:p-8 max-w-xl glow-green bg-white/[0.06] border border-white/[0.12] [animation-delay:0.4s]">
        <p className="text-xs mb-3 font-mono-dm text-white/60">{"// Try it out"}</p>
        <div className="flex items-center justify-between gap-4 p-3 rounded-xl mb-3 bg-white/[0.06] border border-white/[0.12]">
          <span className="text-sm truncate font-mono-dm text-white/70">
            https://example.com/very/long/path/to/some/resource
          </span>
        </div>
        <div className="flex items-center justify-between gap-3 p-3 rounded-xl bg-[#6EE7B7]/10 border border-[#6EE7B7]/20">
          <span className="font-bold font-mono-dm text-sm text-[#6EE7B7]">
            go-short.dev/porto
          </span>
          <button onClick={handleCopy}
            className="hover:cursor-pointer flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-lg transition-all bg-[#6EE7B7]/15 text-[#6EE7B7]">
            {copied ? <Check size={12} /> : <Copy size={12} />}
            {copied ? "Copied!" : "Copy"}
          </button>
        </div>
      </div>
    </section>
  );
}
