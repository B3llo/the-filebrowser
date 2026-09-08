<template>
  <div>
    <header-bar :showBreadcrumb="!isTrash" :base="filesBase">
      <template v-if="isTrash" #default>
        <div
          class="fb-toolbar-title fb-trash-breadcrumb"
          v-if="breadcrumb && breadcrumb.length > 0"
        >
          <button
            v-for="(seg, i) in breadcrumb"
            :key="i"
            class="fb-trash-breadcrumb-item"
            :class="{
              'fb-trash-breadcrumb-item--last': i === breadcrumb.length - 1,
            }"
            @click="emit('navigate-to', seg.url)"
          >
            <span v-if="i > 0" class="fb-trash-breadcrumb-sep">/</span>
            {{ seg.label }}
          </button>
        </div>
        <span class="fb-toolbar-title" v-else>{{ $t("sidebar.trash") }}</span>
        <div
          v-if="isTrash"
          class="fb-trash-view-toggle"
          role="group"
          :aria-label="$t('trash.treeView')"
        >
          <button
            class="fb-tbtn"
            :class="{ 'fb-tbtn--active': !props.flatView }"
            :title="$t('trash.treeView')"
            @click="emit('toggle-flat', false)"
          >
            {{ $t("trash.treeView") }}
          </button>
          <button
            class="fb-tbtn"
            :class="{ 'fb-tbtn--active': !!props.flatView }"
            :title="$t('trash.flatView')"
            @click="emit('toggle-flat', true)"
          >
            {{ $t("trash.flatView") }}
          </button>
        </div>
        <div
          v-if="sourceStore.sources.length > 1"
          class="fb-source-tabs"
          style="flex: 0 0 auto"
        >
          <button
            v-for="src in sourceStore.sources"
            :key="src.id"
            class="fb-tbtn"
            :class="{
              'fb-tbtn--active':
                String(src.id) ===
                (props.sourceId ?? String(sourceStore.activeId)),
            }"
            @click="emit('switch-source', String(src.id))"
          >
            {{ src.name }}
          </button>
        </div>
      </template>
      <template #actions>
        <!-- Inline search — 280px dropdown -->
        <Search ref="searchRef" />

        <!-- Mobile: icon-only search trigger -->
        <button
          class="fb-tbtn fb-search-btn"
          :aria-label="t('buttons.search')"
          @click="searchRef?.focus()"
        >
          <FbIcon name="search" size="18px" />
        </button>

        <!-- Shell toggle -->
        <action
          v-if="!isTrash && headerButtons.shell"
          icon="code"
          :label="t('buttons.shell')"
          @action="layoutStore.toggleShell"
        />

        <!-- 3-way view toggle segmented control -->
        <div class="fb-view-toggle" :title="t('buttons.switchView')">
          <button
            class="fb-tbtn"
            :class="{ 'fb-tbtn--active': currentViewMode === 'list' }"
            :aria-label="t('buttons.listView', 'List view')"
            @click="setView('list')"
          >
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              style="width: 17px; height: 17px"
              aria-hidden="true"
            >
              <path d="M8 6h13" />
              <path d="M8 12h13" />
              <path d="M8 18h13" />
              <path d="M3 6h.01" />
              <path d="M3 12h.01" />
              <path d="M3 18h.01" />
            </svg>
          </button>
          <button
            class="fb-tbtn"
            :class="{ 'fb-tbtn--active': currentViewMode === 'mosaic' }"
            :aria-label="t('buttons.gridView', 'Grid view')"
            @click="setView('mosaic')"
          >
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              style="width: 16px; height: 16px"
              aria-hidden="true"
            >
              <path d="M3 3h7v7H3z" />
              <path d="M14 3h7v7h-7z" />
              <path d="M14 14h7v7h-7z" />
              <path d="M3 14h7v7H3z" />
            </svg>
          </button>
        </div>

        <!-- Sort button — 38×38, bordered, opens dropdown -->
        <div style="position: relative">
          <button
            class="fb-tbtn fb-tbtn--bordered"
            :title="t('files.sort', 'Sort')"
            @click="toggleSortMenu"
          >
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              style="width: 17px; height: 17px"
              aria-hidden="true"
            >
              <path d="M7 4v16" />
              <path d="M3 8l4-4 4 4" />
              <path d="M17 20V4" />
              <path d="M21 16l-4 4-4-4" />
            </svg>
          </button>
          <div
            v-show="sortMenuOpen"
            class="fb-sort-menu"
            @click.self="closeSortMenu"
          >
            <button
              class="fb-sort-item"
              :class="{ 'fb-sort-item--active': currentSortBy === 'name' }"
              @click="handleSortMenuClick('name')"
            >
              <span>{{ t("files.name") }}</span>
              <span v-if="currentSortBy === 'name'" class="fb-sort-arrow">
                {{ currentSortAsc ? "↑" : "↓" }}
              </span>
            </button>
            <button
              class="fb-sort-item"
              :class="{ 'fb-sort-item--active': currentSortBy === 'modified' }"
              @click="handleSortMenuClick('modified')"
            >
              <span>{{ t("files.lastModified") }}</span>
              <span v-if="currentSortBy === 'modified'" class="fb-sort-arrow">
                {{ currentSortAsc ? "↑" : "↓" }}
              </span>
            </button>
            <button
              class="fb-sort-item"
              :class="{ 'fb-sort-item--active': currentSortBy === 'size' }"
              @click="handleSortMenuClick('size')"
            >
              <span>{{ t("files.size") }}</span>
              <span v-if="currentSortBy === 'size'" class="fb-sort-arrow">
                {{ currentSortAsc ? "↑" : "↓" }}
              </span>
            </button>
          </div>
        </div>

        <!-- Select multiple toggle -->
        <button
          class="fb-tbtn fb-tbtn--bordered"
          :class="{ 'fb-tbtn--active': fileStore.multiple }"
          :aria-label="t('buttons.selectMultiple')"
          :title="t('buttons.selectMultiple')"
          @click="toggleMultipleSelection"
        >
          <FbIcon name="select-all" size="18px" />
        </button>

        <!-- Details panel toggle — split-panel icon -->
        <button
          class="fb-tbtn fb-tbtn--bordered"
          :class="{ 'fb-tbtn--active': layoutStore.showDetails }"
          :title="t('buttons.info')"
          @click="layoutStore.toggleDetails()"
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="1.9"
            stroke-linecap="round"
            stroke-linejoin="round"
            style="width: 17px; height: 17px"
            aria-hidden="true"
          >
            <path
              d="M3 5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"
            />
            <path d="M15 3v18" />
          </svg>
        </button>

        <!-- Selection actions popup -->
        <template v-if="!isTrash">
          <SelectionActionsPopup
            v-if="!isMobile"
            :header-buttons="{
              download: headerButtons.download,
              share: headerButtons.share,
              move: headerButtons.move,
              rename: headerButtons.rename,
              delete: headerButtons.delete,
            }"
            :download="download"
          />
        </template>
        <template v-else>
          <button
            v-if="fileStore.req && fileStore.req.items.length > 0"
            class="fb-tbtn fb-tbtn--danger"
            :title="$t('trash.emptyTrash')"
            @click="emit('empty-trash')"
          >
            {{ $t("trash.emptyTrash") }}
          </button>
        </template>

        <!-- New button dropdown -->
        <div
          v-if="!isTrash"
          class="fb-new-btn-wrap"
          v-show="headerButtons.upload || authStore.user?.perm.create"
        >
          <button
            class="fb-tbtn fb-tbtn--accent"
            :title="t('buttons.new', 'New')"
            @click="newMenuOpen = !newMenuOpen"
          >
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.2"
              stroke-linecap="round"
              stroke-linejoin="round"
              style="width: 17px; height: 17px"
              aria-hidden="true"
            >
              <path d="M12 5v14" />
              <path d="M5 12h14" />
            </svg>
            <span>{{ t("buttons.new", "New") }}</span>
          </button>
          <div
            v-show="newMenuOpen"
            class="fb-new-menu"
            @click="newMenuOpen = false"
          >
            <button
              v-if="authStore.user?.perm.create"
              @click.stop="showHoverAndClose('newDir')"
              class="fb-new-menu-item"
            >
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linecap="round"
                stroke-linejoin="round"
                style="width: 16px; height: 16px"
                aria-hidden="true"
              >
                <path
                  d="M20 19a2 2 0 0 0 2-2v-6a2 2 0 0 0-2-2h-7.6a1 1 0 0 1-.8-.4L10.3 6.9a1 1 0 0 0-.8-.4H4a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2Z"
                />
                <path d="M12 11v5" />
                <path d="M9.5 13.5h5" />
              </svg>
              <span>{{ t("sidebar.newFolder") }}</span>
            </button>
            <button
              v-if="authStore.user?.perm.create"
              @click.stop="showHoverAndClose('newFile')"
              class="fb-new-menu-item"
            >
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linecap="round"
                stroke-linejoin="round"
                style="width: 16px; height: 16px"
                aria-hidden="true"
              >
                <path
                  d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"
                />
                <path d="M14 2v6h6" />
                <path d="M12 12v6" />
                <path d="M9 15h6" />
              </svg>
              <span>{{ t("sidebar.newFile") }}</span>
            </button>
            <button
              v-if="headerButtons.upload"
              @click.stop="uploadAndClose()"
              class="fb-new-menu-item"
            >
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linecap="round"
                stroke-linejoin="round"
                style="width: 16px; height: 16px"
              >
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <path d="M17 8l-5-5-5 5" />
                <path d="M12 3v12" />
              </svg>
              <span>{{ t("buttons.upload") }}</span>
            </button>
          </div>
        </div>
      </template>
    </header-bar>

    <template v-if="!isTrash">
      <SelectionBar
        v-if="!isMobile && fileStore.multiple && fileStore.selectedCount > 0"
        :header-buttons="{
          download: headerButtons.download,
          share: headerButtons.share,
          move: headerButtons.move,
          rename: headerButtons.rename,
          delete: headerButtons.delete,
          star: true,
        }"
        :download="download"
      />
    </template>
    <template v-else>
      <SelectionBar
        v-if="fileStore.multiple && fileStore.selected.length > 0"
        :header-buttons="{
          download: false,
          share: false,
          move: false,
          rename: false,
          delete: true,
          star: false,
          restore: true,
        }"
        :download="noop"
        :delete-action="trashDeletePermanent"
        :restore-action="trashRestore"
      />
    </template>

    <div
      v-if="isMobile && !isTrash"
      id="file-selection"
      :class="{
        'file-selection-margin-bottom': fileStore.multiple,
      }"
    >
      <span v-if="fileStore.selectedCount > 0">
        {{ t("prompts.filesSelected", fileStore.selectedCount) }}
      </span>
      <action
        v-if="headerButtons.share"
        icon="share"
        :label="t('buttons.share')"
        show="share"
      />
      <action
        v-if="headerButtons.rename"
        icon="mode_edit"
        :label="t('buttons.rename')"
        show="rename"
      />
      <action
        v-if="headerButtons.copy"
        icon="content_copy"
        :label="t('buttons.copyFile')"
        show="copy"
      />
      <action
        v-if="headerButtons.move"
        icon="forward"
        :label="t('buttons.moveFile')"
        show="move"
      />
      <action
        v-if="headerButtons.delete"
        icon="delete"
        :label="t('buttons.delete')"
        show="delete"
      />
    </div>

    <div v-if="layoutStore.loading" :class="currentViewMode">
      <!-- Mosaic skeleton -->
      <div v-if="currentViewMode === 'mosaic'" class="fb-skeleton-wrap">
        <div class="fb-skeleton fb-skeleton-label"></div>
        <div class="fb-skeleton-folders">
          <div v-for="i in 4" :key="'folder-' + i" class="fb-skeleton-folder">
            <div class="fb-skeleton fb-skeleton-folder-icon"></div>
            <div class="fb-skeleton fb-skeleton-folder-name"></div>
          </div>
        </div>
        <div class="fb-skeleton fb-skeleton-label"></div>
        <div class="fb-skeleton-files">
          <div v-for="i in 8" :key="'file-' + i" class="fb-skeleton-file">
            <div class="fb-skeleton-file-thumb"></div>
            <div class="fb-skeleton-file-info">
              <div class="fb-skeleton fb-skeleton-file-name"></div>
              <div class="fb-skeleton fb-skeleton-file-meta"></div>
            </div>
          </div>
        </div>
      </div>
      <!-- List skeleton -->
      <div v-else class="fb-skeleton-wrap">
        <div class="fb-skeleton-col-header">
          <div class="fb-skeleton fb-skeleton-col-label"></div>
          <div class="fb-skeleton fb-skeleton-col-label"></div>
          <div class="fb-skeleton fb-skeleton-col-label"></div>
          <div class="fb-skeleton fb-skeleton-col-label"></div>
        </div>
        <div v-for="i in 10" :key="'row-' + i" class="fb-skeleton-row">
          <div class="fb-skeleton fb-skeleton-row-icon"></div>
          <div class="fb-skeleton fb-skeleton-row-name"></div>
          <div class="fb-skeleton fb-skeleton-row-size"></div>
          <div class="fb-skeleton fb-skeleton-row-modified"></div>
        </div>
      </div>
    </div>
    <template v-else>
      <div v-if="isEmpty" class="fb-empty">
        <div class="fb-empty-icon">
          <FbIcon :name="emptyIconName" size="40px" />
        </div>
        <p class="fb-empty-title">{{ emptyTitle }}</p>
        <p class="fb-empty-sub">{{ emptySub }}</p>
        <input
          style="display: none"
          type="file"
          id="upload-input"
          @change="uploadInput($event)"
          multiple
        />
        <input
          style="display: none"
          type="file"
          id="upload-folder-input"
          @change="uploadInput($event)"
          webkitdirectory
          multiple
        />
      </div>
      <div
        v-else
        id="listing"
        ref="listing"
        class="file-icons"
        data-clear-on-click="true"
        :class="currentViewMode"
        @click="handleEmptyAreaClick"
        @contextmenu="showContextMenu"
      >
        <div>
          <div class="fb-col-header">
            <div>
              <p
                :class="{ active: nameSorted }"
                class="name"
                role="button"
                tabindex="0"
                @click="sort('name')"
                :title="t('files.sortByName')"
                :aria-label="t('files.sortByName')"
              >
                <span>{{ t("files.name") }}</span>
                <FbIcon :name="nameIcon" size="14px" />
              </p>

              <p
                :class="{ active: sizeSorted }"
                class="size"
                role="button"
                tabindex="0"
                @click="sort('size')"
                :title="t('files.sortBySize')"
                :aria-label="t('files.sortBySize')"
              >
                <span>{{ t("files.size") }}</span>
                <FbIcon :name="sizeIcon" size="14px" />
              </p>
              <p
                :class="{ active: modifiedSorted }"
                class="modified"
                role="button"
                tabindex="0"
                @click="sort('modified')"
                :title="
                  isTrash ? t('trash.deletedAt') : t('files.sortByLastModified')
                "
                :aria-label="
                  isTrash ? t('trash.deletedAt') : t('files.sortByLastModified')
                "
              >
                <span>{{
                  isTrash ? t("trash.deletedAt") : t("files.lastModified")
                }}</span>
                <FbIcon :name="modifiedIcon" size="14px" />
              </p>
            </div>
          </div>
        </div>

        <h2 data-clear-on-click="true" v-if="fileStore.req?.numDirs ?? false">
          {{ t("files.folders") }}
        </h2>
        <div
          v-if="fileStore.req?.numDirs ?? false"
          class="fb-items fb-items--folders"
          data-clear-on-click="true"
          @contextmenu="showContextMenu"
        >
          <item
            v-for="item in dirs"
            :key="base64(item.name)"
            v-bind:index="item.index"
            v-bind:name="isTrash ? cleanTrashName(item.name) : item.name"
            v-bind:isDir="item.isDir"
            v-bind:url="item.url"
            v-bind:modified="item.modified"
            v-bind:type="item.type"
            v-bind:size="item.size"
            v-bind:path="item.path"
            v-bind:onOpen="isTrash ? handleTrashOpen : undefined"
            v-bind:subtitle="
              isTrash && props.flatView ? trashSubtitle(item) : undefined
            "
          >
          </item>
        </div>

        <h2 data-clear-on-click="true" v-if="fileStore.req?.numFiles ?? false">
          {{ t("files.files") }}
        </h2>
        <div
          v-if="fileStore.req?.numFiles ?? false"
          class="fb-items fb-items--files"
          data-clear-on-click="true"
          @contextmenu="showContextMenu"
        >
          <item
            v-for="item in files"
            :key="base64(item.name)"
            v-bind:index="item.index"
            v-bind:name="isTrash ? cleanTrashName(item.name) : item.name"
            v-bind:isDir="item.isDir"
            v-bind:url="item.url"
            v-bind:modified="item.modified"
            v-bind:type="item.type"
            v-bind:size="item.size"
            v-bind:path="item.path"
            v-bind:preview="item.preview"
            v-bind:subtitle="
              isTrash && props.flatView ? trashSubtitle(item) : undefined
            "
            v-bind:noOpen="
              (isTrash && !item.isDir && !trashSubPath) || undefined
            "
            v-bind:onOpen="isTrash ? handleTrashOpen : undefined"
          >
          </item>
        </div>
        <context-menu
          :show="isContextMenuVisible"
          :pos="contextMenuPos"
          @hide="hideContextMenu"
        >
          <template v-if="isTrash">
            <action
              :fbIcon="'arrow-back'"
              :label="$t('trash.restore')"
              @action="trashRestore"
            />
            <hr class="fb-menu-divider" />
            <action
              :fbIcon="'delete'"
              :label="$t('trash.deletePermanent')"
              @action="trashDeletePermanent"
              :danger="true"
            />
          </template>
          <template v-else-if="isContextMenuOnItem">
            <action
              v-if="headerButtons.share"
              :fbIcon="'share'"
              :label="t('buttons.share')"
              show="share"
            />
            <action
              v-if="headerButtons.download"
              :fbIcon="'download'"
              :label="t('buttons.download')"
              @action="download"
              :counter="fileStore.selectedCount"
            />
            <action
              v-if="isContextMenuOnArchive && authStore.user?.perm.modify"
              :fbIcon="'archive'"
              :label="t('buttons.extract')"
              @action="extractArchive"
            />
            <hr class="fb-menu-divider" />
            <action
              v-if="headerButtons.rename"
              :fbIcon="'rename'"
              :label="t('buttons.rename')"
              show="rename"
            />
            <action
              v-if="headerButtons.copy"
              :fbIcon="'copy'"
              :label="t('buttons.copyFile')"
              show="copy"
            />
            <action
              v-if="headerButtons.move"
              :fbIcon="'move'"
              :label="t('buttons.moveFile')"
              show="move"
            />
            <hr class="fb-menu-divider" />
            <action
              :fbIcon="'info'"
              :label="t('buttons.info')"
              @action="layoutStore.showDetails = true"
            />
            <action
              :fbIcon="'star'"
              :label="
                isStarredContextItem ? t('files.unstar') : t('files.star')
              "
              @action="toggleStar"
            />
            <action
              v-if="isContextMenuOnFolder"
              :fbIcon="'palette'"
              :label="t('contextMenu.color')"
              @action="openColorPicker"
            />
            <hr class="fb-menu-divider" />
            <action
              v-if="headerButtons.delete"
              :fbIcon="'trash'"
              :label="t('files.moveToTrash')"
              @action="moveToTrash"
            />
            <action
              v-if="headerButtons.delete"
              :fbIcon="'delete'"
              :label="t('buttons.delete')"
              show="delete"
              :danger="true"
            />
          </template>
          <template v-else>
            <action
              v-if="authStore.user?.perm.create"
              :fbIcon="'new-folder'"
              :label="t('sidebar.newFolder')"
              @action="showHoverAndClose('newDir')"
            />
            <action
              v-if="authStore.user?.perm.create"
              :fbIcon="'new-file'"
              :label="t('sidebar.newFile')"
              @action="showHoverAndClose('newFile')"
            />
            <action
              v-if="headerButtons.upload"
              :fbIcon="'upload'"
              :label="t('buttons.upload')"
              @action="uploadAndClose()"
            />
          </template>
        </context-menu>

        <div
          v-if="showColorPicker"
          class="fb-color-picker-overlay"
          @click="showColorPicker = false"
        >
          <div
            class="fb-color-picker-container"
            :style="{
              left: colorPickerPos.x + 'px',
              top: colorPickerPos.y + 'px',
            }"
            @click.stop
          >
            <FolderColorPicker
              :currentColor="getCurrentFolderColor()"
              @select="handleColorSelect"
              @remove="handleColorRemove"
            />
          </div>
        </div>

        <input
          style="display: none"
          type="file"
          id="upload-input"
          @change="uploadInput($event)"
          multiple
        />
        <input
          style="display: none"
          type="file"
          id="upload-folder-input"
          @change="uploadInput($event)"
          webkitdirectory
          multiple
        />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useClipboardStore } from "@/stores/clipboard";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useExtractStore } from "@/stores/extract";
