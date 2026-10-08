export function GlobalGrid() {
  return (
    <div className="fixed inset-0 pointer-events-none z-[-1] overflow-hidden">
      <div className="absolute inset-0 bg-background" />

      <div className="absolute top-[-10%] right-[10%] w-[800px] h-[800px] rounded-full bg-primary/[0.04] blur-[140px]" />
      
      <div className="absolute top-[40%] left-[-10%] w-[600px] h-[600px] rounded-full bg-primary/[0.03] blur-[120px]" />

      <div className="absolute bottom-[-10%] right-[-5%] w-[700px] h-[700px] rounded-full bg-primary/[0.03] blur-[130px]" />

      <div className="absolute inset-0 w-full flex justify-center opacity-[0.6]">
        <div className="w-full h-full max-w-7xl px-4 md:px-6 relative flex justify-between">
          {[...Array(6)].map((_, i) => (
            <div key={i} className="h-full w-px bg-border/40" />
          ))}
        </div>
      </div>

      <div className="absolute inset-0 bg-[linear-gradient(to_bottom,transparent_99%,rgba(255,255,255,0.04)_100%)] bg-[length:100%_80px]" />
      
      <div 
        className="absolute inset-0 opacity-[0.03]" 
        style={{ backgroundImage: "url('https://grainy-gradients.vercel.app/noise.svg')" }} 
      />
    </div>
  )
}
