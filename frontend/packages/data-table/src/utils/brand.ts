import type { CSSProperties } from "react";

/**
 * Builds CSS custom properties derived from a brand color.
 * Used to style primary buttons, filter active state, and other accents.
 */
export function buildBrandStyle(brandColor: string): CSSProperties {
  return {
    ["--dt-brand" as any]: brandColor,
    ["--dt-brand-hover" as any]: `color-mix(in srgb, ${brandColor} 85%, #000)`,
    ["--dt-brand-soft" as any]: `color-mix(in srgb, ${brandColor} 10%, #fff)`,
    ["--dt-brand-softer" as any]: `color-mix(in srgb, ${brandColor} 6%, #fff)`,
    ["--dt-brand-border" as any]: `color-mix(in srgb, ${brandColor} 35%, #fff)`,
  } as CSSProperties;
}
