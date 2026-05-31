"use client";

import React, { createContext, useContext, useEffect, useState, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useUserStore } from "@/store/useUserStore";

interface WebSocketContextType {
  isConnected: boolean;
}

const WebSocketContext = createContext<WebSocketContextType>({
  isConnected: false,
});

export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const [isConnected, setIsConnected] = useState(false);
  const queryClient = useQueryClient();
  const user = useUserStore((s) => s.user);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    // Only establish connection if user is logged in
    if (!user) {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
        setIsConnected(false);
      }
      return;
    }

    let active = true;
    let reconnectTimeout: NodeJS.Timeout;

    function connect() {
      if (!active) return;

      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      const wsUrl = `${protocol}//${window.location.host}/api/wallet/ws`;

      console.log("[WebSocket] Connecting to:", wsUrl);
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log("[WebSocket] Connection established successfully");
        if (active) {
          setIsConnected(true);
        }
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === "WALLET_UPDATE") {
            console.log("[WebSocket] Real-time wallet update event received. Invalidating React-Query...");
            queryClient.invalidateQueries({ queryKey: ["wallet"] });
          }
        } catch (err) {
          console.error("[WebSocket] Failed to parse incoming socket event:", err);
        }
      };

      ws.onclose = (event) => {
        console.log("[WebSocket] Connection closed. Reconnecting in 5 seconds...", event);
        if (active) {
          setIsConnected(false);
          wsRef.current = null;
          reconnectTimeout = setTimeout(connect, 5000);
        }
      };

      ws.onerror = (err) => {
        if (active) {
          console.warn("[WebSocket] Connection error:", err);
        }
        ws.close();
      };
    }

    connect();

    return () => {
      active = false;
      clearTimeout(reconnectTimeout);
      if (wsRef.current) {
        const ws = wsRef.current;
        if (ws.readyState === WebSocket.CONNECTING) {
          // Gracefully close once the handshake completes to bypass native browser warnings
          ws.onopen = () => {
            ws.close();
          };
          ws.onerror = null;
          ws.onclose = null;
          ws.onmessage = null;
        } else {
          ws.close();
        }
        wsRef.current = null;
      }
    };
  }, [user, queryClient]);

  return (
    <WebSocketContext.Provider value={{ isConnected }}>
      {children}
    </WebSocketContext.Provider>
  );
}

export function useWebSocket() {
  return useContext(WebSocketContext);
}
