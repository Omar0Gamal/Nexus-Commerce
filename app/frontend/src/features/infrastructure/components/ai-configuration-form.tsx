"use client"

import { useState } from "react"
import { Eye, EyeOff, KeyRound, Lock, ShieldAlert, CheckCircle2, Loader2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from "@/components/ui/input-group"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { usePutApiV1PlatformAdminAiConfig } from "@/api/endpoints/admin/admin"

export function AiConfigurationForm() {
  const [apiKey, setApiKey] = useState("")
  const [reveal, setReveal] = useState(false)
  const [confirmed, setConfirmed] = useState(false)
  const [rotated, setRotated] = useState(false)

  const { mutate, isPending } = usePutApiV1PlatformAdminAiConfig()

  const canSubmit = apiKey.trim().length >= 12 && confirmed && !isPending

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    if (!canSubmit) return

    mutate({ data: { api_key: apiKey } }, {
      onSuccess: () => {
        setRotated(true)
        setApiKey("")
        setConfirmed(false)
        setReveal(false)
      },
      onError: () => {
        console.error("Failed to rotate AI API key")
      }
    })
  }

  return (
    <Card className="max-w-2xl">
      <CardHeader>
        <div className="flex items-center gap-2.5">
          <span className="flex size-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
            <KeyRound className="size-4.5" />
          </span>
          <div className="flex flex-col gap-1">
            <CardTitle>AI Configuration</CardTitle>
            <CardDescription>
              Rotate the platform key used for AI content and SEO generation.
            </CardDescription>
          </div>
        </div>
      </CardHeader>

      <form onSubmit={handleSubmit}>
        <CardContent>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="ai-api-key">Rotate AI API Keys</FieldLabel>
              <InputGroup>
                <InputGroupAddon>
                  <Lock className="text-muted-foreground" />
                </InputGroupAddon>
                <InputGroupInput
                  id="ai-api-key"
                  type={reveal ? "text" : "password"}
                  autoComplete="off"
                  placeholder="sk-nexus-••••••••••••••••••••"
                  value={apiKey}
                  onChange={(e) => {
                    setApiKey(e.target.value)
                    setRotated(false)
                  }}
                  disabled={isPending}
                />
                <InputGroupAddon align="inline-end">
                  <InputGroupButton
                    type="button"
                    size="icon-xs"
                    aria-label={reveal ? "Hide API key" : "Show API key"}
                    aria-pressed={reveal}
                    onClick={() => setReveal((v) => !v)}
                    disabled={isPending}
                  >
                    {reveal ? <EyeOff /> : <Eye />}
                  </InputGroupButton>
                </InputGroupAddon>
              </InputGroup>
              <FieldDescription>
                The new key is encrypted at rest and never displayed again after saving.
              </FieldDescription>
            </Field>

            <Alert>
              <ShieldAlert />
              <AlertTitle>Rotating invalidates the current key immediately</AlertTitle>
              <AlertDescription>
                All tenant AI generation jobs will use the new key on their next request.
              </AlertDescription>
            </Alert>

            <Field orientation="horizontal">
              <Checkbox
                id="confirm-rotate"
                checked={confirmed}
                onCheckedChange={(value) => setConfirmed(value === true)}
                disabled={isPending}
              />
              <FieldLabel htmlFor="confirm-rotate" className="font-normal">
                I understand this will immediately revoke the existing AI API key across all
                tenants.
              </FieldLabel>
            </Field>

            {rotated ? (
              <Alert>
                <CheckCircle2 />
                <AlertTitle>API key rotated</AlertTitle>
                <AlertDescription>
                  The AI engine will pick up the new key within a few seconds.
                </AlertDescription>
              </Alert>
            ) : null}
          </FieldGroup>
        </CardContent>

        <CardFooter className="mt-6 flex items-center justify-between gap-3 border-t pt-6">
          <p className="text-xs text-muted-foreground">
            Requires platform admin privileges.
          </p>
          <Button type="submit" variant="destructive" disabled={!canSubmit}>
            {isPending ? <Loader2 className="mr-2 size-4 animate-spin" /> : <KeyRound data-icon="inline-start" />}
            {isPending ? "Rotating..." : "Rotate key"}
          </Button>
        </CardFooter>
      </form>
    </Card>
  )
}
