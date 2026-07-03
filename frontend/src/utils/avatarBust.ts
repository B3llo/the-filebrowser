import { ref } from "vue";

// Shared across every AvatarBadge instance: the server always serves the
// same URL for a given user id, so after an upload we bump this to force
// all badges (sidebar, details panel, profile) to re-fetch the new image
// instead of showing a stale one from the browser cache.
export const avatarBust = ref(0);

export function bumpAvatarBust() {
  avatarBust.value = Date.now();
}
