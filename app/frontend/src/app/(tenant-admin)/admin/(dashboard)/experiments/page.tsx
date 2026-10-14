"use client"

import { useState } from "react"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Beaker, TrendingUp, Users, Plus, Trophy, AlertCircle, CheckCircle2 } from "lucide-react"
import {
  useGetApiV1AnalyticsAbTests,
  postApiV1AnalyticsAbTests,
  postApiV1AnalyticsAbTestsIdConclude,
  type ABTestItem,
} from "@/api/endpoints/analytics/analytics"

export default function ExperimentsPage() {
  const { data: abTestsResponse, isLoading, isError, refetch } = useGetApiV1AnalyticsAbTests()
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [isCreating, setIsCreating] = useState(false)
  const [concludingId, setConcludingId] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [metric, setMetric] = useState<"click" | "cart_add" | "purchase">("purchase")
  const [trafficSplit, setTrafficSplit] = useState(50)

  const experiments: ABTestItem[] = abTestsResponse?.data || []

  const handleToggle = async (experiment: ABTestItem) => {
    setActionError(null)
    if (experiment.status === "active") {
      try {
        setConcludingId(experiment.id)
        await postApiV1AnalyticsAbTestsIdConclude(experiment.id)
        await refetch()
      } catch (err: any) {
        setActionError(err.message || "Failed to conclude experiment")
      } finally {
        setConcludingId(null)
      }
    }
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return

    setActionError(null)
    setIsCreating(true)
    try {
      await postApiV1AnalyticsAbTests({
        name: name.trim(),
        description: description.trim(),
        metric,
        traffic_split: Number(trafficSplit),
        variant_a: { name: "Original / Control", type: "baseline" },
        variant_b: { name: "Variation B", type: "test" },
      })
      setName("")
      setDescription("")
      setMetric("purchase")
      setTrafficSplit(50)
      setIsDialogOpen(false)
      await refetch()
    } catch (err: any) {
      setActionError(err.message || "Failed to create experiment")
    } finally {
      setIsCreating(false)
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Experiments</h1>
          <p className="text-muted-foreground mt-1">
            Manage your storefront A/B tests and statistical conversion experiments
          </p>
        </div>

        <Button onClick={() => setIsDialogOpen(true)} className="w-fit gap-2">
          <Plus className="size-4" />
          New Experiment
        </Button>

        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogContent className="sm:max-w-[500px]">
            <form onSubmit={handleCreate}>
              <DialogHeader>
                <DialogTitle>Create New Experiment</DialogTitle>
                <DialogDescription>
                  Deploy an A/B test variant to test hypothesis against baseline conversions.
                </DialogDescription>
              </DialogHeader>

              <div className="grid gap-4 py-4">
                <div className="grid gap-2">
                  <Label htmlFor="name">Experiment Name</Label>
                  <Input
                    id="name"
                    placeholder="e.g. 1-Click Checkout CTA"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    required
                  />
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="description">Hypothesis / Description</Label>
                  <Textarea
                    id="description"
                    placeholder="Describe what variation you are testing and the expected conversion lift..."
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    rows={3}
                  />
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="grid gap-2">
                    <Label htmlFor="metric">Target Metric</Label>
                    <Select
                      value={metric}
                      onValueChange={(selectedMetric) => {
                        if (selectedMetric) setMetric(selectedMetric as "click" | "cart_add" | "purchase")
                      }}
                    >
                      <SelectTrigger id="metric">
                        <SelectValue placeholder="Select metric" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="purchase">Purchase (Orders)</SelectItem>
                        <SelectItem value="cart_add">Add to Cart</SelectItem>
                        <SelectItem value="click">CTR (Clicks)</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="grid gap-2">
                    <Label htmlFor="traffic">Traffic Split (% Variant B)</Label>
                    <Input
                      id="traffic"
                      type="number"
                      min={10}
                      max={90}
                      value={trafficSplit}
                      onChange={(e) => setTrafficSplit(Number(e.target.value))}
                    />
                  </div>
                </div>
              </div>

              <DialogFooter>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setIsDialogOpen(false)}
                >
                  Cancel
                </Button>
                <Button type="submit" disabled={isCreating || !name.trim()}>
                  {isCreating ? "Deploying..." : "Launch Experiment"}
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {actionError && (
        <div className="p-4 rounded-xl bg-destructive/10 border border-destructive/20 text-destructive text-sm flex items-center gap-2">
          <AlertCircle className="size-4 shrink-0" />
          <span>{actionError}</span>
        </div>
      )}

      {isLoading ? (
        <div className="grid grid-cols-1 gap-6">
          {Array.from({ length: 3 }).map((_, i) => (
            <Card key={i} className="card-elevated border-border/40 p-6 space-y-4">
              <Skeleton className="h-6 w-1/3" />
              <Skeleton className="h-4 w-2/3" />
              <div className="flex gap-4 pt-4">
                <Skeleton className="h-16 w-32 rounded-lg" />
                <Skeleton className="h-16 w-32 rounded-lg" />
              </div>
            </Card>
          ))}
        </div>
      ) : isError ? (
        <Card className="p-12 text-center border-dashed border-border/60">
          <AlertCircle className="size-8 text-destructive mx-auto mb-3" />
          <h3 className="text-lg font-semibold">Failed to load experiments</h3>
          <p className="text-sm text-muted-foreground mt-1 mb-4">
            An error occurred while fetching experiment data from the backend.
          </p>
          <Button variant="outline" onClick={() => refetch()}>
            Retry
          </Button>
        </Card>
      ) : experiments.length === 0 ? (
        <Card className="p-12 text-center border-dashed border-border/60">
          <Beaker className="size-10 text-muted-foreground/60 mx-auto mb-3" />
          <h3 className="text-lg font-semibold">No active experiments</h3>
          <p className="text-sm text-muted-foreground mt-1 mb-4 max-w-sm mx-auto">
            You have not launched any A/B tests yet. Create your first experiment to test hypotheses and improve sales conversion.
          </p>
          <Button onClick={() => setIsDialogOpen(true)} className="gap-2">
            <Plus className="size-4" />
            Launch First Experiment
          </Button>
        </Card>
      ) : (
        <div className="grid grid-cols-1 gap-6">
          {experiments.map((experiment) => {
            const isActive = experiment.status === "active"
            const uplift = experiment.conversion_uplift

            return (
              <Card key={experiment.id} className="card-elevated border-border/40">
                <CardHeader className="flex flex-row items-start justify-between pb-2">
                  <div className="space-y-1">
                    <CardTitle className="flex items-center gap-2">
                      <Beaker className="size-5 text-primary" />
                      {experiment.name}
                      {experiment.winner && experiment.winner !== "inconclusive" && (
                        <span className="flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full bg-amber-500/10 text-amber-500 border border-amber-500/20">
                          <Trophy className="size-3" />
                          Winner: Variant {experiment.winner.toUpperCase()}
                        </span>
                      )}
                    </CardTitle>
                    {experiment.description && (
                      <CardDescription>{experiment.description}</CardDescription>
                    )}
                  </div>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <span
                        className={`size-2 rounded-full ${
                          isActive ? "bg-emerald-500 animate-pulse" : "bg-muted-foreground/40"
                        }`}
                      />
                      {isActive ? "Active" : "Concluded"}
                    </div>
                    {isActive ? (
                      <Switch
                        checked={true}
                        disabled={concludingId === experiment.id}
                        onCheckedChange={() => handleToggle(experiment)}
                        title="Toggle to conclude experiment"
                      />
                    ) : (
                      <span className="text-xs text-muted-foreground flex items-center gap-1">
                        <CheckCircle2 className="size-3.5 text-muted-foreground" />
                        Finalized
                      </span>
                    )}
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="flex flex-wrap items-center gap-6 mt-4 p-4 rounded-lg bg-muted/20 border border-border/20">
                    <div className="flex flex-col gap-1 min-w-[120px]">
                      <span className="text-xs text-muted-foreground font-medium uppercase tracking-wider">
                        Participants
                      </span>
                      <div className="flex items-center gap-2 text-xl font-semibold">
                        <Users className="size-4 text-muted-foreground" />
                        {(experiment.participants || 0).toLocaleString()}
                      </div>
                    </div>

                    <div className="flex flex-col gap-1 min-w-[140px]">
                      <span className="text-xs text-muted-foreground font-medium uppercase tracking-wider">
                        Conversion Uplift
                      </span>
                      <div className="flex items-center gap-2 text-xl font-semibold">
                        <TrendingUp
                          className={`size-4 ${
                            !uplift
                              ? "text-muted-foreground"
                              : uplift > 0
                              ? "text-emerald-500"
                              : "text-destructive"
                          }`}
                        />
                        <span
                          className={
                            !uplift
                              ? "text-muted-foreground"
                              : uplift > 0
                              ? "text-emerald-500"
                              : "text-destructive"
                          }
                        >
                          {uplift ? `${uplift > 0 ? "+" : ""}${uplift}%` : "0.0%"}
                        </span>
                      </div>
                    </div>

                    <div className="flex flex-col gap-1 min-w-[100px]">
                      <span className="text-xs text-muted-foreground font-medium uppercase tracking-wider">
                        Metric Goal
                      </span>
                      <span className="text-sm font-medium capitalize mt-1 text-foreground">
                        {experiment.metric.replace("_", " ")}
                      </span>
                    </div>

                    <div className="flex flex-col gap-1 min-w-[100px]">
                      <span className="text-xs text-muted-foreground font-medium uppercase tracking-wider">
                        Split (A / B)
                      </span>
                      <span className="text-sm font-medium mt-1 text-foreground">
                        {100 - (experiment.traffic_split || 50)}% / {experiment.traffic_split || 50}%
                      </span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}
    </div>
  )
}
