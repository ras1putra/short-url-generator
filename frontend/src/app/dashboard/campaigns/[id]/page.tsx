import CampaignDetailClient from "./CampaignDetailClient";

export default async function CampaignPage({ params }: { params: Promise<{ id: string }> }) {
  const resolvedParams = await params;
  return <CampaignDetailClient id={resolvedParams.id} />;
}
