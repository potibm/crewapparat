import { createLogger } from "@core/logger/logger";
import { UserManager } from "oidc-client-ts";
import { AuthProvider } from "react-admin";

const log = createLogger("Auth");

let userManager: UserManager | null = null;
let activeCallbackPromise: Promise<void> | null = null;

export interface CanAccessParams {
  action: string;
  resource: string;
  record?: unknown;
}

export const configureOidc = (authority: string, clientId: string) => {
  userManager = new UserManager({
    authority,
    client_id: clientId,
    // No dedicated callback path: React Admin lives at the root and processes the
    // `?code=` parameter on the login page itself. A `/auth-callback` route
    // would lead nowhere, because no such route exists.
    redirect_uri: globalThis.location.origin + "/auth-callback",
    post_logout_redirect_uri: `${globalThis.location.origin}/`,
    response_type: "code",
    scope: "openid profile email groups federated:id",
  });
};

interface MeData {
  id: string;
  handle: string;
  realname: string;
  email: string;
  pronouns?: string;
  avatar_url?: string;
  is_admin: boolean;
}

let meCache: { data: MeData; expires: number } | null = null;
let mePromise: Promise<MeData | null> | null = null;
const ME_TTL_MS = 60_000;

export const linkMe = async (): Promise<boolean> => {
  try {
    const token = await getAccessToken();
    const res = await fetch("/api/admin/me/link", {
      method: "POST",
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    });
    return res.ok;
  } catch (error) {
    log.warn("Failed to link /me", error);
    return false;
  }
};

const fetchMe = async (): Promise<MeData | null> => {
  const now = Date.now();
  if (meCache && meCache.expires > now) return meCache.data;
  if (mePromise) return mePromise;

  mePromise = (async () => {
    try {
      const token = await getAccessToken();
      const res = await fetch("/api/admin/me", {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      });
      if (res.ok) {
        const data = (await res.json()) as MeData;
        meCache = { data, expires: Date.now() + ME_TTL_MS };
        return data;
      }

      // First-login linking per concept §4.1.5: try to link once, then retry.
      if (res.status === 404) {
        const linked = await linkMe();
        if (linked) {
          const retry = await fetch("/api/admin/me", {
            headers: token ? { Authorization: `Bearer ${token}` } : {},
          });
          if (retry.ok) {
            const data = (await retry.json()) as MeData;
            meCache = { data, expires: Date.now() + ME_TTL_MS };
            return data;
          }
        }
      }

      return null;
    } catch (error) {
      log.warn("Failed to fetch /me", error);
      return null;
    } finally {
      mePromise = null;
    }
  })();

  return mePromise;
};

export const invalidateMe = (): void => {
  meCache = null;
  mePromise = null;
};

export const authProvider: AuthProvider = {
  login: () => {
    if (!userManager) throw new Error("OIDC not configured");
    return userManager.signinRedirect();
  },

  handleCallback: async () => {
    if (!userManager) throw new Error("OIDC not configured");

    if (activeCallbackPromise) {
      return activeCallbackPromise;
    }

    activeCallbackPromise = userManager
      .signinRedirectCallback()
      .then(() => {})
      .catch((error) => {
        log.error("Authentication error in callback:", error);
        throw error instanceof Error ? error : new Error(String(error));
      })
      .finally(() => {
        activeCallbackPromise = null;
      });

    return activeCallbackPromise;
  },

  checkAuth: async () => {
    if (!userManager) throw new Error("OIDC not configured");

    const user = await userManager.getUser();
    if (!user || user.expired) {
      throw new Error("User not authenticated or token expired");
    }
  },

  checkError: async (error) => {
    if (error.status === 401 || error.status === 403) {
      throw new Error("Unauthorized API access");
    }
  },

  logout: async () => {
    invalidateMe();
    if (!userManager) return;

    try {
      const user = await userManager.getUser();
      if (user) {
        await userManager.signoutRedirect();
      } else {
        await userManager.removeUser();
      }
    } catch (error) {
      if (
        error instanceof Error &&
        error.message.includes("No end session endpoint")
      ) {
        log.warn(
          "IdP does not support end session endpoint, clearing local data.",
        );
        await userManager.removeUser();
      } else {
        throw error instanceof Error ? error : new Error(String(error));
      }
    }
  },

  getIdentity: async () => {
    const data = await fetchMe();
    if (data) {
      return {
        id: data.id,
        fullName: data.realname || data.handle,
        avatar: data.avatar_url,
      };
    }

    if (!userManager) throw new Error("OIDC not configured");

    const user = await userManager.getUser();
    if (user?.profile) {
      return {
        id: user.profile.sub,
        fullName: user.profile.name || user.profile.preferred_username,
        avatar: user.profile.picture,
      };
    }
    throw new Error("No identity profile found");
  },

  getPermissions: async () => {
    const data = await fetchMe();
    if (data) return data.is_admin ? ["admin"] : [];

    if (!userManager) return [];

    const user = await userManager.getUser();
    return user ? user.profile.roles : [];
  },

  canAccess: async ({
    action,
    resource,
    record,
  }: CanAccessParams): Promise<boolean> => {
    // An unlinked (not provisioned) user must see nothing.
    const me = await fetchMe();
    if (!me) return false;

    const permissions = (await authProvider.getPermissions?.({})) ?? [];
    const isAdmin = Array.isArray(permissions) && permissions.includes("admin");

    if (isAdmin) return true;

    if (action === "list") {
      return true;
    }

    if (resource === "organizers" && action === "edit") {
      // useEditController calls canAccess without a record -> allow load,
      // OrganizerEdit enforces ownership afterwards.
      if (!record) return true;
      const permissions = (await authProvider.getPermissions?.({})) ?? [];
      const isAdmin =
        Array.isArray(permissions) && permissions.includes("admin");
      if (isAdmin) return true;
      const identity = await authProvider.getIdentity?.();
      const recordId = (record as { id?: string } | null | undefined)?.id;
      return recordId === identity?.id;
    }
    if (resource === "organizers" && (action === "show" || action === "view")) {
      return true;
    }

    return false;
  },
};

export const getAccessToken = async (): Promise<string | null> => {
  if (!userManager) {
    log.warn("getAccessToken: userManager is null");
    return null;
  }
  const user = await userManager.getUser();
  if (!user) {
    log.warn("getAccessToken: no user found");
    return null;
  }
  if (user.expired) {
    log.warn("getAccessToken: user token expired");
    return null;
  }
  const token = user.access_token ?? user.id_token ?? null;
  if (!token) {
    log.warn("getAccessToken: no access_token or id_token found");
  }
  return token;
};
