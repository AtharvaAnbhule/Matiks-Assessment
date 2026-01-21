import React from "react";
import { StyleSheet } from "react-native";
import { RootNavigator } from "./src/navigation/RootNavigator";

/**
 * Root Application Component
 * Entry point for the React Native Expo application
 *
 * Architecture:
 * - Navigation: Bottom tab navigation with stack support
 * - Screens: Leaderboard, Search, Profile
 * - Services: API layer with Axios
 * - Hooks: Custom hooks for data fetching and state management
 * - Features:
 *   - Real-time rank updates (5s polling)
 *   - Debounced search (500ms)
 *   - Pagination support
 *   - Pull-to-refresh
 *   - Error handling and loading states
 *   - Responsive UI design
 */
export default function App() {
  return <RootNavigator />;
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#fff",
    alignItems: "center",
    justifyContent: "center",
  },
});