import { useSourceStore } from "@/stores/source";

import { users, files as api } from "@/api";
import { enableExec } from "@/utils/constants";
import * as upload from "@/utils/upload";
import { throttle } from "lodash-es";
import { Base64 } from "js-base64";
import {
  isStarred,
  toggleStarred,
  starVersion,
  sortStarredFirst,
} from "@/utils/starred";
import { fileKind } from "@/utils/fileKind";
import { StatusError } from "@/api/utils";
import {
  buildTrashInfo,
  cleanTrashName,
  isTrashPath,
  originalPathFromTrash,
  parentDir,
  shouldShowInTrash,
  trashFilesRoot,
  trashInfoRoot,
  trashInfoUrl,
  trashSourcePathFromUrl,
  withVersionSuffix,
} from "@/utils/trash";

import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";
import Item from "@/components/files/ListingItem.vue";
import ContextMenu from "@/components/ContextMenu.vue";
import SelectionActionsPopup from "@/components/SelectionActionsPopup.vue";
import SelectionBar from "@/components/SelectionBar.vue";
import FbIcon from "@/components/FbIcon.vue";
import FolderColorPicker from "@/components/FolderColorPicker.vue";
import type { IconName } from "@/utils/icons";
import Search from "@/components/Search.vue";
import {
  computed,
  inject,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
} from "vue";
import { useRoute, useRouter, onBeforeRouteUpdate } from "vue-router";
import { useI18n } from "vue-i18n";
import { storeToRefs } from "pinia";
import { removePrefix } from "@/api/utils";

