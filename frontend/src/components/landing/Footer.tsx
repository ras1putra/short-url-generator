"use client";

import Link from "next/link";
import { Link2 } from "lucide-react";

export default function Footer() {
  return (
    <footer className="px-4 sm:px-6 md:px-12 py-10 border-t border-white/[0.06]">
      <div className="max-w-7xl mx-auto flex flex-col md:flex-row items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 rounded-md flex items-center justify-center bg-[#6EE7B7]">
            <Link2 size={12} className="text-[#0A0A0A] stroke-[2.5]" />
          </div>
          <span className="font-bold text-sm tracking-tight">go-short</span>
        </div>
        <p className="text-xs font-mono-dm text-white/50">
            Built with care. Designed for you.
          </p>
          <div className="flex gap-4 sm:gap-6 text-xs text-white/60">
          <Link href="/login"className="hover:text-white transition-colors">Login</Link>
          <Link href="/register"className="hover:text-white transition-colors">Register</Link>
        </div>
      </div>
    </footer>
  );
}
