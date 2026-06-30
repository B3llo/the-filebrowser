<template>
  <div id="editor-container">
    <header-bar>
      <span class="fb-preview-type">{{ editorTypeLabel }}</span>
      <title>{{ fileName }}</title>

      <template #actions>
        <action
          v-if="
            authStore.user?.perm.modify &&
            fileStore.req?.type !== 'textImmutable'
          "
          id="save-button"
          fb-icon="save"
          size="18px"
          :label="t('buttons.save')"
          @action="save()"
        />
        <action
          v-if="isMarkdownFile"
          fb-icon="eye"
          size="18px"
          :label="isPreview ? t('buttons.editAsText') : t('buttons.preview')"
          @action="togglePreview()"
        />
        <action
          fb-icon="x"
          size="18px"
          :label="t('buttons.close')"
          @action="close()"
        />
      </template>
    </header-bar>

    <div class="loading delayed" v-if="layoutStore.loading">
      <div class="spinner">
        <div class="bounce1"></div>
        <div class="bounce2"></div>
        <div class="bounce3"></div>
      </div>
    </div>
    <template v-else>
      <div
        v-show="isPreview && isMarkdownFile"
        id="preview-container"
        class="md_preview markdown-body"
        v-html="previewContent"
      ></div>
      <form v-show="!isPreview || !isMarkdownFile" id="editor"></form>
    </template>
  </div>
</template>

<script setup lang="ts">
import { files as api } from "@/api";
import buttons from "@/utils/buttons";
import url from "@/utils/url";
import ace, { Ace, version as ace_version } from "ace-builds";
import "ace-builds/src-noconflict/ext-language_tools";
import modelist from "ace-builds/src-noconflict/ext-modelist";
import DOMPurify from "dompurify";

import Action from "@/components/header/Action.vue";
import HeaderBar from "@/components/header/HeaderBar.vue";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { getEditorTheme } from "@/utils/theme";
import { marked } from "marked";
import markedKatex from "marked-katex-extension";
import {
  computed,
  inject,
  onBeforeUnmount,
  onMounted,
  ref,
  watchEffect,
} from "vue";
import { useI18n } from "vue-i18n";
import { onBeforeRouteUpdate, useRoute, useRouter } from "vue-router";

const $showError = inject<IToastError>("$showError")!;

const fileStore = useFileStore();
const authStore = useAuthStore();
const layoutStore = useLayoutStore();

const { t } = useI18n();

const route = useRoute();
const router = useRouter();

const editor = ref<Ace.Editor | null>(null);
const fontSize = ref(parseInt(localStorage.getItem("editorFontSize") || "14"));

// Start directly in edit mode (never open in preview first — #2).
const isPreview = ref(false);
const previewContent = ref("");

const isMarkdownFile =
  (fileStore.req?.extension.toLowerCase() === ".md" ||
    fileStore.req?.extension.toLowerCase() === ".markdown") ??
  false;

const katexOptions = {
  output: "mathml" as const,
  throwOnError: false,
};
marked.use(markedKatex(katexOptions));

const fileName = computed(() => fileStore.req?.name ?? "");

const editorTypeLabel = computed(() => {
  if (!fileStore.req) return "";
  if (isMarkdownFile) return "Markdown";
  const modeName = modelist.getModeForPath(fileStore.req.name).mode as string;
  const lang = modeName.split("/").pop() ?? "";
  return lang.charAt(0).toUpperCase() + lang.slice(1) || "Text";
});

onMounted(() => {
  window.addEventListener("keydown", keyEvent);
  window.addEventListener("beforeunload", handlePageChange);

  const fileContent = fileStore.req?.content || "";

  watchEffect(async () => {
    if (isMarkdownFile && isPreview.value) {
      const new_value = editor.value?.getValue() || "";
      try {
        const stripped = new_value.replace(/^---\n[\s\S]*?\n---\n?/, "");
        previewContent.value = DOMPurify.sanitize(await marked(stripped));
      } catch (error) {
        console.error("Failed to convert content to HTML:", error);
        previewContent.value = "";
      }
    }
  });

  ace.config.set(
    "basePath",
    `https://cdn.jsdelivr.net/npm/ace-builds@${ace_version}/src-min-noconflict/`
  );

  if (!layoutStore.loading) {
    initEditor(fileContent);
  } else {
    const unwatch = watchEffect(() => {
      if (!layoutStore.loading) {
        setTimeout(() => {
          initEditor(fileContent);
          unwatch();
        }, 50);
      }
    });
  }
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", keyEvent);
  window.removeEventListener("beforeunload", handlePageChange);
  editor.value?.destroy();
});

onBeforeRouteUpdate((to, from, next) => {
  if (editor.value?.session.getUndoManager().isClean()) {
    next();
    return;
  }

  layoutStore.showHover({
    prompt: "discardEditorChanges",
    confirm: (event: Event) => {
      event.preventDefault();
      next();
    },
    saveAction: async () => {
      await save();
      next();
    },
  });
});

const initEditor = (fileContent: string) => {
  editor.value = ace.edit("editor", {
    value: fileContent,
    showPrintMargin: false,
    readOnly: fileStore.req?.type === "textImmutable",
    theme: getEditorTheme(authStore.user?.aceEditorTheme ?? ""),
    mode: modelist.getModeForPath(fileStore.req!.name).mode,
    wrap: true,
    enableBasicAutocompletion: true,
    enableLiveAutocompletion: true,
    enableSnippets: true,
  });

  editor.value.setFontSize(fontSize.value);
  editor.value.setOption(
    "fontFamily",
    '"SF Mono", "Monaco", "Menlo", "Ubuntu Mono", "Cascadia Code", "Consolas", "Liberation Mono", monospace'
  );
  editor.value.focus();
};

const keyEvent = (event: KeyboardEvent) => {
  if (event.code === "Escape") {
    close();
  }

  if (!event.ctrlKey && !event.metaKey) {
    return;
  }

  if (event.key !== "s") {
    return;
  }

  event.preventDefault();
  save();
};

const handlePageChange = (event: BeforeUnloadEvent) => {
  if (!editor.value?.session.getUndoManager().isClean()) {
    event.preventDefault();
    event.returnValue = true;
  }
};

const save = async (throwError?: boolean) => {
  const button = "save";
  buttons.loading("save");

  try {
    await api.put(route.path, editor.value?.getValue());
    editor.value?.session.getUndoManager().markClean();
    buttons.success(button);
  } catch (e: any) {
    buttons.done(button);
    $showError(e);
    if (throwError) throw e;
  }
};

const close = () => {
  if (!editor.value?.session.getUndoManager().isClean()) {
    layoutStore.showHover({
      prompt: "discardEditorChanges",
      confirm: (event: Event) => {
        event.preventDefault();
        editor.value?.session.getUndoManager().reset();
        finishClose();
      },
      saveAction: async () => {
        try {
          await save(true);
          finishClose();
        } catch {}
      },
    });
    return;
  }
  finishClose();
};

const finishClose = () => {
  const uri = url.removeLastDir(route.path) + "/";
  router.push({ path: uri });
};

const togglePreview = () => {
  isPreview.value = !isPreview.value;
};
</script>

<style scoped>
#editor-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

#editor,
#preview-container {
  flex: 1;
  overflow: auto;
}
</style>
