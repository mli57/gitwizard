/*
Polls the running container's port until the app responds or 30 seconds pass
*/
package docker

import(
	"net/http"
	"time"
	"fmt"
)

func WaitForApp(port int) error {
	count := 0
	for { // poll the app every second until it responds or we time out
		count++
		response, err := http.Get(fmt.Sprintf("http://localhost:%d", port))
		time.Sleep(time.Second)

		if count >= 30 && err != nil{ // if no response in 30 secs, error
			return err
		}
		if err == nil {
			response.Body.Close()
			return nil
		}

	}

}