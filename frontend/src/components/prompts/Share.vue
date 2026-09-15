<template>
  <div class="card floating" id="share">
    <div class="card-title">
      <h2>{{ $t("buttons.share") }}</h2>
    </div>

    <div v-if="showShareViaOS" class="card-content">
      <button
        class="button button--block button--blue fb-share-via-os"
        :aria-label="$t('buttons.shareViaOS')"
        :title="$t('buttons.shareViaOS')"
        @click="shareViaOS"
      >
        <i class="material-icons">ios_share</i>
        <span>{{ $t("buttons.shareViaOS") }}</span>
      </button>
      <p class="fb-share-via-os-hint">{{ $t("prompts.shareViaOSHint") }}</p>
    </div>

    <div class="fb-share-tabs">
      <button
        class="button button--flat"
        :class="{ 'button--blue': tab === 'links' }"
        @click="tab = 'links'"
      >
        {{ $t("grants.links") }}
      </button>
      <button
        class="button button--flat"
        :class="{ 'button--blue': tab === 'people' }"
        @click="() => switchGrantTab()"
      >
        {{ $t("grants.people") }}
      </button>
    </div>

    <template v-if="tab === 'people'">
      <div class="card-content">
        <div v-if="grants.length > 0">
          <table>
            <tbody>
              <tr v-for="grant in grants" :key="grant.id">
                <td>
                  <span class="fb-grant-user">{{
                    grant.granteeUsername || grant.granteeID
                  }}</span>
                  <span class="fb-grant-role">{{ grantRoleLabel(grant) }}</span>
                </td>
                <td class="small">
                  <button
                    class="action"
                    @click="deleteGrant($event, grant)"
                    :aria-label="$t('grants.revoke')"
                    :title="$t('grants.revoke')"
                  >
                    <i class="material-icons">delete</i>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <p v-else>{{ $t("grants.onlyYou") }}</p>

        <p>{{ $t("grants.searchPlaceholder") }}</p>
        <input
          class="input input--block"
          type="text"
          v-model.trim="searchQuery"
          @input="onSearchInput"
        />
        <div v-if="searchResults.length > 0" class="fb-grant-results">
          <button
            v-for="u in searchResults"
            :key="u.id"
            class="button button--flat button--block"
            @click="() => pickUser(u)"
          >
            {{ u.username }}
          </button>
        </div>
        <div v-if="selectedUser" class="fb-grant-picked">
          <span>{{ selectedUser.username }}</span>
          <div class="fb-settings-seg">
            <button
              class="fb-settings-seg-btn"
              :class="{ 'fb-settings-seg-btn--active': grantRole === 'viewer' }"
              @click="grantRole = 'viewer'"
            >
              {{ $t("grants.viewer") }}
            </button>
            <button
              class="fb-settings-seg-btn"
              :class="{ 'fb-settings-seg-btn--active': grantRole === 'editor' }"
              @click="grantRole = 'editor'"
            >
              {{ $t("grants.editor") }}
            </button>
          </div>
        </div>
      </div>

      <div class="card-action">
        <button
          class="button button--flat button--grey"
          @click="closeHovers"
          :aria-label="$t('buttons.close')"
          :title="$t('buttons.close')"
        >
          {{ $t("buttons.close") }}
        </button>
        <button
          class="button button--flat button--blue"
          @click="submitGrant"
          :disabled="!selectedUser"
          :aria-label="$t('buttons.share')"
          :title="$t('buttons.share')"
        >
          {{ $t("buttons.share") }}
        </button>
      </div>
    </template>

    <template v-else-if="listing">
      <div class="card-content">
        <table>
          <tr>
            <th>#</th>
            <th>{{ $t("settings.shareDuration") }}</th>
            <th></th>
            <th></th>
            <th></th>
          </tr>

          <tr v-for="link in links" :key="link.hash">
            <td>{{ link.hash }}</td>
            <td>
              <template v-if="link.expire !== 0">{{
                humanTime(link.expire)
              }}</template>
              <template v-else>{{ $t("permanent") }}</template>
            </td>
            <td class="small">
              <button
                class="action"
                :aria-label="$t('buttons.copyToClipboard')"
                :title="$t('buttons.copyToClipboard')"
                @click="copyToClipboard(buildLink(link))"
              >
                <i class="material-icons">content_paste</i>
              </button>
            </td>
            <td class="small">
              <button
                class="action"
                :aria-label="$t('buttons.copyDownloadLinkToClipboard')"
                :title="$t('buttons.copyDownloadLinkToClipboard')"
                :disabled="!!link.password_hash"
                @click="copyToClipboard(buildDownloadLink(link))"
              >
                <i class="material-icons">content_paste_go</i>
              </button>
            </td>
            <td class="small">
              <button
                class="action"
                @click="deleteLink($event, link)"
                :aria-label="$t('buttons.delete')"
                :title="$t('buttons.delete')"
              >
                <i class="material-icons">delete</i>
              </button>
            </td>
          </tr>
        </table>
      </div>

      <div class="card-action">
        <button
          class="button button--flat button--grey"
          @click="closeHovers"
          :aria-label="$t('buttons.close')"
          :title="$t('buttons.close')"
          tabindex="2"
        >
          {{ $t("buttons.close") }}
        </button>
        <button
          id="focus-prompt"
          class="button button--flat button--blue"
          @click="() => switchListing()"
          :aria-label="$t('buttons.new')"
          :title="$t('buttons.new')"
          tabindex="1"
        >
          {{ $t("buttons.new") }}
        </button>
      </div>
    </template>

    <template v-else>
      <div class="card-content">
        <p>{{ $t("settings.shareDuration") }}</p>
        <div class="fb-share-duration">
          <input
            class="input fb-share-duration-input"
            type="number"
            :min="0"
            :max="2147483647"
            v-model.number="time"
            @keyup.enter="submit"
            tabindex="1"
          />
          <div class="fb-settings-seg fb-share-duration-seg">
            <button
              class="fb-settings-seg-btn"
              :class="{ 'fb-settings-seg-btn--active': unit === 'seconds' }"
              @click="unit = 'seconds'"
              tabindex="2"
            >
              {{ $t("time.seconds") }}
            </button>
            <button
              class="fb-settings-seg-btn"
              :class="{ 'fb-settings-seg-btn--active': unit === 'minutes' }"
              @click="unit = 'minutes'"
              tabindex="2"
            >
              {{ $t("time.minutes") }}
            </button>
            <button
              class="fb-settings-seg-btn"
              :class="{ 'fb-settings-seg-btn--active': unit === 'hours' }"
              @click="unit = 'hours'"
              tabindex="2"
            >
              {{ $t("time.hours") }}
            </button>
            <button
              class="fb-settings-seg-btn"
              :class="{ 'fb-settings-seg-btn--active': unit === 'days' }"
              @click="unit = 'days'"
              tabindex="2"
            >
              {{ $t("time.days") }}
            </button>
          </div>
        </div>
        <p>{{ $t("prompts.optionalPassword") }}</p>
        <input
          class="input input--block"
          type="password"
          v-model.trim="password"
          tabindex="3"
        />
      </div>

      <div class="card-action">
        <button
          class="button button--flat button--grey"
          @click="() => switchListing()"
          :aria-label="$t('buttons.cancel')"
          :title="$t('buttons.cancel')"
          tabindex="5"
        >
          {{ $t("buttons.cancel") }}
        </button>
        <button
          id="focus-prompt"
          class="button button--flat button--blue"
          @click="submit"
          :aria-label="$t('buttons.share')"
          :title="$t('buttons.share')"
          tabindex="4"
        >
          {{ $t("buttons.share") }}
        </button>
      </div>
    </template>
  </div>