const props = defineProps<{
  isTrash?: boolean;
  sourceId?: string;
  searchQuery?: string;
  sortBy?: string;
  sortAsc?: boolean;
  trashSubPath?: string;
  breadcrumb?: { label: string; url: string }[];
  flatView?: boolean;
}>();

const emit = defineEmits<{
  "update:searchQuery": [value: string];
  "update:sortBy": [value: string];
  "update:sortAsc": [value: boolean];
  restore: [];
  "delete-permanent": [];
  "empty-trash": [];
  "switch-source": [id: string];
  "navigate-to": [url: string];
  "toggle-flat": [value: boolean];
}>();

const showLimit = ref<number>(50);
let fadeResetTimer: number | null = null;
const width = ref<number>(window.innerWidth);
const itemWeight = ref<number>(0);
const isContextMenuVisible = ref<boolean>(false);
const contextMenuPos = ref<{ x: number; y: number }>({ x: 0, y: 0 });
const isContextMenuOnItem = ref<boolean>(false);
const isContextMenuOnFolder = ref<boolean>(false);
const isContextMenuOnArchive = computed(() => {
  const items = fileStore.req?.items;
  if (!items || fileStore.selected.length !== 1) return false;
  const item = items[fileStore.selected[0]];
  if (!item || item.isDir) return false;
  return fileKind(item) === "archive";
});
const showColorPicker = ref<boolean>(false);
const colorPickerPos = ref<{ x: number; y: number }>({ x: 0, y: 0 });
const newMenuOpen = ref<boolean>(false);
const sortMenuOpen = ref<boolean>(false);

