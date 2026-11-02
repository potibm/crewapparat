import { useEffect, useState } from "react";
import { useDataProvider, useRedirect } from "react-admin";
import { useParams } from "react-router";

export const SlugRedirect = () => {
  const { slug } = useParams();
  const dataProvider = useDataProvider();
  const redirect = useRedirect();
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    if (!slug) return;

    dataProvider
      .getList("organizers", {
        filter: { slug },
        pagination: { page: 1, perPage: 1 },
        sort: { field: "id", order: "ASC" },
      })
      .then(({ data }) => {
        if (data && data.length > 0) {
          redirect(`/organizers/${data[0].id}/show`);
        } else {
          setNotFound(true);
        }
      })
      .catch(() => setNotFound(true));
  }, [slug, dataProvider, redirect]);

  if (!slug || notFound) {
    return <div>Organizer not found.</div>;
  }

  return <div>Loading…</div>;
};

export default SlugRedirect;
