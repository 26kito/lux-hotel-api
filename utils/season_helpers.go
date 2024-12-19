package utils

import "time"

// Constants for price adjustments
const PeakSeasonIncrease = 0.50 // 50% increase for peak season
const LowSeasonDiscount = 0.10  // 10% discount for low season

// CheckSeason determines if the given date falls into peak, low, or normal season.
func CheckSeason(date time.Time) string {
	month := date.Month()

	switch month {
	case time.December, time.January, time.July: // Peak season months
		return "peak"
	case time.February, time.March, time.April, time.May, time.September, time.October, time.November: // Low season months
		return "low"
	default:
		return "normal"
	}
}

// AdjustPrice adjusts the price based on the season
func AdjustPrice(basePrice float64, checkInDate time.Time) float64 {
	season := CheckSeason(checkInDate)

	switch season {
	case "peak":
		return basePrice + (basePrice * PeakSeasonIncrease)
	case "low":
		return basePrice - (basePrice * LowSeasonDiscount)
	default:
		return basePrice // No adjustment for normal season
	}
}
