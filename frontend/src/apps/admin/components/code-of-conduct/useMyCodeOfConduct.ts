import { useCallback, useEffect, useState } from "react";
import {
  fetchMe,
  getAccessToken,
  invalidateMe,
} from "../../providers/authProvider";
import type { CodeOfConductRecord } from "../../resources/code-of-conduct/code-of-conduct.types";

export type MyCodeOfConduct = CodeOfConductRecord;

type GateState = "checking" | "accepted" | "required";

export const useMyCodeOfConduct = () => {
  const [state, setState] = useState<GateState>("checking");
  const [codeOfConduct, setCodeOfConduct] = useState<MyCodeOfConduct | null>(
    null,
  );

  const load = useCallback(async () => {
    setState("checking");

    const me = await fetchMe();

    // Unlinked users cannot accept anything, and React Admin handles the
    // missing permissions; there is nothing to gate.
    if (!me || !me.code_of_conduct || me.code_of_conduct_accepted_at) {
      setState("accepted");

      return;
    }

    setCodeOfConduct(me.code_of_conduct);
    setState("required");
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const accept = useCallback(async () => {
    if (!codeOfConduct) return;

    const token = await getAccessToken();

    if (!token) throw new Error("Not authenticated");

    const response = await fetch("/api/admin/me/code-of-conduct/accept", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ code_of_conduct_id: codeOfConduct.id }),
    });

    if (response.status === 409) {
      // A different version was published in the meantime: reload the prompt.
      invalidateMe();
      await load();

      throw new Error(
        "The Code of Conduct changed. Please review the current version and confirm again.",
      );
    }

    if (!response.ok) {
      throw new Error(
        `Failed to accept the Code of Conduct (${response.status})`,
      );
    }

    invalidateMe();
    await load();
  }, [codeOfConduct, load]);

  return { state, codeOfConduct, accept };
};
