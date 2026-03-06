package consts

import "math/rand/v2"

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.3124.85",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:136.0) Gecko/20100101 Firefox/136.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 OPR/118.0.0.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 " +
		"(KHTML, like Gecko) Version/17.10 Safari/605.1.15",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.3124.85",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14.7; rv:136.0) Gecko/20100101 Firefox/136.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64; rv:136.0) Gecko/20100101 Firefox/136.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 OPR/118.0.0.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_7 like Mac OS X) AppleWebKit/605.1.15 " +
		"(KHTML, like Gecko) CriOS/134.0.6998.99 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 18_3_2 like Mac OS X) AppleWebKit/605.1.15 " +
		"(KHTML, like Gecko) Version/18.3.1 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_7_2 like Mac OS X) AppleWebKit/605.1.15 " +
		"(KHTML, like Gecko) Version/18.0 EdgiOS/134.3124.77 Mobile/15E148 Safari/605.1.15",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_4 like Mac OS X) AppleWebKit/605.1.15 " +
		"(KHTML, like Gecko) FxiOS/136.0 Mobile/15E148 Safari/605.1.15",
	"Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/134.0.0.0 Mobile Safari/537.3",
	"Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/134.0.6998.135 Mobile Safari/537.36 EdgA/134.0.3124.68",
	"Mozilla/5.0 (Android 15; Mobile; rv:136.0) Gecko/136.0 Firefox/136.0",
	"Mozilla/5.0 (Linux; Android 10; SM-G970F) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/134.0.6998.135 Mobile Safari/537.36 OPR/76.2.4027.73374",
	"Mozilla/5.0 (iPad; CPU OS 17_7_2 like Mac OS X) AppleWebKit/605.1.15 " +
		"(KHTML, like Gecko) Version/18.3 Mobile/15E148 Safari/604.1",
}

// GetRandomUserAgent returns a random user agent string from the pool.
func GetRandomUserAgent() string {
	// #nosec G404 -- Weak random is acceptable for selecting a random user agent from a pool
	return userAgents[rand.IntN(len(userAgents))]
}
