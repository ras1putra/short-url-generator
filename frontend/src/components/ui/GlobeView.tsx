"use client";

import { useMemo, useRef, useState, useCallback, useEffect } from "react";
import Globe, { GlobeMethods } from "react-globe.gl";
import * as THREE from "three";
import { getCountryCoord, getCountryName } from "@/lib/countries";

export interface GlobePoint {
  country: string;
  count: number;
}

interface GlobeViewProps {
  points: GlobePoint[];
  height?: number;
}

const NEON = new THREE.Color(0x6EE7B7);

export default function GlobeView({ points, height = 420 }: GlobeViewProps) {
  const globeEl = useRef<GlobeMethods | undefined>(undefined);
  const containerRef = useRef<HTMLDivElement>(null);
  const [width, setWidth] = useState(700);
  const [ready, setReady] = useState(false);

  const updateWidth = useCallback(() => {
    if (containerRef.current) {
      setWidth(containerRef.current.clientWidth);
    }
  }, []);

  useEffect(() => {
    updateWidth();
    window.addEventListener("resize", updateWidth);
    return () => window.removeEventListener("resize", updateWidth);
  }, [updateWidth]);

  const handleGlobeReady = useCallback(() => {
    if (globeEl.current) {
      globeEl.current.pointOfView({ lat: 20, lng: 0, altitude: 2.2 }, 2000);
      const controls = globeEl.current.controls();
      controls.autoRotate = true;
      controls.autoRotateSpeed = 0.35;
      controls.enableDamping = true;
      controls.dampingFactor = 0.1;
    }
    Promise.resolve().then(() => setReady(true));
  }, []);

  const globeMaterial = useMemo(() => {
    const mat = new THREE.MeshPhongMaterial({
      color: new THREE.Color(0x0a0a0a),
      emissive: NEON,
      emissiveIntensity: 0.08,
      transparent: true,
      opacity: 0.95,
      shininess: 15,
      specular: new THREE.Color(0x222222),
    });
    return mat;
  }, []);

  const globePoints = useMemo(() => {
    return points
      .map((p) => {
        const coord = getCountryCoord(p.country);
        if (!coord) return null;
        return {
          lat: coord.lat,
          lng: coord.lng,
          label: `${getCountryName(p.country)}: ${p.count.toLocaleString()} clicks`,
          count: p.count,
        };
      })
      .filter(Boolean) as { lat: number; lng: number; label: string; count: number }[];
  }, [points]);

  const maxCount = useMemo(
    () => Math.max(...globePoints.map((p) => p.count), 1),
    [globePoints]
  );

  const arcsData = useMemo(() => {
    if (globePoints.length < 1) return [];
    const sorted = [...globePoints].sort((a, b) => b.count - a.count);
    const top = sorted.slice(0, Math.min(5, sorted.length));
    const center: [number, number] = [39.0, -98.5];
    const arcs: { startLat: number; startLng: number; endLat: number; endLng: number; count: number }[] = [];
    for (const p of top) {
      arcs.push({
        startLat: center[0],
        startLng: center[1],
        endLat: p.lat,
        endLng: p.lng,
        count: p.count,
      });
    }
    return arcs;
  }, [globePoints]);

  const customPointsData = useMemo(() => {
    return globePoints.map((p) => ({
      ...p,
      size: Math.max(0.4, (p.count / maxCount) * 1.2),
    }));
  }, [globePoints, maxCount]);

  const ringsData = useMemo(() => {
    return globePoints.map((p) => ({
      lat: p.lat,
      lng: p.lng,
      label: p.label,
      maxR: Math.max(1.5, (p.count / maxCount) * 8),
      propagationSpeed: Math.max(1, (p.count / maxCount) * 4),
      repeatPeriod: 600,
    }));
  }, [globePoints, maxCount]);

  if (points.length === 0) {
    return null;
  }

  return (
    <div ref={containerRef} style={{ height: `${height}px`, width: "100%", position: "relative" }} className="rounded-2xl overflow-hidden">
      {!ready && (
        <div className="absolute inset-0 z-10 flex items-center justify-center bg-[#0a0a0a]/80 backdrop-blur-sm rounded-2xl">
          <div className="flex flex-col items-center gap-3">
            <div className="w-10 h-10 border-2 border-[#6EE7B7]/30 border-t-[#6EE7B7] rounded-full animate-spin" />
            <p className="text-[#6EE7B7]/60 text-xs font-mono-dm uppercase tracking-widest">Loading globe</p>
          </div>
        </div>
      )}
      <Globe
        ref={globeEl}
        width={width}
        height={height}
        backgroundColor="rgba(0,0,0,0)"
        globeImageUrl="/earth-dark.jpg"
        globeMaterial={globeMaterial}
        atmosphereColor="#6EE7B7"
        atmosphereAltitude={0.25}
        showGraticules={true}
        showAtmosphere={true}
        onGlobeReady={handleGlobeReady}
        pointsData={customPointsData}
        pointLat="lat"
        pointLng="lng"
        pointAltitude={0.018}
        pointRadius={((obj: object) => (obj as (typeof customPointsData)[number]).size) as unknown as number}
        pointColor={() => "#6EE7B7"}
        pointsMerge={true}
        arcsData={arcsData}
        arcStartLat="startLat"
        arcStartLng="startLng"
        arcEndLat="endLat"
        arcEndLng="endLng"
        arcColor={() => (t: number) => `rgba(110,231,183,${Math.max(0, 1 - t)})`}
        arcAltitude={0.2}
        arcStroke={0.6}
        arcDashLength={0.4}
        arcDashGap={0.2}
        arcDashAnimateTime={2000}
        ringsData={ringsData}
        ringLat="lat"
        ringLng="lng"
        ringAltitude={0.01}
        ringColor={() => (t: number) => `rgba(110,231,183,${Math.max(0, 1 - t)})`}
        ringMaxRadius="maxR"
        ringPropagationSpeed="propagationSpeed"
        ringRepeatPeriod="repeatPeriod"
      />
      <div
        className="pointer-events-none absolute inset-x-0 bottom-0 h-20"
        style={{ background: "linear-gradient(to top, rgba(5,5,5,0.8), transparent)" }}
      />
      <div
        className="pointer-events-none absolute inset-x-0 top-0 h-10"
        style={{ background: "linear-gradient(to bottom, rgba(5,5,5,0.8), transparent)" }}
      />
      <div
        className="pointer-events-none absolute inset-0 rounded-2xl"
        style={{ boxShadow: "inset 0 0 60px rgba(110,231,183,0.04)" }}
      />
    </div>
  );
}