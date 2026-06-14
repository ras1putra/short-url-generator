import { useState, useEffect } from "react";
import { useConfig } from "@/hooks/useConfig";
import { useReadContract, useWriteContract, useWaitForTransactionReceipt } from "wagmi";
import { useWalletConnection } from "@/hooks/wallet/useWalletConnection";
import { ERC20_ABI } from "@/lib/paymentGateway";
import { NFT_PASS_ABI } from "@/lib/nftPass";
import { toast } from "sonner";
import { classifyWalletError } from "@/lib/wallet";

export const ACTION_APPROVE = "approve" as const;
export const ACTION_MINT = "mint" as const;
export type NFTPassAction = typeof ACTION_APPROVE | typeof ACTION_MINT;

export function useNFTPass() {
  const { data: appConfig, isLoading: isConfigLoading, isError } = useConfig();
  const tokenSymbol = appConfig?.token_symbol ?? "";
  const { address, isConnected, connectWallet, isConnecting } = useWalletConnection();

  const [isMintingProcess, setIsMintingProcess] = useState(false);
  const [currentTxHash, setCurrentTxHash] = useState<`0x${string}` | null>(null);
  const [pendingAction, setPendingAction] = useState<NFTPassAction | null>(null);

  // Check if user already holds the NFT Pass
  const { data: nftBalance, refetch: refetchNFTBalance } = useReadContract({
    address: appConfig?.contract_nft_pass as `0x${string}`,
    abi: NFT_PASS_ABI,
    functionName: "balanceOf",
    args: address ? [address as `0x${string}`] : undefined,
    query: {
      enabled: !!address && !!appConfig?.contract_nft_pass,
    },
  });

  // Read mintPrice from the contract dynamically
  const { data: rawMintPrice } = useReadContract({
    address: appConfig?.contract_nft_pass as `0x${string}`,
    abi: NFT_PASS_ABI,
    functionName: "mintPrice",
    query: {
      enabled: !!appConfig?.contract_nft_pass,
    },
  });

  const price = rawMintPrice !== undefined ? BigInt(rawMintPrice.toString()) : BigInt(0);
  const displayPrice = rawMintPrice !== undefined ? (Number(rawMintPrice) / 1e18).toString() : "...";

  // Check SURL Balance
  const { data: surlBalance, refetch: refetchSurlBalance } = useReadContract({
    address: appConfig?.contract_token as `0x${string}`,
    abi: ERC20_ABI,
    functionName: "balanceOf",
    args: address ? [address as `0x${string}`] : undefined,
    query: {
      enabled: !!address && !!appConfig?.contract_token,
    },
  });

  // Check Allowance
  const { data: allowance, refetch: refetchAllowance } = useReadContract({
    address: appConfig?.contract_token as `0x${string}`,
    abi: ERC20_ABI,
    functionName: "allowance",
    args: address && appConfig?.contract_nft_pass ? [address as `0x${string}`, appConfig.contract_nft_pass as `0x${string}`] : undefined,
    query: {
      enabled: !!address && !!appConfig?.contract_token && !!appConfig?.contract_nft_pass,
    },
  });

  const { mutateAsync } = useWriteContract();

  // Watch transaction completion
  const { isLoading: isTxWaiting } = useWaitForTransactionReceipt({
    hash: currentTxHash || undefined,
    query: {
      enabled: !!currentTxHash,
    }
  });

  const handleMintPass = async () => {
    if (!isConnected || !address) {
      try {
        await connectWallet();
      } catch (err) {
        toast.error(classifyWalletError(err, "connect"));
      }
      return;
    }

    if (!appConfig?.contract_nft_pass || !appConfig?.contract_token) {
      toast.error("Web3 configuration is missing.");
      return;
    }

    // Check balance
    if (surlBalance !== undefined && BigInt(surlBalance.toString()) < price) {
      toast.error(`Insufficient ${tokenSymbol} balance. You need at least ${displayPrice} ${tokenSymbol} to mint.`);
      return;
    }

    setIsMintingProcess(true);

    try {
      // Approve if needed
      const currentAllowance = allowance !== undefined ? BigInt(allowance.toString()) : BigInt(0);
      if (currentAllowance < price) {
        setPendingAction(ACTION_APPROVE);
        toast.info(`Approving ${tokenSymbol} spend for NFT Pass...`);
        const hash = await mutateAsync({
          address: appConfig.contract_token as `0x${string}`,
          abi: ERC20_ABI,
          functionName: "approve",
          args: [appConfig.contract_nft_pass as `0x${string}`, price],
        });
        setCurrentTxHash(hash);
        return;
      }

      // Mint
      setPendingAction(ACTION_MINT);
      toast.info("Minting your NFT Pass...");
      const hash = await mutateAsync({
        address: appConfig.contract_nft_pass as `0x${string}`,
        abi: NFT_PASS_ABI,
        functionName: "mint",
      });
      setCurrentTxHash(hash);
    } catch (err) {
      toast.error(classifyWalletError(err));
      setIsMintingProcess(false);
      setPendingAction(null);
      setCurrentTxHash(null);
    }
  };

  // Monitor transaction status
  useEffect(() => {
    if (currentTxHash && !isTxWaiting) {
      const action = pendingAction;
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setCurrentTxHash(null);
      setPendingAction(null);

      if (action === ACTION_APPROVE) {
        toast.success("Allowance approved successfully!");
        refetchAllowance().then(() => {
          handleMintPass();
        });
      } else if (action === ACTION_MINT) {
        toast.success("NFT Pass minted successfully!");
        refetchNFTBalance();
        refetchSurlBalance();
        setIsMintingProcess(false);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isTxWaiting, currentTxHash, pendingAction]);

  const hasNFT = nftBalance !== undefined && BigInt(nftBalance.toString()) > BigInt(0);
  const isBusy = isMintingProcess || isTxWaiting || isConnecting;

  return {
    isConnected,
    isConnecting,
    isConfigLoading,
    isError,
    tokenSymbol,
    displayPrice,
    hasNFT,
    isBusy,
    pendingAction,
    handleMintPass,
  };
}
