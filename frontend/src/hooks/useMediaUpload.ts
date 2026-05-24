"use client";

import { useState, useCallback } from "react";
import { api } from "@/lib/api";
import {
  API_MEDIA_UPLOAD,
  API_MEDIA_CROP_VIDEO,
  MAX_IMAGE_SIZE,
  MAX_VIDEO_SIZE,
} from "@/lib/constants";

interface UploadResult {
  url: string;
}

interface UploadOptions {
  targetRatio?: number;
}

export function useMediaUpload() {
  const [isUploading, setIsUploading] = useState(false);
  const [progress, setProgress] = useState(0);

  const upload = useCallback(
    async (file: File, options?: UploadOptions): Promise<UploadResult | null> => {
      const isImage = file.type.startsWith("image/");
      const isVideo = file.type.startsWith("video/");

      if (!isImage && !isVideo) {
        throw new Error("Unsupported file type. Please upload PNG, JPG, WEBP, GIF, or MP4/WEBM video.");
      }

      const sizeCap = isVideo ? MAX_VIDEO_SIZE : MAX_IMAGE_SIZE;
      if (file.size > sizeCap) {
        throw new Error(`File size exceeds limit. Maximum allowed size is ${Math.round(sizeCap / (1024 * 1024))}MB.`);
      }

      setIsUploading(true);
      setProgress(0);

      const formData = new FormData();
      formData.append("file", file);

      if (options?.targetRatio) {
        formData.append("target_ratio", options.targetRatio.toString());
      }

      try {
        const response = await api.post<{ data: { url: string } }>(
          API_MEDIA_UPLOAD,
          formData,
          {
            headers: { "Content-Type": "multipart/form-data" },
            onUploadProgress: (progressEvent) => {
              const total = progressEvent.total || file.size;
              setProgress(Math.round((progressEvent.loaded * 100) / total));
            },
          }
        );

        if (response.data?.data?.url) {
          return { url: response.data.data.url };
        }

        throw new Error("Upload failed - no URL returned");
      } finally {
        setIsUploading(false);
        setProgress(0);
      }
    },
    []
  );

  const uploadVideoCrop = useCallback(
    async (file: File, x: number, y: number, w: number, h: number): Promise<UploadResult | null> => {
      if (file.size > MAX_VIDEO_SIZE) {
        throw new Error(`File size exceeds limit. Maximum allowed size is ${Math.round(MAX_VIDEO_SIZE / (1024 * 1024))}MB.`);
      }

      setIsUploading(true);
      setProgress(0);

      const formData = new FormData();
      formData.append("file", file);
      formData.append("x", String(x));
      formData.append("y", String(y));
      formData.append("w", String(w));
      formData.append("h", String(h));

      try {
        const response = await api.post<{ data: { url: string } }>(
          API_MEDIA_CROP_VIDEO,
          formData,
          {
            headers: { "Content-Type": "multipart/form-data" },
            onUploadProgress: (progressEvent) => {
              const total = progressEvent.total || file.size;
              setProgress(Math.round((progressEvent.loaded * 100) / total));
            },
          }
        );

        if (response.data?.data?.url) {
          return { url: response.data.data.url };
        }

        throw new Error("Video crop failed - no URL returned");
      } finally {
        setIsUploading(false);
        setProgress(0);
      }
    },
    []
  );

  return { upload, uploadVideoCrop, isUploading, progress };
}
