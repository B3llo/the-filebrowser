import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const mobileCss = readFileSync(resolve(__dirname, "../mobile.css"), "utf8");

const normalizedCss = mobileCss.replace(/\s+/g, " ");

describe("mobile file listing styles", () => {
  it("drops the Size column consistently (rows + header) and realigns the grid at ≤736px", () => {
    // Size is hidden for both rows and header so the two stay in sync…
    expect(normalizedCss).toContain(
      "#listing.list .item .size { display: none;"
    );
    // …and the grid is narrowed to a 2-column layout so the Modified cell
    // doesn't slip into the Size track.
    expect(normalizedCss).toContain("grid-template-columns: 1fr 168px;");
  });

  it("drops the Modified column too at ≤450px, leaving only Name", () => {
    expect(normalizedCss).toContain(
      "#listing.list .item .modified { display: none;"
    );
    expect(normalizedCss).toContain("grid-template-columns: 1fr;");
  });

  it("does not use legacy percentage-width column hacks", () => {
    // Old layout set explicit widths on header cells; the grid now owns sizing.
    expect(normalizedCss).not.toMatch(
      /\.item\.header \.(name|size|modified) \{[^}]*width:/
    );
  });
});
