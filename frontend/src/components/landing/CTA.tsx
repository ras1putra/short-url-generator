"use client";

import Link from "next/link";
import { ArrowRight, Megaphone } from "lucide-react";

export default function CTA() {
  return (
    <section className="py-32 px-6 md:px-12 max-w-7xl mx-auto text-center">
      <div className="relative">
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
          <div className="w-[600px] h-[300px] rounded-full bg-[radial-gradient(ellipse,rgba(110,231,183,0.06)_0%,transparent_70%)]" />
        </div>
        <p className="font-mono-dm text-xs mb-6 tracking-widest uppercase text-[#6EE7B7]">
          Ready to ship?
        </p>
        <h2 className="text-5xl md:text-7xl font-black tracking-tight leading-none mb-8">
          SHORT LINKS.<br />SERIOUS<br />INFRASTRUCTURE.
        </h2>
        <p className="text-lg mb-10 max-w-md mx-auto text-white/70">
          Register in seconds. Start shortening immediately.
          No credit card. No limits on the free plan.
        </p>
        <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
          <Link href="/register" className="btn-primary px-10 py-4 rounded-xl text-base inline-flex items-center gap-2">
            Create free account <ArrowRight size={18} />
          </Link>
          <Link href="/register/advertiser" className="px-8 py-4 rounded-xl text-sm font-bold tracking-tight inline-flex items-center gap-2 bg-[#22D3EE] text-[#0A0A0A] hover:bg-[#67E8F9] hover:-translate-y-0.5 hover:shadow-[0_8px_30px_rgba(34,211,238,0.3)] transition-all duration-200">
            <Megaphone size={16} />
            Advertise with us
          </Link>
          <Link href="/login" className="px-8 py-4 rounded-xl text-sm transition-colors border border-white/20 text-white/70 hover:text-white hover:border-white/40">
            Already have an account
          </Link>
        </div>
      </div>
    </section>
  );
}
