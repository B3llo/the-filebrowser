<template>
  <div class="code-preview">
    <pre class="code-content"><code v-html="highlightedCode"></code></pre>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import hljs from "highlight.js";
import "highlight.js/styles/github-dark.css";

interface Props {
  content: string;
  language: string;
  filename?: string;
}

const props = withDefaults(defineProps<Props>(), {
  filename: "",
});

const highlightedCode = ref("");

const highlightCode = async () => {
  if (!props.content) {
    highlightedCode.value = "";
    return;
  }

  try {
    let lang = props.language;
    // Map some languages to highlight.js names
    const langMap: Record<string, string> = {
      typescript: "typescript",
      javascript: "javascript",
      python: "python",
      go: "go",
      rust: "rust",
      java: "java",
      kotlin: "kotlin",
      c: "c",
      cpp: "cpp",
      csharp: "csharp",
      php: "php",
      ruby: "ruby",
      swift: "swift",
      dart: "dart",
      lua: "lua",
      perl: "perl",
      r: "r",
      scala: "scala",
      clojure: "clojure",
      bash: "bash",
      powershell: "powershell",
      html: "html",
      css: "css",
      scss: "scss",
      json: "json",
      xml: "xml",
      yaml: "yaml",
      toml: "ini",
      sql: "sql",
      graphql: "graphql",
      markdown: "markdown",
    };

    lang = langMap[lang.toLowerCase()] || "plaintext";

    if (lang === "plaintext" || !hljs.getLanguage(lang)) {
      // Fallback to auto-detection
      const result = hljs.highlightAuto(props.content);
      highlightedCode.value = result.value;
    } else {
      const result = hljs.highlight(props.content, { language: lang });
      highlightedCode.value = result.value;
    }
  } catch (e) {
    console.error("Failed to highlight code:", e);
    // Escape HTML and display as plain text
    highlightedCode.value = props.content
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;");
  }
};

onMounted(() => {
  highlightCode();
});

watch(
  () => props.content,
  () => {
    highlightCode();
  }
);

watch(
  () => props.language,
  () => {
    highlightCode();
  }
);
</script>

<style scoped>
.code-preview {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--code-bg, #1e1e1e);
  border-radius: 8px;
  overflow: hidden;
}

.code-header {
  padding: 12px 16px;
  background: var(--code-header-bg, #2d2d2d);
  border-bottom: 1px solid var(--border-color, #404040);
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.language-badge {
  background: var(--primary-color, #4a9eff);
  color: white;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  text-transform: uppercase;
}

.line-count {
  color: var(--text-secondary, #888);
  font-size: 12px;
}

.code-content {
  flex: 1;
  margin: 0;
  padding: 16px;
  overflow: auto;
  font-family: "Monaco", "Menlo", "Ubuntu Mono", monospace;
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-primary, #d4d4d4);
  background: transparent;
  white-space: pre;
  tab-size: 2;
}

.code-content code {
  font-family: inherit;
}

/* Custom scrollbar */
.code-content::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.code-content::-webkit-scrollbar-track {
  background: var(--code-header-bg, #2d2d2d);
}

.code-content::-webkit-scrollbar-thumb {
  background: var(--border-color, #404040);
  border-radius: 4px;
}

.code-content::-webkit-scrollbar-thumb:hover {
  background: var(--text-secondary, #666);
}
</style>