const $showError = inject<IToastError>("$showError")!;

const clipboardStore = useClipboardStore();
const authStore = useAuthStore();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();
const extractStore = useExtractStore();
const sourceStore = useSourceStore();

const { req } = storeToRefs(fileStore);

const route = useRoute();
const router = useRouter();

const filesBase = computed(() =>
  props.isTrash
    ? `/files/${props.sourceId ?? sourceStore.activeId}`
    : `/files/${route.params.sourceId ?? 0}`
);

onBeforeRouteUpdate(() => {
  hideContextMenu();
});

const { t } = useI18n();

const listing = ref<HTMLElement | null>(null);
const searchRef = ref<InstanceType<typeof Search> | null>(null);

const nameSorted = computed(() =>
  props.isTrash
    ? props.sortBy === "name"
    : fileStore.req
      ? fileStore.req.sorting.by === "name"
      : false
);

const sizeSorted = computed(() =>
  props.isTrash
    ? props.sortBy === "size"
    : fileStore.req
      ? fileStore.req.sorting.by === "size"
      : false
);

const modifiedSorted = computed(() =>
  props.isTrash
    ? props.sortBy === "modified"
    : fileStore.req
      ? fileStore.req.sorting.by === "modified"
      : false
);

const ascOrdered = computed(() =>
  props.isTrash
    ? !!props.sortAsc
    : fileStore.req
      ? fileStore.req.sorting.asc
      : false
);

const currentSortBy = computed(() =>
  props.isTrash
    ? (props.sortBy ?? "name")
    : (fileStore.req?.sorting.by ?? "name")
);
const currentSortAsc = computed(() =>
  props.isTrash ? (props.sortAsc ?? true) : (fileStore.req?.sorting.asc ?? true)
);

const dirs = computed(() => items.value.dirs.slice(0, showLimit.value));

const items = computed(() => {
  const dirs: ResourceItem[] = [];
  const files: ResourceItem[] = [];

  const q =
    props.isTrash && props.searchQuery
      ? props.searchQuery.toLowerCase().trim()
      : "";
  let source = fileStore.req?.items ?? [];

  if (q) {
    source = source.filter((item) =>
      cleanTrashName(item.name).toLowerCase().includes(q)
    );
  }

  if (props.isTrash) {
    const by = props.sortBy ?? "name";
    const asc = props.sortAsc ?? true;
    source = [...source].sort((a, b) => {
      let cmp = 0;
      if (by === "name") {
        cmp = cleanTrashName(a.name).localeCompare(cleanTrashName(b.name));
      } else if (by === "modified") {
        cmp = new Date(a.modified).getTime() - new Date(b.modified).getTime();
      } else if (by === "size") {
        cmp = a.size - b.size;
      }
      return asc ? cmp : -cmp;
    });
  }

  source
    .filter(
      (item) =>
        item.name !== ".Trash" &&
        (!props.isTrash || shouldShowInTrash(item.name))
    )
    .forEach((item) => {
      if (item.isDir) {
        dirs.push(item);
      } else {
        files.push(item);
      }
    });

  // Starred items are pinned to the top of their section, sorted by name.
  // Starring/unstarring re-evaluates this computed via starVersion.
  if (!props.isTrash) {
    starVersion.value;
    return { dirs: sortStarredFirst(dirs), files: sortStarredFirst(files) };
  }

  return { dirs, files };
});

const files = computed((): ResourceItem[] => {
  let _showLimit = showLimit.value - items.value.dirs.length;

  if (_showLimit < 0) _showLimit = 0;

  return items.value.files.slice(0, _showLimit);
});

const nameIcon = computed((): IconName => {
  return nameSorted.value && !ascOrdered.value ? "arrow-up" : "arrow-down";
});

const sizeIcon = computed((): IconName => {
  return sizeSorted.value && ascOrdered.value ? "arrow-down" : "arrow-up";
});

const modifiedIcon = computed((): IconName => {
  return modifiedSorted.value && ascOrdered.value ? "arrow-down" : "arrow-up";
});

const isEmpty = computed(
  () => (fileStore.req?.numDirs ?? 0) + (fileStore.req?.numFiles ?? 0) === 0
);

const emptyIconName = computed<IconName>(() => "empty-folder");

const emptyTitle = computed(() => t("files.emptyFolderTitle"));

const emptySub = computed(() => t("files.emptyFolderSub"));

const currentViewMode = computed(() => {
  const mode = authStore.user?.viewMode ?? "list";
  // Gallery view was removed from the UI; fall back to plain mosaic so any
  // persisted "mosaic gallery" preference still renders cleanly.
  return mode === "mosaic gallery" ? "mosaic" : mode;
});

const headerButtons = computed(() => {
  return {
    upload: authStore.user?.perm.create,
    download: authStore.user?.perm.download,
    shell: authStore.user?.perm.execute && enableExec,
    delete: fileStore.selectedCount > 0 && authStore.user?.perm.delete,
    rename: fileStore.selectedCount === 1 && authStore.user?.perm.rename,
    share:
      fileStore.selectedCount === 1 &&
      authStore.user?.perm.share &&
      authStore.user?.perm.download,
    move: fileStore.selectedCount > 0 && authStore.user?.perm.rename,
    copy: fileStore.selectedCount > 0 && authStore.user?.perm.create,
  };
});

const isMobile = computed(() => {
  return width.value <= 736;
});

watch(req, () => {
  // Reset the show value
  showLimit.value = 50;

  nextTick(() => {
    // Ensures that the listing is displayed
    // How much every listing item affects the window height
    setItemWeight();

    // Scroll to the item opened previously
    if (!revealPreviousItem()) {
      // Fill and fit the window with listing items
      fillWindow(true);
    }
  });
});

// Clear selection when details panel closes
watch(
  () => layoutStore.showDetails,
  (isOpen, wasOpen) => {
    if (wasOpen && !isOpen) {
      fileStore.selected = [];
    }
  }
);

onMounted(() => {
  // Check the columns size for the first time.
  columnsResize();

  // How much every listing item affects the window height
  setItemWeight();

  // Scroll to the item opened previously
  if (!revealPreviousItem()) {
    // Fill and fit the window with listing items
    fillWindow(true);
  }

  // Add the needed event listeners to the window and document.
  window.addEventListener("keydown", keyEvent);
  window.addEventListener("scroll", scrollEvent, true);
  window.addEventListener("resize", windowsResize);
  if (!props.isTrash) {
    document.addEventListener("click", closeNewMenu);
  }

  if (props.isTrash || !authStore.user?.perm.create) return;
  document.addEventListener("dragover", onDragOverGlobal);
  document.addEventListener("dragenter", dragEnter);
  document.addEventListener("dragleave", dragLeave);
  document.addEventListener("dragend", dragEnd);
  document.addEventListener("drop", drop);
});

