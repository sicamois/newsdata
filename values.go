package newsdata

// allowedBinaries defines the valid values for pseudo-boolean fields.
var allowedBinaries = []string{"1", "0"}

// allowedCategories defines the valid categories for the BreakingNewsRequest.
var allowedCategories = []string{
	"business", "crime", "domestic", "education", "entertainment",
	"environment", "food", "health", "lifestyle", "other",
	"politics", "science", "sports", "technology", "top", "tourism", "world",
}

// allowedCountries defines the valid countries for the BreakingNewsRequest.
var allowedCountries = []string{
	"af", "al", "dz", "ad", "ao", "ar",
	"am", "au", "at", "az", "bs", "bh",
	"bd", "bb", "by", "be", "bz", "bj",
	"bm", "bt", "bo", "ba", "bw", "br",
	"bn", "bg", "bf", "bi", "kh", "cm",
	"ca", "cv", "ky", "cf", "td", "cl",
	"cn", "co", "km", "cg", "ck", "cr",
	"hr", "cu", "cw", "cy", "cz", "dk",
	"dj", "dm", "do", "cd", "ec", "eg",
	"sv", "gq", "er", "ee", "sz", "et",
	"fj", "fi", "fr", "pf", "ga", "gm",
	"ge", "de", "gh", "gi", "gr", "gd",
	"gt", "gn", "gy", "ht", "hn", "hk",
	"hu", "is", "in", "id", "ir", "iq",
	"ie", "il", "it", "ci", "jm", "jp",
	"je", "jo", "kz", "ke", "ki", "xk",
	"kw", "kg", "la", "lv", "lb", "ls",
	"lr", "ly", "li", "lt", "lu", "mo",
	"mk", "mg", "mw", "my", "mv", "ml",
	"mt", "mh", "mr", "mu", "mx", "fm",
	"md", "mc", "mn", "me", "ma", "mz",
	"mm", "na", "nr", "np", "nl", "nc",
	"nz", "ni", "ne", "ng", "kp", "no",
	"om", "pk", "pw", "ps", "pa", "pg",
	"py", "pe", "ph", "pl", "pt", "pr",
	"qa", "ro", "ru", "rw", "lc", "sx",
	"ws", "sm", "st", "sa", "sn", "rs",
	"sc", "sl", "sg", "sk", "si", "sb",
	"so", "za", "kr", "es", "lk", "sd",
	"sr", "se", "ch", "sy", "tw", "tj",
	"tz", "th", "tl", "tg", "to", "tt",
	"tn", "tr", "tm", "tv", "ug", "ua",
	"ae", "gb", "us", "uy", "uz", "vu",
	"va", "ve", "vi", "vg", "wo", "ye",
	"zm", "zw",
}

// allowedLanguages defines the valid languages for the BreakingNewsRequest.
var allowedLanguages = []string{
	"af", "sq", "am", "ar", "hy", "as", "az", "bm", "eu", "be", "bn", "bs", "bg", "my", "ca", "ckb", "zh", "hr", "cs", "da", "nl", "en", "et", "pi", "fi", "fr", "gl", "ka", "de", "el", "gu", "ha", "he", "hi", "hu", "is", "id", "it", "jp", "kn", "kz", "kh", "rw", "ko", "ku", "lv", "lt", "lb", "mk", "ms", "ml", "mt", "mi", "mr", "mn", "ne", "no", "or", "ps", "fa", "pl", "pt", "pa", "ro", "ru", "sm", "sr", "sn", "sd", "si", "sk", "sl", "so", "es", "sw", "sv", "tg", "ta", "te", "th", "zht", "tr", "tk", "uk", "ur", "uz", "vi", "cy", "zu",
}

// allowedPriorityDomain defines the valid priority domains for the BreakingNewsRequest.
var allowedPriorityDomains = []string{
	"top", "medium", "low",
}

// allowedSentiment defines the valid sentiment for the BreakingNewsRequest.
var allowedSentiments = []string{
	"positive", "negative", "neutral",
}

// allowedTags defines the valid tags for the BreakingNewsRequest.
var allowedTags = []string{
	"adoption", "blockchain", "coin_fundamental", "competition", "developers_community", "economy", "education", "exchange", "founders_investors", "general", "geopolitics", "global_markets", "government", "liquidity", "mining", "scam", "security_privacy", "sentiments", "supply", "technical_analysis", "technology",
}
