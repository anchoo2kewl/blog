package utils

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// FormatRelativeTime converts a date string to relative time format like "10 months ago"
func FormatRelativeTime(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	// Parse the date - try both RFC3339 and other common formats
	var t time.Time
	var err error

	// Try RFC3339 format first (common for database timestamps)
	if t, err = time.Parse(time.RFC3339, dateStr); err != nil {
		// Try RFC3339Nano format
		if t, err = time.Parse(time.RFC3339Nano, dateStr); err != nil {
			// Try layout with timezone
			if t, err = time.Parse("2006-01-02T15:04:05.999999Z07:00", dateStr); err != nil {
				// If all parsing fails, return original string
				return dateStr
			}
		}
	}

	return RelativeTimeFromTime(t)
}

// RelativeTimeFromTime converts a time.Time to relative time format
func RelativeTimeFromTime(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	if duration < 0 {
		return "in the future"
	}

	// Less than a minute
	if duration < time.Minute {
		seconds := int(duration.Seconds())
		if seconds <= 1 {
			return "just now"
		}
		return fmt.Sprintf("%d seconds ago", seconds)
	}

	// Less than an hour
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	// Less than a day
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	// Less than a week
	if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	// Less than a month (approximately 30 days)
	if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / (7 * 24))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}

	// Less than a year (approximately 365 days)
	if duration < 365*24*time.Hour {
		months := int(duration.Hours() / (30 * 24))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}

	// More than a year
	years := int(duration.Hours() / (365 * 24))
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}

// FormatFriendlyDate formats a date string to "January 2, 2006" format
func FormatFriendlyDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	var t time.Time
	var err error

	// Try RFC3339 format first
	if t, err = time.Parse(time.RFC3339, dateStr); err != nil {
		// Try RFC3339Nano format
		if t, err = time.Parse(time.RFC3339Nano, dateStr); err != nil {
			// Try layout with timezone
			if t, err = time.Parse("2006-01-02T15:04:05.999999Z07:00", dateStr); err != nil {
				// If all parsing fails, return original string
				return dateStr
			}
		}
	}

	return t.Format("January 2, 2006")
}

// CalculateReadingTime estimates reading time based on word count (200 words per minute)
func CalculateReadingTime(content string) int {
	if content == "" {
		return 1
	}

	// Simple word count (split by whitespace)
	words := len(strings.Fields(content))
	
	// Assume 200 words per minute reading speed
	readingMinutes := int(math.Ceil(float64(words) / 200.0))
	
	if readingMinutes < 1 {
		return 1
	}
	
	return readingMinutes
}