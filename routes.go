package rentals

import (
	"encoding/json"
	"log"
	"net/http"
)

func (s *Server) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var userData struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		defer r.Body.Close()
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&userData)
		if err != nil {
			ErrorResponse(w, http.StatusInternalServerError, "Internal Server error")
			log.Printf("[ERROR] %v", err)
			return
		}

		token, err := s.AuthN.Login(userData.Username, userData.Password)
		if err != nil {
			ErrorResponse(w, http.StatusUnauthorized, "Not allowed")
			return
		}

		var returnToken struct {
			Token string `json:"token"`
		}
		returnToken.Token = token

		jsonRes, err := json.Marshal(returnToken)
		if err != nil {
			ErrorResponse(w, http.StatusInternalServerError, "Internal Server error")
			log.Printf("[ERROR] %v", err)
			return
		}

		_, _ = w.Write(jsonRes)
	}
}

func (s *Server) profileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This must exist otherwise the middleware would have rejected it
		token := r.Header["Authorization"][0]
		user := s.AuthN.Verify(token)

		if user == nil {
			ErrorResponse(w, http.StatusUnauthorized, "Not allowed")
			return
		}

		jsonRes, err := json.Marshal(user)
		if err != nil {
			ErrorResponse(w, http.StatusInternalServerError, "Internal Server error")
			log.Printf("[ERROR] %v", err)
			return
		}

		_, _ = w.Write(jsonRes)
	})
}
