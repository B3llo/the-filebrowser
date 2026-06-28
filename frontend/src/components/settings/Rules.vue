<template>
  <div class="fb-settings-rules">
    <div v-for="(rule, index) in rules" :key="index" class="fb-settings-rule">
      <div class="fb-settings-rule-checkbox">
        <input type="checkbox" v-model="rule.regex" :id="'regex-' + index" />
        <label :for="'regex-' + index">Regex</label>
      </div>

      <div class="fb-settings-rule-checkbox">
        <input type="checkbox" v-model="rule.allow" :id="'allow-' + index" />
        <label :for="'allow-' + index">Allow</label>
      </div>

      <div class="fb-settings-rule-input">
        <input
          @keypress.enter.prevent
          type="text"
          v-if="rule.regex"
          v-model="rule.regexp.raw"
          :placeholder="$t('settings.insertRegex')"
          class="input"
          style="width: 100%"
        />
        <input
          @keypress.enter.prevent
          type="text"
          v-else
          v-model="rule.path"
          :placeholder="$t('settings.insertPath')"
          class="input"
          style="width: 100%"
        />
      </div>

      <button
        class="fb-settings-rule-remove"
        @click="remove($event, index)"
        :aria-label="$t('buttons.delete')"
      >
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          style="width: 14px; height: 14px"
        >
          <line x1="18" y1="6" x2="6" y2="18" />
          <line x1="6" y1="6" x2="18" y2="18" />
        </svg>
      </button>
    </div>

    <div>
      <button class="button button--flat" @click="create" default="false">
        {{ $t("buttons.new") }}
      </button>
    </div>
  </div>
</template>

<script>
export default {
  name: "rules-textarea",
  props: ["rules"],
  methods: {
    remove(event, index) {
      event.preventDefault();
      const rules = [...this.rules];
      rules.splice(index, 1);
      this.$emit("update:rules", [...rules]);
    },
    create(event) {
      event.preventDefault();

      this.$emit("update:rules", [
        ...this.rules,
        {
          allow: true,
          path: "",
          regex: false,
          regexp: {
            raw: "",
          },
        },
      ]);
    },
  },
};
</script>
