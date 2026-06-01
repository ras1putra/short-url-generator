interface Turnstile {
  getResponse(widgetId?: string): string | undefined;
  render(element: string | HTMLElement, options: Record<string, unknown>): string;
  remove(widgetId: string): void;
}

interface Window {
  turnstile?: Turnstile;
}

