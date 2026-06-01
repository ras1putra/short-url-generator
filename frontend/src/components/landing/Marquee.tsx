"use client";

const MARQUEE_KEYWORDS = [
  "Fast Redirects", "Analytics", "QR Codes", "Custom Slugs", "Earn Crypto",
  "Real-time Tracking", "Geo Insights", "Secure Links", "Team Ready", "API Access"
];

export default function Marquee() {
  return (
    <div className="overflow-hidden py-6 border-y border-white/5">
      <div className="marquee">
        {[...Array(2)].map((_, i) => (
          <div key={i} className="flex items-center gap-12 pr-12">
            {MARQUEE_KEYWORDS.map((t) => (
              <span key={t} className="text-sm font-bold tracking-widest uppercase font-mono-dm text-white/30 whitespace-nowrap">
                {t} &nbsp;·
              </span>
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}
