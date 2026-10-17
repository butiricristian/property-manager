import { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";

export function OAuthCallbackPage() {
  const [searchParams] = useSearchParams();
  const { handleOAuthToken } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    const token = searchParams.get("token");
    if (token) {
      handleOAuthToken(token);
    }
    navigate("/", { replace: true });
  }, [searchParams, handleOAuthToken, navigate]);

  return <p>Signing in...</p>;
}
