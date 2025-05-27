import { useUserStore } from "@/store/useUserStore";

/**
 * Utility function to handle automatic logout when session expires
 * This is separated from components to avoid circular dependencies
 */
export const handleAutoLogout = () => {
  console.log("Session expired. Auto-logging out user.");

  const { logout } = useUserStore.getState();

  // Clear user state - this will automatically trigger WebSocket disconnection
  // in the GlobalWebSocketProvider's useEffect when isAuthenticated becomes false
  logout();

  // Redirect to login page using replace to prevent going back to protected page
  if (typeof window !== "undefined") {
    // Only redirect if not already on the login page
    if (
      window.location.pathname !== "/login" &&
      window.location.pathname !== "/register"
    ) {
      window.location.replace("/login");
    }
  }
};

/**
 * Check if an error response indicates an expired session
 */
export const isSessionExpiredError = (status: number): boolean => {
  return status === 401;
};
