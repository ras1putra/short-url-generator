declare module "gif.js" {
  const GIF: {
    new (options: {
      width: number;
      height: number;
      workers: number;
      quality: number;
      repeat: number;
      workerScript: string;
      transparent: number | null;
      background?: string;
      dither?: boolean | string;
      debug?: boolean;
    }): {
      addFrame(
        image: ImageData | CanvasRenderingContext2D | HTMLImageElement,
        options?: { delay?: number; copy?: boolean },
      ): void;
      on(event: "finished", callback: (blob: Blob) => void): void;
      on(event: "abort", callback: () => void): void;
      on(event: string, callback: (args: unknown) => void): void;
      render(): void;
    };
  };
  export default GIF;
}
