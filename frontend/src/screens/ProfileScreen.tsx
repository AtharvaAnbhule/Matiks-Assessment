import React, { useState, useCallback } from "react";
import {
  View,
  TextInput,
  StyleSheet,
  ActivityIndicator,
  Text,
  TouchableOpacity,
  SafeAreaView,
  ScrollView,
  Alert,
} from "react-native";
import { userAPI } from "../services/api";

const ProfileScreen: React.FC = () => {
  const [userId, setUserId] = useState("");
  const [username, setUsername] = useState("");
  const [initialRating, setInitialRating] = useState("1000");
  const [userCreated, setUserCreated] = useState(false);
  const [loading, setLoading] = useState(false);
  const [newRating, setNewRating] = useState("");
  const [updatingRating, setUpdatingRating] = useState(false);
  const [userInfo, setUserInfo] = useState<any>(null);

  const handleCreateUser = useCallback(async () => {
    if (!userId.trim()) {
      Alert.alert("Error", "User ID is required");
      return;
    }

    if (!username.trim() || username.length < 3) {
      Alert.alert("Error", "Username must be at least 3 characters");
      return;
    }

    const rating = parseInt(initialRating);
    if (isNaN(rating) || rating < 100 || rating > 5000) {
      Alert.alert("Error", "Rating must be between 100 and 5000");
      return;
    }

    try {
      setLoading(true);
      const result = await userAPI.createUser(userId, username, rating);
      setUserInfo(result);
      setUserCreated(true);
      setNewRating(rating.toString());
      Alert.alert("Success", `User ${username} created successfully!`);
    } catch (error: any) {
      Alert.alert(
        "Error",
        error?.response?.data?.message || "Failed to create user",
      );
    } finally {
      setLoading(false);
    }
  }, [userId, username, initialRating]);

  const handleUpdateRating = useCallback(async () => {
    if (!userInfo?.id) {
      Alert.alert("Error", "User not loaded");
      return;
    }

    const rating = parseInt(newRating);
    if (isNaN(rating) || rating < 100 || rating > 5000) {
      Alert.alert("Error", "Rating must be between 100 and 5000");
      return;
    }

    try {
      setUpdatingRating(true);
      const result = await userAPI.updateRating(userInfo.id, rating);
      setUserInfo(result);
      Alert.alert("Success", `Rating updated to ${rating}!`);
    } catch (error: any) {
      Alert.alert(
        "Error",
        error?.response?.data?.message || "Failed to update rating",
      );
    } finally {
      setUpdatingRating(false);
    }
  }, [userInfo?.id, newRating]);

  const handleFetchUser = useCallback(async () => {
    if (!userId.trim()) {
      Alert.alert("Error", "User ID is required");
      return;
    }

    try {
      setLoading(true);
      const result = await userAPI.getUser(userId);
      setUserInfo(result);
      setNewRating(result.rating.toString());
    } catch (error: any) {
      Alert.alert("Error", error?.response?.data?.message || "User not found");
    } finally {
      setLoading(false);
    }
  }, [userId]);

  const handleReset = useCallback(() => {
    setUserId("");
    setUsername("");
    setInitialRating("1000");
    setUserCreated(false);
    setUserInfo(null);
    setNewRating("");
  }, []);

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView showsVerticalScrollIndicator={false}>
        <View style={styles.header}>
          <Text style={styles.title}>Profile</Text>
          <Text style={styles.subtitle}>Manage your account</Text>
        </View>

        {!userCreated ? (
          <View style={styles.formSection}>
            <Text style={styles.sectionTitle}>Create New User</Text>

            <View style={styles.inputGroup}>
              <Text style={styles.label}>User ID</Text>
              <TextInput
                style={styles.input}
                placeholder="Enter unique user ID"
                placeholderTextColor="#CCCCCC"
                value={userId}
                onChangeText={setUserId}
                editable={!loading}
              />
            </View>

            <View style={styles.inputGroup}>
              <Text style={styles.label}>Username</Text>
              <TextInput
                style={styles.input}
                placeholder="Enter display name (3-50 chars)"
                placeholderTextColor="#CCCCCC"
                value={username}
                onChangeText={setUsername}
                maxLength={50}
                editable={!loading}
              />
              <Text style={styles.helperText}>
                {username.length}/50 characters
              </Text>
            </View>

            <View style={styles.inputGroup}>
              <Text style={styles.label}>Initial Rating</Text>
              <TextInput
                style={styles.input}
                placeholder="Enter rating (100-5000)"
                placeholderTextColor="#CCCCCC"
                value={initialRating}
                onChangeText={setInitialRating}
                keyboardType="number-pad"
                editable={!loading}
              />
              <Text style={styles.helperText}>
                Range: 100 - 5000 (current: {initialRating})
              </Text>
            </View>

            <TouchableOpacity
              style={[
                styles.button,
                styles.primaryButton,
                loading && styles.disabledButton,
              ]}
              onPress={handleCreateUser}
              disabled={loading}>
              {loading ? (
                <ActivityIndicator color="#FFFFFF" size="small" />
              ) : (
                <Text style={styles.buttonText}>Create User</Text>
              )}
            </TouchableOpacity>

            <View style={styles.divider} />
            <Text style={styles.sectionTitle}>Or Load Existing User</Text>

            <View style={styles.inputGroup}>
              <Text style={styles.label}>User ID</Text>
              <TextInput
                style={styles.input}
                placeholder="Enter user ID to load"
                placeholderTextColor="#CCCCCC"
                value={userId}
                onChangeText={setUserId}
                editable={!loading}
              />
            </View>

            <TouchableOpacity
              style={[
                styles.button,
                styles.secondaryButton,
                loading && styles.disabledButton,
              ]}
              onPress={handleFetchUser}
              disabled={loading}>
              <Text style={styles.secondaryButtonText}>Load User</Text>
            </TouchableOpacity>
          </View>
        ) : (
          <View style={styles.profileSection}>
            <Text style={styles.sectionTitle}>User Profile</Text>

            <View style={styles.infoCard}>
              <View style={styles.infoItem}>
                <Text style={styles.infoLabel}>Username</Text>
                <Text style={styles.infoValue}>{userInfo?.username}</Text>
              </View>

              <View style={styles.infoDivider} />

              <View style={styles.infoItem}>
                <Text style={styles.infoLabel}>User ID</Text>
                <Text style={styles.infoValue} selectable={true}>
                  {userInfo?.id}
                </Text>
              </View>

              <View style={styles.infoDivider} />

              <View style={styles.infoItem}>
                <Text style={styles.infoLabel}>Current Rank</Text>
                <Text style={styles.infoValue}>#{userInfo?.rank}</Text>
              </View>

              <View style={styles.infoDivider} />

              <View style={styles.infoItem}>
                <Text style={styles.infoLabel}>Current Rating</Text>
                <Text style={styles.infoValue}>{userInfo?.rating}</Text>
              </View>
            </View>

            <View style={styles.updateSection}>
              <Text style={styles.sectionTitle}>Update Rating</Text>

              <View style={styles.inputGroup}>
                <Text style={styles.label}>New Rating</Text>
                <TextInput
                  style={styles.input}
                  placeholder="Enter new rating (100-5000)"
                  placeholderTextColor="#CCCCCC"
                  value={newRating}
                  onChangeText={setNewRating}
                  keyboardType="number-pad"
                  editable={!updatingRating}
                />
              </View>

              <TouchableOpacity
                style={[
                  styles.button,
                  styles.primaryButton,
                  updatingRating && styles.disabledButton,
                ]}
                onPress={handleUpdateRating}
                disabled={updatingRating}>
                {updatingRating ? (
                  <ActivityIndicator color="#FFFFFF" size="small" />
                ) : (
                  <Text style={styles.buttonText}>Update Rating</Text>
                )}
              </TouchableOpacity>
            </View>

            <TouchableOpacity
              style={[styles.button, styles.dangerButton]}
              onPress={handleReset}>
              <Text style={styles.buttonText}>Load Different User</Text>
            </TouchableOpacity>
          </View>
        )}
      </ScrollView>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#F5F5F5",
  },
  header: {
    backgroundColor: "#007AFF",
    paddingHorizontal: 16,
    paddingVertical: 24,
  },
  title: {
    fontSize: 28,
    fontWeight: "bold",
    color: "#FFFFFF",
  },
  subtitle: {
    fontSize: 14,
    color: "#E0E0E0",
    marginTop: 4,
  },
  formSection: {
    paddingHorizontal: 16,
    paddingVertical: 20,
  },
  profileSection: {
    paddingHorizontal: 16,
    paddingVertical: 20,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: "bold",
    color: "#000000",
    marginBottom: 16,
    marginTop: 12,
  },
  inputGroup: {
    marginBottom: 16,
  },
  label: {
    fontSize: 14,
    fontWeight: "600",
    color: "#333333",
    marginBottom: 8,
  },
  input: {
    borderWidth: 1,
    borderColor: "#E0E0E0",
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 12,
    fontSize: 16,
    color: "#000000",
    backgroundColor: "#FFFFFF",
  },
  helperText: {
    fontSize: 12,
    color: "#999999",
    marginTop: 4,
  },
  button: {
    paddingHorizontal: 16,
    paddingVertical: 14,
    borderRadius: 8,
    alignItems: "center",
    marginBottom: 12,
  },
  primaryButton: {
    backgroundColor: "#007AFF",
  },
  secondaryButton: {
    backgroundColor: "#FFFFFF",
    borderWidth: 1,
    borderColor: "#007AFF",
  },
  dangerButton: {
    backgroundColor: "#FF9500",
  },
  disabledButton: {
    opacity: 0.6,
  },
  buttonText: {
    color: "#FFFFFF",
    fontWeight: "600",
    fontSize: 16,
  },
  secondaryButtonText: {
    color: "#007AFF",
    fontWeight: "600",
    fontSize: 16,
  },
  divider: {
    height: 1,
    backgroundColor: "#E0E0E0",
    marginVertical: 20,
  },
  infoCard: {
    backgroundColor: "#FFFFFF",
    borderRadius: 12,
    paddingHorizontal: 16,
    paddingVertical: 16,
    marginBottom: 20,
    shadowColor: "#000000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  infoItem: {
    paddingVertical: 12,
  },
  infoLabel: {
    fontSize: 12,
    color: "#999999",
    marginBottom: 4,
  },
  infoValue: {
    fontSize: 16,
    fontWeight: "600",
    color: "#000000",
  },
  infoDivider: {
    height: 1,
    backgroundColor: "#F0F0F0",
  },
  updateSection: {
    marginBottom: 20,
  },
});

export default ProfileScreen;
