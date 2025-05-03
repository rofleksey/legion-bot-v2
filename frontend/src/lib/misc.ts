export function errorToString(e: any) {
  console.error(e);

  if (e?.response?.data) {
    return e.response.data;
  }

  if (e?.message) {
    return e.message;
  }

  if (e?.toString?.()) {
    return e.toString?.();
  }

  return "error"
}

export function ymReachGoal(goal: string) {
  try {
    // @ts-ignore
    window.ym(101556361, 'reachGoal', goal);
  } catch (e) {
    console.error(e);
  }
}

export function formatDuration(nanoseconds: number): string {
  // Treat negative numbers as 0
  const safeNanoseconds = Math.max(0, nanoseconds);

  // Convert nanoseconds to seconds
  const totalSeconds = Math.floor(safeNanoseconds / 1e9);

  // Calculate hours, minutes, and seconds
  const hours = Math.floor(totalSeconds / 3600);
  const remainingSeconds = totalSeconds % 3600;
  const minutes = Math.floor(remainingSeconds / 60);
  const seconds = remainingSeconds % 60;

  // Format each component to two digits
  const formattedHours = hours.toString().padStart(2, '0');
  const formattedMinutes = minutes.toString().padStart(2, '0');
  const formattedSeconds = seconds.toString().padStart(2, '0');

  return `${formattedHours}:${formattedMinutes}:${formattedSeconds}`;
}
