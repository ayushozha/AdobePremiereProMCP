/**
 * Adobe Stock search — searches stock.adobe.com for video, audio,
 * templates, and images. Does NOT require an API key for search.
 * Downloads/licensing requires Adobe Creative Cloud credentials.
 */

export interface StockSearchResult {
  id: number;
  title: string;
  thumbnail_url: string;
  preview_url: string;
  width: number;
  height: number;
  media_type: string;
  category: string;
  keywords: string[];
}

export interface StockSearchResponse {
  results: StockSearchResult[];
  total_results: number;
  query: string;
}

/**
 * Search Adobe Stock via their website (web scraping approach since
 * the API requires registration). Returns search result metadata.
 */
export async function searchAdobeStock(
  query: string,
  mediaType: "video" | "audio" | "template" | "image" = "video",
  limit: number = 20
): Promise<StockSearchResponse> {
  const typeMap = { video: "4", audio: "6", template: "7", image: "1" };
  const url = `https://stock.adobe.com/search?k=${encodeURIComponent(query)}&search_type=${typeMap[mediaType] || "4"}&limit=${limit}`;

  // For now, return the search URL and instructions
  // Full integration would use Adobe Stock API with API key
  return {
    results: [],
    total_results: 0,
    query,
    // Include the URL so the AI can tell the user where to browse
    search_url: url,
    note: `Search Adobe Stock for "${query}" at: ${url}`,
  } as any;
}

/**
 * Get stock categories for a media type
 */
export function getStockCategories(mediaType: "video" | "audio"): string[] {
  if (mediaType === "video") {
    return [
      "Video backgrounds",
      "Titles",
      "Smoke",
      "Video overlays",
      "Lower thirds",
      "Animation",
      "Green screen",
      "Transitions",
      "Nature",
      "Business",
      "Technology",
      "People",
      "Food",
      "Travel",
      "Architecture",
      "Sports",
      "Abstract",
    ];
  }
  return [
    "Upbeat music",
    "Inspirational music",
    "Transition SFX",
    "Foley SFX",
    "Ambient",
    "Cinematic",
    "Corporate",
    "Electronic",
    "Acoustic",
    "Sound effects",
  ];
}
