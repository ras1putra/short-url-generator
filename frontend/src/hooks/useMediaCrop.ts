"use client";

import { useState, useCallback } from "react";
import { type Area } from "react-easy-crop";
import { toast } from "sonner";
import { RATIO_TOLERANCE, MAX_IMAGE_SIZE, MAX_VIDEO_SIZE } from "@/lib/constants";
import { useMediaUpload } from "@/hooks/useMediaUpload";
import { parseGIF, decompressFrames } from "gifuct-js";

interface GIFOptions {
  width: number;
  height: number;
  workers: number;
  quality: number;
  repeat: number;
  workerScript: string;
  transparent: number | null;
}

interface GIFInstance {
  addFrame(
    image: ImageData | CanvasRenderingContext2D | HTMLImageElement,
    options?: { delay?: number; copy?: boolean },
  ): void;
  on(event: "finished", callback: (blob: Blob) => void): void;
  on(event: "abort", callback: () => void): void;
  on(event: string, callback: (args: unknown) => void): void;
  render(): void;
}

interface GIFConstructor {
  new (options: GIFOptions): GIFInstance;
}

function createImage(src: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const img = document.createElement("img");
    img.crossOrigin = "anonymous";
    img.onload = () => resolve(img);
    img.onerror = () => reject(new Error("Failed to load image"));
    img.src = src;
  });
}

async function createCroppedBlob(imageSrc: string, croppedAreaPixels: Area, mimeType?: string): Promise<Blob> {
  const image = await createImage(imageSrc);
  const canvas = document.createElement("canvas");
  canvas.width = croppedAreaPixels.width;
  canvas.height = croppedAreaPixels.height;
  const ctx = canvas.getContext("2d")!;
  ctx.drawImage(
    image,
    croppedAreaPixels.x,
    croppedAreaPixels.y,
    croppedAreaPixels.width,
    croppedAreaPixels.height,
    0,
    0,
    croppedAreaPixels.width,
    croppedAreaPixels.height,
  );
  const format = mimeType === "image/jpeg" ? "image/jpeg" : "image/png";
  return new Promise((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (blob) resolve(blob);
      else reject(new Error("Canvas toBlob failed"));
    }, format);
  });
}

async function cropGif(file: File, croppedAreaPixels: Area): Promise<Blob> {
  const arrayBuffer = await file.arrayBuffer();
  const gifParsed = parseGIF(arrayBuffer);
  const frames = decompressFrames(gifParsed, true);

  if (!frames.length) throw new Error("No frames found in GIF");

  const fullWidth = gifParsed.lsd.width;
  const fullHeight = gifParsed.lsd.height;
  const { x, y, width: cropW, height: cropH } = croppedAreaPixels;

  const GIF = (await import("gif.js")).default as unknown as GIFConstructor;

  return new Promise<Blob>((resolve, reject) => {
    const gif = new GIF({
      width: Math.round(cropW),
      height: Math.round(cropH),
      workers: 2,
      quality: 10,
      repeat: 0,
      workerScript: "/gif.worker.js",
      transparent: null,
    });

    gif.on("finished", (blob: Blob) => resolve(blob));
    gif.on("abort", () => reject(new Error("GIF encoding aborted")));

    const fullCanvas = document.createElement("canvas");
    fullCanvas.width = fullWidth;
    fullCanvas.height = fullHeight;
    const fullCtx = fullCanvas.getContext("2d")!;

    const cropCanvas = document.createElement("canvas");
    cropCanvas.width = Math.round(cropW);
    cropCanvas.height = Math.round(cropH);
    const cropCtx = cropCanvas.getContext("2d")!;

    let previousCanvasState: ImageData | null = null;

    for (let i = 0; i < frames.length; i++) {
      const frame = frames[i];

      if (i > 0) {
        const prev = frames[i - 1];
        if (prev.disposalType === 2) {
          fullCtx.clearRect(prev.dims.left, prev.dims.top, prev.dims.width, prev.dims.height);
        } else if (prev.disposalType === 3 && previousCanvasState) {
          fullCtx.putImageData(previousCanvasState, 0, 0);
        }
      }

      if (frame.disposalType === 3) {
        previousCanvasState = fullCtx.getImageData(0, 0, fullWidth, fullHeight);
      } else {
        previousCanvasState = null;
      }

      const imageData = fullCtx.createImageData(frame.dims.width, frame.dims.height);
      imageData.data.set(frame.patch);
      fullCtx.putImageData(imageData, frame.dims.left, frame.dims.top);

      cropCtx.clearRect(0, 0, cropCanvas.width, cropCanvas.height);
      cropCtx.drawImage(
        fullCanvas,
        Math.round(x), Math.round(y), Math.round(cropW), Math.round(cropH),
        0, 0, cropCanvas.width, cropCanvas.height,
      );

      gif.addFrame(cropCtx.getImageData(0, 0, cropCanvas.width, cropCanvas.height), {
        delay: frame.delay || 100,
        copy: true,
      });
    }

    gif.render();
  });
}

async function captureVideoFrame(file: File): Promise<string> {
  const video = document.createElement("video");
  video.preload = "metadata";
  video.muted = true;
  video.playsInline = true;
  const src = URL.createObjectURL(file);
  video.src = src;
  await video.load();

  return new Promise((resolve, reject) => {
    video.onloadeddata = () => { video.currentTime = 0.5; };
    video.onseeked = () => {
      const canvas = document.createElement("canvas");
      canvas.width = video.videoWidth;
      canvas.height = video.videoHeight;
      const ctx = canvas.getContext("2d")!;
      ctx.drawImage(video, 0, 0);
      URL.revokeObjectURL(src);
      resolve(canvas.toDataURL("image/jpeg"));
    };
    video.onerror = () => { URL.revokeObjectURL(src); reject(new Error("Failed to load video")); };
  });
}

