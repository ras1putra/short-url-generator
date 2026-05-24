"use client";

import React, { useState, useRef } from "react";
import Cropper from "react-easy-crop";
import { Upload, X, FileVideo, Image as ImageIcon, Loader2, CropIcon, SlidersHorizontal } from "lucide-react";
import { useMediaCrop } from "@/hooks/useMediaCrop";
import {
  MEDIA_ACCEPT_TYPES,
} from "@/lib/constants";
import { isVideoUrl, isGifUrl } from "@/lib/media";

interface MediaUploaderProps {
  value: string;
  onChange: (url: string) => void;
  targetRatio?: number;
  recommendedResolution?: string;
  className?: string;
}

export function MediaUploader({ value, onChange, targetRatio, recommendedResolution, className = "" }: MediaUploaderProps) {
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const {
    processFile,
    showCrop,
    cropFileSrc,
    crop,
    zoom,
    onCropComplete,
    setCrop,
    setZoom,
    handleCropConfirm,
    handleCropCancel,
    isUploading,
    progress,
  } = useMediaCrop(onChange, targetRatio);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      processFile(e.target.files[0]);
    }
    e.target.value = "";
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      processFile(e.dataTransfer.files[0]);
    }
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = () => {
    setIsDragging(false);
  };

  const triggerSelect = () => {
    fileInputRef.current?.click();
  };

  return (
    <div className={`relative w-full ${className}`}>
      {value ? (
        <div className="relative rounded-2xl border border-white/[0.08] bg-white/[0.02] p-3 flex flex-col items-center gap-3 overflow-hidden group">
          {isVideoUrl(value) ? (
            <div className="relative rounded-lg overflow-hidden bg-black/45 w-full flex justify-center max-h-48">
              <video
                src={value}
                muted
                playsInline
                preload="metadata"
                className="max-h-48 object-contain w-full rounded-lg pointer-events-none"
              />
              <div className="absolute top-2 left-2 px-2 py-0.5 rounded bg-black/60 backdrop-blur-md border border-white/10 flex items-center gap-1.5 text-[10px] font-mono text-cyan-400">
                <FileVideo size={10} />
                VIDEO AD
              </div>
            </div>
          ) : isGifUrl(value) ? (
            <div className="relative rounded-lg overflow-hidden bg-black/45 w-full flex items-center justify-center max-h-48">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={value}
                alt="Ad Creative Preview"
                className="max-h-48 w-full object-contain group-hover:scale-[1.02] transition-transform duration-500"
              />
              <div className="absolute top-2 left-2 px-2 py-0.5 rounded bg-black/60 backdrop-blur-md border border-white/10 flex items-center gap-1.5 text-[10px] font-mono text-cyan-400">
                <ImageIcon size={10} />
                GIF AD
              </div>
            </div>
          ) : (
            <div className="relative rounded-lg overflow-hidden bg-black/45 w-full flex items-center justify-center max-h-48">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={value}
                alt="Ad Creative Preview"
                className="max-h-48 w-full object-contain group-hover:scale-[1.02] transition-transform duration-500"
              />
              <div className="absolute top-2 left-2 px-2 py-0.5 rounded bg-black/60 backdrop-blur-md border border-white/10 flex items-center gap-1.5 text-[10px] font-mono text-cyan-400">
                <ImageIcon size={10} />
                IMAGE AD
              </div>
            </div>
          )}

          <div className="flex items-center justify-between w-full px-2 py-1">
            <span className="text-[11px] font-mono text-white/30 truncate max-w-[80%]">
              {value}
            </span>
            <button
              type="button"
              onClick={() => onChange("")}
              className="p-1 rounded-lg bg-red-500/10 text-red-400 border border-red-500/25 hover:bg-red-500/20 hover:text-white transition-all cursor-pointer"
              title="Remove Creative"
            >
              <X size={14} />
            </button>
          </div>
        </div>
      ) : (
        <div
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
          onClick={triggerSelect}
          className={`flex flex-col items-center justify-center border-2 border-dashed rounded-2xl p-8 text-center cursor-pointer transition-all duration-300 min-h-[160px] ${isDragging
              ? "border-[#22D3EE] bg-[#22D3EE]/[0.03] scale-[0.99] shadow-[0_0_20px_rgba(34,211,238,0.05)]"
              : "border-white/10 bg-white/[0.01] hover:border-white/20 hover:bg-white/[0.02]"
            }`}
        >
          <input
            type="file"
            ref={fileInputRef}
            onChange={handleFileChange}
            accept={`${MEDIA_ACCEPT_TYPES.images},${MEDIA_ACCEPT_TYPES.videos}`}
            className="hidden"
          />

          {isUploading ? (
            <div className="flex flex-col items-center gap-3.5 w-full max-w-[80%]">
              <Loader2 className="animate-spin text-[#22D3EE]" size={28} />
              <div className="w-full bg-white/5 rounded-full h-1.5 overflow-hidden">
                <div
                  className="bg-gradient-to-r from-cyan-400 to-[#22D3EE] h-1.5 rounded-full transition-all duration-300"
                  style={{ width: `${progress}%` }}
                />
              </div>
              <span className="text-xs font-mono text-white/50">
                Uploading creative... {progress}%
              </span>
            </div>
          ) : (
            <div className="flex flex-col items-center gap-2">
              <div className="p-3 rounded-full bg-white/[0.03] border border-white/[0.06] text-white/40 group-hover:text-white/60 transition-colors">
                <Upload size={20} />
              </div>
              <div>
                <p className="text-xs font-bold text-white/80">
                  Drag & drop ad creative or <span className="text-[#22D3EE] underline hover:text-[#67E8F9]">browse</span>
                </p>
                <p className="text-xs text-white/30 mt-1 font-mono-dm leading-relaxed">
                  Images/GIFs (max 5MB) • MP4/WEBM Videos (max 50MB)
                  {recommendedResolution && (
                    <> &mdash; Recommended <span className="text-[#22D3EE]">{recommendedResolution}</span></>
                  )}
                </p>
              </div>
            </div>
          )}
        </div>
      )}

      {showCrop && cropFileSrc && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-4">
          <div className="relative w-full max-w-2xl rounded-2xl border border-white/[0.08] bg-[#0A0A0A] p-6">
            <h3 className="text-lg font-bold text-white/90 mb-4 flex items-center gap-2">
              <CropIcon size={18} className="text-[#22D3EE]" />
              Crop Ad Creative
              {recommendedResolution && (
                <span className="text-xs font-mono text-white/40 ml-auto">
                  Target: {recommendedResolution}
                </span>
              )}
            </h3>

            <div className="relative w-full h-[400px] rounded-xl overflow-hidden bg-black/60">
              <Cropper
                image={cropFileSrc}
                crop={crop}
                zoom={zoom}
                aspect={targetRatio}
                onCropChange={setCrop}
                onZoomChange={setZoom}
                onCropComplete={onCropComplete}
              />
            </div>

            <div className="flex items-center gap-3 mt-4 px-1">
              <SlidersHorizontal size={14} className="text-white/40 shrink-0" />
              <input
                type="range"
                min={1}
                max={3}
                step={0.05}
                value={zoom}
                onChange={(e) => setZoom(Number(e.target.value))}
                className="flex-1 accent-[#22D3EE] h-1 rounded-full appearance-none bg-white/10 cursor-pointer"
              />
              <span className="text-xs font-mono text-white/40 w-8 text-right">{zoom.toFixed(2)}x</span>
            </div>

            <div className="flex justify-end gap-3 mt-6">
              <button
                type="button"
                onClick={handleCropCancel}
                disabled={isUploading}
                className="px-5 py-2.5 text-sm rounded-xl border border-white/10 text-white/60 hover:text-white/80 hover:bg-white/[0.03] transition-all cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={handleCropConfirm}
                disabled={isUploading}
                className="btn-primary flex items-center gap-2 px-5 py-2.5 text-sm tracking-wider uppercase cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isUploading ? <Loader2 size={14} className="animate-spin" /> : <CropIcon size={14} />}
                {isUploading ? "Processing..." : "Apply Crop"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
