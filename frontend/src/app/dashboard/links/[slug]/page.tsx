import LinkDetailPage from "@/components/links/LinkDetailPage";

export default async function LinkStatsPage({ params }: { params: Promise<{ slug: string }> }) {
  const resolvedParams = await params;

  return <LinkDetailPage slug={resolvedParams.slug} />;
}
