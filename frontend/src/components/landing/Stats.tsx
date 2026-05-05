"use client";

const STATS_DATA = [
  { num: "Fast", unit: "", label: "Instant redirects every time" },
  { num: "Global", unit: "", label: "Reach audiences worldwide" },
  { num: "Secure", unit: "", label: "Your links are always safe" },
  { num: "Simple", unit: "", label: "No fuss, just short links" },
];

export default function Stats() {
  return (
    <section id="stats" className="py-24 px-6 md:px-12 max-w-7xl mx-auto">
      <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
        {STATS_DATA.map(({ num, unit, label }) => (
          <div key={label} className="p-6 rounded-2xl bg-white/[0.05] border border-white/[0.12]">
            <div className="flex items-baseline gap-1 mb-1">
              <span className="text-4xl font-black stat-number">{num}</span>
              <span className="text-sm font-mono-dm text-white/60">{unit}</span>
            </div>
            <p className="text-sm text-white/70">{label}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
