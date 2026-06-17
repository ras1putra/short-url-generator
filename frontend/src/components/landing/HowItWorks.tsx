"use client";

const STEPS = [
  {
    step: "01",
    title: "Create Your Account",
    desc: "Sign up in seconds. No credit card required, no complicated setup. Just a simple account ready to go.",
  },
  {
    step: "02",
    title: "Shorten Your Link",
    desc: "Paste any long URL and get a clean, short link instantly. Add a custom name or set an expiration date if you want.",
  },
  {
    step: "03",
    title: "Share & Watch It Grow",
    desc: "Share your link anywhere — social media, email, print. Watch clicks roll in with detailed insights about your audience.",
  },
];

export default function HowItWorks() {
  return (
    <section id="how"className="py-12 sm:py-16 lg:py-24 px-4 sm:px-6 md:px-12 max-w-7xl mx-auto">
      <div className="mb-8 sm:mb-12 lg:mb-16">
        <p className="text-xs font-mono-dm mb-4 tracking-widest uppercase text-[#6EE7B7]">
          How it works
        </p>
        <h2 className="text-3xl sm:text-4xl md:text-6xl font-black tracking-tight leading-none">
          THREE STEPS.<br />DONE.
        </h2>
      </div>

      <div className="grid md:grid-cols-3 gap-4 sm:gap-6">
        {STEPS.map(({ step, title, desc }) => (
          <div key={step} className="relative p-8 rounded-2xl bg-white/[0.05] border border-white/[0.12]">
            <span className="font-black text-7xl absolute -top-4 right-6 text-white/[0.08] leading-none font-syne">
              {step}
            </span>
            <div className="relative">
              <span className="font-mono-dm text-xs mb-4 block text-[#6EE7B7]">STEP {step}</span>
              <h3 className="text-2xl font-black mb-3 tracking-tight text-white/90">{title}</h3>
              <p className="text-sm leading-relaxed text-white/70">{desc}</p>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
