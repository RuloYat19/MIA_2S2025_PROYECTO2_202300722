import { create } from "zustand";
import { DiskState } from "../types/disks";

export const useDisksStore = create<DiskState>((set, get) => ({
  disks: JSON.parse(localStorage.getItem("disks") || "[]"),
  loading: false,
  error: null,

  addDisk: (filePath) => {
    const { disks } = get();
    const fileName = filePath.split("/").pop() || filePath;
    if (disks.some((disk) => disk.filePath === filePath)) {
      set({ error: "El disco ya ha sido agregado." });
      return;
    }

    const newDisks = [...disks, { fileName, filePath }];
    set({ disks: newDisks, error: null });
    localStorage.setItem("disks", JSON.stringify(newDisks));
  },

  addDisksFromFolder: (filePaths) => {
    const { disks } = get();
    set({ loading: true });

    const newDisks = filePaths
      .filter((filePath) => !disks.some((disk) => disk.filePath === filePath))
      .map((filePath) => {
        const fileName = filePath.split("/").pop() || filePath;
        return { fileName, filePath };
      });

    const updatedDisks = [...disks, ...newDisks];
    set({ disks: updatedDisks, loading: false });
    localStorage.setItem("disks", JSON.stringify(updatedDisks));
  },

  setLoading: (isLoading) => set({ loading: isLoading }),

  removeDisk: (diskIndex: number) => {
    const { disks } = get();
    const updatedDisks = disks.filter((_, index) => index !== diskIndex);
    set({ disks: updatedDisks, error: null });
    localStorage.setItem("disks", JSON.stringify(updatedDisks));
  },
  
  clearDisks: () => {
    set({ disks: [], error: null });
    localStorage.removeItem("disks");
  },

}));