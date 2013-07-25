package http

import (
	"github.com/purzelrakete/bandit"
	bhttp "github.com/purzelrakete/bandit/http"
	"log"
	"test"
)

func main() {
	bandit := bandit.Softmax(5, 0.1)
	http.HandleFunc("/", bHttp.BanditHandler(bandit))
	http.ListenAndServe(httpPort, nil)
}
