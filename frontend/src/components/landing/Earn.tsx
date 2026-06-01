"use client";

import { Coins, Eye, Wallet, TrendingUp } from "lucide-react";

const EARN_FEATURES = [
  {
    icon: <Eye size={20} />,
    title: "Monetize Any Link",
    desc: "Toggle monetization on any short link. Visitors see a short ad before being redirected, and you earn tokens.",
  },
  {
    icon: <Coins size={20} />,
    title: "Earn Crypto Rewards",
    desc: "Each view earns you tokens deposited directly to your wallet. No middlemen, no delays — just straight to your balance.",
  },
  {
    icon: <TrendingUp size={20} />,
    title: "Track Your Earnings",
    desc: "See exactly how much each link earns. Real-time stats show impressions, payouts, and your total balance at a glance.",
  },
  {
    icon: <Wallet size={20} />,
    title: "Connect Your Wallet",
    desc: "Link any EVM-compatible wallet to withdraw your earnings. Full control over your funds at all times.",
  },
];

export default function Earn() {
  return (
    <section id="earn" className="py-24 px-6 md:px-12 max-w-7xl mx-auto">
      <div className="mb-16">
        <p className="text-xs font-mono-dm mb-4 tracking-widest uppercase text-[#6EE7B7]">
          Earn crypto
        </p>
        <h2 className="text-4xl md:text-6xl font-black tracking-tight leading-none">
          SHORT LINKS.<br />EARN TOKENS.
        </h2>
      </div>

      <div className="grid md:grid-cols-2 gap-4">
        {EARN_FEATURES.map(({ icon, title, desc }) => (
          <div key={title} className="p-6 rounded-2xl bg-white/[0.05] border border-white/[0.12]">
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
