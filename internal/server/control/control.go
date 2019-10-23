/*
 * Copyright 2019 Banco Bilbao Vizcaya Argentaria, S.A.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package control

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/BBVA/kapow/internal/server/model"
	"github.com/BBVA/kapow/internal/server/user"
)

// Run must start the control server in a specific address
func Run(bindAddr string) {
	log.Fatal(http.ListenAndServe(bindAddr, configRouter()))
}

func configRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/routes/{id}", removeRoute).
		Methods("DELETE")
	r.HandleFunc("/routes", listRoutes).
		Methods("GET")
	r.HandleFunc("/routes", addRoute).
		Methods("POST")
	return r
}

var funcRemove func(id string) error = user.Routes.Delete

func removeRoute(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	if err := funcRemove(id); err != nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	res.WriteHeader(http.StatusNoContent)
}

var funcList func() []model.Route = user.Routes.List

func listRoutes(res http.ResponseWriter, req *http.Request) {

	list := funcList()

	listBytes, _ := json.Marshal(list)
	res.Header().Set("Content-Type", "application/json")
	_, _ = res.Write(listBytes)
}

var funcAdd func(model.Route) model.Route = user.Routes.Append

func addRoute(res http.ResponseWriter, req *http.Request) {
	var route model.Route

	payload, _ := ioutil.ReadAll(req.Body)
	err := json.Unmarshal(payload, &route)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	if route.Method == "" {
		res.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if route.Pattern == "" {
		res.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	created := funcAdd(route)
	createdBytes, _ := json.Marshal(created)

	res.WriteHeader(http.StatusCreated)
	res.Header().Set("Content-Type", "application/json")
	_, _ = res.Write(createdBytes)
}
