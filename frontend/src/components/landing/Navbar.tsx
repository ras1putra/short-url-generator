"use client";

import { useState } from "react";
import Link from "next/link";
import { Link2, Megaphone, Menu, X, Zap, BarChart3 } from "lucide-react";

const NAV_LINKS = [
  { href: "/#features", label: "Features", icon: Zap },
  { href: "/#how", label: "How it works", icon: BarChart3 },
  { href: "/#stats", label: "Plus points", icon: Zap },
  { href: "/register/advertiser", label: "Advertise", icon: Megaphone },
];

export default function Navbar() {
  const [open, setOpen] = useState(false);

  return (
    <nav className="fixed top-0 left-0 right-0 z-50 flex items-center justify-between px-4 sm:px-6 md:px-12 py-5 bg-[#0A0A0A]/85 backdrop-blur-xl border-b border-white/5">
      <div className="flex items-center gap-2">
        <div className="w-8 h-8 rounded-lg flex items-center justify-center bg-[#6EE7B7]">
          <Link2 size={16} className="text-[#0A0A0A] stroke-[2.5]" />
        </div>
        <span className="text-lg font-bold tracking-tight">go-short</span>
      </div>

      <button
        onClick={() => setOpen(!open)}
        className="md:hidden relative z-50 w-10 h-10 rounded-xl flex items-center justify-center bg-white/[0.03] border border-white/[0.08] hover:bg-white/[0.06] transition-colors cursor-pointer"
        aria-label="Toggle menu"
      >
        {open ? <X size={18} className="text-white"/> : <Menu size={18} className="text-white/70" />}
      </button>

      <div className="hidden md:flex items-center gap-8 text-sm text-white/80">
        <Link href="/#features"className="hover:text-white transition-colors">Features</Link>
        <Link href="/#how"className="hover:text-white transition-colors">How it works</Link>
        <Link href="/#stats"className="hover:text-white transition-colors">Plus points</Link>
        <Link href="/register/advertiser"className="hover:text-white transition-colors flex items-center gap-1.5">
          <Megaphone size={14} />
          Advertise
        </Link>
      </div>

      <div className="hidden md:flex items-center gap-3">
        <Link href="/login"className="text-sm px-4 py-2 rounded-lg text-white/80 hover:text-white transition-colors">
          Login
        </Link>
        <Link href="/register"className="btn-primary text-sm px-5 py-2 rounded-lg inline-block">
          Get started →
        </Link>
      </div>

      <div
        className={`absolute top-full left-0 right-0 z-40 md:hidden overflow-hidden transition-all duration-300 ${
          open ? "max-h-80 opacity-100" : "max-h-0 opacity-0"
        }`}
      >
        <div className="mx-4 sm:mx-6 mb-3 rounded-2xl bg-[#0A0A0A]/95 backdrop-blur-xl border border-white/[0.08] shadow-xl overflow-hidden">
          {NAV_LINKS.map(({ href, label, icon: Icon }) => (
            <Link
              key={href}
              href={href}
              onClick={() => setOpen(false)}
              className="flex items-center gap-3 px-5 py-3.5 text-sm transition-colors text-white/60 hover:text-white hover:bg-white/[0.02] border-b border-white/[0.04] last:border-b-0"
            >
              <Icon size={16} className="text-white/40" />
              <span className="font-medium">{label}</span>
            </Link>
          ))}
          <div className="border-t border-white/[0.06] mt-1 pt-1 pb-2 px-2">
            <Link
              href="/login"
              onClick={() => setOpen(false)}
              className="flex items-center justify-center px-4 py-2.5 rounded-lg text-sm font-medium text-white/80 hover:text-white hover:bg-white/[0.04] transition-colors"
            >
              Login
            </Link>
            <Link
              href="/register"
              onClick={() => setOpen(false)}
              className="flex items-center justify-center px-4 py-2.5 rounded-lg text-sm font-bold tracking-tight btn-primary mt-1"
            >
              Get started
            </Link>
          </div>
        </div>
      </div>
    </nav>
  );
}
