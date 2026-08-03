<template>
  <div
    class="card floating upload-card"
    @dragover.prevent="onDragOver"
    @dragleave="onDragLeave"
    @drop.prevent="onDrop"
  >
    <div class="card-title">
      <h2>{{ t("prompts.upload") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ t("prompts.uploadMessage") }}</p>
    </div>

    <div
      class="upload-dropzone"
      :class="{ 'upload-dropzone--over': dropActive }"
      aria-hidden="true"
    >
      <i class="material-icons">file_upload</i>
      <span>{{ t("prompts.uploadDropHint") }}</span>
    </div>

    <div class="card-action full">
      <div
        @click="uploadFile"
        @keypress.enter="uploadFile"
        class="action"
        id="focus-prompt"
        tabindex="1"
      >
        <i class="material-icons">insert_drive_file</i>
        <div class="title">{{ t("buttons.file") }}</div>
      </div>
      <div
        @click="uploadFolder"
        @keypress.enter="uploadFolder"
        class="action"
        tabindex="2"
      >
        <i class="material-icons">folder</i>
        <div class="title">{{ t("buttons.folder") }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import { useLayoutStore } from "@/stores/layout";

import * as upload from "@/utils/upload";

const { t } = useI18n();
const route = useRoute();

const layoutStore = useLayoutStore();

const dropActive = ref<boolean>(false);

const onDragOver = (event: DragEvent) => {
  if (event.dataTransfer) event.dataTransfer.dropEffect = "copy";
  dropActive.value = true;
};

const onDragLeave = () => {
  dropActive.value = false;
};

const onDrop = async (event: DragEvent) => {
  dropActive.value = false;
  const dt = event.dataTransfer;
  if (dt === null || dt.files.length <= 0) return;

  const files = (await upload.scanFiles(dt)) as UploadList;
  if (files.length === 0) return;

  const path = route.path.endsWith("/") ? route.path : route.path + "/";

  const conflict = await upload.checkConflict(files, path);

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
          upload.handleFiles(files, path);
        }
      },
    });

    return;
  }

  upload.handleFiles(files, path);
};

// TODO: this is a copy of the same function in FileListing.vue
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
          upload.handleFiles(uploadFiles, path);
        }
      },
    });

    return;
  }

  upload.handleFiles(uploadFiles, path);
};

const openUpload = (isFolder: boolean) => {
  const input = document.createElement("input");
  input.type = "file";
  input.multiple = true;
  input.webkitdirectory = isFolder;
  // TODO: call the function in FileListing.vue instead
  input.onchange = uploadInput;
  input.click();
};

const uploadFile = () => {
  openUpload(false);
};
const uploadFolder = () => {
  openUpload(true);
};
</script>
