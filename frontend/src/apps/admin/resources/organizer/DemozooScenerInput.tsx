import {
  AutocompleteInput,
  NumberInput,
  ReferenceInput,
  useGetList,
} from "react-admin";

const countryFlag = (code?: string): string => {
  if (!code || code.length !== 2) return "";
  const base = 0x1f1e6;
  return String.fromCodePoint(
    ...code
      .toUpperCase()
      .split("")
      .map((char) => base + char.charCodeAt(0) - 65),
  );
};

export const DemozooScenerInput = () => {
  const { data, isLoading } = useGetList("demozoo-sceners", {
    pagination: { page: 1, perPage: 1 },
    sort: { field: "name", order: "ASC" },
    filter: {},
  });

  if (isLoading) {
    return <NumberInput source="demozoo_id" helperText="Demozoo scener ID" />;
  }

  const hasSceners = (data?.length ?? 0) > 0;
  if (!hasSceners) {
    return <NumberInput source="demozoo_id" helperText="Demozoo scener ID" />;
  }

  return (
    <ReferenceInput source="demozoo_id" reference="demozoo-sceners">
      <AutocompleteInput
        optionText={(record: {
          id?: number;
          name?: string;
          groups?: string[];
          country_code?: string;
        }) => {
          const groups = record?.groups?.length
            ? ` / ${record.groups.join(", ")}`
            : "";
          const country = record?.country_code
            ? `  ${countryFlag(record.country_code)}`
            : "";
          return `${record?.name ?? ""} (${record?.id ?? ""})${groups}${country}`;
        }}
        inputText={(record: { id?: number; name?: string } | null) =>
          record ? `${record.name ?? ""} (${record.id ?? ""})` : ""
        }
        matchSuggestion={(
          filter: string,
          choice: { id?: number; name?: string },
        ) => {
          const numeric = /^\d+$/.test(filter ?? "");
          if (numeric) {
            return (
              String(choice?.id ?? "") === (filter ?? "") ||
              (choice?.name ?? "")
                .toLowerCase()
                .includes((filter ?? "").toLowerCase())
            );
          }
          return (choice?.name ?? "")
            .toLowerCase()
            .includes((filter ?? "").toLowerCase());
        }}
        helperText="Demozoo scener"
        filterToQuery={(search: string) => ({ q: search })}
      />
    </ReferenceInput>
  );
};
