// @vitest-environment jsdom
import { describe, it, expect } from "vitest";
import { createApp, h } from "vue";
import BaseModal from "@/components/prompts/BaseModal.vue";

describe("BaseModal ESC keydown lifecycle", () => {
  it("removes its window keydown listener on unmount so ESC is not swallowed", () => {
    const closed: string[] = [];
    const container = document.createElement("div");
    document.body.appendChild(container);

    const app = createApp({
      render: () => h(BaseModal, { onClosed: () => closed.push("closed") }),
    });
    app.mount(container);

    let sentinelFired = 0;
    const sentinel = () => sentinelFired++;
    window.addEventListener("keydown", sentinel);

    // While the modal is open, ESC closes it and (intentionally) stops other
    // window handlers from running.
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    expect(closed).toEqual(["closed"]);
    expect(sentinelFired).toBe(0);

    // After the modal closes/unmounts, the listener must be gone: ESC should
    // reach other window handlers (e.g. the preview's close handler) and must
    // not keep emitting "closed".
    app.unmount();
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));

    expect(closed).toEqual(["closed"]);
    expect(sentinelFired).toBe(2);

    window.removeEventListener("keydown", sentinel);
    container.remove();
  });
});
