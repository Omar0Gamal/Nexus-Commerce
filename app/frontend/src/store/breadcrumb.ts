import { create } from "zustand"

interface BreadcrumbStore {
  titleOverrides: Record<string, string>
  setTitleOverride: (segment: string, title: string) => void
  clearTitleOverride: (segment: string) => void
}

export const useBreadcrumbStore = create<BreadcrumbStore>((set) => ({
  titleOverrides: {},
  setTitleOverride: (segment, title) =>
    set((state) => ({
      titleOverrides: {
        ...state.titleOverrides,
        [segment.toLowerCase()]: title,
      },
    })),
  clearTitleOverride: (segment) =>
    set((state) => {
      const copy = { ...state.titleOverrides }
      delete copy[segment.toLowerCase()]
      return { titleOverrides: copy }
    }),
}))
