"use client";

import Link from "next/link";
import { Link2, Megaphone } from "lucide-react";

export default function Navbar() {
  return (
    <nav className="fixed top-0 left-0 right-0 z-50 flex items-center justify-between px-6 md:px-12 py-5 bg-[#0A0A0A]/85 backdrop-blur-xl border-b border-white/5">
      <div className="flex items-center gap-2">
        <div className="w-8 h-8 rounded-lg flex items-center justify-center bg-[#6EE7B7]">
          <Link2 size={16} className="text-[#0A0A0A] stroke-[2.5]" />
        </div>
        <span className="text-lg font-bold tracking-tight">go-short</span>
      </div>

      <div className="hidden md:flex items-center gap-8 text-sm text-white/80">
        <a href="#features" className="hover:text-white transition-colors">Features</a>
        <a href="#how" className="hover:text-white transition-colors">How it works</a>
        <a href="#stats" className="hover:text-white transition-colors">Plus points</a>
        <Link href="/register/advertiser" className="hover:text-white transition-colors flex items-center gap-1.5">
          <Megaphone size={14} />
          Advertise
        </Link>
      </div>

      <div className="flex items-center gap-3">
        <Link href="/login" className="text-sm px-4 py-2 rounded-lg text-white/80 hover:text-white transition-colors">
          Login
        </Link>
        <Link href="/register" className="btn-primary text-sm px-5 py-2 rounded-lg inline-block">
          Get started →
        </Link>
      </div>
    </nav>
  );
}
