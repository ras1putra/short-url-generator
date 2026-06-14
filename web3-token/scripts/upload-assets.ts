import * as fs from "fs";
import * as path from "path";

async function main() {
  let ipfsUrl = process.env.IPFS_API_URL || "http://ipfs:5001";

  try {
    await fetch(`${ipfsUrl}/api/v0/version`, { method: "POST" });
  } catch (e) {
    console.log(`Could not reach IPFS at ${ipfsUrl}, trying localhost fallback...`);
    ipfsUrl = "http://127.0.0.1:5001";
    try {
      await fetch(`${ipfsUrl}/api/v0/version`, { method: "POST" });
    } catch (err) {
      throw new Error(`IPFS API is not reachable at ${process.env.IPFS_API_URL || "http://ipfs:5001"} or http://127.0.0.1:5001. Ensure your Kubo node is running and port 5001 is open.`);
    }
  }

  console.log(`Connecting to IPFS API at: ${ipfsUrl}`);

  const videoPath = path.join(__dirname, "../assets/nft_video.mp4");
  if (!fs.existsSync(videoPath)) {
    throw new Error(`Asset not found at ${videoPath}`);
  }

  console.log("Uploading nft_video.mp4 to IPFS...");
  const videoBuffer = fs.readFileSync(videoPath);
  const videoBlob = new Blob([videoBuffer]);
  const videoForm = new FormData();
  videoForm.append("file", videoBlob, "nft_video.mp4");

  const videoResponse = await fetch(`${ipfsUrl}/api/v0/add?cid-version=1`, {
    method: "POST",
    body: videoForm,
  });

  if (!videoResponse.ok) {
    throw new Error(`Failed to upload video: ${videoResponse.statusText}`);
  }

  const videoData = (await videoResponse.json()) as { Hash: string };
  const videoCID = videoData.Hash;
  console.log(`Video uploaded successfully! CID: ${videoCID}`);

  const metadata = {
    name: "AdBypass NFT Pass",
    description: "Premium NFT Pass to bypass redirect ads",
    image: `ipfs://${videoCID}`,
    animation_url: `ipfs://${videoCID}`,
    attributes: [
      {
        trait_type: "Access",
        value: "Unlimited AdBypass"
      }
    ]
  };

  console.log("Uploading metadata JSON to IPFS...");
  const metadataBlob = new Blob([JSON.stringify(metadata, null, 2)], { type: "application/json" });
  const metadataForm = new FormData();
  metadataForm.append("file", metadataBlob, "metadata.json");

  const metadataResponse = await fetch(`${ipfsUrl}/api/v0/add?cid-version=1`, {
    method: "POST",
    body: metadataForm,
  });

  if (!metadataResponse.ok) {
    throw new Error(`Failed to upload metadata: ${metadataResponse.statusText}`);
  }

  const metadataData = (await metadataResponse.json()) as { Hash: string };
  const metadataCID = metadataData.Hash;
  const metadataURI = `ipfs://${metadataCID}`;
  console.log(`Metadata uploaded successfully! CID: ${metadataCID}`);
  console.log(`Metadata URI: ${metadataURI}`);

  const outputPath = path.join(__dirname, "../metadata-cid.json");
  fs.writeFileSync(outputPath, JSON.stringify({ metadataURI, metadataCID, videoCID }, null, 2));
  console.log(`Saved CIDs to ${outputPath}`);
}

main().catch((error) => {
  console.error("IPFS Upload Failed:", error);
  process.exitCode = 1;
});
