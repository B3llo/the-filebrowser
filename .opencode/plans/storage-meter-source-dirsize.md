# Storage Meter: Show Source Folder Size + Disk Usage

## Problem

The sidebar storage meter shows `{used} of {total} used` — this is the **OS-level disk partition usage** (`gopsutil/disk`), not the actual size of the current data source folder. When a source is a 10 GB folder on a 11 GB disk, the user sees "90% used" and thinks it's the folder, when it's actually the whole disk.

## Goal

Show **two pieces of information** in the sidebar storage meter, compact layout:

```
┌─────────────────────────────┐
│ Storage                     │
│ 📁 Projetos    10.2 GB      │  ← source folder size (recursive dirSize)
│ 🏠 Disk        45% · 1 TB   │  ← disk partition usage (existing)
└─────────────────────────────┘
```

## Files to Modify

| File | Changes |
|---|---|
| `frontend/src/components/Sidebar.vue` | Template, script (setup, computed, methods, watchers), CSS |
| `frontend/src/i18n/en.json` | Add `"disk": "Disk"` key |

## Detailed Specification

### 1. Template (lines 127-149)

Replace the existing storage meter block with:

```vue
<!-- Storage meter -->
<div v-if="isLoggedIn && !disableUsedPercentage" class="fb-sidebar-storage">
  <div class="fb-sidebar-storage-header">
    <span class="fb-sidebar-storage-label">{{
      $t("sidebar.storage")
    }}</span>
  </div>
  <div class="fb-sidebar-source-row">
    <span class="fb-sidebar-row-label">
      <fb-icon name="folder" size="14px" />
      {{ activeSourceName || $t("sidebar.myFiles") }}
    </span>
    <span class="fb-sidebar-row-value">{{ sourceSizeFormatted }}</span>
  </div>
  <div class="fb-sidebar-disk-row">
    <span class="fb-sidebar-row-label">
      <fb-icon name="home" size="14px" />
      {{ $t("sidebar.disk") }}
    </span>
    <span class="fb-sidebar-row-value">
      {{ usage.usedPercentage }}% · {{ usage.total }}
    </span>
  </div>
</div>
```

### 2. Script — Setup

In `export default { setup() { ... } }`, add reactive refs alongside the existing `usage`:

```js
import { reactive, ref } from "vue";

setup() {
  const usage = reactive(USAGE_DEFAULT);
  const sourceDirSize = ref(null);
  const loadingSourceDirSize = ref(false);
  return { usage, sourceDirSize, loadingSourceDirSize, usageAbortController: new AbortController() };
},
```

### 3. Script — Computed

Add `sourceSizeFormatted` after `isAdmin`:

```js
sourceSizeFormatted() {
  if (this.sourceDirSize === null) return "—";
  if (this.loadingSourceDirSize) return "…";
  return prettyBytes(this.sourceDirSize, { binary: true });
},
```

### 4. Script — Methods

Add `fetchSourceDirSize()` after `fetchUsage()`:

```js
async fetchSourceDirSize() {
  const source = this.active;
  if (!source || source.path === undefined || source.path === "") {
    this.sourceDirSize = null;
    this.loadingSourceDirSize = false;
    return;
  }
  this.loadingSourceDirSize = true;
  try {
    const res = await api.dirSize(`/files/${source.id}/`);
    this.sourceDirSize = res.size;
  } catch (e) {
    console.warn("failed to fetch source dir size", e);
    this.sourceDirSize = null;
  } finally {
    this.loadingSourceDirSize = false;
  }
},
```

### 5. Script — Watchers

Replace the `$route` watcher to **only** call `fetchUsage`, and add a separate watcher on `active` for `fetchSourceDirSize`:

```js
watch: {
  $route: {
    handler() {
      this.fetchUsage();
      this.closeHovers();
    },
    immediate: true,
  },
  "active": {
    handler() {
      this.fetchSourceDirSize();
    },
    immediate: false,
  },
},
```

### 6. CSS

Replace the existing `.fb-sidebar-storage-detail` rule and add new row styles:

```css
/* Keep existing */
.fb-sidebar-storage-detail {
  margin-top: 6px;
  font-size: 11.5px;
  color: var(--dim);
}

/* New rows */
.fb-sidebar-source-row,
.fb-sidebar-disk-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
  font-size: 12.5px;
}

.fb-sidebar-source-row {
  margin-top: 2px;
}

.fb-sidebar-disk-row {
  border-top: 1px solid var(--border);
  margin-top: 4px;
  padding-top: 6px;
}

.fb-sidebar-row-label {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--text);
  font-weight: 500;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.fb-sidebar-row-value {
  color: var(--dim);
  font-weight: 500;
  flex-shrink: 0;
  margin-left: 8px;
}
```

### 7. i18n

In `frontend/src/i18n/en.json`, add `"disk"` inside the `"sidebar"` block:

```json
"disk": "Disk",
```

## Behavior Matrix

| Scenario | Source row | Disk row |
|---|---|---|
| Source with concrete path | `📁 Name    10.2 GB` | `🏠 Disk    45% · 1 TB` |
| Implicit source (id=0, "My Files", no path) | `📁 My Files    —` | `🏠 Disk    45% · 1 TB` |
| Loading dirSize | `📁 Name    …` | `🏠 Disk    45% · 1 TB` |
| Error fetching dirSize | `📁 Name    —` | `🏠 Disk    45% · 1 TB` |
| Sidebar collapsed | hidden (existing CSS) | hidden |

## Implementation Notes

- Use `api.dirSize()` from `@/api/files` — it already exists, calls `GET /api/resources/dirsize/{path}`
- The path `/files/{source.id}/` is passed to `removePrefix()` which strips the `/files/{id}` prefix — this works correctly for the implicit source too (just won't have a `path` field)
- No backend changes needed — `dirSize` already walks the directory recursively
- Performance: `dirSize` can be slow for large directories — the loading state (`"…"`) handles this
- The `home` icon already exists in the icon registry (`frontend/src/utils/icons.ts`) — no new icon needed
- To test: after making changes, run `cd frontend && pnpm build && cd .. && go build -o filebrowser . && rm -f filebrowser.db && ./filebrowser`