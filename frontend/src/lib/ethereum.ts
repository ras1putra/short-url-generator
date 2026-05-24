export function getEthereum() {
  if (typeof window === "undefined") return undefined;
  return (window as unknown as { ethereum?: { request: (args: unknown) => Promise<unknown>; isMetaMask?: boolean } }).ethereum;
}
