export function roundCoord4(value: number): number {
  return Math.round(value * 1e4) / 1e4;
}

export function formatDistanceKm(distanceKm: number | null | undefined): string {
  if (distanceKm === null || distanceKm === undefined || Number.isNaN(distanceKm)) {
    return '—';
  }
  if (distanceKm < 1) {
    const meters = Math.round(distanceKm * 1000);
    return `${meters} m`;
  }
  const rounded = Math.round(distanceKm * 10) / 10;
  return `${rounded} km`;
}

