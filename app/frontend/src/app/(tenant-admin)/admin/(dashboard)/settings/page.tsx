import { SettingsForm } from "@/features/tenant-admin/components/SettingsForm"
import { StorefrontEditorForm } from "@/features/tenant-admin/components/StorefrontEditorForm"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

export default function AdminSettingsPage() {
  return (
    <div className="flex flex-col gap-12">
      <div className="flex flex-col gap-3">
        <span className="rounded-full px-3 py-1 text-[10px] uppercase tracking-[0.2em] font-medium text-zinc-400 border border-zinc-800 w-max">
          Configuration
        </span>
        <h1 className="text-4xl font-semibold tracking-tight text-zinc-50">Settings</h1>
        <p className="text-zinc-400 text-lg">
          Manage your store's configuration, domains, and appearance.
        </p>
      </div>
      
      <Tabs defaultValue="general" className="w-full">
        <TabsList className="mb-8 w-full justify-start bg-transparent border-b border-white/[0.05] rounded-none p-0 h-auto space-x-6">
          <TabsTrigger 
            value="general" 
            className="rounded-none bg-transparent data-[state=active]:bg-transparent data-[state=active]:shadow-none data-[state=active]:border-b-2 data-[state=active]:border-zinc-100 px-1 py-3 text-sm text-zinc-400 data-[state=active]:text-zinc-100"
          >
            General Settings
          </TabsTrigger>
          <TabsTrigger 
            value="appearance" 
            className="rounded-none bg-transparent data-[state=active]:bg-transparent data-[state=active]:shadow-none data-[state=active]:border-b-2 data-[state=active]:border-zinc-100 px-1 py-3 text-sm text-zinc-400 data-[state=active]:text-zinc-100"
          >
            Storefront Appearance
          </TabsTrigger>
        </TabsList>
        
        <TabsContent value="general" className="mt-0 outline-none">
          <div className="p-1.5 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
            <div className="rounded-2xl bg-zinc-900/50 p-6 md:p-8 lg:p-12 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)]">
              <SettingsForm />
            </div>
          </div>
        </TabsContent>
        
        <TabsContent value="appearance" className="mt-0 outline-none">
          <div className="p-1.5 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
            <div className="rounded-2xl bg-zinc-900/50 p-6 md:p-8 lg:p-12 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)]">
              <StorefrontEditorForm />
            </div>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}