onBeforeUnmount(() => {
  // Remove event listeners before destroying this page.
  window.removeEventListener("keydown", keyEvent);
  window.removeEventListener("scroll", scrollEvent, true);
  window.removeEventListener("resize", windowsResize);
  stopDragScroll();
  if (fadeResetTimer) {
    window.clearTimeout(fadeResetTimer);
    fadeResetTimer = null;
  }
  if (!props.isTrash) {
    document.removeEventListener("click", closeNewMenu);
  }

  if (props.isTrash || !authStore.user?.perm.create) return;
  document.removeEventListener("dragover", onDragOverGlobal);
  document.removeEventListener("dragenter", dragEnter);
  document.removeEventListener("dragleave", dragLeave);
  document.removeEventListener("dragend", dragEnd);
  document.removeEventListener("drop", drop);
});

const base64 = (name: string) => Base64.encodeURI(name);

const closeNewMenu = (e: MouseEvent) => {
  const wrap = document.querySelector(".fb-new-btn-wrap");
  if (wrap && !wrap.contains(e.target as Node)) {
    newMenuOpen.value = false;
  }
};

const keyEvent = (event: KeyboardEvent) => {
  // No prompts are shown
  if (layoutStore.currentPrompt !== null) {
    return;
  }

  if (event.key === "Escape") {
    // Reset files selection.
    fileStore.selected = [];
  }

  if (event.key === "Delete") {
    if (!authStore.user?.perm.delete || fileStore.selectedCount == 0) return;

    if (props.isTrash) {
      emit("delete-permanent");
    } else {
      // Show delete prompt.
      layoutStore.showHover("delete");
    }
  }

  if (event.key === "F2") {
    if (!authStore.user?.perm.rename || fileStore.selectedCount !== 1) return;

    // Show rename prompt.
    layoutStore.showHover("rename");
  }

  // Ctrl is pressed
  if (!event.ctrlKey && !event.metaKey) {
    return;
  }

  switch (event.key) {
    case "f":
    case "F":
      if (event.shiftKey) {
        event.preventDefault();
        searchRef.value?.focus();
      }
      break;
    case "c":
    case "x":
      copyCut(event);
      break;
    case "v":
      paste(event);
      break;
    case "a":
      event.preventDefault();
      for (const file of items.value.files) {
        if (fileStore.selected.indexOf(file.index) === -1) {
          fileStore.selected.push(file.index);
        }
      }
      for (const dir of items.value.dirs) {
        if (fileStore.selected.indexOf(dir.index) === -1) {
          fileStore.selected.push(dir.index);
        }
      }
      break;
    case "s":
      event.preventDefault();
      document.getElementById("download-button")?.click();
      break;
  }
};

// --- Drag auto-scroll -----------------------------------------------------
// While dragging, moving the pointer near the top or bottom edge of the
// scrollable area scrolls it, so items can be dropped on folders far off
// screen without dropping first. Follows the conventions of
// react-beautiful-dnd / dnd-kit auto-scrollers: an edge zone sized as a
// percentage of the container height (clamped to absolute pixels), and a
// scroll speed that accelerates as the pointer approaches the true edge.

let dragScrollRaf: number | null = null;
let dragClientY = -1;
let dragLastMove = 0;

const SCROLL_ZONE_FRACTION = 0.18;
const SCROLL_ZONE_MIN = 48;
const SCROLL_ZONE_MAX = 180;
const SCROLL_MAX_SPEED_FRACTION = 0.05;
const SCROLL_MAX_SPEED_MIN = 12;
const SCROLL_MAX_SPEED_MAX = 40;
const SCROLL_STALE_MS = 300;

const getDragScroller = (): HTMLElement => {
  // The listing scrolls inside `main`; when the view is short the window
  // itself scrolls instead.
  const main = document.querySelector("main");
  if (main && main.scrollHeight > main.clientHeight + 4) {
    return main as HTMLElement;
  }
  return document.documentElement;
};

const stopDragScroll = () => {
  dragClientY = -1;
  if (dragScrollRaf !== null) {
    cancelAnimationFrame(dragScrollRaf);
    dragScrollRaf = null;
  }
};

const onDragOverGlobal = (event: DragEvent) => {
  event.preventDefault();
  dragClientY = event.clientY;
  dragLastMove = Date.now();
  if (dragScrollRaf === null) {
    dragScrollRaf = requestAnimationFrame(dragScrollTick);
  }
};

const dragScrollTick = () => {
  dragScrollRaf = null;

  // Safety net: if no dragover arrives for a while the drag ended without a
  // proper dragend/drop (e.g. Esc-cancel quirks) — stop scrolling.
  if (Date.now() - dragLastMove > SCROLL_STALE_MS) {
    stopDragScroll();
    return;
  }

  const y = dragClientY;
  if (y < 0) return;

  const scroller = getDragScroller();
  const rect = scroller.getBoundingClientRect();
  const viewH = scroller.clientHeight || window.innerHeight;
  if (viewH <= 0) return;

  const zone = Math.min(
    Math.max(viewH * SCROLL_ZONE_FRACTION, SCROLL_ZONE_MIN),
    SCROLL_ZONE_MAX
  );
  const maxSpeed = Math.min(
    Math.max(viewH * SCROLL_MAX_SPEED_FRACTION, SCROLL_MAX_SPEED_MIN),
    SCROLL_MAX_SPEED_MAX
  );

  const fromTop = y - rect.top;
  const fromBottom = rect.bottom - y;
  let speed = 0;
  if (fromTop >= 0 && fromTop < zone) {
    const t = 1 - fromTop / zone;
    speed = -Math.max(1, Math.round(maxSpeed * t * t));
  } else if (fromBottom >= 0 && fromBottom < zone) {
    const t = 1 - fromBottom / zone;
    speed = Math.max(1, Math.round(maxSpeed * t * t));
  }

  if (speed !== 0) {
    scroller.scrollTop += speed;
    dragScrollRaf = requestAnimationFrame(dragScrollTick);
  }
};

const copyCut = (event: Event | KeyboardEvent): void => {
  if ((event.target as HTMLElement).tagName?.toLowerCase() === "input") return;

  if (fileStore.req === null) return;

  const items = [];

  for (const i of fileStore.selected) {
    items.push({
      from: fileStore.req.items[i].url,
      name: fileStore.req.items[i].name,
      size: fileStore.req.items[i].size,
      isDir: fileStore.req.items[i].isDir,
      modified: fileStore.req.items[i].modified,
    });
  }

  if (items.length === 0) {
    return;
  }

  clipboardStore.$patch({
    key: (event as KeyboardEvent).key,
    items,
    path: route.path,
  });
};

const paste = async (event: Event) => {
  if ((event.target as HTMLElement).tagName?.toLowerCase() === "input") return;

  // TODO router location should it be
  const items: any[] = [];

  for (const item of clipboardStore.items) {
    const from = item.from.endsWith("/") ? item.from.slice(0, -1) : item.from;
    const to = route.path + encodeURIComponent(item.name);
    items.push({
      from,
      to,
      name: item.name,
      size: item.size,
      isDir: item.isDir,
      modified: item.modified,
      overwrite: false,
      rename: clipboardStore.path == route.path,
    });
  }

  if (items.length === 0) {
    return;
  }

  const preselect = removePrefix(route.path) + items[0].name;

  let action = (overwrite?: boolean, rename?: boolean) => {
    api
      .copy(items, overwrite, rename)
      .then(() => {
        fileStore.preselect = preselect;
        fileStore.reload = true;
      })
      .catch($showError);
  };

  if (clipboardStore.key === "x") {
    action = (overwrite, rename) => {
      api
        .move(items, overwrite, rename)
        .then(() => {
          clipboardStore.resetClipboard();
          fileStore.preselect = preselect;
          fileStore.reload = true;
        })
        .catch($showError);
    };
  }

  const path = route.path.endsWith("/") ? route.path : route.path + "/";
  const conflict = await upload.checkConflict(items, path, true);

  if (conflict.length > 0) {
    layoutStore.showHover({
      prompt: "resolve-conflict",
      props: {
        conflict: conflict,
      },
      confirm: (event: Event, result: Array<ConflictingResource>) => {
        event.preventDefault();
        layoutStore.closeHovers();
        for (let i = result.length - 1; i >= 0; i--) {
          const item = result[i];
          if (item.checked.length == 2) {
            items[item.index].rename = true;
          } else if (item.checked.length == 1 && item.checked[0] == "origin") {
            items[item.index].overwrite = true;
          } else {
            items.splice(item.index, 1);
          }
        }
        if (items.length > 0) {
          action();
        }
      },
    });

    return;
  }

  action(false, false);
};

