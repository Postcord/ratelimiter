# Ratelimiter

A Discord ratelimiter intended to be used with net/http clients.

## Example

```go
package main

import (
	"net/http"
	"os"
	"time"

	"github.com/Postcord/ratelimiter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	client := http.Client{}
	rl := ratelimiter.NewRatelimiter()

	req, err := http.NewRequest("GET", "https://discord.com/api/guilds/775135665602953286/members/109710323094683648", nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create request")
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	req.Header.Add("Authorization", "Bot "+os.Getenv("TOKEN"))
	req.Header.Add("User-Agent", "RatelimitTestClient/1.0 (https://github.com/Postcord/ratelimiter)")

	for i := 0; i < 10; i++ {
        // Request a reservation, and wait if required
		err := rl.Limit(req)
		if err != nil {
			continue
		}
		start := time.Now()
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		length := time.Since(start)
		log.Info().Int("status", resp.StatusCode).Dur("duration", length).Msg("Request successful")
        // Update the ratelimiter with the response from Discord
		rl.Update(resp)
		resp.Body.Close()
	}
}
```