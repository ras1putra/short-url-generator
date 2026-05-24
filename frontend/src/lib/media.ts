export function isVideoUrl(url: string): boolean {
  return /\.(mp4|webm|ogg)([?#]|$)/i.test(url);
}

export function isGifUrl(url: string): boolean {
  return /\.gif([?#]|$)/i.test(url);
}
