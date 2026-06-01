interface Turnstile {
  getResponse(widgetId?: string): string | undefined;
  render(element: string | HTMLElement, options: Record<string, unknown>): string;
}

interface Window {
  turnstile?: Turnstile;
}
