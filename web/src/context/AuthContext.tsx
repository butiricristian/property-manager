import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  type ReactNode,
} from "react";
import type { User } from "@/types";
import {
  login as apiLogin,
  register as apiRegister,
  guestLogin as apiGuestLogin,
  upgradeGuest as apiUpgradeGuest,
  getCurrentUser,
  setAuthToken,
  getAuthToken,
} from "@/api/client";

interface AuthContextType {
  user: User | null;
  loading: boolean;
  isGuest: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  loginAsGuest: () => Promise<void>;
  upgradeGuest: (email: string, password: string) => Promise<void>;
  loginWithGoogle: () => void;
  handleOAuthToken: (token: string) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = getAuthToken();
    if (!token) {
      setLoading(false);
      return;
    }
    getCurrentUser()
      .then(setUser)
      .catch(() => {
        setAuthToken(null);
      })
      .finally(() => setLoading(false));
  }, []);

  const handleOAuthToken = useCallback((token: string) => {
    setAuthToken(token);
    getCurrentUser()
      .then(setUser)
      .catch(() => setAuthToken(null));
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const res = await apiLogin(email, password);
    setAuthToken(res.token);
    setUser(res.user);
  }, []);

  const register = useCallback(async (email: string, password: string) => {
    const res = await apiRegister(email, password);
    setAuthToken(res.token);
    setUser(res.user);
  }, []);

  const loginAsGuest = useCallback(async () => {
    const res = await apiGuestLogin();
    setAuthToken(res.token);
    setUser(res.user);
  }, []);

  const upgradeGuest = useCallback(async (email: string, password: string) => {
    const res = await apiUpgradeGuest(email, password);
    setAuthToken(res.token);
    setUser(res.user);
  }, []);

  const loginWithGoogle = useCallback(() => {
    const url = import.meta.env.DEV
      ? "http://localhost:8080/api/auth/google"
      : "/api/auth/google";
    window.location.href = url;
  }, []);

  const logout = useCallback(() => {
    setAuthToken(null);
    setUser(null);
  }, []);

  const isGuest = user?.user_type === "guest";

  return (
    <AuthContext.Provider
      value={{
        user,
        loading,
        isGuest,
        login,
        register,
        loginAsGuest,
        upgradeGuest,
        loginWithGoogle,
        handleOAuthToken,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
