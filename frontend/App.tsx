import React from "react";
import { StyleSheet } from "react-native";
import { RootNavigator } from "./src/navigation/RootNavigator";

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
