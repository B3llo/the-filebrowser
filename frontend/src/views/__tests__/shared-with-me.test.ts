// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { createApp, h, nextTick } from "vue";
import { createPinia, setActivePinia } from "pinia";

const push = vi.fn();
vi.mock("vue-router", async (importOriginal) => {
  const mod = await importOriginal<typeof import("vue-router")>();
  return { ...mod, useRouter: () => ({ push }) };
});

vi.mock("@/api", () => ({
  grants: { sharedWithMe: vi.fn() },
}));

// @/api/utils pulls the real router (via utils/auth); stub StatusError so the
// view test never loads Files.vue -> DetailsPanel -> pdfjs (needs DOMMatrix).
vi.mock("@/api/utils", () => ({
  StatusError: class StatusError extends Error {
    status?: number;
    constructor(message?: string, status?: number) {
      super(message);
      this.status = status;
    }
  },
}));

vi.mock("@/views/Errors.vue", () => ({
  default: { template: "<div class='errors-stub' />" },
}));

vi.mock("@/components/FbIcon.vue", () => ({
  default: { template: "<span class='fbicon-stub' />" },
}));

import i18n from "@/i18n";

function stubWindow() {
  vi.stubGlobal("window", {
    FileBrowser: { Name: "File Browser", BaseURL: "" },
    location: { origin: "http://localhost" },
  });
}

async function mount(grantsData: unknown[]) {
  stubWindow();
  vi.resetModules();
  const { default: View } = await import("@/views/SharedWithMe.vue");
  const { grants } = await import("@/api");
  vi.mocked(grants.sharedWithMe).mockResolvedValue(grantsData as any);

  setActivePinia(createPinia());
  const container = document.createElement("div");
  document.body.appendChild(container);
  const app = createApp({ render: () => h(View) });
  app.use(createPinia());
  app.use(i18n);
  app.provide("$showError", vi.fn());
  app.mount(container);
  await nextTick();
  await new Promise((r) => setTimeout(r, 0));
  await nextTick();
  return { app, container };
}

describe("SharedWithMe view", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  it("lists grants and navigates into Files on click", async () => {
    const { app, container } = await mount([
      {
        id: 1,
        path: "/docs",
        ownerID: 1,
        granteeID: 2,
        role: "viewer",
        ownerUsername: "alice",
      },
    ]);

    expect(container.textContent).toContain("docs");
    expect(container.textContent).toContain("alice");

    (container.querySelector(".fb-grant-row") as HTMLElement).click();
    expect(push).toHaveBeenCalledWith({
      path: expect.stringContaining("/docs"),
    });
    app.unmount();
    container.remove();
  });

  it("shows the empty state when nothing is shared", async () => {
    const { app, container } = await mount([]);

    expect(container.querySelector(".fb-settings-empty")).not.toBeNull();
    app.unmount();
    container.remove();
  });
});
