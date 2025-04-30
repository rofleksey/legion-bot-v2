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
