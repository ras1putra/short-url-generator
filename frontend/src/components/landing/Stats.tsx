"use client";

const STATS_DATA = [
  { num: "Fast", label: "Instant redirects every time" },
  { num: "Global", label: "Reach audiences worldwide" },
  { num: "Secure", label: "Your links are always safe" },
  { num: "Simple", label: "No fuss, just short links" },
];

export default function Stats() {
  return (
    <section id="stats"className="py-12 sm:py-16 lg:py-24 px-4 sm:px-6 md:px-12 max-w-7xl mx-auto">
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 sm:gap-6">
        {STATS_DATA.map(({ num, label }) => (
          <div key={label} className="p-4 sm:p-6 rounded-2xl bg-white/[0.05] border border-white/[0.12]">
            <div className="flex items-baseline gap-1 mb-1">
              <span className="text-2xl sm:text-3xl lg:text-4xl font-black stat-number">{num}</span>
            </div>
            <p className="text-sm text-white/70">{label}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
