import { AxiosError } from "axios";

function extractErrorMessage(err: unknown): string {
  if (err instanceof AxiosError) {
    return err.response?.data?.message ?? "";
  }
  if (typeof err === "object" && err !== null && "message" in err) {
    return (err as { message: string }).message;
  }
  if (err instanceof Error) return err.message;
  return String(err);
}

export function classifyWalletError(err: unknown, context: "connect" | "deposit" | "withdraw" = "deposit"): string {
  const originalMsg = extractErrorMessage(err);
  const msg = originalMsg.toLowerCase();

  if (msg.includes("rejected") || msg.includes("denied")) return "User rejected the request";
  if (msg.includes("insufficient funds")) return "Insufficient balance";
  if (msg.includes("cooldownactive") || msg.includes("cooldown") || msg.includes("already claimed"))
    return "Already claimed. Please wait 24 hours between claims.";
  if (msg.includes("deadlineexpired")) return "Claim deadline expired. Please try again.";
  if (msg.includes("exceedsmaxclaim")) return "Claim amount exceeds the maximum allowed.";
  if (msg.includes("insufficientfaucetbalance")) return "Faucet is empty. Contact support.";
  if (msg.includes("invalidsignature")) return "Invalid claim signature. Please try again.";
  if (msg.includes("execution reverted")) return "Transaction reverted by contract";
  if (msg.includes("no wallet")) return originalMsg;

  if (originalMsg) return originalMsg;
  return context === "connect" ? "Failed to connect wallet" : "Transaction failed";
}

export function formatBalance(val: any): string {
  const num = Number(val || 0);
  return num.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 8 });
}