</template>

<script>
import { mapActions, mapState } from "pinia";
import { useFileStore } from "@/stores/file";
import { useAuthStore } from "@/stores/auth";
import * as api from "@/api/index";
import dayjs from "dayjs";
import { useLayoutStore } from "@/stores/layout";
import { copy } from "@/utils/clipboard";
import { canShareFiles, shareViaOS } from "@/utils/nativeShare";
import { removePrefix } from "@/api/utils";

export default {
  name: "share",
  props: {
    initialTab: {
      type: String,
      default: "links",
    },
  },
  data: function () {
    return {
      time: 0,
      unit: "hours",
      links: [],
      clip: null,
      password: "",
      listing: true,
      tab: "links",
      grants: [],
      searchQuery: "",
      searchResults: [],
      selectedUser: null,
      grantRole: "viewer",
      searchTimer: null,
    };
  },
  inject: ["$showError", "$showSuccess"],
  computed: {
    ...mapState(useFileStore, [
      "req",
      "selected",
      "selectedCount",
      "isListing",
    ]),
    ...mapState(useAuthStore, ["user"]),
    showShareViaOS() {
      return Boolean(this.user?.perm?.download) && canShareFiles();
    },
    shareTargets() {
      if (this.isListing && this.selectedCount === 1) {
        const item = this.req.items[this.selected[0]];
        return [{ name: item.name, url: item.url, isDir: item.isDir }];
      }

      const path = this.$route.path;
      const segments = path.split("/").filter(Boolean);
      const name =
        segments.length > 0
          ? decodeURIComponent(segments[segments.length - 1])
          : "file";

      return [{ name, url: path, isDir: false }];
    },
    url() {
      if (!this.isListing) {
        return this.$route.path;
      }

      if (this.selectedCount === 0 || this.selectedCount > 1) {
        // This shouldn't happen.
        return;
      }

      return this.req.items[this.selected[0]].url;
    },
  },
  watch: {
    initialTab(newVal) {
      if (newVal === "people" || newVal === "links") {
        this.tab = newVal;
        if (newVal === "people") {
          this.loadGrants().catch((e) => this.$showError(e));
        }
      }
    },
  },
  async beforeMount() {
    if (this.initialTab === "people") {
      this.tab = "people";
    }
    try {
      const links = await api.share.get(this.url);
      this.links = links;
      this.sort();

      if (this.links.length == 0) {
        this.listing = false;
      }
    } catch (e) {
      this.$showError(e);
    }
    try {
      await this.loadGrants();
    } catch (e) {
      this.$showError(e);
    }
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
    shareViaOS: async function () {
      try {
        const outcome = await shareViaOS(this.shareTargets);

        if (outcome === "shared") {
          this.$showSuccess(this.$t("success.sharedViaOS"));
        } else if (outcome === "unsupported") {
          api.files.download(null, this.shareTargets[0].url);
          this.$showError(this.$t("errors.shareUnsupported"));
        }
      } catch (e) {
        this.$showError(e);
      }
    },
    copyToClipboard: function (text) {
      copy({ text }).then(
        () => {
          // clipboard successfully set
          this.$showSuccess(this.$t("success.linkCopied"));
        },
        () => {
          // clipboard write failed
          copy({ text }, { permission: true }).then(
            () => {
              // clipboard successfully set
              this.$showSuccess(this.$t("success.linkCopied"));
            },
            (e) => {
              // clipboard write failed
              this.$showError(e);
            }
          );
        }
      );
    },
    submit: async function () {
      try {
        let res = null;

        if (!this.time) {
          res = await api.share.create(this.url, this.password);
        } else {
          res = await api.share.create(
            this.url,
            this.password,
            this.time,
            this.unit
          );
        }

        this.links.push(res);
        this.sort();

        this.time = 0;
        this.unit = "hours";
        this.password = "";

        this.listing = true;
      } catch (e) {
        this.$showError(e);
      }
    },
    deleteLink: async function (event, link) {
      event.preventDefault();
      try {
        await api.share.remove(link.hash);
        this.links = this.links.filter((item) => item.hash !== link.hash);

        if (this.links.length == 0) {
          this.listing = false;
        }
      } catch (e) {
        this.$showError(e);
      }
    },
    humanTime(time) {
      return dayjs(time * 1000).fromNow();
    },
    buildLink(share) {
      return api.share.getShareURL(share);
    },
    buildDownloadLink(share) {
      return api.pub.getDownloadURL(
        {
          hash: share.hash,
          path: "",
        },
        true
      );
    },
    sort() {
      this.links = this.links.sort((a, b) => {
        if (a.expire === 0) return -1;
        if (b.expire === 0) return 1;
        return new Date(a.expire) - new Date(b.expire);
      });
    },
    switchListing() {
      if (this.links.length == 0 && !this.listing) {
        this.closeHovers();
      }

      this.listing = !this.listing;
    },
    normalizedGrantPath() {
      if (!this.url) return undefined;
      return removePrefix(this.url);
    },
    switchGrantTab() {
      this.tab = "people";
      this.loadGrants().catch((e) => this.$showError(e));
    },
    async loadGrants() {
      const path = this.normalizedGrantPath();
      if (!path) {
        this.grants = [];
        return;
      }
      const all = await api.grants.list();
      this.grants = all.filter((g) => g.path === path);
    },
    onSearchInput() {
      if (this.searchTimer) clearTimeout(this.searchTimer);
      this.searchTimer = setTimeout(() => {
        this.runUserSearch().catch((e) => this.$showError(e));
      }, 250);
    },
    async runUserSearch() {
      const q = (this.searchQuery || "").trim();
      if (q.length < 2) {
        this.searchResults = [];
        return;
      }
      this.searchResults = await api.users.searchUsers(q);
    },
    pickUser(u) {
      this.selectedUser = u;
      this.searchResults = [];
      this.searchQuery = u.username;
    },
    grantRoleLabel(grant) {
      return grant.role === "editor"
        ? this.$t("grants.editor")
        : this.$t("grants.viewer");
    },
    async submitGrant() {
      try {
        if (!this.selectedUser) return;
        const path = this.normalizedGrantPath();
        if (!path) return;
        const res = await api.grants.create({
          path,
          // Send the exact user ID from the picker (unambiguous even for
          // numeric usernames like "123", which the API would parse as IDs).
          grantee: String(this.selectedUser.id),
          role: this.grantRole,
        });
        this.grants.push(res);
        this.selectedUser = null;
        this.searchQuery = "";
        this.grantRole = "viewer";
        this.$showSuccess(this.$t("grants.grantCreated"));
      } catch (e) {
        if (e && e.status === 409) {
          this.$showError(this.$t("grants.alreadyShared"));
        } else {
          this.$showError(e);
        }
      }
    },
    async deleteGrant(event, grant) {
      event.preventDefault();
      try {
        await api.grants.remove(grant.id);
        this.grants = this.grants.filter((item) => item.id !== grant.id);
        this.$showSuccess(this.$t("grants.grantRevoked"));
      } catch (e) {
        this.$showError(e);
      }
    },
  },
};
</script>
