package resiliency

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RunJitterBackoff solves the "Thundering Herd" problem.
// If an Apple datacenter database crashes, 50,000 proxies lose connection simultaneously.
// When the database reboots, if all 50,000 proxies retry at the EXACT same time
// (e.g., fixed 2-second backoff), they will instantly crash the database again.
//
// Jitter is mathematical randomness applied to the exponential backoff calculation.
// This spreads the retry spikes out over a bell curve, allowing the database to survive.
func RunJitterBackoff() {
	fmt.Println("=== Exponential Backoff with Jitter ===")

	baseWait := 100.0 // Milliseconds
	maxWait := 5000.0 // 5 Seconds hard cap
	maxRetries := 5

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 1. Calculate standard Exponential Backoff: Base * (2 ^ Attempt)
		exponentialDelay := baseWait * math.Pow(2, float64(attempt))

		// Cap it to prevent waiting 10 minutes on attempt 15!
		if exponentialDelay > maxWait {
			exponentialDelay = maxWait
		}

		// 2. Add Full Jitter
		// Standard calculation: Random float between 0 and ExponentialDelay
		jitteredDelay := rand.Float64() * exponentialDelay

		waitDuration := time.Duration(jitteredDelay) * time.Millisecond

		fmt.Printf("[Retry Attempt %d] Standard Exp Delay: %v ms | Applied Jitter Delay: %v\n",
			attempt, int(exponentialDelay), waitDuration)

		// 3. Fake Wait
		// time.Sleep(waitDuration)
	}

	fmt.Println("Retries exhausted or Network resolved.")
}
