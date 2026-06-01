"use client";

import { Shield, Zap, BarChart3, QrCode, Globe, Clock, Megaphone, Coins } from "lucide-react";

const FEATURES_DATA = [
  {
    icon: <Shield size={20} />,
    title: "Secure by Default",
    desc: "Your account is protected with industry-standard security. Automatic session management keeps you logged in without compromising safety.",
  },
  {
    icon: <Zap size={20} />,
    title: "Lightning Fast",
    desc: "Every link redirects instantly, even under heavy traffic. Your visitors never wait — your campaigns never slow down.",
  },
  {
    icon: <BarChart3 size={20} />,
    title: "Powerful Analytics",
    desc: "Know your audience inside out. Track where clicks come from, what devices they use, and how they find you — all in real time.",
  },
  {
    icon: <QrCode size={20} />,
    title: "QR Codes Built In",
    desc: "Every link comes with a ready-to-use QR code. Perfect for print, packaging, or in-store displays. Customize the size to fit your needs.",
  },
  {
    icon: <Coins size={20} />,
    title: "Earn Crypto",
    desc: "Monetize your links and earn tokens when visitors see ads before being redirected. Withdraw anytime to your connected wallet.",
  },
  {
    icon: <Megaphone size={20} />,
    title: "Ad Campaigns",
    desc: "Launch and manage ad campaigns with full budget control. Set your CPM, choose categories, and track spend in real time.",
  },
  {
    icon: <Globe size={20} />,
    title: "Global Reach Insights",
    desc: "See where in the world your links are being clicked. We respect privacy while giving you the geographic insights that matter.",
  },
  {
    icon: <Clock size={20} />,
    title: "Expiring Links",
    desc: "Set a lifespan for any link. Perfect for limited-time offers, event registrations, or time-sensitive campaigns. Expired links gracefully redirect visitors.",
  },
];

export default function Features() {
  return (
    <section id="features" className="py-24 px-6 md:px-12 max-w-7xl mx-auto">
      <div className="mb-16">
        <p className="text-xs font-mono-dm mb-4 tracking-widest uppercase text-[#6EE7B7]">
          Features
        </p>
        <h2 className="text-4xl md:text-6xl font-black tracking-tight leading-none">
          EVERYTHING YOU<br />NEED TO SCALE.
        </h2>
      </div>

      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
        {FEATURES_DATA.map(({ icon, title, desc }) => (
          <div key={title} className="card-hover p-6 rounded-2xl bg-white/[0.05] border border-white/[0.12]">
            <div className="w-10 h-10 rounded-xl flex items-center justify-center mb-5 bg-[#6EE7B7]/10 text-[#6EE7B7]">
              {icon}
            </div>
            <h3 className="font-bold text-lg mb-2 tracking-tight text-white/90">{title}</h3>
            <p className="text-sm leading-relaxed text-white/70">{desc}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
