package handler

import (
	"encoding/json"
	"net/http"
)

type Users struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func UserViewHandler(users map[string]Users) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		userId := r.URL.Query().Get("user_id")
		if userId == "" {
			http.Error(rw, "user_id is empty", http.StatusBadRequest)
			return
		}

		user, ok := users[userId]
		if !ok {
			http.Error(rw, "user not found", http.StatusNotFound)
			return
		}

		jsonUser, err := json.Marshal(user)
		if err != nil {
			http.Error(rw, "can't provide a json. internal error",
				http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write(jsonUser)

	}
}

//func main() {
//	users := make(map[string]Users)
//	u1 := Users{
//		ID:        "u1",
//		FirstName: "Misha",
//		LastName:  "Popov",
//	}
//	u2 := Users{
//		ID:        "u2",
//		FirstName: "Sasha",
//		LastName:  "Popov",
//	}
//	users["u1"] = u1
//	users["u2"] = u2
//
//	http.HandleFunc("/users", UserViewHandler(users))
//	log.Fatal(http.ListenAndServe(":8080", nil))
//}
