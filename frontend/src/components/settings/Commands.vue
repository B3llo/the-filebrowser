<template>
  <div>
    <h3
      style="
        margin: 20px 0 8px;
        font-size: 14px;
        font-weight: 600;
        color: var(--text);
      "
    >
      {{ $t("settings.userCommands") }}
    </h3>
    <p class="small" style="margin-bottom: 12px">
      {{ $t("settings.userCommandsHelp") }} <code>git svn hg</code>.
    </p>
    <input class="input input--block" type="text" v-model.trim="raw" />
  </div>
</template>

<script>
export default {
  name: "commands",
  props: ["commands"],
  computed: {
    raw: {
      get() {
        return this.commands.join(" ");
      },
      set(value) {
        if (value !== "") {
          this.$emit("update:commands", value.split(" "));
        } else {
          this.$emit("update:commands", []);
        }
      },
    },
  },
};
</script>
