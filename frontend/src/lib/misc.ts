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