function getMediaDimensions(file: File): Promise<{ width: number; height: number }> {
  return new Promise((resolve, reject) => {
    if (file.type.startsWith("video/")) {
      const video = document.createElement("video");
      video.preload = "metadata";
      video.onloadedmetadata = () => {
        URL.revokeObjectURL(video.src);
        resolve({ width: video.videoWidth, height: video.videoHeight });
      };
      video.onerror = () => {
        URL.revokeObjectURL(video.src);
        reject(new Error("Failed to load video metadata"));
      };
      video.src = URL.createObjectURL(file);
    } else {
      const img = document.createElement("img");
      img.onload = () => {
        URL.revokeObjectURL(img.src);
        resolve({ width: img.naturalWidth, height: img.naturalHeight });
      };
      img.onerror = () => {
        URL.revokeObjectURL(img.src);
        reject(new Error("Failed to load image"));
      };
      img.src = URL.createObjectURL(file);
    }
  });
}

export function useMediaCrop(onUploadSuccess: (url: string) => void, targetRatio?: number) {
  const { upload, uploadVideoCrop, isUploading, progress } = useMediaUpload();

  const [showCrop, setShowCrop] = useState(false);
  const [cropFileSrc, setCropFileSrc] = useState("");
  const [cropOrigFile, setCropOrigFile] = useState<File | null>(null);
  const [crop, setCrop] = useState({ x: 0, y: 0 });
  const [zoom, setZoom] = useState(1);
  const [croppedAreaPixels, setCroppedAreaPixels] = useState<Area | null>(null);

  const onCropComplete = useCallback((_: Area, croppedPixels: Area) => {
    setCroppedAreaPixels(croppedPixels);
  }, []);

  const handleUpload = useCallback(async (file: File) => {
    try {
      const result = await upload(file, { targetRatio });
      if (result) {
        onUploadSuccess(result.url);
        toast.success("Media uploaded successfully!");
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to upload media.";
      toast.error(msg);
    }
  }, [upload, targetRatio, onUploadSuccess]);

  const processFile = useCallback(async (file: File) => {
    const isVideo = file.type.startsWith("video/");
    const sizeCap = isVideo ? MAX_VIDEO_SIZE : MAX_IMAGE_SIZE;
    if (file.size > sizeCap) {
      toast.error(`File size exceeds limit. Maximum allowed size is ${Math.round(sizeCap / (1024 * 1024))}MB.`);
      return;
    }

    if (!targetRatio) {
      await handleUpload(file);
      return;
    }

    const mediaDims = await getMediaDimensions(file);
    const actualRatio = mediaDims.width / mediaDims.height;

    if (Math.abs(actualRatio - targetRatio) / targetRatio > RATIO_TOLERANCE) {
      if (file.type.startsWith("video/")) {
        const frameSrc = await captureVideoFrame(file);
        setCropFileSrc(frameSrc);
      } else {
        const src = URL.createObjectURL(file);
        setCropFileSrc(src);
      }
      setCropOrigFile(file);
      setCrop({ x: 0, y: 0 });
      setZoom(1);
      setCroppedAreaPixels(null);
      setShowCrop(true);
      return;
    }

    await handleUpload(file);
  }, [targetRatio, handleUpload]);

  const handleCropConfirm = useCallback(async () => {
    if (!cropFileSrc || !croppedAreaPixels || !cropOrigFile) return;

    try {
      const isVideo = cropOrigFile.type.startsWith("video/");
      const isGif = !isVideo && (cropOrigFile.type === "image/gif" || /\.gif$/i.test(cropOrigFile.name));

      if (isVideo) {
        const { x, y, width: w, height: h } = croppedAreaPixels;
        const result = await uploadVideoCrop(cropOrigFile, Math.round(x), Math.round(y), Math.round(w), Math.round(h));
        if (result) {
          URL.revokeObjectURL(cropFileSrc);
          setShowCrop(false);
          setCropFileSrc("");
          setCropOrigFile(null);
          onUploadSuccess(result.url);
          toast.success("Video cropped and uploaded successfully!");
        }
        return;
      }

      let croppedFile: File;

      if (isGif) {
        const blob = await cropGif(cropOrigFile, croppedAreaPixels);
        croppedFile = new File([blob], cropOrigFile.name, { type: "image/gif" });
      } else {
        const origType = cropOrigFile.type;
        const format = origType === "image/jpeg" ? "jpeg" : "png";
        const mimeType = origType === "image/jpeg" ? "image/jpeg" : "image/png";
        const blob = await createCroppedBlob(cropFileSrc, croppedAreaPixels, origType);
        croppedFile = new File([blob], cropOrigFile.name.replace(/\.[^.]+$/, `.${format}`), { type: mimeType });
      }

      URL.revokeObjectURL(cropFileSrc);
      setShowCrop(false);
      setCropFileSrc("");
      setCropOrigFile(null);

      await handleUpload(croppedFile);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to crop media.";
      toast.error(msg);
    }
  }, [cropFileSrc, croppedAreaPixels, cropOrigFile, handleUpload, uploadVideoCrop, onUploadSuccess]);

  const handleCropCancel = useCallback(() => {
    if (cropFileSrc) URL.revokeObjectURL(cropFileSrc);
    setShowCrop(false);
    setCropFileSrc("");
    setCropOrigFile(null);
  }, [cropFileSrc]);

  return {
    processFile,
    showCrop,
    cropFileSrc,
    crop,
    zoom,
    croppedAreaPixels,
    onCropComplete,
    setCrop,
    setZoom,
    handleCropConfirm,
    handleCropCancel,
    isUploading,
    progress,
  };
}
