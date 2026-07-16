export const OFFICE_THUMB_MAX_SIZE = 5 * 1024 * 1024;

export function officeThumbKind(
  name: string,
  size: number
): "docx" | "sheet" | false {
  if (size > OFFICE_THUMB_MAX_SIZE) return false;
  const lower = name.toLowerCase();
  if (lower.endsWith(".docx")) return "docx";
  if (lower.endsWith(".xlsx") || lower.endsWith(".ods")) return "sheet";
  return false;
}

export async function fetchDocxText(
  url: string,
  maxLines = 12
): Promise<string | null> {
  try {
    const res = await fetch(url);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const buf = await res.arrayBuffer();
    const mod: any = await import("mammoth");
    const mammoth = mod.default ?? mod;
    const result = await mammoth.extractRawText({ arrayBuffer: buf });
    const lines = result.value
      .split("\n")
      .filter((l: string) => l.trim())
      .slice(0, maxLines);
    return lines.join("\n");
  } catch (e) {
    console.warn("Failed to fetch docx thumb:", e);
    return null;
  }
}

export async function fetchSheetGrid(
  url: string,
  maxRows = 8,
  maxCols = 6
): Promise<string[][] | null> {
  try {
    const res = await fetch(url);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const buf = await res.arrayBuffer();
    const mod: any = await import("xlsx");
    const XLSX = mod.default ?? mod;
    const wb = XLSX.read(buf, { type: "array" });
    const rows: any[][] = XLSX.utils.sheet_to_json(
      wb.Sheets[wb.SheetNames[0]],
      { header: 1, defval: "", blankrows: false }
    );
    return rows
      .slice(0, maxRows)
      .map((r) => (r as any[]).slice(0, maxCols).map((c) => String(c ?? "")));
  } catch (e) {
    console.warn("Failed to fetch sheet thumb:", e);
    return null;
  }
}
