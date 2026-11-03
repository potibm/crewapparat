import { useGetList } from "react-admin";
import type { CodeOfConductRecord } from "./code-of-conduct.types";

export const useLatestPublishedCodeOfConduct = () =>
  useGetList<CodeOfConductRecord>("code-of-conducts", {
    filter: { status: "published" },
    // "published" is the backend's sort alias for published_at; passing
    // "published_at" would be rejected with "unsupported sort column".
    sort: { field: "published", order: "DESC" },
    pagination: { page: 1, perPage: 1 },
  });
