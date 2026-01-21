import React, { useState, useCallback, useRef, useEffect } from "react";
import {
  View,
  TextInput,
  StyleSheet,
  ActivityIndicator,
  Text,
  TouchableOpacity,
  SafeAreaView,
  Keyboard,
} from "react-native";
import { useSearch, useUserRank } from "../hooks/useAPI";

const SearchScreen: React.FC = () => {
  const { query, result, loading, error, debouncedSearch, clearSearch } =
    useSearch();
  const { user: userDetails, refreshing } = useUserRank(
    result?.user?.id || null,
  );
  const [isFocused, setIsFocused] = useState(false);
  const searchInputRef = useRef<TextInput>(null);

  const handleSearchChange = useCallback(
    (text: string) => {
      debouncedSearch(text);
    },
    [debouncedSearch],
  );

  const handleClearSearch = useCallback(() => {
    clearSearch();
    searchInputRef.current?.clear();
  }, [clearSearch]);

  const renderSearchResult = () => {
    if (!result?.found) {
      return (
        <View style={styles.noResultContainer}>
          <Text style={styles.noResultText}>No user found</Text>
          <Text style={styles.noResultSubtext}>
            Try searching with a different username
          </Text>
        </View>
      );
    }

    const displayUser = userDetails || result.user;

    return (
      <View style={styles.resultCard}>
        <View style={styles.rankBadge}>
          <Text style={styles.rankBadgeText}>#{displayUser.rank}</Text>
        </View>

        <View style={styles.userInfoSection}>
          <Text style={styles.usernameDisplay}>{displayUser.username}</Text>
          <Text style={styles.ratingDisplay}>Rating: {displayUser.rating}</Text>
        </View>

        <View style={styles.liveIndicator}>
          <View style={styles.liveDot} />
          <Text style={styles.liveText}>Live</Text>
        </View>

        {refreshing && (
          <View style={styles.refreshingIndicator}>
            <ActivityIndicator size="small" color="#007AFF" />
          </View>
        )}

        <View style={styles.statsContainer}>
          <View style={styles.statItem}>
            <Text style={styles.statLabel}>Your Rank</Text>
            <Text style={styles.statValue}>#{displayUser.rank}</Text>
          </View>
          <View style={styles.statDivider} />
          <View style={styles.statItem}>
            <Text style={styles.statLabel}>Rating</Text>
            <Text style={styles.statValue}>{displayUser.rating}</Text>
          </View>
        </View>

        <View style={styles.progressSection}>
          <Text style={styles.progressLabel}>Ranking Level</Text>
          <View style={styles.progressBarContainer}>
            <View
              style={[
                styles.progressBar,
                {
                  width: `${((displayUser.rating - 100) / (5000 - 100)) * 100}%`,
                },
              ]}
            />
          </View>
          <View style={styles.progressLabels}>
            <Text style={styles.progressMinLabel}>Min: 100</Text>
            <Text style={styles.progressMaxLabel}>Max: 5000</Text>
          </View>
        </View>
      </View>
    );
  };

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>Search Player</Text>
        <Text style={styles.subtitle}>Find your global rank</Text>
      </View>

      <View
        style={[
          styles.searchContainer,
          isFocused && styles.searchContainerFocused,
        ]}>
        <Text style={styles.searchIcon}>🔍</Text>
        <TextInput
          ref={searchInputRef}
          style={styles.searchInput}
          placeholder="Enter username..."
          placeholderTextColor="#CCCCCC"
          onChangeText={handleSearchChange}
          onFocus={() => setIsFocused(true)}
          onBlur={() => setIsFocused(false)}
          returnKeyType="search"
          onSubmitEditing={Keyboard.dismiss}
        />
        {query && (
          <TouchableOpacity onPress={handleClearSearch}>
            <Text style={styles.clearButton}>✕</Text>
          </TouchableOpacity>
        )}
      </View>

      {loading && (
        <View style={styles.statusContainer}>
          <ActivityIndicator size="large" color="#007AFF" />
          <Text style={styles.statusText}>Searching...</Text>
        </View>
      )}

      {error && (
        <View style={styles.errorContainer}>
          <Text style={styles.errorIcon}>⚠️</Text>
          <Text style={styles.errorText}>{error}</Text>
        </View>
      )}

      {result && !loading && (
        <View style={styles.resultsContainer}>{renderSearchResult()}</View>
      )}

      {!result && !loading && (
        <View style={styles.emptyStateContainer}>
          <Text style={styles.emptyStateIcon}>🔎</Text>
          <Text style={styles.emptyStateText}>
            {query ? "Searching..." : "Search by username"}
          </Text>
          <Text style={styles.emptyStateSubtext}>
            {query ? "" : "Enter a username to find player rank"}
          </Text>
        </View>
      )}
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
  searchContainer: {
    flexDirection: "row",
    alignItems: "center",
    marginHorizontal: 16,
    marginVertical: 16,
    paddingHorizontal: 12,
    borderRadius: 8,
    backgroundColor: "#FFFFFF",
    borderWidth: 2,
    borderColor: "#E0E0E0",
  },
  searchContainerFocused: {
    borderColor: "#007AFF",
  },
  searchIcon: {
    fontSize: 20,
    marginRight: 8,
  },
  searchInput: {
    flex: 1,
    paddingVertical: 12,
    fontSize: 16,
    color: "#000000",
  },
  clearButton: {
    fontSize: 20,
    color: "#666666",
    padding: 8,
  },
  statusContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  statusText: {
    marginTop: 12,
    fontSize: 16,
    color: "#666666",
  },
  errorContainer: {
    marginHorizontal: 16,
    marginTop: 12,
    paddingHorizontal: 16,
    paddingVertical: 12,
    backgroundColor: "#FFEBEE",
    borderRadius: 8,
    flexDirection: "row",
    alignItems: "center",
  },
  errorIcon: {
    fontSize: 20,
    marginRight: 12,
  },
  errorText: {
    flex: 1,
    fontSize: 14,
    color: "#D32F2F",
  },
  resultsContainer: {
    flex: 1,
    paddingHorizontal: 16,
    paddingVertical: 16,
    justifyContent: "center",
  },
  resultCard: {
    backgroundColor: "#FFFFFF",
    borderRadius: 12,
    padding: 20,
    shadowColor: "#000000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  rankBadge: {
    alignSelf: "flex-start",
    backgroundColor: "#FFD700",
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 20,
    marginBottom: 16,
  },
  rankBadgeText: {
    fontSize: 16,
    fontWeight: "bold",
    color: "#000000",
  },
  userInfoSection: {
    marginBottom: 16,
  },
  usernameDisplay: {
    fontSize: 24,
    fontWeight: "bold",
    color: "#000000",
    marginBottom: 4,
  },
  ratingDisplay: {
    fontSize: 16,
    color: "#666666",
  },
  liveIndicator: {
    flexDirection: "row",
    alignItems: "center",
    marginBottom: 16,
  },
  liveDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    backgroundColor: "#4CAF50",
    marginRight: 6,
  },
  liveText: {
    fontSize: 12,
    color: "#4CAF50",
    fontWeight: "600",
  },
  refreshingIndicator: {
    position: "absolute",
    top: 20,
    right: 20,
  },
  statsContainer: {
    flexDirection: "row",
    backgroundColor: "#F5F5F5",
    borderRadius: 8,
    paddingVertical: 12,
    marginBottom: 16,
  },
  statItem: {
    flex: 1,
    alignItems: "center",
  },
  statDivider: {
    width: 1,
    backgroundColor: "#E0E0E0",
  },
  statLabel: {
    fontSize: 12,
    color: "#999999",
    marginBottom: 4,
  },
  statValue: {
    fontSize: 18,
    fontWeight: "bold",
    color: "#007AFF",
  },
  progressSection: {
    marginTop: 16,
  },
  progressLabel: {
    fontSize: 14,
    fontWeight: "600",
    color: "#333333",
    marginBottom: 8,
  },
  progressBarContainer: {
    height: 8,
    backgroundColor: "#E0E0E0",
    borderRadius: 4,
    overflow: "hidden",
    marginBottom: 8,
  },
  progressBar: {
    height: "100%",
    backgroundColor: "#007AFF",
    borderRadius: 4,
  },
  progressLabels: {
    flexDirection: "row",
    justifyContent: "space-between",
  },
  progressMinLabel: {
    fontSize: 12,
    color: "#999999",
  },
  progressMaxLabel: {
    fontSize: 12,
    color: "#999999",
  },
  noResultContainer: {
    alignItems: "center",
    paddingVertical: 32,
  },
  noResultText: {
    fontSize: 18,
    fontWeight: "600",
    color: "#666666",
    marginBottom: 8,
  },
  noResultSubtext: {
    fontSize: 14,
    color: "#999999",
  },
  emptyStateContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  emptyStateIcon: {
    fontSize: 64,
    marginBottom: 16,
  },
  emptyStateText: {
    fontSize: 18,
    fontWeight: "600",
    color: "#333333",
    marginBottom: 8,
  },
  emptyStateSubtext: {
    fontSize: 14,
    color: "#999999",
  },
});

export default SearchScreen;
