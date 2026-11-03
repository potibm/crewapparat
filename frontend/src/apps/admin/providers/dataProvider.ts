import jsonServerProvider from "ra-data-json-server";
import {
  CreateParams,
  DataProvider,
  UpdateParams,
  fetchUtils,
} from "react-admin";
import { getAccessToken } from "./authProvider";

const httpClient = async (url: string, options: fetchUtils.Options = {}) => {
  if (!options.headers) {
    options.headers = new Headers({ Accept: "application/json" });
  }

  const headers =
    options.headers instanceof Headers
      ? options.headers
      : new Headers(options.headers);

  const token = await getAccessToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  options.headers = headers;

  return fetchUtils.fetchJson(url, options);
};

const baseProvider = jsonServerProvider("/api/admin", httpClient);

interface OrganizerFormData {
  handle?: string;
  realname?: string;
  email?: string;
  phone?: string;
  slack_handle?: string;
  slack_profile_url?: string;
  mediawiki_profile_url?: string;
  pronouns?: string;
  languages?: string[];
  demozoo_id?: number | string;
  first_year_as_organizer?: number | string;
  avatar_url?: string;
  remove_avatar?: boolean;
  media_upload?: {
    rawFile?: File;
  };
  [key: string]: unknown;
}

const ORGANIZER_RESOURCE = "organizers";

const handleOrganizerFileUpload = async (
  resource: string,
  params: UpdateParams | CreateParams,
  method: "POST" | "PUT",
) => {
  const data = params.data as OrganizerFormData;

  if (!data.media_upload?.rawFile) {
    const removed =
      data.media_upload === null &&
      !!(params as UpdateParams).previousData?.avatar_url;
    const { media_upload: _media_upload, ...rest } = data;
    // Preserve the existing avatar URL when the form does not carry one
    // (e.g. on edits where no new file was selected), mirroring the
    // billedapparat behaviour of not wiping unchanged media.
    if (removed) {
      rest.remove_avatar = true;
      delete (rest as { avatar_url?: string }).avatar_url;
    } else if (rest.avatar_url === undefined) {
      const previousAvatar = (params as UpdateParams).previousData?.avatar_url;
      if (previousAvatar) {
        rest.avatar_url = previousAvatar;
      }
    }
    const newParams = { ...params, data: rest };

    if (method === "POST") {
      return baseProvider.create(resource, newParams as CreateParams);
    }
    return baseProvider.update(resource, newParams as UpdateParams);
  }

  const formData = new FormData();

  formData.append("handle", data.handle || "");
  formData.append("realname", data.realname || "");
  formData.append("email", data.email || "");
  formData.append("phone", data.phone || "");
  formData.append("slack_handle", data.slack_handle || "");
  formData.append("slack_profile_url", data.slack_profile_url || "");
  formData.append("mediawiki_profile_url", data.mediawiki_profile_url || "");
  formData.append("pronouns", data.pronouns || "");
  (data.languages || []).forEach((lang) => formData.append("languages", lang));
  if (data.demozoo_id != null && data.demozoo_id !== "") {
    formData.append("demozoo_id", String(data.demozoo_id));
  }
  if (data.first_year_as_organizer != null && data.first_year_as_organizer !== "") {
    formData.append("first_year_as_organizer", String(data.first_year_as_organizer));
  }
  if (data.avatar_url) {
    formData.append("avatar_url", data.avatar_url);
  }
  if (data.remove_avatar) {
    formData.append("remove_avatar", "true");
  }

  formData.append("media_upload", data.media_upload.rawFile);

  const id = "id" in params ? params.id : "";

  const url =
    method === "PUT"
      ? `/api/admin/${ORGANIZER_RESOURCE}/${id}`
      : `/api/admin/${ORGANIZER_RESOURCE}`;

  const token = await getAccessToken();

  const headers = new Headers();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  return fetch(url, {
    method: method,
    headers: headers,
    body: formData,
  })
    .then(async (response) => {
      if (!response.ok) {
        const errBody = await response.text();
        throw new Error(errBody || response.statusText);
      }
      return response.json();
    })
    .then((json) => ({ data: json }));
};

export const dataProvider: DataProvider = {
  ...baseProvider,

  create: (resource, params) => {
    if (resource === ORGANIZER_RESOURCE) {
      return handleOrganizerFileUpload(resource, params, "POST");
    }
    return baseProvider.create(resource, params);
  },

  update: (resource, params) => {
    if (resource === ORGANIZER_RESOURCE) {
      return handleOrganizerFileUpload(resource, params, "PUT");
    }
    return baseProvider.update(resource, params);
  },
};
