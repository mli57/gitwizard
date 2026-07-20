/*
Finds which port the app is on, and waits for it to finish starting.

Compose publishes several ports (app, database, cache) without saying which is
which, so we try them all until one answers like a website.
*/
package docker

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"
)

// without a timeout a database port hangs us forever: it accepts the connection
// and then waits for a message it understands, which an HTTP request isn't
var probeClient = &http.Client{Timeout: 2 * time.Second}

// reports whether anything answered, and whether it was a website. 
// Kept separate so the caller can tell "nothing running" from "running, but not a web app".
func probe(port int) (responded, servesPage bool) {
	// not "localhost": Windows resolves that to IPv6 first, but Docker publishes on IPv4
	response, err := probeClient.Get(fmt.Sprintf("http://127.0.0.1:%d", port))
	if err != nil {
		return false, false
	}
	defer response.Body.Close()

	// HasPrefix because the header usually reads "text/html; charset=utf-8"
	isHTML := strings.HasPrefix(response.Header.Get("Content-Type"), "text/html")

	// a crashed app often serves a tidy HTML error page, so the status has to be OK too
	return true, isHTML && response.StatusCode == http.StatusOK
}

func WaitForApp(ports []int) (int, error) {
	if len(ports) == 0 {
		return 0, fmt.Errorf("no ports to check")
	}
	slices.Sort(ports) // lowest port wins a tie, so runs are repeatable

	sawResponse := false
	// re-check every port each round, since the app may still be starting
	for round := 0; round < 30; round++ {
		for _, port := range ports {
			responded, servesPage := probe(port)
			if servesPage {
				return port, nil
			}
			if responded {
				sawResponse = true
			}
		}
		time.Sleep(time.Second)
	}

	if sawResponse {
		return 0, fmt.Errorf("something is running on %v, but none of it serves a web page - this looks like an API, not a web app", ports)
	}
	return 0, fmt.Errorf("nothing responded on %v - the app may have crashed on startup (try `docker compose logs`)", ports)
}
