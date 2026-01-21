import React, { useEffect, useState } from "react";
import { NavigationContainer } from "@react-navigation/native";
import { createNativeStackNavigator } from "@react-navigation/native-stack";
import { createBottomTabNavigator } from "@react-navigation/bottom-tabs";
import { View, Text, StyleSheet, ActivityIndicator } from "react-native";
import LeaderboardScreen from "../screens/LeaderboardScreen";
import SearchScreen from "../screens/SearchScreen";
import ProfileScreen from "../screens/ProfileScreen";
import { userAPI } from "../services/api";

const Stack = createNativeStackNavigator();
const Tab = createBottomTabNavigator();

const SplashScreen: React.FC<{ onFinish: () => void }> = ({ onFinish }) => {
  const [loading, setLoading] = useState(true);
  const [apiReady, setApiReady] = useState(false);

  useEffect(() => {
    const checkAPI = async () => {
      try {
        const isHealthy = await userAPI.checkHealth();
        setApiReady(isHealthy);
      } catch (error) {
        console.warn("API health check failed:", error);

        setApiReady(true);
      } finally {
        setLoading(false);

        setTimeout(onFinish, 1000);
      }
    };

    checkAPI();
  }, [onFinish]);

  return (
    <View style={styles.splashContainer}>
      <Text style={styles.splashTitle}>🏆</Text>
      <Text style={styles.splashAppName}>Leaderboard</Text>
      <Text style={styles.splashSubtitle}>Scalable Ranking System</Text>
      {!apiReady && (
        <View style={styles.warningContainer}>
          <Text style={styles.warningText}>⚠️ API Unavailable</Text>
        </View>
      )}
      <ActivityIndicator size="large" color="#007AFF" style={styles.spinner} />
    </View>
  );
};

const TabNavigator: React.FC = () => {
  return (
    <Tab.Navigator
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: "#007AFF",
        tabBarInactiveTintColor: "#999999",
        tabBarStyle: styles.tabBar,
        tabBarLabelStyle: styles.tabBarLabel,
      }}>
      <Tab.Screen
        name="Leaderboard"
        component={LeaderboardScreen}
        options={{
          tabBarLabel: "Leaderboard",
          tabBarIcon: ({ color }) => <Text style={{ fontSize: 24 }}>📊</Text>,
        }}
      />
      <Tab.Screen
        name="Search"
        component={SearchScreen}
        options={{
          tabBarLabel: "Search",
          tabBarIcon: ({ color }) => <Text style={{ fontSize: 24 }}>🔍</Text>,
        }}
      />
      <Tab.Screen
        name="Profile"
        component={ProfileScreen}
        options={{
          tabBarLabel: "Profile",
          tabBarIcon: ({ color }) => <Text style={{ fontSize: 24 }}>👤</Text>,
        }}
      />
    </Tab.Navigator>
  );
};

export const RootNavigator: React.FC = () => {
  const [isSplashFinished, setIsSplashFinished] = useState(false);

  if (!isSplashFinished) {
    return <SplashScreen onFinish={() => setIsSplashFinished(true)} />;
  }

  return (
    <NavigationContainer>
      <Stack.Navigator screenOptions={{ headerShown: false }}>
        <Stack.Screen name="Main" component={TabNavigator} />
      </Stack.Navigator>
    </NavigationContainer>
  );
};

const styles = StyleSheet.create({
  splashContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
    backgroundColor: "#007AFF",
  },
  splashTitle: {
    fontSize: 80,
    marginBottom: 20,
  },
  splashAppName: {
    fontSize: 32,
    fontWeight: "bold",
    color: "#FFFFFF",
    marginBottom: 8,
  },
  splashSubtitle: {
    fontSize: 14,
    color: "#E0E0E0",
    marginBottom: 40,
  },
  warningContainer: {
    marginBottom: 20,
  },
  warningText: {
    fontSize: 12,
    color: "#FFD700",
    fontWeight: "600",
  },
  spinner: {
    marginTop: 20,
  },
  tabBar: {
    backgroundColor: "#FFFFFF",
    borderTopWidth: 1,
    borderTopColor: "#F0F0F0",
    paddingBottom: 4,
    height: 60,
  },
  tabBarLabel: {
    fontSize: 11,
    marginTop: -4,
  },
});
