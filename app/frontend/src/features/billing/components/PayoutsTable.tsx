import { Download, FileX } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import type { Payout } from "../hooks/useBillingData"

interface PayoutsTableProps {
  payouts: Payout[]
}

export function PayoutsTable({ payouts }: PayoutsTableProps) {
  const handleDownloadCsv = () => {
    if (!payouts.length) return;
    
    const headers = ["Payout ID", "Date", "Amount", "Status", "Destination Bank"]
    const csvContent = [
      headers.join(","),
      ...payouts.map((payout: Payout) => 
        [payout.id, payout.date, payout.amount.replace(/,/g, ''), payout.status, payout.bank].join(",")
      )
    ].join("\n")

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement("a")
    const url = URL.createObjectURL(blob)
    link.setAttribute("href", url)
    link.setAttribute("download", `platform-payouts-${new Date().toISOString().split('T')[0]}.csv`)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  return (
    <section aria-label="Recent Payouts">
      <Card className="card-elevated border-border/40 dark:border-border/30 bg-card relative overflow-hidden">
        {/* Decorative top accent */}
        <div className="absolute top-0 left-0 right-0 h-[2px] bg-gradient-to-r from-emerald-500 via-blue-500 to-violet-500" />

        <CardHeader className="flex flex-row items-center justify-between pt-5">
          <div className="flex flex-col gap-1.5">
            <CardTitle className="text-base">Recent Payouts</CardTitle>
            <CardDescription>Platform fees paid out to your primary bank account.</CardDescription>
          </div>
          <Button variant="outline" size="sm" onClick={handleDownloadCsv} disabled={!payouts.length} className="shadow-sm">
            <Download className="mr-2 size-4" />
            Download CSV
          </Button>
        </CardHeader>
        <CardContent>
          {payouts.length === 0 ? (
             <div className="flex flex-col h-40 items-center justify-center border border-dashed rounded-lg bg-muted/10 text-muted-foreground text-sm">
               <FileX className="size-8 text-muted-foreground/40 mb-2" />
               <p>No recent payouts found.</p>
             </div>
          ) : (
            <div className="rounded-lg border border-border/40 overflow-hidden">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/40 dark:bg-muted/20 hover:bg-muted/40 dark:hover:bg-muted/20 border-b border-border/40">
                    <TableHead>Payout ID</TableHead>
                    <TableHead>Date</TableHead>
                    <TableHead>Destination</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Amount</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {payouts.map((payout: Payout) => (
                    <TableRow key={payout.id} className="transition-colors hover:bg-muted/30 dark:hover:bg-muted/15">
                      <TableCell className="font-mono text-xs text-muted-foreground">{payout.id}</TableCell>
                      <TableCell className="text-sm">{payout.date}</TableCell>
                      <TableCell className="text-sm">{payout.bank}</TableCell>
                      <TableCell>
                        <Badge variant="outline" className="bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border-emerald-500/20 font-semibold text-[11px]">
                          {payout.status}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right font-semibold tabular-nums">{payout.amount}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </section>
  )
}