const columnsResize = () => {
  // No-op: mosaic view now uses CSS Grid (auto-fill / minmax) for column
  // sizing, so per-item inline widths are no longer required.
};

const scrollEvent = throttle((e?: Event) => {
  const totalItems =
    (fileStore.req?.numDirs ?? 0) + (fileStore.req?.numFiles ?? 0);

  // All items are displayed
  if (showLimit.value >= totalItems) return;

  // Support both window scroll (mobile) and container scroll (desktop)
  const scroller =
    (e?.target instanceof HTMLElement && e.target.id !== "fb-layout"
      ? e.target
      : document.querySelector("main")) ?? document.documentElement;

  const scrollTop = scroller.scrollTop ?? window.scrollY;
  const scrollHeight = scroller.scrollHeight ?? document.body.offsetHeight;
  const clientHeight = scroller.clientHeight ?? window.innerHeight;
  const currentPos = scrollTop + clientHeight;
  const triggerPos = scrollHeight - clientHeight * 0.25;

  if (currentPos > triggerPos) {
    const showQuantity = Math.ceil((clientHeight * 2) / itemWeight.value);
    showLimit.value += showQuantity;
  }
}, 100);

const fadeItems = () => {
  // When the user starts dragging an item, put every
  // file on the listing with 50% opacity.
  const items = document.getElementsByClassName("item");
  Array.from(items).forEach((file: Element) => {
    (file as HTMLElement).style.opacity = "0.5";
  });

  // Watchdog: whatever happens (Esc cancel, quirks in Chrome's drag events,
  // drops outside the listing), the fade must never stick. Every dragenter
  // renews the timer, so it only fires while the drag is idle or over.
  if (fadeResetTimer) window.clearTimeout(fadeResetTimer);
  fadeResetTimer = window.setTimeout(() => {
    fadeResetTimer = null;
    stopDragScroll();
    resetOpacity();
  }, 4000);
};

const dragEnter = (event: DragEvent) => {
  event.preventDefault();

  // Only dim the listing for internal item drags (which carry text/plain).
  // Dragging files from the OS to upload must never fade the background.
  const types = event.dataTransfer?.types ?? [];
  const isInternalDrag =
    types.includes("text/plain") && !types.includes("Files");
  if (isInternalDrag) {
    fadeItems();
  }
};

const dragLeave = (event: DragEvent) => {
  event.preventDefault();
  // Only reset when the drag actually leaves the document. Moving between
  // elements fires document-level dragleave with a relatedTarget still
  // inside the document — resetting there would flicker the fade.
  const related = event.relatedTarget as Node | null;
  if (related === null || !document.contains(related)) {
    resetDragUi();
  }
};

const dragEnd = () => {
  resetDragUi();
};

const clearDropTargets = () => {
  document
    .querySelectorAll("#listing .item.fb-drop-target")
    .forEach((el) => el.classList.remove("fb-drop-target"));
};

const resetDragUi = () => {
  stopDragScroll();
  if (fadeResetTimer) {
    window.clearTimeout(fadeResetTimer);
    fadeResetTimer = null;
  }
  resetOpacity();
  clearDropTargets();
};

const drop = async (event: DragEvent) => {
  event.preventDefault();
  resetDragUi();

  const dt = event.dataTransfer;
  let el: HTMLElement | null = event.target as HTMLElement;

  // The upload modal handles its own drops — don't start a second upload here.
  if ((event.target as HTMLElement).closest?.(".upload-card")) return;

  if (fileStore.req === null || dt === null || dt.files.length <= 0) return;

  for (let i = 0; i < 5; i++) {
    if (el !== null && !el.classList.contains("item")) {
      el = el.parentElement;
    }
  }

  const files: UploadList = (await upload.scanFiles(dt)) as UploadList;
  let path = route.path.endsWith("/") ? route.path : route.path + "/";

  if (
    el !== null &&
    el.classList.contains("item") &&
    el.dataset.dir === "true"
  ) {
    // Get url from ListingItem instance
    // TODO: Don't know what is happening here
    path = el.__vue__.url;

    try {
      (await api.fetch(path)).items;
    } catch (error: any) {
      $showError(error);
      return;
    }
  }

  const conflict = await upload.checkConflict(files, path);

  const preselect = removePrefix(path) + (files[0].fullPath || files[0].name);

  if (conflict.length > 0) {
    layoutStore.showHover({
      prompt: "resolve-conflict",
      props: {
        conflict: conflict,
        isUploadAction: true,
      },
      confirm: (event: Event, result: Array<ConflictingResource>) => {
        event.preventDefault();
        layoutStore.closeHovers();
        for (let i = result.length - 1; i >= 0; i--) {
          const item = result[i];
          if (item.checked.length == 2) {
            continue;
          } else if (item.checked.length == 1 && item.checked[0] == "origin") {
            files[item.index].overwrite = true;
          } else {
            files.splice(item.index, 1);
          }
        }
        if (files.length > 0) {
          upload.handleFiles(files, path, true);
          fileStore.preselect = preselect;
        }
      },
    });

    return;
  }

  upload.handleFiles(files, path);
  fileStore.preselect = preselect;
};

const uploadInput = async (event: Event) => {
  const files = (event.currentTarget as HTMLInputElement)?.files;
  if (files === null) return;

  const folder_upload = !!files[0].webkitRelativePath;

  const uploadFiles: UploadList = [];
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    const fullPath = folder_upload ? file.webkitRelativePath : undefined;
    uploadFiles.push({
      file,
      name: file.name,
      size: file.size,
      isDir: false,
      fullPath,
    });
  }

  const path = route.path.endsWith("/") ? route.path : route.path + "/";
  const conflict = await upload.checkConflict(uploadFiles, path);

  if (conflict.length > 0) {
    layoutStore.showHover({
      prompt: "resolve-conflict",
      props: {
        conflict: conflict,
        isUploadAction: true,
      },
      confirm: (event: Event, result: Array<ConflictingResource>) => {
        event.preventDefault();
        layoutStore.closeHovers();
        for (let i = result.length - 1; i >= 0; i--) {
          const item = result[i];
          if (item.checked.length == 2) {
            continue;
          } else if (item.checked.length == 1 && item.checked[0] == "origin") {
            uploadFiles[item.index].overwrite = true;
          } else {
            uploadFiles.splice(item.index, 1);
          }
        }
        if (uploadFiles.length > 0) {
          upload.handleFiles(uploadFiles, path, true);
        }
      },
    });

    return;
  }

  upload.handleFiles(uploadFiles, path);
};

const resetOpacity = () => {
  const items = document.getElementsByClassName("item");

  Array.from(items).forEach((file: Element) => {
    (file as HTMLElement).style.opacity = "1";
  });
};

const sort = async (by: string, explicitAsc?: boolean) => {
  if (props.isTrash) {
    let asc = false;
    if (explicitAsc !== undefined) {
      asc = explicitAsc;
    } else if (props.sortBy === by) {
      asc = !props.sortAsc;
    } else {
      asc = true;
    }
    emit("update:sortBy", by);
    emit("update:sortAsc", asc);
    return;
  }

  let asc = false;

  if (explicitAsc !== undefined) {
    asc = explicitAsc;
  } else if (by === "name") {
    if (nameIcon.value === "arrow-up") {
      asc = true;
    }
  } else if (by === "size") {
    if (sizeIcon.value === "arrow-up") {
      asc = true;
    }
  } else if (by === "modified") {
    if (modifiedIcon.value === "arrow-up") {
      asc = true;
    }
  }

  try {
    if (authStore.user?.id) {
      await users.update({ id: authStore.user?.id, sorting: { by, asc } }, [
        "sorting",
      ]);
    }
  } catch (e: any) {
    $showError(e);
  }

  fileStore.reload = true;
};

