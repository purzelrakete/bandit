package http

import (
	"encoding/json"
	"github.com/purzelrakete/bandit"
	"net/http"
)

// OOBHandler
func OOBHandler(b bandit.Bandit, campaign bandit.Campaign) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "text/json")

		selected := b.SelectArm()
		variant, err := bandit.SelectVariant(campaign, selected)
		if err != nil {
			http.Error(w, "could not select variant", http.StatusInternalServerError)
			return
		}

		json, err := json.Marshal(variant)
		if err != nil {
			http.Error(w, "could not marshal variant", http.StatusInternalServerError)
			return
		}

		w.Write(json)
	}
}
