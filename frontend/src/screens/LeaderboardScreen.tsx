import React, { useState, useCallback } from "react";
import {
  View,
  TextInput,
  StyleSheet,
  ActivityIndicator,
  FlatList,
  Text,
  TouchableOpacity,
  RefreshControl,
  SafeAreaView,
  ScrollView,
} from "react-native";
import { useLeaderboard } from "../hooks/useAPI";

/**
 * Leaderboard Screen Component
 * Displays paginated leaderboard with real-time ranking
 * Features:
 * - Pagination support
 * - Pull-to-refresh
 * - Loading states
 * - Error handling
 * - Responsive layout
 */
const LeaderboardScreen: React.FC = () => {
  const {
    data,
    loading,
    error,
    currentPage,
    nextPage,
    prevPage,
    refresh,
    fetch: fetchLeaderboard,
  } = useLeaderboard();
  const [refreshing, setRefreshing] = useState(false);

  // Always fetch leaderboard on mount
  React.useEffect(() => {
    fetchLeaderboard(1);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Handle pull-to-refresh gesture
  const onRefresh = useCallback(() => {
    setRefreshing(true);
    Promise.resolve(fetchLeaderboard(currentPage)).finally(() =>
      setRefreshing(false),
    );
  }, [fetchLeaderboard, currentPage]);

  // Render individual leaderboard entry
  const renderEntry = ({ item, index }: { item: any; index: number }) => {
    const isHighRank = item.rank <= 10; // Highlight top 10
    const isMediumRank = item.rank <= 100;
    // Use a color palette for avatars
    const avatarColors = [
      "#FFD700",
      "#FF8C00",
      "#1E90FF",
      "#32CD32",
      "#FF69B4",
      "#8A2BE2",
      "#00CED1",
      "#FF6347",
      "#20B2AA",
      "#FFB6C1",
    ];
    const avatarColor = avatarColors[item.rank % avatarColors.length];

    return (
      <View
        style={[
          styles.entryContainer,
          isHighRank && styles.topRankEntry,
          isMediumRank && !isHighRank && styles.mediumRankEntry,
        ]}>
        <View style={styles.rankContainer}>
          <Text style={[styles.rankText, isHighRank && styles.topRankText]}>
            #{item.rank}
          </Text>
        </View>
        <View style={styles.avatarContainer}>
          <View style={[styles.avatarCircle, { backgroundColor: avatarColor }]}>
            <Text style={styles.avatarText}>
              {item.username?.charAt(0)?.toUpperCase() || "?"}
            </Text>
          </View>
        </View>
        <View style={styles.userInfoContainer}>
          <Text style={styles.usernameText} numberOfLines={1}>
            {item.username}
          </Text>
          <Text style={styles.ratingText}>Rating: {item.rating}</Text>
        </View>
        <View style={styles.badgeContainer}>
          {isHighRank && <Text style={styles.badgeText}>⭐ Top 10</Text>}
        </View>
      </View>
    );
  };

  // Render pagination controls
  const renderPagination = () => (
    <View style={styles.paginationContainer}>
      <TouchableOpacity
        style={[
          styles.paginationButton,
          currentPage === 1 && styles.disabledButton,
        ]}
        onPress={prevPage}
        disabled={currentPage === 1}>
        <Text style={styles.paginationButtonText}>← Previous</Text>
      </TouchableOpacity>

      <Text style={styles.pageInfo}>
        Page {data?.page} of{" "}
        {Math.ceil((data?.total || 0) / (data?.page_size || 50))}
      </Text>

      <TouchableOpacity
        style={[
          styles.paginationButton,
          !data?.has_more && styles.disabledButton,
        ]}
        onPress={nextPage}
        disabled={!data?.has_more}>
        <Text style={styles.paginationButtonText}>Next →</Text>
      </TouchableOpacity>
    </View>
  );

  if (loading && !data) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#007AFF" />
          <Text style={styles.loadingText}>Loading leaderboard...</Text>
        </View>
      </SafeAreaView>
    );
  }

  if (error && !data) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.errorContainer}>
          <Text style={styles.errorText}>Error: {error}</Text>
          <TouchableOpacity style={styles.retryButton} onPress={refresh}>
            <Text style={styles.retryButtonText}>Retry</Text>
          </TouchableOpacity>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.headerShadow}>
        <View style={styles.headerContent}>
          <Text style={styles.title}>🏆 Leaderboard</Text>
          <Text style={styles.subtitle}>Total Players: {data?.total || 0}</Text>
        </View>
      </View>

      <FlatList
        data={data?.entries || []}
        renderItem={renderEntry}
        keyExtractor={(item, index) => `${item.rank}-${index}`}
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={styles.emptyIcon}>😶‍🌫️</Text>
            <Text style={styles.emptyText}>No leaderboard entries yet</Text>
            <Text style={styles.emptySubtext}>
              Be the first to join the leaderboard!
            </Text>
          </View>
        }
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
        ListFooterComponent={
          data && data.entries.length > 0 ? renderPagination() : null
        }
        scrollEnabled={true}
        style={styles.listContainer}
        contentContainerStyle={{ paddingBottom: 32 }}
      />
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#F4F7FA",
    alignItems: "center",
    width: "100%",
  },
  headerShadow: {
    backgroundColor: "#fff",
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.08,
    shadowRadius: 8,
    elevation: 4,
    borderBottomLeftRadius: 18,
    borderBottomRightRadius: 18,
    marginBottom: 8,
    width: "100%",
    maxWidth: 600,
    alignSelf: "center",
  },
  headerContent: {
    backgroundColor: "#007AFF",
    borderBottomLeftRadius: 18,
    borderBottomRightRadius: 18,
    paddingHorizontal: 24,
    paddingVertical: 28,
    alignItems: "center",
    width: "100%",
  },
  title: {
    fontSize: 32,
    fontWeight: "bold",
    color: "#fff",
    letterSpacing: 1.2,
    marginBottom: 2,
    textAlign: "center",
  },
  subtitle: {
    fontSize: 16,
    color: "#E0E0E0",
    marginTop: 2,
    fontWeight: "500",
    textAlign: "center",
  },
  listContainer: {
    flex: 1,
    width: "100%",
    maxWidth: 600,
    alignSelf: "center",
    paddingHorizontal: 0,
  },
  entryContainer: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 22,
    paddingVertical: 18,
    marginHorizontal: 18,
    marginVertical: 10,
    borderRadius: 22,
    backgroundColor: "#fff",
    borderLeftWidth: 0,
    shadowColor: "#007AFF",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.13,
    shadowRadius: 12,
    elevation: 4,
    width: "100%",
    maxWidth: 600,
    alignSelf: "center",
    // Gradient effect (mobile only, fallback for web)
    overflow: "hidden",
  },
  avatarContainer: {
    marginRight: 16,
    alignItems: "center",
    justifyContent: "center",
  },
  avatarCircle: {
    width: 44,
    height: 44,
    borderRadius: 22,
    alignItems: "center",
    justifyContent: "center",
    marginRight: 2,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
    elevation: 2,
  },
  avatarText: {
    color: "#fff",
    fontWeight: "bold",
    fontSize: 22,
    letterSpacing: 0.5,
  },
  topRankEntry: {
    backgroundColor: "#FFF9E6",
    borderWidth: 2,
    borderColor: "#FFD700",
    shadowColor: "#FFD700",
    shadowOpacity: 0.18,
    elevation: 6,
  },
  mediumRankEntry: {
    backgroundColor: "#F0F8FF",
    borderWidth: 1,
    borderColor: "#4169E1",
    shadowColor: "#4169E1",
    shadowOpacity: 0.12,
    elevation: 4,
  },
  rankContainer: {
    width: 60,
    justifyContent: "center",
    alignItems: "center",
  },
  rankText: {
    fontSize: 22,
    fontWeight: "bold",
    color: "#666666",
    letterSpacing: 0.5,
    textAlign: "center",
  },
  topRankText: {
    color: "#FFD700",
    fontSize: 26,
    textShadowColor: "#FFD700",
    textShadowOffset: { width: 0, height: 1 },
    textShadowRadius: 2,
    textAlign: "center",
  },
  userInfoContainer: {
    flex: 1,
    marginHorizontal: 18,
    minWidth: 120,
  },
  usernameText: {
    fontSize: 20,
    fontWeight: "700",
    color: "#222",
    letterSpacing: 0.2,
    textAlign: "left",
  },
  ratingText: {
    fontSize: 15,
    color: "#666",
    marginTop: 2,
    fontWeight: "500",
    textAlign: "left",
  },
  badgeContainer: {
    marginLeft: 12,
  },
  badgeText: {
    fontSize: 15,
    fontWeight: "bold",
    color: "#FFD700",
    textShadowColor: "#FFD700",
    textShadowOffset: { width: 0, height: 1 },
    textShadowRadius: 2,
  },
  paginationContainer: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    paddingHorizontal: 32,
    paddingVertical: 20,
    marginTop: 14,
    width: "100%",
    maxWidth: 600,
    alignSelf: "center",
  },
  paginationButton: {
    paddingHorizontal: 22,
    paddingVertical: 14,
    backgroundColor: "#007AFF",
    borderRadius: 12,
    shadowColor: "#007AFF",
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.12,
    shadowRadius: 2,
    elevation: 1,
  },
  disabledButton: {
    backgroundColor: "#CCCCCC",
    opacity: 0.5,
  },
  paginationButtonText: {
    color: "#FFFFFF",
    fontWeight: "700",
    fontSize: 16,
    letterSpacing: 0.5,
  },
  pageInfo: {
    fontSize: 16,
    fontWeight: "700",
    color: "#333",
    textAlign: "center",
  },
  loadingContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  loadingText: {
    marginTop: 12,
    fontSize: 16,
    color: "#666666",
  },
  errorContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
    paddingHorizontal: 32,
  },
  errorText: {
    fontSize: 16,
    color: "#D32F2F",
    textAlign: "center",
    marginBottom: 16,
  },
  retryButton: {
    paddingHorizontal: 24,
    paddingVertical: 12,
    backgroundColor: "#007AFF",
    borderRadius: 8,
  },
  retryButtonText: {
    color: "#FFFFFF",
    fontWeight: "600",
  },
  emptyContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
    paddingVertical: 48,
    paddingHorizontal: 18,
    width: "100%",
    maxWidth: 600,
    alignSelf: "center",
  },
  emptyText: {
    fontSize: 20,
    color: "#999",
    fontWeight: "600",
    marginBottom: 10,
    textAlign: "center",
  },
  emptyIcon: {
    fontSize: 44,
    marginBottom: 10,
    textAlign: "center",
  },
  emptySubtext: {
    fontSize: 15,
    color: "#aaa",
    marginTop: 2,
    textAlign: "center",
  },
});

export default LeaderboardScreen;