const toggleMultipleSelection = () => {
  fileStore.toggleMultiple();
  layoutStore.closeHovers();
};

const windowsResize = throttle(() => {
  columnsResize();
  width.value = window.innerWidth;

  // Listing element is not displayed
  if (listing.value == null) return;

  // How much every listing item affects the window height
  setItemWeight();

  // Fill but not fit the window
  fillWindow();
}, 100);

const download = () => {
  if (fileStore.req === null) return;

  if (
    fileStore.selectedCount === 1 &&
    !fileStore.req.items[fileStore.selected[0]].isDir
  ) {
    api.download(null, fileStore.req.items[fileStore.selected[0]].url);
    return;
  }

  layoutStore.showHover({
    prompt: "download",
    confirm: (format: any) => {
      layoutStore.closeHovers();

      const files = [];

      if (fileStore.selectedCount > 0 && fileStore.req !== null) {
        for (const i of fileStore.selected) {
          files.push(fileStore.req.items[i].url);
        }
      } else {
        files.push(route.path);
      }

      api.download(format, ...files);
    },
  });
};

const extractArchive = async () => {
  hideContextMenu();
  const items = fileStore.req?.items;
  if (!items || fileStore.selected.length !== 1) return;
  const item = items[fileStore.selected[0]];
  if (!item) return;
  extractStore.start(item.url, item.name);
  try {
    await api.extract(item.url);
    extractStore.finish(item.url);
    fileStore.reload = true;
  } catch (e: any) {
    extractStore.fail(item.url);
    $showError(e);
  }
};

const setView = async (mode: ViewModeType) => {
  if (authStore.user?.viewMode === mode) return;
  layoutStore.closeHovers();
  const data = { id: authStore.user?.id, viewMode: mode };
  users.update(data, ["viewMode"]).catch($showError);
  authStore.updateUser(data);
  setItemWeight();
  fillWindow();
};

const toggleSortMenu = () => {
  sortMenuOpen.value = !sortMenuOpen.value;
};

const closeSortMenu = () => {
  sortMenuOpen.value = false;
};

const handleSortMenuClick = async (by: string) => {
  if (props.isTrash) {
    const curBy = props.sortBy ?? "name";
    const curAsc = props.sortAsc ?? true;
    const asc = curBy === by ? !curAsc : true;
    sortMenuOpen.value = false;
    await sort(by, asc);
    return;
  }
  const cur = fileStore.req?.sorting;
  if (!cur) return;
  const asc = cur.by === by ? !cur.asc : true;
  sortMenuOpen.value = false;
  await sort(by, asc);
};

const showHoverAndClose = (prompt: string) => {
  newMenuOpen.value = false;
  layoutStore.showHover(prompt);
};

const uploadAndClose = () => {
  newMenuOpen.value = false;
  uploadFunc();
};

const uploadFunc = () => {
  if (
    typeof window.DataTransferItem !== "undefined" &&
    typeof DataTransferItem.prototype.webkitGetAsEntry !== "undefined"
  ) {
    layoutStore.showHover("upload");
  } else {
    document.getElementById("upload-input")?.click();
  }
};

const setItemWeight = () => {
  // Listing element is not displayed
  if (listing.value === null || fileStore.req === null) return;

  let itemQuantity = fileStore.req.numDirs + fileStore.req.numFiles;
  if (itemQuantity > showLimit.value) itemQuantity = showLimit.value;

  // How much every listing item affects the window height
  itemWeight.value = listing.value.offsetHeight / itemQuantity;
};

const fillWindow = (fit = false) => {
  if (fileStore.req === null) return;

  const totalItems = fileStore.req.numDirs + fileStore.req.numFiles;

  // More items are displayed than the total
  if (showLimit.value >= totalItems && !fit) return;

  const container = document.querySelector("main") ?? document.documentElement;
  const viewHeight = container.clientHeight || window.innerHeight;

  if (itemWeight.value <= 0) {
    showLimit.value = Math.max(50, totalItems);
    return;
  }

  // Quantity of items needed to fill 2x of the viewport height
  const showQuantity = Math.ceil(
    (viewHeight + viewHeight * 2) / itemWeight.value
  );

  // Less items to display than current
  if (showLimit.value > showQuantity && !fit) return;

  // Set the number of displayed items
  showLimit.value = showQuantity > totalItems ? totalItems : showQuantity;
};

const revealPreviousItem = () => {
  if (!fileStore.req || !fileStore.oldReq) return;

  const index = fileStore.selected[0];
  if (index === undefined) return;

  showLimit.value =
    index + Math.ceil((window.innerHeight * 2) / itemWeight.value);

  nextTick(() => {
    const items = document.querySelectorAll("#listing .item");
    items[index].scrollIntoView({ block: "center" });
  });

  return true;
};

const showContextMenu = (event: MouseEvent) => {
  event.preventDefault();
  const target = event.target as HTMLElement;
  isContextMenuOnItem.value = !!target.closest(".item");
  isContextMenuOnFolder.value = !!target.closest('[data-dir="true"]');
  if (props.isTrash && isContextMenuOnItem.value) {
    const itemEl = target.closest(".item") as HTMLElement | null;
    const itemPath = itemEl?.getAttribute("data-path");
    const items = fileStore.req?.items ?? [];
    const found = itemPath ? items.find((it) => it.path === itemPath) : null;
    if (found && !fileStore.selected.includes(found.index)) {
      fileStore.selected = [found.index];
    }
  } else if (!isContextMenuOnItem.value) {
    fileStore.selected = [];
    isContextMenuOnFolder.value = false;
  }
  isContextMenuVisible.value = true;
  showColorPicker.value = false;
  contextMenuPos.value = {
    x: event.clientX + 8,
    y: event.clientY + 8,
  };
};

const hideContextMenu = () => {
  isContextMenuVisible.value = false;
};

const isStarredContextItem = computed(() => {
  starVersion.value; // subscribe so it re-evaluates after toggleStarred
  const items = fileStore.req?.items;
  if (!items || fileStore.selected.length === 0) return false;
  return fileStore.selected.some((idx) => {
    const item = items[idx];
    return item ? isStarred(item.url) : false;
  });
});

const toggleStar = () => {
  hideContextMenu();
  const items = fileStore.req?.items;
  if (!items || fileStore.selected.length === 0) return;
  for (const idx of fileStore.selected) {
    const item = items[idx];
    if (item) {
      toggleStarred({ url: item.url, name: item.name, type: item.type });
    }
  }
};

/**
 * Move one item, resolving name collisions client-side with `(n)` suffixes so
 * the caller always knows the final destination (needed for the trash sidecar).
 * Returns the full `/files/<id>/...` destination URL.
 */
const moveWithUniqueName = async (
  fromUrl: string,
  destDirUrl: string,
  name: string,
  isDir: boolean
): Promise<string> => {
  const slash = isDir ? "/" : "";
  for (let attempt = 0; attempt < 1000; attempt++) {
    const candidate = withVersionSuffix(name, attempt);
    const to = `${destDirUrl}${candidate}${slash}`;
    try {
      await api.move([{ from: fromUrl, to }]);
      return to;
    } catch (e: any) {
      if (e instanceof StatusError && e.status === 409) continue;
      throw e;
    }
  }
  throw new Error(`could not find a free name for ${name}`);
};

