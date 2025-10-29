import {create} from "zustand";
import {AuthState} from "../types/auth";

export const useAuth = create<AuthState>((set) => ({
    isLogged: false,
    Login: () => set({ isLogged: true }),
    Logout: () => set({ isLogged: false }),
}));