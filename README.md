# Ratelimiter

A Discord ratelimiter intended to be used with net/http clients using time/x/rate.

## Example

```go
package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/Postcord/ratelimiter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const GetMemberFmt = "https://discord.com/api/guilds/%s/members/%s"

func main() {
	guild := flag.String("guild", "", "Guild ID")
	member := flag.String("member", "", "Member ID")
	token := flag.String("token", "", "Discord API Token")
	flag.Parse()

	if *guild == "" || *member == "" || *token == "" {
		log.Fatal().Msg("Please provide a guild ID, member ID, and token")
	}

	client := http.Client{}
	rl := ratelimiter.NewRatelimiter()

	req, err := http.NewRequest("GET", fmt.Sprintf(GetMemberFmt, *guild, *member), nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create request")
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	req.Header.Add("Authorization", "Bot "+*token)
	req.Header.Add("User-Agent", "KelwingTestClient/1.0 (Linux)")

	for i := 0; i < 10; i++ {
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
		rl.Update(resp)
		resp.Body.Close()
	}
}

```
