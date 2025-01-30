package config

type Filter string
type LikeStatus string

// Use constants for string-based enums
const (
	Food     Filter = "FOOD"
	Travel   Filter = "TRAVEL"
	Shopping Filter = "SHOPPING"
)

const (
	Liked    LikeStatus = "LIKED"
	NotLiked LikeStatus = "NOT_LIKED"
)
