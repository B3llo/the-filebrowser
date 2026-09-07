// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { createApp, h, nextTick } from "vue";
import { createPinia, setActivePinia } from "pinia";

vi.mock("@/api", () => ({
  share: { list: vi.fn(), getShareURL: vi.fn() },
  grants: { list: vi.fn(), remove: vi.fn(), patch: vi.fn() },
  users: { getAll: vi.fn() },
}));

vi.mock("@/api/utils", () => ({
  StatusError: class StatusError extends Error {
    status?: number;
    constructor(message?: string, status?: number) {
      super(message);
      this.status = status;
    }
  },
}));

vi.mock("@/utils/clipboard", () => ({
  copy: vi.fn().mockResolvedValue(undefined),
}));

vi.mock("@/views/Errors.vue", () => ({
  default: { template: "<div class='errors-stub' />" },
}));

import i18n from "@/i18n";
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";

function stubWindow() {
  vi.stubGlobal("window", {
    FileBrowser: { Name: "File Browser", BaseURL: "" },
    location: { origin: "http://localhost" },
  });
}

async function mountAs(admin: boolean) {
  stubWindow();
  vi.resetModules();
  const { default: View } = await import("@/views/settings/Shares.vue");
  const { share, grants, users } = await import("@/api");
  vi.mocked(share.list).mockResolvedValue([]);
  vi.mocked(users.getAll).mockResolvedValue([
    { id: 1, username: "alice" },
    { id: 2, username: "bob" },
  ] as any);
  vi.mocked(grants.list).mockResolvedValue([
    {
      id: 1,
      path: "/docs",
      ownerID: 1,
      granteeID: 2,
      role: "viewer",
      expire: 0,
      ownerUsername: "alice",
      granteeUsername: "bob",
    },
  ] as any);

  const pinia = createPinia();
  setActivePinia(pinia);
  (useAuthStore() as any).user = { perm: { admin } };

  const container = document.createElement("div");
  document.body.appendChild(container);
  const app = createApp({ render: () => h(View) });
  app.use(pinia);
  app.use(i18n);
  app.provide("$showError", vi.fn());
  app.provide("$showSuccess", vi.fn());
  app.mount(container);
  await nextTick();
  await new Promise((r) => setTimeout(r, 0));
  await nextTick();
  return { app, container, api: { share, grants, users } };
}

describe("Shares admin grants tab", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  it("switches to access tab and revokes through confirm", async () => {
    const { app, container, api } = await mountAs(true);

    const tabs = [...container.querySelectorAll(".fb-share-tabs button")];
    expect(tabs).toHaveLength(2);
    (tabs[1] as HTMLElement).click();
    await nextTick();

    expect(container.textContent).toContain("/docs");
    expect(container.textContent).toContain("bob");

    const revokeBtn = container.querySelector(
      "tbody tr button[title]"
    ) as HTMLElement;
    revokeBtn.click();
    await nextTick();

    const layoutStore = useLayoutStore();
    const confirm = layoutStore.currentPrompt
      ?.confirm as unknown as () => Promise<void>;
    expect(layoutStore.currentPromptName).toBe("share-delete");
    await confirm();
    await nextTick();

    expect(api.grants.remove).toHaveBeenCalledWith(1);
    expect(container.textContent).not.toContain("/docs");
    app.unmount();
    container.remove();
  });

  it("changes role inline", async () => {
    const { app, container, api } = await mountAs(true);
    vi.mocked(api.grants.patch).mockResolvedValue({
      id: 1,
      role: "editor",
    } as any);

    (
      container.querySelectorAll(".fb-share-tabs button")[1] as HTMLElement
    ).click();
    await nextTick();

    const select = container.querySelector(
      "tbody tr select"
    ) as HTMLSelectElement;
    select.value = "editor";
    select.dispatchEvent(new Event("change"));
    await nextTick();
    await new Promise((r) => setTimeout(r, 0));

    expect(api.grants.patch).toHaveBeenCalledWith(1, { role: "editor" });
    app.unmount();
    container.remove();
  });
});
