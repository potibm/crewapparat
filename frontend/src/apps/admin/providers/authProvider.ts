import { createLogger } from "@core/logger/logger";
import { UserManager } from "oidc-client-ts";
import { AuthProvider } from "react-admin";

const log = createLogger("Auth");

let userManager: UserManager | null = null;
let activeCallbackPromise: Promise<void> | null = null;

export const configureOidc = (authority: string, clientId: string) => {
  userManager = new UserManager({
    authority,
    client_id: clientId,
    // Kein eigener Callback-Pfad: React Admin hängt am Root und verarbeitet den
    // `?code=`-Parameter auf der Login-Seite selbst. Ein `/auth-callback` würde
    // ins Leere routen, weil es keine Route dafür gibt.
    redirect_uri: globalThis.location.origin + "/auth-callback",
    post_logout_redirect_uri: `${globalThis.location.origin}/`,
    response_type: "code",
    scope: "openid profile email groups",
  });
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
    try {
      const token = await getAccessToken();
      const res = await fetch("/api/admin/me", {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      });
      if (res.ok) {
        const data = await res.json();
        return {
          id: data.id,
          fullName: data.realname || data.handle,
          avatar: data.avatar_url,
        };
      }
    } catch (error) {
      log.warn("Failed to fetch /me", error);
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
    try {
      const token = await getAccessToken();
      const res = await fetch("/api/admin/me", {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      });
      if (res.ok) {
        const data = await res.json();
        return data.is_admin ? ["admin"] : [];
      }
    } catch (error) {
      log.warn("Failed to fetch /me for permissions", error);
    }

    if (!userManager) return [];

    const user = await userManager.getUser();
    return user ? user.profile.roles : [];
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
