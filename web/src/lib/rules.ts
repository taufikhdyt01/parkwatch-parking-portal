import { api } from "./api";
import type { Ruleset, RuleVersion } from "./types";

export function getActiveRules(): Promise<RuleVersion> {
  return api<RuleVersion>("/rules/active");
}

export function getRuleVersions(): Promise<RuleVersion[]> {
  return api<{ versions: RuleVersion[] }>("/rules").then((r) => r.versions);
}

export function publishRules(ruleset: Ruleset): Promise<RuleVersion> {
  return api<RuleVersion>("/rules", {
    method: "POST",
    body: JSON.stringify(ruleset),
  });
}
