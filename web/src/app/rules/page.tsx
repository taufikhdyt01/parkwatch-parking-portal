"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import { AppHeader } from "@/components/app-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useAuth } from "@/contexts/auth-context";
import { formatDateTime, formatIDR } from "@/lib/format";
import { getActiveRules, getRuleVersions, publishRules } from "@/lib/rules";
import {
  VIOLATION_TYPES,
  VIOLATION_TYPE_LABELS,
  type Ruleset,
} from "@/lib/types";

export default function RulesPage() {
  const router = useRouter();
  const { user, loading } = useAuth();
  const queryClient = useQueryClient();

  const activeQuery = useQuery({
    queryKey: ["rules", "active"],
    queryFn: getActiveRules,
    enabled: !!user,
  });
  const versionsQuery = useQuery({
    queryKey: ["rules", "versions"],
    queryFn: getRuleVersions,
    enabled: !!user,
  });

  const [draft, setDraft] = useState<Ruleset | null>(null);
  useEffect(() => {
    if (activeQuery.data) setDraft(structuredClone(activeQuery.data.ruleset));
  }, [activeQuery.data]);

  const publish = useMutation({
    mutationFn: publishRules,
    onSuccess: (v) => {
      toast.success(`Published version ${v.version}`);
      queryClient.invalidateQueries({ queryKey: ["rules"] });
    },
    onError: (e) =>
      toast.error(e instanceof Error ? e.message : "Publish failed"),
  });

  // Officers only — members are redirected to their dashboard.
  useEffect(() => {
    if (!loading && (!user || user.role !== "officer")) router.replace("/");
  }, [loading, user, router]);

  if (loading || !user || user.role !== "officer") {
    return (
      <main className="flex flex-1 items-center justify-center">
        <p className="text-muted-foreground text-sm">Loading…</p>
      </main>
    );
  }

  function setBase(type: string, value: number) {
    setDraft((d) =>
      d ? { ...d, base_amounts: { ...d.base_amounts, [type]: value } } : d,
    );
  }
  function setTime(field: keyof Ruleset["time_multiplier"], value: number) {
    setDraft((d) =>
      d ? { ...d, time_multiplier: { ...d.time_multiplier, [field]: value } } : d,
    );
  }
  function setTier(index: number, field: "min_prior_unpaid" | "multiplier", value: number) {
    setDraft((d) => {
      if (!d) return d;
      const tiers = d.repeat_multiplier.tiers.map((t, i) =>
        i === index ? { ...t, [field]: value } : t,
      );
      return { ...d, repeat_multiplier: { tiers } };
    });
  }

  return (
    <>
      <AppHeader />
      <main className="mx-auto w-full max-w-4xl flex-1 px-6 py-8">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-xl font-semibold">Fine rules</h1>
            <p className="text-muted-foreground mt-1 text-sm">
              {activeQuery.data
                ? `Active version ${activeQuery.data.version}. Publishing a new version never changes fines already issued.`
                : "Loading active ruleset…"}
            </p>
          </div>
          <Button variant="ghost" size="sm" onClick={() => router.push("/")}>
            ← Dashboard
          </Button>
        </div>

        <Tabs defaultValue="publish">
          <TabsList>
            <TabsTrigger value="publish">Edit &amp; publish</TabsTrigger>
            <TabsTrigger value="history">
              Version history
              {versionsQuery.data ? ` (${versionsQuery.data.length})` : ""}
            </TabsTrigger>
          </TabsList>

          <TabsContent value="publish" className="mt-4">
            {!draft ? (
              <p className="text-muted-foreground text-sm">Loading…</p>
            ) : (
              <form
                onSubmit={(e) => {
                  e.preventDefault();
                  publish.mutate(draft);
                }}
                className="space-y-6"
              >
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">Base amounts (IDR)</CardTitle>
                    <CardDescription>Per violation type.</CardDescription>
                  </CardHeader>
                  <CardContent className="grid gap-4 sm:grid-cols-2">
                    {VIOLATION_TYPES.map((type) => (
                      <div key={type} className="space-y-2">
                        <Label htmlFor={`base-${type}`}>
                          {VIOLATION_TYPE_LABELS[type]}
                        </Label>
                        <Input
                          id={`base-${type}`}
                          type="number"
                          min={0}
                          value={draft.base_amounts[type] ?? 0}
                          onChange={(e) => setBase(type, Number(e.target.value))}
                        />
                      </div>
                    ))}
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">Time multiplier</CardTitle>
                    <CardDescription>
                      Day window is [day start, night start); outside it uses the
                      night multiplier.
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="grid gap-4 sm:grid-cols-2">
                    <Field label="Day start hour">
                      <Input
                        type="number"
                        min={0}
                        max={23}
                        value={draft.time_multiplier.day_start_hour}
                        onChange={(e) => setTime("day_start_hour", Number(e.target.value))}
                      />
                    </Field>
                    <Field label="Night start hour">
                      <Input
                        type="number"
                        min={0}
                        max={23}
                        value={draft.time_multiplier.night_start_hour}
                        onChange={(e) => setTime("night_start_hour", Number(e.target.value))}
                      />
                    </Field>
                    <Field label="Day multiplier">
                      <Input
                        type="number"
                        step="0.1"
                        min={0}
                        value={draft.time_multiplier.day_multiplier}
                        onChange={(e) => setTime("day_multiplier", Number(e.target.value))}
                      />
                    </Field>
                    <Field label="Night multiplier">
                      <Input
                        type="number"
                        step="0.1"
                        min={0}
                        value={draft.time_multiplier.night_multiplier}
                        onChange={(e) => setTime("night_multiplier", Number(e.target.value))}
                      />
                    </Field>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">Repeat multiplier</CardTitle>
                    <CardDescription>
                      Based on prior unpaid violations on the same plate in the last
                      90 days.
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    {draft.repeat_multiplier.tiers.map((tier, i) => (
                      <div key={i} className="grid grid-cols-2 gap-4">
                        <Field label="Prior unpaid ≥">
                          <Input
                            type="number"
                            min={0}
                            value={tier.min_prior_unpaid}
                            onChange={(e) => setTier(i, "min_prior_unpaid", Number(e.target.value))}
                          />
                        </Field>
                        <Field label="Multiplier">
                          <Input
                            type="number"
                            step="0.1"
                            min={0}
                            value={tier.multiplier}
                            onChange={(e) => setTier(i, "multiplier", Number(e.target.value))}
                          />
                        </Field>
                      </div>
                    ))}
                  </CardContent>
                </Card>

                <div className="flex items-center gap-3">
                  <Button type="submit" disabled={publish.isPending}>
                    {publish.isPending ? "Publishing…" : "Publish new version"}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() =>
                      activeQuery.data &&
                      setDraft(structuredClone(activeQuery.data.ruleset))
                    }
                  >
                    Reset
                  </Button>
                </div>
              </form>
            )}
          </TabsContent>

          <TabsContent value="history" className="mt-4">
            <Card>
              <CardContent className="pt-6">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Version</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Base amounts</TableHead>
                      <TableHead>Published by</TableHead>
                      <TableHead>When</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {versionsQuery.data?.map((v) => (
                      <TableRow key={v.id}>
                        <TableCell className="font-medium">v{v.version}</TableCell>
                        <TableCell>
                          {v.is_active ? (
                            <Badge>Active</Badge>
                          ) : (
                            <Badge variant="secondary">Superseded</Badge>
                          )}
                        </TableCell>
                        <TableCell className="text-muted-foreground text-xs">
                          {VIOLATION_TYPES.map((t) => (
                            <div key={t}>
                              {VIOLATION_TYPE_LABELS[t]}:{" "}
                              {formatIDR(v.ruleset.base_amounts[t] ?? 0)}
                            </div>
                          ))}
                        </TableCell>
                        <TableCell className="text-sm">{v.created_by}</TableCell>
                        <TableCell className="text-muted-foreground text-sm">
                          {formatDateTime(v.created_at)}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </main>
    </>
  );
}

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-2">
      <Label>{label}</Label>
      {children}
    </div>
  );
}
