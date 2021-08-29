package registry

import (
	"log"
	"net/http"
	"time"
)

// Heartbeat send heart beat signal to notify the registry that client is still alive
func Heartbeat(registry, addr string, interval time.Duration) {
	if interval == 0 {
		// interval should be less than timeout setting
		interval = defaultTimeout - time.Duration(1)*time.Minute
	}

	var err error
	err = sendHeartbeat(registry, addr)
	go func() {
		t := time.NewTicker(interval)
		for err == nil {
			<- t.C
			// periodically send heartbeat to registry
			err = sendHeartbeat(registry, addr)
		}
	}()
}

func sendHeartbeat(registry, addr string) error {
	log.Println(addr, "send heart beat to registry", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", registry, nil)
	req.Header.Set(serverHTTPHead, addr)
	if _, err := httpClient.Do(req); err != nil {
		log.Println("rpc server: heart beat err:", err)
		return err
	}

	return nil
}
