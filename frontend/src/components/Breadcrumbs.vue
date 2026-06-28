<template>
  <div class="fb-breadcrumbs">
    <component
      :is="element"
      :to="base || '/files'"
      class="fb-crumb"
      :class="{ 'fb-crumb--active': items.length === 0 }"
      :aria-label="t('files.home')"
    >
      {{ t("sidebar.myFiles") }}
    </component>

    <span v-for="(link, index) in items" :key="index" class="fb-crumb-seg">
      <svg
        class="fb-crumb-chevron"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path d="M9 6l6 6-6 6" />
      </svg>
      <component
        :is="element"
        :to="link.url"
        class="fb-crumb"
        :class="{ 'fb-crumb--active': index === items.length - 1 }"
      >
        {{ link.name }}
      </component>
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";

const { t } = useI18n();
const route = useRoute();

const props = defineProps<{
  base: string;
  noLink?: boolean;
}>();

const items = computed(() => {
  const relativePath = route.path.replace(props.base, "");
  const parts = relativePath.split("/");

  if (parts[0] === "") parts.shift();
  if (parts[parts.length - 1] === "") parts.pop();

  const breadcrumbs: BreadCrumb[] = [];

  for (let i = 0; i < parts.length; i++) {
    if (i === 0) {
      breadcrumbs.push({
        name: decodeURIComponent(parts[i]),
        url: props.base + "/" + parts[i] + "/",
      });
    } else {
      breadcrumbs.push({
        name: decodeURIComponent(parts[i]),
        url: breadcrumbs[i - 1].url + parts[i] + "/",
      });
    }
  }

  if (breadcrumbs.length > 3) {
    while (breadcrumbs.length !== 4) breadcrumbs.shift();
    breadcrumbs[0].name = "...";
  }

  return breadcrumbs;
});

const element = computed(() => (props.noLink ? "span" : "router-link"));
</script>

<style>
.fb-breadcrumbs {
  display: flex;
  align-items: center;
  gap: 0;
  font-size: 16px;
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
}

.fb-crumb-seg {
  display: flex;
  align-items: center;
}

.fb-crumb-chevron {
  width: 15px;
  height: 15px;
  flex: 0 0 auto;
  color: var(--faint);
  margin: 0 2px;
}

.fb-crumb {
  font-size: 16px;
  font-weight: 500;
  color: var(--dim);
  padding: 5px 8px;
  border-radius: 7px;
  border: none;
  background: transparent;
  cursor: pointer;
  text-decoration: none;
  white-space: nowrap;
  display: inline-block;
  transition:
    background 0.1s,
    color 0.1s;
  font-family: inherit;
}

.fb-crumb:hover {
  background: var(--hover);
  color: var(--text);
}

.fb-crumb--active {
  font-weight: 650;
  color: var(--text);
  cursor: default;
  pointer-events: none;
}
</style>
