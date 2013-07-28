package http

import (
	"encoding/json"
	"github.com/purzelrakete/bandit"
	"net/http"
)

// APIResponse is the json response on the /test endpoint
type APIResponse struct {
	UID      int    `json:"uid"`
	Campaign string `json:"campaign"`
	URL      string `json:"url"`
	Tag      string `json:"tag"`
}

// OOBHandler response to /test/:campaign with the provided bandits.
func OOBHandler(tests bandit.Tests) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "text/json")

		name := r.URL.Query().Get(":campaign")
		test, ok := tests[name]
		if ok != true {
			http.Error(w, "campaign not found", http.StatusInternalServerError)
			return
		}

		selected := test.Bandit.SelectArm()
		variant, err := bandit.SelectVariant(test.Campaign, selected)
		if err != nil {
			http.Error(w, "could not select variant", http.StatusInternalServerError)
			return
		}

		json, err := json.Marshal(APIResponse{
			UID:      0,
			Campaign: test.Campaign.Name,
			URL:      variant.URL,
			Tag:      variant.Tag,
		})

		if err != nil {
			http.Error(w, "could not marshal variant", http.StatusInternalServerError)
			return
		}

		w.Write(json)
	}
}
