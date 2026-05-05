import CreateLinkForm from "@/components/links/CreateLinkForm";
import LinkTable from "@/components/links/LinkTable";
import DashboardGlobe from "@/components/links/DashboardGlobe";

export default function DashboardPage() {
  return (
    <div>
      <div className="mb-8">
        <h1 className="text-3xl font-black tracking-tight text-white">Dashboard</h1>
        <p className="mt-2 text-white/50 font-mono-dm text-sm">{"// Create, manage, and track your shortened URLs"}</p>
      </div>

      <DashboardGlobe />

      <div className="mt-8">
        <CreateLinkForm />
      </div>

      <div className="mt-12">
        <h2 className="text-xl font-bold text-white/90 mb-6">Your Links</h2>
        <LinkTable />
      </div>
    </div>
  );
}