const moveToTrash = async () => {
  hideContextMenu();
  const items = fileStore.req?.items;
  if (!items || fileStore.selected.length === 0) return;

  // Never trash from inside the trash view, and never trash the trash itself.
  const candidates = fileStore.selected
    .map((idx) => items[idx])
    .filter(
      (item) =>
        Boolean(item) && item.name !== ".Trash" && !isTrashPath(item.path ?? "")
    );
  if (candidates.length === 0) return;

  const batchId = Date.now();
  const sourceId = String(route.params.sourceId ?? sourceStore.activeId ?? "0");
  const filesRoot = `${trashFilesRoot(filesBase.value)}/`;
  const infoRoot = `${trashInfoRoot(filesBase.value)}/`;

  // Ensure the mirrored trash roots exist (ignore when they already do).
  for (const dir of [filesRoot, infoRoot]) {
    try {
      await api.post(dir);
    } catch {
      // The move below surfaces any real error.
    }
  }

  let trashed = 0;
  let lastError: any = null;
  for (const item of candidates) {
    // Mirror the original folder structure inside .Trash/files so files deleted
    // from the same folder land together instead of scattered at the root.
    const originalPath: string = item.path;
    const destDirUrl =
      parentDir(originalPath) === "/"
        ? filesRoot
        : `${filesRoot}${parentDir(originalPath).slice(1)}/`;
    try {
      await api.post(destDirUrl);
    } catch {
      // May already exist — the move surfaces real errors.
    }
    try {
      const destUrl = await moveWithUniqueName(
        item.url,
        destDirUrl,
        item.name,
        item.isDir
      );
      // Sidecar keyed by the ACTUAL trash destination (post-rename), holding the
      // original path for an exact restore. Best effort: trashing must succeed
      // even if the sidecar write fails (fallback restores to the source root).
      try {
        const trashSourcePath = trashSourcePathFromUrl(
          destUrl,
          filesBase.value
        );
        const infoUrl =
          trashSourcePath !== null
            ? trashInfoUrl(filesBase.value, trashSourcePath)
            : null;
        if (infoUrl !== null) {
          try {
            await api.post(`${parentDir(infoUrl)}/`);
          } catch {
            // Parent may already exist.
          }
          await api.post(
            infoUrl,
            new Blob(
              [
                JSON.stringify(
                  buildTrashInfo({
                    originalPath,
                    name: item.name,
                    isDir: item.isDir,
                    deletedAt: new Date().toISOString(),
                    sourceId,
                    batchId,
                  })
                ),
              ],
              { type: "application/json" }
            ),
            true
          );
        }
      } catch (sidecarError) {
        console.warn(
          "trash sidecar write failed for",
          originalPath,
          sidecarError
        );
      }
      trashed++;
    } catch (e: any) {
      lastError = e;
    }
  }

  if (trashed > 0) fileStore.reload = true;
  if (lastError !== null && trashed === 0) {
    $showError(lastError);
  } else if (lastError !== null) {
    console.warn("some items could not be moved to trash", lastError);
  }
};

const openColorPicker = () => {
  // Position the color picker to the right of the context menu
  const menuWidth = 214; // approximate context menu width
  let x = contextMenuPos.value.x + menuWidth + 8;
  let y = contextMenuPos.value.y;

  // Keep within viewport
  if (x + 200 > window.innerWidth) {
    x = contextMenuPos.value.x - 200 - 8;
  }
  if (y + 250 > window.innerHeight) {
    y = window.innerHeight - 260;
  }

  colorPickerPos.value = { x, y };
  showColorPicker.value = true;
};

const handleColorSelect = async (color: string) => {
  if (!fileStore.selected || fileStore.selected.length === 0) return;
  const selectedIndex = fileStore.selected[0];
  const items = fileStore.req?.items;
  if (!items || selectedIndex === undefined) return;
  const selectedItem = items[selectedIndex];
  if (!selectedItem || !selectedItem.isDir) return;

  const path = selectedItem.path;
  try {
    const user = authStore.user;
    if (!user) return;

    const folderColors = { ...(user.folderColors || {}) };
    folderColors[path] = color;

    await users.update({ id: user.id, folderColors }, ["folderColors"]);
    if (authStore.user) {
      authStore.user.folderColors = folderColors;
    }
  } catch (error) {
    console.error("Failed to update folder color:", error);
  }
  showColorPicker.value = false;
  hideContextMenu();
};

const handleColorRemove = async () => {
  if (!fileStore.selected || fileStore.selected.length === 0) return;
  const selectedIndex = fileStore.selected[0];
  const items = fileStore.req?.items;
  if (!items || selectedIndex === undefined) return;
  const selectedItem = items[selectedIndex];
  if (!selectedItem || !selectedItem.isDir) return;

  const path = selectedItem.path;
  try {
    const user = authStore.user;
    if (!user) return;

    const folderColors = { ...(user.folderColors || {}) };
    delete folderColors[path];

    await users.update({ id: user.id, folderColors }, ["folderColors"]);
    if (authStore.user) {
      authStore.user.folderColors = folderColors;
    }
  } catch (error) {
    console.error("Failed to remove folder color:", error);
  }
  showColorPicker.value = false;
  hideContextMenu();
};

const getCurrentFolderColor = (): string | undefined => {
  if (!fileStore.selected || fileStore.selected.length === 0) return undefined;
  const selectedIndex = fileStore.selected[0];
  const items = fileStore.req?.items;
  if (!items || selectedIndex === undefined) return undefined;
  const selectedItem = items[selectedIndex];
  if (!selectedItem || !selectedItem.isDir) return undefined;

  const user = authStore.user;
  if (!user || !user.folderColors) return undefined;

  return user.folderColors[selectedItem.path];
};

const handleEmptyAreaClick = (e: MouseEvent) => {
  const target = e.target;
  if (!(target instanceof HTMLElement)) return;

  if (target.dataset.clearOnClick === "true") {
    fileStore.selected = [];
  }
};

/** Original location shown under flat-view trash rows (mirror → parent, legacy → root). */
const trashSubtitle = (item: { path: string }): string => {
  const original = originalPathFromTrash(item.path ?? "");
  if (original !== null) return parentDir(original);
  return "/";
};

const noop = () => {};

const handleTrashOpen = (url: string, isDir: boolean) => {
  if (isDir) {
    emit("navigate-to", url);
  } else {
    router.push({ path: url });
  }
};

const trashRestore = () => {
  hideContextMenu();
  emit("restore");
};

const trashDeletePermanent = () => {
  hideContextMenu();
  emit("delete-permanent");
};
</script>
<style scoped>
#listing {
  min-height: calc(100vh - 8rem);
}

.fb-trash-breadcrumb {
  display: flex;
  align-items: center;
  gap: 2px;
  min-width: 0;
  overflow: hidden;
}

.fb-trash-breadcrumb-item {
  background: none;
  border: none;
  color: var(--dim);
  cursor: pointer;
  font: inherit;
  font-size: 0.95em;
  padding: 2px 4px;
  border-radius: 4px;
  white-space: nowrap;
}

.fb-trash-breadcrumb-item:hover {
  color: var(--text);
  background: var(--hover);
}

.fb-trash-breadcrumb-item--last {
  color: var(--text);
  cursor: default;
}

.fb-trash-breadcrumb-item--last:hover {
  background: none;
}

.fb-trash-breadcrumb-sep {
  margin-right: 4px;
  color: var(--border-strong);
}

.fb-trash-view-toggle {
  display: flex;
  align-items: center;
  gap: 2px;
  background: var(--hover);
  border-radius: 10px;
  padding: 3px;
  margin-left: 12px;
  flex: 0 0 auto;
}

.fb-trash-view-toggle .fb-tbtn {
  border: none;
  background: transparent;
  height: 30px;
  padding: 0 10px;
  font-size: 12.5px;
  border-radius: 7px;
  white-space: nowrap;
}

.file-selection-margin-bottom {
  margin-bottom: 3.5rem;
}

.fb-color-picker-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
}

.fb-color-picker-container {
  position: fixed;
  z-index: 1001;
}
</style>
